# Membership System Setup Guide

This guide explains how to set up and configure the Stripe-based membership system for Pact.

## Overview

The membership system allows users to:
- Subscribe to premium features via Stripe
- Manage their subscription (update payment, pause, cancel) through Stripe Customer Portal
- Have their membership status automatically synced via webhooks

## Database Schema

Three new fields added to the `users` table:
- `stripe_customer_id` - Stripe's customer identifier (set on first subscription)
- `stripe_subscription_id` - Current subscription identifier
- `subscription_status` - Status string (active, canceled, trialing, past_due, etc.)
- `is_member` - Boolean flag (1 = active member, 0 = not a member)

## Environment Variables Required

Add these to your `.env` file:

```env
STRIPE_SECRET_KEY=sk_test_...           # Your Stripe secret API key
STRIPE_PUBLISHABLE_KEY=pk_test_...      # Your Stripe publishable key
STRIPE_MONTH_PRICE_ID=price_...         # Stripe Price ID for monthly subscription
STRIPE_WEBHOOK_SECRET=whsec_...         # Webhook signing secret
BASE_URL=http://localhost:8080          # Your application's base URL
```

## Stripe Setup Steps

### 1. Create Product & Price
1. Go to Stripe Dashboard → Products
2. Create a new product (e.g., "Pact Premium Membership")
3. Add a recurring price (e.g., $9.99/month)
4. Copy the Price ID (starts with `price_...`) → use as `STRIPE_MONTH_PRICE_ID`

### 2. Get API Keys
1. Go to Stripe Dashboard → Developers → API Keys
2. Copy "Publishable key" → `STRIPE_PUBLISHABLE_KEY`
3. Copy "Secret key" → `STRIPE_SECRET_KEY`

### 3. Configure Webhooks
1. Go to Stripe Dashboard → Developers → Webhooks
2. Click "Add endpoint"
3. Set endpoint URL: `https://yourdomain.com/webhook/stripe`
4. Select events to listen for:
   - `checkout.session.completed`
   - `customer.subscription.created`
   - `customer.subscription.updated`
   - `customer.subscription.deleted`
5. Copy the "Signing secret" (starts with `whsec_...`) → `STRIPE_WEBHOOK_SECRET`

**For local development:**
- Use Stripe CLI to forward webhooks: `stripe listen --forward-to localhost:8080/webhook/stripe`
- This will give you a temporary webhook secret to use locally

## Architecture

### Routes
- `GET /membership` - Membership page (shows subscribe or manage buttons based on status)
- `POST /createSession` - Creates Stripe Checkout session for new subscriptions
- `POST /createPortalSession` - Creates Stripe Customer Portal session for managing subscriptions
- `POST /webhook/stripe` - Webhook endpoint for Stripe events (no auth required)

### Templates
- `membershipPage.html` - Full page layout
- `membershipContent.html` - HTMX fragment (same content, different wrapper)

### Handlers (internal/stripe/handlers.go)
- `ServeMembershipPage` - Renders page with user's current subscription status
- `HandleCreateCheckoutSession` - Creates checkout session, redirects to Stripe
- `HandleCreatePortalSession` - Creates portal session, redirects to Stripe
- `HandleStripeWebhook` - Processes webhook events, updates database

### Webhook Event Handlers
- `handleCheckoutSessionCompleted` - Saves Stripe customer ID on first successful checkout
- `handleSubscriptionUpdated` - Updates subscription status and is_member flag
- `handleSubscriptionDeleted` - Revokes member access on cancellation

## User Flow

### New Subscription
1. User visits `/membership` → sees pricing and "Subscribe Now" button
2. Clicks button → `startCheckout()` calls `/createSession`
3. Redirects to Stripe Checkout page
4. User enters payment details, completes checkout
5. Stripe sends `checkout.session.completed` webhook → saves customer ID
6. Stripe sends `customer.subscription.created` webhook → sets `is_member=1`, saves subscription ID
7. User redirected back to `/membership?session_id=...`
8. Page now shows "Active Member" view with "Manage Subscription" button

