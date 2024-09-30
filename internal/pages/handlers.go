package pages

import (
	"net/http"
)

func ServeGuestPage(w http.ResponseWriter, r *http.Request) {
	data := TemplateData{
		Data: map[string]string{
			"Title": "Pact",
		}}
	RenderLayoutTemplate(w, "guestPage", data)
}
func ServeGuestContent(w http.ResponseWriter, r *http.Request) {
	data := TemplateData{
		Data: map[string]string{
			"Title": "Pact",
		}}
	RenderTemplateFraction(w, "guest", data)
}
