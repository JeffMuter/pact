package pages

import (
	"net/http"
)

func ServeHomePage(w http.ResponseWriter, r *http.Request) { // show login form page
	data := TemplateData{
		Data: map[string]string{
			"Title": "Pact",
		}}
	RenderLayoutTemplate(w, "homePage", data)
}
func ServeHomeContent(w http.ResponseWriter, r *http.Request) { // show login form page
	data := TemplateData{
		Data: map[string]string{
			"Title": "Pact",
		}}
	RenderTemplateFraction(w, "guest", data)
}
