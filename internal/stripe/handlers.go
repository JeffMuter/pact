package stripe

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"pact/database"
	"pact/internal/pages"

	"github.com/joho/godotenv"
	"github.com/stripe/stripe-go/v79"
	billingportal "github.com/stripe/stripe-go/v79/billingportal/session"
	checkout "github.com/stripe/stripe-go/v79/checkout/session"
	"github.com/stripe/stripe-go/v79/webhook"
)

var DB *database.Queries

func SetDB(db *database.Queries) {
	DB = db
}

// ServeMembershipPage renders the membership page based on user's subscription status
func ServeMembershipPage(w http.ResponseWriter, r *http.Request) {
	if DB == nil {
		log.Println("ERROR: Stripe DB is nil - not initialized")
		http.Error(w, "Configuration error", http.StatusInternalServerError)
		return
	}
	
	userId := r.Context().Value("userID").(int)
	
	user, err := DB.GetUserById(context.Background(), int64(userId))
	if err != nil {
		log.Printf("Error getting user %d: %v", userId, err)
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	err = godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file")
	}
	
	publishableKey := os.Getenv("STRIPE_PUBLISHABLE_KEY")
	
	// Determine membership status
	isMember := user.IsMember == 1
	var subscriptionStatus string
	if user.SubscriptionStatus.Valid {
		subscriptionStatus = user.SubscriptionStatus.String
	}
	
	data := pages.TemplateData{
		Data: map[string]any{
			"Title":                "Membership",
			"IsMember":             isMember,
			"SubscriptionStatus":   subscriptionStatus,
			"StripePublishableKey": publishableKey,
			"HasStripeCustomer":    user.StripeCustomerID.Valid,
		},
	}
	
	// Check if this is an HTMX request
	if r.Header.Get("HX-Request") == "true" {
		pages.RenderTemplateFraction(w, "membershipContent", data)
		return
	}
	
	pages.RenderLayoutTemplate(w, r, "membershipPage", data)
}

// HandleCreateCheckoutSession creates a Stripe checkout session for new subscriptions
func HandleCreateCheckoutSession(w http.ResponseWriter, r *http.Request) {
	userId := r.Context().Value("userID").(int)
	
	user, err := DB.GetUserById(context.Background(), int64(userId))
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	err = godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file")
	}

	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")
	priceID := os.Getenv("STRIPE_MONTH_PRICE_ID")
	publishableKey := os.Getenv("STRIPE_PUBLISHABLE_KEY")
	
	// Build success/cancel URLs
	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	params := &stripe.CheckoutSessionParams{
		Mode: stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(priceID),
				Quantity: stripe.Int64(1),
			},
		},
		SuccessURL: stripe.String(baseURL + "/membership?session_id={CHECKOUT_SESSION_ID}"),
		CancelURL:  stripe.String(baseURL + "/membership"),
		CustomerEmail: stripe.String(user.Email),
		ClientReferenceID: stripe.String(fmt.Sprintf("%d", userId)),
	}

	// If user already has a Stripe customer, reuse it
	if user.StripeCustomerID.Valid {
		params.Customer = stripe.String(user.StripeCustomerID.String)
	}

	s, err := checkout.New(params)
	if err != nil {
		log.Printf("Error creating checkout session: %v", err)
		http.Error(w, "Unable to create checkout session", http.StatusInternalServerError)
		return
	}

	// Return JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"sessionId":            s.ID,
		"url":                  s.URL,
		"stripePublishableKey": publishableKey,
	})
}

// HandleCreatePortalSession creates a Stripe Customer Portal session for managing subscriptions
func HandleCreatePortalSession(w http.ResponseWriter, r *http.Request) {
	userId := r.Context().Value("userID").(int)
	
	user, err := DB.GetUserById(context.Background(), int64(userId))
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	if !user.StripeCustomerID.Valid {
		http.Error(w, "No Stripe customer found", http.StatusBadRequest)
		return
	}

	err = godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file")
	}

	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")
	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	params := &stripe.BillingPortalSessionParams{
		Customer:  stripe.String(user.StripeCustomerID.String),
		ReturnURL: stripe.String(baseURL + "/membership"),
	}

	s, err := billingportal.New(params)
	if err != nil {
		log.Printf("Error creating portal session: %v", err)
		http.Error(w, "Unable to create portal session", http.StatusInternalServerError)
		return
	}

	// Return JSON response with portal URL
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"url": s.URL,
	})
}

