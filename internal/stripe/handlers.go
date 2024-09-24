package stripe

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/stripe/stripe-go/v72"
	"github.com/stripe/stripe-go/v72/customer"
	"github.com/stripe/stripe-go/v72/subscription"
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
