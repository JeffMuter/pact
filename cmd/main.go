package pact

import (
	"fmt"
	"net/http"
	"pact/internal/render"
)

func main() {
	mux := render.Router()
	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		fmt.Printf("error with listen&serve: %v", err)
	}
}
