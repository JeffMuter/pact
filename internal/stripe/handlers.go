package stripe

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"pact/internal/pages"

	"github.com/joho/godotenv"
	"github.com/stripe/stripe-go/checkout/session"
	"github.com/stripe/stripe-go/v79"
	"github.com/stripe/stripe-go/v79/customer"
	"github.com/stripe/stripe-go/v79/subscription"
)

type SubscriptionRequest struct {
	Email           string `json:"email"`
	PaymentMethodID string `json:"paymentMethodID"`
	PriceID         string `json:"priceID"`
}

func handleCreateSubscription(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req SubscriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Create a new customer
	customerParams := &stripe.CustomerParams{
		Email:         stripe.String(req.Email),
		PaymentMethod: stripe.String(req.PaymentMethodID),
		InvoiceSettings: &stripe.CustomerInvoiceSettingsParams{
			DefaultPaymentMethod: stripe.String(req.PaymentMethodID),
		},
	}
	newCustomer, err := customer.New(customerParams)
	if err != nil {
		log.Printf("Error creating customer: %v", err)
		http.Error(w, "Error creating customer", http.StatusInternalServerError)
		return
	}

	// Create the subscription
	subscriptionParams := &stripe.SubscriptionParams{
		Customer: stripe.String(newCustomer.ID),
		Items: []*stripe.SubscriptionItemsParams{
			{
				Price: stripe.String(req.PriceID),
			},
		},
	}
	newSubscription, err := subscription.New(subscriptionParams)
	if err != nil {
		log.Printf("Error creating subscription: %v", err)
		http.Error(w, "Error creating subscription", http.StatusInternalServerError)
		return
	}

	// Return the subscription ID
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"subscriptionID": newSubscription.ID})
}

func ServeMembershipPage(w http.ResponseWriter, r *http.Request) {
	data := pages.TemplateData{
		Data: map[string]string{
			"Title": "Membership",
		}}
	fmt.Println("membership page handler ran")
	pages.RenderLayoutTemplate(w, "stripePage", data)
}

func ServeMembershipForm(w http.ResponseWriter, r *http.Request) {
	data := pages.TemplateData{
		Data: map[string]string{
			"Title": "Membership",
		}}
	fmt.Println("membership form handler ran")
	pages.RenderLayoutTemplate(w, "stripePage", data)
}

func CreateStripeSubSession() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("error loading env package")
	}
	stripe.Key = os.Getenv("STRIPE_KEY")
	params := &stripe.CheckoutSessionParams{
		Mode: stripe.String(string(stripe.CheckoutSessionModeSubscription)), LineItems: []*stripe.CheckoutSessionLineItemParams{&stripe.CheckoutSessionLineItemParams{Price: stripe.String("{{PRICE_ID}}"), Quantity: stripe.Int64(1)}},
		UIMode:    stripe.String(string(stripe.CheckoutSessionUIModeEmbedded)),
		ReturnURL: stripe.String("https://example.com/checkout/return?session_id={CHECKOUT_SESSION_ID}"),
	}
	result, err := session.New(params)
}
