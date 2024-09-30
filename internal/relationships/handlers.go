package relationships

import (
	"fmt"
	"net/http"
	"pact/internal/pages"
)

func ServePageContent(w http.ResponseWriter, r *http.Request) {
	fmt.Println("serving relationships ")
	//	err := godotenv.Load()
	//	if err != nil {
	//		fmt.Println("error getting godotenv to load in serve stripe form...")
	//		return
	//	}
	//	publishableId := os.Getenv("STRIPE_PUBLISHABLE_KEY")
	data := pages.TemplateData{
		Data: map[string]string{
			"Title": "Relationships",

			//			"StripePublishableKey": publishableId,
		},
	}
	pages.RenderTemplateFraction(w, "stripeForm", data)
}
