package connections

import (
	"net/http"
	"pact/internal/pages"
)

func ServeConnectionsContent(w http.ResponseWriter, r *http.Request) {
	data := pages.TemplateData{
		Data: map[string]string{
			"Title": "Relationships",
		},
	}
	pages.RenderTemplateFraction(w, "connections", data)
}
