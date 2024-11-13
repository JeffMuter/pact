package relationships

import (
	"net/http"
	"pact/internal/pages"
)

func ServePageContent(w http.ResponseWriter, r *http.Request) {
	data := pages.TemplateData{
		Data: map[string]string{
			"Title": "Relationships",
		},
	}
	pages.RenderTemplateFraction(w, "connections", data)
}
