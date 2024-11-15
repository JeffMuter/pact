package connections

import (
	"net/http"
	"pact/internal/pages"
)

func ServeConnectionsContent(w http.ResponseWriter, r *http.Request) {
	// should have a user id added in the context of this req here. lets check
	// got get list of requests.

	// go get connections, if existing, if not, need to know.

	data := pages.TemplateData{
		Data: map[string]string{
			"Title": "Relationships",
		},
	}
	pages.RenderTemplateFraction(w, "connections", data)
}
