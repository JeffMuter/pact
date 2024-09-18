package models

// TemplateData represents the data used to render any dynamic page.
type TemplateData struct {
	Page       Page
	Posts      []Post
	FormAction string
}

// Page represents general page metadata.
type Page struct {
	Title   string
	Heading string
}

// UserForm represents the structure for a user-related form.
type UserForm struct {
	Request string
}

// Posts is a collection of Post entries.
type Posts struct {
	Posts []Post
}

// Post represents individual blog or content posts.
type Post struct {
	Title string
}
