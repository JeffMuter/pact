package pages

import (
	"net/http"
)

func ServeHomePage(w http.ResponseWriter, r *http.Request) {

	data := TemplateData{
		Data: map[string]string{
			"Title": "Home",
		}}

	RenderTemplate(w, "./templates/index.html", data)
}
