// Stripe checkout and portal functions
async function startCheckout() {
    try {
        const response = await fetch('/createSession', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            }
        });
        
        const data = await response.json();
        
        if (data.url) {
            // Redirect to Stripe Checkout
            window.location.href = data.url;
        } else {
            alert('Error creating checkout session');
        }
    } catch (error) {
        console.error('Error:', error);
        alert('Error starting checkout process');
    }
}

async function openCustomerPortal() {
    try {
        const response = await fetch('/createPortalSession', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            }
        });
        
        const data = await response.json();
        
        if (data.url) {
            // Redirect to Stripe Customer Portal
            window.location.href = data.url;
        } else {
            alert('Error opening customer portal');
        }
    } catch (error) {
        console.error('Error:', error);
        alert('Error opening customer portal');
    }
}
