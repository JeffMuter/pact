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

func ServeGuestNavbar(w http.ResponseWriter, r *http.Request) {
	data := TemplateData{}
	RenderTemplateFraction(w, "guestNavbar", data)
}

func ServeLoggedInNavbar(w http.ResponseWriter, r *http.Request) {
	data := TemplateData{}
	RenderTemplateFraction(w, "guestNavbar", data)
}

func ServeMemberNavbar(w http.ResponseWriter, r *http.Request) {
	data := TemplateData{}
	RenderTemplateFraction(w, "guestNavbar", data)
}
