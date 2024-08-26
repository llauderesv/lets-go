package main

import (
	"fmt"
	"net/http"

	"github.com/bmizerany/pat"
	"github.com/justinas/alice"
)

type TestRoute struct{}

func (t *TestRoute) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello World")
}

// Create application routes using mux
func (app *application) routes() http.Handler {
	// Create a middleware chain containing our 'standard' middleware.
	// which will be used for every request our application receives.
	standardMiddleware := alice.New(app.recoverPanic, app.logRequest, secureHeaders)

	mux := pat.New()
	mux.Get("/", http.HandlerFunc(app.home))
	mux.Get("/snippets/create", http.HandlerFunc(app.createSnippetForm))
	mux.Get("/users", http.HandlerFunc(app.users))
	// mux.Get("/snippets/create", http.HandlerFunc(app.users))
	mux.Post("/snippets/create", http.HandlerFunc(app.createSnippet))
	mux.Get("/snippets/:id", http.HandlerFunc(app.showSnippet))

	// Create a file server which serves files out of the "./ui/static" directory
	// Note that the path given to the http.Dir function is relative to the provider
	// directory root
	fileServer := http.FileServer(http.Dir("./ui/static"))
	mux.Get("/static/", http.StripPrefix("/static", fileServer))

	return standardMiddleware.Then(mux)
}
