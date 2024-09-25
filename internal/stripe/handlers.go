package stripe

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"pact/internal/pages"

	"github.com/joho/godotenv"
	"github.com/stripe/stripe-go/v79"
	"github.com/stripe/stripe-go/v79/checkout/session"
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

func createStripeSubSession() (*stripe.CheckoutSession, error) {
	fmt.Println("createStripeSession begins...")
	err := godotenv.Load()
	if err != nil {
		return nil, fmt.Errorf("error loading env package")
	}

	stripe.Key = os.Getenv("STRIPE_KEY")
	priceID := os.Getenv("STRIPE_MONTH_PRICE_ID")

	params := &stripe.CheckoutSessionParams{
		Mode: stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			&stripe.CheckoutSessionLineItemParams{
				Price:    stripe.String(priceID),
				Quantity: stripe.Int64(1),
			},
		},
		UIMode:    stripe.String(string(stripe.CheckoutSessionUIModeEmbedded)),
		ReturnURL: stripe.String("https://www.localhost:8081?session_id={CHECKOUT_SESSION_ID}"),
	}

	params.PaymentMethodTypes = stripe.StringSlice([]string{"card"})

	result, err := session.New(params)
	if err != nil {
		return nil, fmt.Errorf("failed to create stripe session: %w", err)
	}
	fmt.Println("create stripe session ended successfully.")
	fmt.Println("resultID: " + result.URL)
	return result, nil
}

func HandleCreateCheckoutSession(w http.ResponseWriter, r *http.Request) {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("error getting godotenv to load in create checkout process...")
		return
	}

	publishableId := os.Getenv("STRIPE_PUBLISHABLE_KEY")
	fmt.Println("serve stripe form publishable id: " + publishableId)

	session, err := createStripeSubSession()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		fmt.Println("error http internal server error while handling chechout session...")
		return
	}

	// Return JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"sessionId":            session.ID,
		"stripePublishableKey": publishableId,
	})
}

func ServeStripeForm(w http.ResponseWriter, r *http.Request) {
	fmt.Println("serving stripe form")
	err := godotenv.Load()
	if err != nil {
		fmt.Println("error getting godotenv to load in serve stripe form...")
		return
	}
	publishableId := os.Getenv("STRIPE_PUBLISHABLE_KEY")
	fmt.Println("serve stripe form publishable id: " + publishableId)
	data := pages.TemplateData{
		Data: map[string]string{
			"Title":                "Membership",
			"StripePublishableKey": publishableId,
		},
	}
	pages.RenderTemplateFraction(w, "stripeForm", data)
}
