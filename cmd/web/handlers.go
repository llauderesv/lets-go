package main

import (
	"fmt"
	"net/http"
	"strconv"

	"vincellauderes.net/snippetbox/pkg/forms"
	"vincellauderes.net/snippetbox/pkg/models"
)

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		app.notFound(w)
		return
	}

	s, err := app.snippets.Latest()
	if err != nil {
		app.serverError(w, err)
		return
	}

	// panic("oops! something went wrong") // Deliberate panic

	app.render(w, r, "home.page.tmpl", &templateData{
		Snippets: s,
	})
}

func (app *application) showSnippet(w http.ResponseWriter, r *http.Request) {
	// Extract the value of the id parameter from the query string and try to
	// convert it to an integer using the strconv.Atoi() function. If it can't
	// be converted to an integer, or the value is less than 1, we return a 404
	// not found response.
	id, err := strconv.Atoi(r.URL.Query().Get(":id"))
	if err != nil || id < 1 {
		app.notFound(w)
		return
	}

	s, err := app.snippets.Get(id)
	if err == models.ErrNoRecord {
		app.notFound(w)
		return
	} else if err != nil {
		app.serverError(w, err)
		return
	}

	// data := &templateData{Snippet: s}

	// Write the snippet data as a plain-text HTTP response body.
	// fmt.Fprintf(w, "%v", s)

	app.render(w, r, "show.page.tmpl", &templateData{
		Snippet: s,
	})
	// files := []string{
	// 	"./ui/html/show.page.tmpl",
	// 	"./ui/html/base.layout.tmpl",
	// 	"./ui/html/footer.partial.tmpl",
	// }

	// ts, err := template.ParseFiles(files...)
	// if err != nil {
	// 	app.serverError(w, err)
	// 	return
	// }

	// if err = ts.Execute(w, data); err != nil {
	// 	app.serverError(w, err)
	// }

}

func (app *application) users(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"name": "Alex Banks"}`)
}

func (app *application) createSnippet(w http.ResponseWriter, r *http.Request) {
	// First we call r.ParseForm() which adds any data in POST request bodies
	// to the r.PostForm map. This also works in the same way for PUT and PATCH
	// requests. If there are any errors, we use our app.ClientError helper to
	// a 400 Bad Request response to the user.
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form := forms.New(r.PostForm)
	form.Required("title", "content", "expires")
	form.MaxLength("title", 100)
	form.PermittedValues("expires", "365", "7", "1")

	if !form.Valid() {
		app.render(w, r, "create.page.tmpl", &templateData{Form: form})
		return
	}

	// Use the r.PostForm.Get() method to retrieve the relevant data fields
	// from the r.PostForm map.

	title := form.Get("title")
	content := form.Get("content")
	expires := form.Get("expires")

	// Initialize a map to hold any validation errors.
	// errors := make(map[string]string)

	// Check that the title field is not blank and is not more than 100 characters
	// long. If it fails either of those checks, add a message to the errors
	// map using the field name as the key.
	// if strings.TrimSpace(title) == "" {
	// 	errors["title"] = "This field cannot be blank"
	// } else if utf8.RuneCountInString(title) > 100 {
	// 	errors["title"] = "This field is too long (maximum is 100 characters)"
	// }

	// // Check that the Content field isn't blank.
	// if strings.TrimSpace(content) == "" {
	// 	errors["content"] = "This field cannot be blank"
	// }

	// if strings.TrimSpace(expires) == "" {
	// 	errors["expires"] = "This field cannot be blank"
	// } else if expires != "365" && expires != "7" && expires != "1" {
	// 	errors["expires"] = "This field is invalid"
	// }

	// // If there are any errors, dump them in a plain text HTTP response and ret
	// // from the handler.
	// if len(errors) > 0 {
	// 	// fmt.Fprint(w, errors)
	// 	app.render(w, r, "create.page.tmpl", &templateData{
	// 		Form: forms.New(nil),
	// 	})
	// 	return
	// }

	id, err := app.snippets.Insert(title, content, expires)
	if err != nil {
		app.serverError(w, err)
		return
	}

	// Use the Put() method to add a string value ("Your snippet was saved
	// successfully!") and the corresponding key ("flash") to the session
	// data. Note that if there's no existing session for the current user
	// (or their session has expired) then a new, empty, session for them
	// will automatically be created by the session middleware.
	app.sessions.Put(r, "flash", "Snippet Successfully created!")

	app.infoLog.Println(id)

	// Redirect the user to the relevant page for the snippet.
	http.Redirect(w, r, fmt.Sprintf("/snippets/%d", id), http.StatusSeeOther)
}

func (app *application) createSnippetForm(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, "create.page.tmpl", &templateData{Form: forms.New(nil)})
}

func (app *application) signupUserForm(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, "signup.page.tmpl", &templateData{
		Form: forms.New(nil),
	})
}

func (app *application) signupUser(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
	}

	// Validate the form contents using the form helper we made earlier
	form := forms.New(r.PostForm)
	form.Required("name", "email", "password")
	form.MatchesPattern("email", forms.EmailRX)
	form.MinLength("password", 10)

	if !form.Valid() {
		app.render(w, r, "signup.page.tmpl", &templateData{Form: form})
		return
	}

	err = app.userss.Insert(form.Get("name"), form.Get("email"), form.Get("password"))
	if err == models.ErrDuplicateEmail {
		form.Errors.Add("email", "Address is already in use")
		app.render(w, r, "signup.page.tmpl", &templateData{Form: form})
		return
	} else if err != nil {
		app.serverError(w, err)
		return
	}

	app.sessions.Put(r, "flash", "Your signup was successful. Please log in.")
	http.Redirect(w, r, "/user/login", http.StatusSeeOther)
}

func (app *application) loginUserForm(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, "login.page.tmpl", &templateData{
		Form: forms.New(nil),
	})
}

func (app *application) loginUser(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form := forms.New(r.PostForm)
	id, err := app.userss.Authenticate(form.Get("email"), form.Get("password"))
	if err == models.ErrInvalidCredentials {
		form.Errors.Add("generic", "Email or Password is incorrect")
		app.render(w, r, "login.page.tmpl", &templateData{Form: form})
		return
	} else if err != nil {
		app.serverError(w, err)
		return
	}

	app.sessions.Put(r, "userID", id)

	http.Redirect(w, r, "/snippets/create", http.StatusSeeOther)
}

func (app *application) logoutUser(w http.ResponseWriter, r *http.Request) {
	app.sessions.Remove(r, "userID")

	app.sessions.Put(r, "flash", "You've been logged out successfully!")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