### Manage Subscription
1. Active member visits `/membership` → sees "Manage Subscription" button
2. Clicks button → `openCustomerPortal()` calls `/createPortalSession`
3. Redirects to Stripe Customer Portal
4. User can:
   - Update payment method
   - View invoices
   - Cancel subscription
   - Update billing info
5. Any changes trigger webhooks → database synced automatically
6. User clicks "Return to [Your Site]" → back to `/membership`

### Cancellation
1. User cancels via Customer Portal
2. Stripe sends `customer.subscription.deleted` webhook
3. `handleSubscriptionDeleted` sets `is_member=0`, `subscription_status='canceled'`
4. User loses access to premium features

## Subscription Status Logic

Member access granted when:
- `subscription_status` = `"active"` OR `"trialing"`
- `is_member` = 1

Member access revoked when:
- `subscription_status` = `"canceled"`, `"unpaid"`, `"past_due"` (configurable)
- `is_member` = 0

## Testing

### Test Mode
- Use Stripe test mode keys (start with `sk_test_` and `pk_test_`)
- Use test card: `4242 4242 4242 4242` (any future expiry, any CVC)
- Trigger webhooks with Stripe CLI: `stripe trigger customer.subscription.created`

### Local Webhook Testing
```bash
# Install Stripe CLI
stripe login

# Forward webhooks to local server
stripe listen --forward-to localhost:8080/webhook/stripe

# Use the webhook secret printed by the command in your .env
```

### Manual Testing Checklist
- [ ] Non-member sees "Subscribe Now" button and pricing
- [ ] Clicking "Subscribe Now" redirects to Stripe Checkout
- [ ] Completing checkout returns to app with success message
- [ ] User's membership page shows "Active Member"
- [ ] Clicking "Manage Subscription" opens Customer Portal
- [ ] Canceling subscription revokes member access
- [ ] Re-subscribing restores member access

## Security Notes

1. **Webhook Signature Verification**: All webhooks are verified using `stripe.webhook.ConstructEvent()` with the webhook secret. Unsigned requests are rejected.

2. **Authentication**: Checkout and portal endpoints require authentication via `AuthMiddleware`. Only logged-in users can create sessions.

3. **Customer Ownership**: When creating a checkout session, we set `ClientReferenceID` to the user's ID. When creating a portal session, we verify the customer belongs to the authenticated user.

4. **Never Trust Client-Side**: Membership status is ONLY updated via webhooks from Stripe, never from client requests.

## Troubleshooting

### Webhook not firing
- Check Stripe Dashboard → Developers → Webhooks → [Your endpoint] → Logs
- Verify endpoint URL is correct and publicly accessible
- For local dev, ensure Stripe CLI is running and forwarding

### Subscription not activating
- Check logs for webhook processing errors
- Verify `STRIPE_WEBHOOK_SECRET` matches your endpoint
- Check database: `SELECT * FROM users WHERE user_id = X;`
- Stripe Dashboard → Customers → find customer → check subscription status

### Customer Portal not opening
- Verify user has `stripe_customer_id` set in database
- Check browser console for JavaScript errors
- Verify `STRIPE_SECRET_KEY` is correct

### Payment failing
- Test mode: only test cards work
- Live mode: check card details, billing address
- Check Stripe Dashboard → Payments for decline reason

## Migration

To apply database changes to existing databases:

```bash
sqlite3 database/database.db < database/migrations/003_stripe_subscription_fields.sql
```

Or rebuild from schema:
```bash
sqlite3 database/database.db < database/schema.sql
```

## Next Steps / Future Enhancements

- [ ] Add subscription analytics dashboard
- [ ] Email notifications for subscription events (requires email service)
- [ ] Trial period configuration
- [ ] Multiple subscription tiers
- [ ] Annual subscription discount
- [ ] Dunning/retry logic for failed payments
- [ ] Grace period after cancellation
- [ ] Referral program
