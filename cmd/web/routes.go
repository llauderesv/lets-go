package main

import (
	"encoding/json"
	"net/http"

	"github.com/bmizerany/pat"
	"github.com/justinas/alice"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

// In Go, struct is just like a class in other Programming Language.
// To use struct you need to create an instance of that Struct
// t := TestRoute{}
type TestRoute struct{}

func (t *TestRoute) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	d := map[string]string{
		"message": "Hello World",
	}

	json, err := json.Marshal(d)
	check(err)

	w.Write(json)
}

// Create application routes using mux
func (app *application) routes() http.Handler {
	// Create a middleware chain containing our 'standard' middleware.
	// which will be used for every request our application receives.
	standardMiddleware := alice.New(app.recoverPanic, app.logRequest, secureHeaders)

	t := &TestRoute{}

	// Create a new middleware chain containing the middleware specific to
	// our dynamic application routes. For now, this chain will only contain
	// the sessions middleware but we'll add more to it later.
	dynamicMiddleWare := alice.New(app.sessions.Enable)

	mux := pat.New()
	mux.Get("/", dynamicMiddleWare.ThenFunc(app.home))
	mux.Get("/testroute", dynamicMiddleWare.ThenFunc(t.ServeHTTP))
	mux.Get("/snippets/create", dynamicMiddleWare.ThenFunc(app.createSnippetForm))
	mux.Get("/users", dynamicMiddleWare.ThenFunc(app.users))
	// mux.Get("/snippets/create", dynamicMiddleWare.ThenFunc(app.users))
	mux.Post("/snippets/create", dynamicMiddleWare.ThenFunc(app.createSnippet))
	mux.Get("/snippets/:id", dynamicMiddleWare.ThenFunc(app.showSnippet))

	// Create a file server which serves files out of the "./ui/static" directory
	// Note that the path given to the http.Dir function is relative to the provider
	// directory root
	fileServer := http.FileServer(http.Dir("./ui/static"))
	mux.Get("/static/", http.StripPrefix("/static", fileServer))

	return standardMiddleware.Then(mux)
}
