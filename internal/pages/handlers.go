package pages

import (
	"net/http"
)

func ServeHomePage(w http.ResponseWriter, r *http.Request) {

	data := TemplateData{
		Data: map[string]string{
			"Title": "Home",
		}}

	TemplateRenderer(w, "defaultLayout", "index.html", data)
}
