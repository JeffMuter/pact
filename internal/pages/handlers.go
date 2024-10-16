package pages

import (
	"net/http"
)

func ServeDescriptionPage(w http.ResponseWriter, r *http.Request) {
	data := TemplateData{
		Data: map[string]string{
			"Title": "Pact",
		}}
	RenderLayoutTemplate(w, r, "guestPage", data)
}
func ServeDescriptionContent(w http.ResponseWriter, r *http.Request) {
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

func ServeRegisteredNavbar(w http.ResponseWriter, r *http.Request) {
	data := TemplateData{}
	RenderTemplateFraction(w, "registeredNavbar", data)
}

func ServeMemberNavbar(w http.ResponseWriter, r *http.Request) {
	data := TemplateData{}
	RenderTemplateFraction(w, "memberNavbar", data)
}

func ServeBucketsPage(w http.ResponseWriter, r *http.Request) {
	data := TemplateData{
		Data: map[string]string{
			"Title": "Buckets",
		}}
	RenderLayoutTemplate(w, r, "bucketsPage", data)
}
func ServeBucketsContent(w http.ResponseWriter, r *http.Request) {
	data := TemplateData{
		Data: map[string]string{
			"Title": "Buckets",
		}}
	RenderTemplateFraction(w, "buckets", data)
}
