package handlers

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"text/template"
)

type TemplateData struct {
	Page       Page
	Posts      []Post
	FormAction string
}

type Page struct {
	Title   string
	Heading string
}

type UserForm struct {
	Request string
}

type Posts struct {
	Posts []Post
}

type Post struct {
	Title string
}

func RenderTemplate(w http.ResponseWriter, templatePath string, data TemplateData) {
	t, err := template.ParseFiles(templatePath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = t.Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

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

func ServePostPage(w http.ResponseWriter, r *http.Request) { // generate a post page
	pathToTemplates := "../../templates/posts/"
	postName := strings.TrimPrefix(r.URL.Path, "/post/")
	postData := Page{Title: postName, Heading: postName}
	if postName == "" {
		http.NotFound(w, r)
		return
	}
	postPath := pathToTemplates + postName

	data := TemplateData{
		Page: postData,
	}

	RenderTemplate(w, postPath, data)
}

func ServeLoginPage(w http.ResponseWriter, r *http.Request) { // show login form page
	pageData := Page{Title: "Login", Heading: "Login"}
	data := TemplateData{Page: pageData, FormAction: "/login"}
	RenderTemplate(w, "../../templates/login.html", data)
}

func ServeRegistrationPage(w http.ResponseWriter, r *http.Request) { // registration form page
	pageData := Page{Title: "Register", Heading: "Register"}
	data := TemplateData{Page: pageData, FormAction: "/registeruser"}
	RenderTemplate(w, "../../templates/login.html", data)
}
