package pages

import (
	"net/http"
)

func ServeHomePage(w http.ResponseWriter, r *http.Request) { // show login form page
	data := TemplateData{
		Data: map[string]string{
			"Title": "Home",
		}}
	RenderLayoutTemplate(w, "homePage", data)
}
