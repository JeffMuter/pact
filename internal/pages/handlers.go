package pages

import (
	"fmt"
	"net/http"
	"os"
)

func ServeHomePage(w http.ResponseWriter, r *http.Request) {
	postPath := "../../templates/posts/"
	var posts []Post

	entries, err := os.ReadDir(postPath)
	if err != nil {
		fmt.Println("homehandler could ned read dir path")
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		post := Post{Title: entry.Name()}
		posts = append(posts, post)
	}

	postData := posts
	pageData := &Page{Title: "File-Serving | Home", Heading: "File-Transfer Server"}

	data := TemplateData{
		Posts: postData,
		Page:  *pageData,
	}

	RenderTemplate(w, "../../templates/index.html", data)
}