// HandleStripeWebhook processes Stripe webhook events
func HandleStripeWebhook(w http.ResponseWriter, r *http.Request) {
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file")
	}

	webhookSecret := os.Getenv("STRIPE_WEBHOOK_SECRET")
	
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading webhook body: %v", err)
		http.Error(w, "Error reading request body", http.StatusBadRequest)
		return
	}

	// Verify webhook signature
	event, err := webhook.ConstructEvent(payload, r.Header.Get("Stripe-Signature"), webhookSecret)
	if err != nil {
		log.Printf("Error verifying webhook signature: %v", err)
		http.Error(w, "Invalid signature", http.StatusBadRequest)
		return
	}

	// Handle the event
	switch event.Type {
	case "checkout.session.completed":
		var session stripe.CheckoutSession
		err := json.Unmarshal(event.Data.Raw, &session)
		if err != nil {
			log.Printf("Error parsing webhook JSON: %v", err)
			http.Error(w, "Error parsing webhook", http.StatusBadRequest)
			return
		}
		handleCheckoutSessionCompleted(session)

	case "customer.subscription.created", "customer.subscription.updated":
		var subscription stripe.Subscription
		err := json.Unmarshal(event.Data.Raw, &subscription)
		if err != nil {
			log.Printf("Error parsing webhook JSON: %v", err)
			http.Error(w, "Error parsing webhook", http.StatusBadRequest)
			return
		}
		handleSubscriptionUpdated(subscription)

	case "customer.subscription.deleted":
		var subscription stripe.Subscription
		err := json.Unmarshal(event.Data.Raw, &subscription)
		if err != nil {
			log.Printf("Error parsing webhook JSON: %v", err)
			http.Error(w, "Error parsing webhook", http.StatusBadRequest)
			return
		}
		handleSubscriptionDeleted(subscription)

	default:
		log.Printf("Unhandled event type: %s", event.Type)
	}

	w.WriteHeader(http.StatusOK)
}

// handleCheckoutSessionCompleted processes successful checkout
func handleCheckoutSessionCompleted(session stripe.CheckoutSession) {
	// Get user ID from client_reference_id
	if session.ClientReferenceID == "" {
		log.Println("No client_reference_id in checkout session")
		return
	}

	var userId int64
	fmt.Sscanf(session.ClientReferenceID, "%d", &userId)

	if userId == 0 {
		log.Println("Invalid user ID in client_reference_id")
		return
	}

	// Update user with Stripe customer ID
	err := DB.UpdateUserStripeCustomer(context.Background(), database.UpdateUserStripeCustomerParams{
		StripeCustomerID: sql.NullString{String: session.Customer.ID, Valid: true},
		UserID:           userId,
	})
	if err != nil {
		log.Printf("Error updating user with Stripe customer: %v", err)
		return
	}

	log.Printf("User %d subscribed successfully with customer %s", userId, session.Customer.ID)
}

// handleSubscriptionUpdated processes subscription creation or updates
func handleSubscriptionUpdated(subscription stripe.Subscription) {
	// Find user by Stripe customer ID
	user, err := DB.GetUserByStripeCustomerId(context.Background(), 
		sql.NullString{String: subscription.Customer.ID, Valid: true})
	if err != nil {
		log.Printf("Error finding user for customer %s: %v", subscription.Customer.ID, err)
		return
	}

	// Determine if user should have member access
	isMember := int64(0)
	if subscription.Status == "active" || subscription.Status == "trialing" {
		isMember = 1
	}

	// Update user subscription status
	err = DB.UpdateUserSubscription(context.Background(), database.UpdateUserSubscriptionParams{
		StripeSubscriptionID: sql.NullString{String: subscription.ID, Valid: true},
		SubscriptionStatus:   sql.NullString{String: string(subscription.Status), Valid: true},
		IsMember:             isMember,
		UserID:               user.UserID,
	})
	if err != nil {
		log.Printf("Error updating user subscription: %v", err)
		return
	}

	log.Printf("User %d subscription updated: status=%s, is_member=%d", 
		user.UserID, subscription.Status, isMember)
}

// handleSubscriptionDeleted processes subscription cancellations
func handleSubscriptionDeleted(subscription stripe.Subscription) {
	// Find user by Stripe customer ID
	user, err := DB.GetUserByStripeCustomerId(context.Background(), 
		sql.NullString{String: subscription.Customer.ID, Valid: true})
	if err != nil {
		log.Printf("Error finding user for customer %s: %v", subscription.Customer.ID, err)
		return
	}

	// Revoke member access
	err = DB.UpdateUserSubscription(context.Background(), database.UpdateUserSubscriptionParams{
		StripeSubscriptionID: sql.NullString{String: subscription.ID, Valid: true},
		SubscriptionStatus:   sql.NullString{String: "canceled", Valid: true},
		IsMember:             0,
		UserID:               user.UserID,
	})
	if err != nil {
		log.Printf("Error revoking user membership: %v", err)
		return
	}

	log.Printf("User %d subscription canceled", user.UserID)
}
