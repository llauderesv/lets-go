package main

import (
	"crypto/tls"
	"database/sql"
	"flag"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/golangcollege/sessions"
	"vincellauderes.net/snippetbox/pkg/models/mysql"
)

type Vince int

// In Go a type struct
type Config struct {
	Addr      string
	StaticDir string
	Dsn       string
	Secret    string
}

// Define an application struct to hold the application wide dependencies for the web application
// For now we'll only include fields for the two custom logger
// we'll add more to it as the build progresses.
type application struct {
	errorLog      *log.Logger
	infoLog       *log.Logger
	sessions      *sessions.Session
	snippets      *mysql.SnippetModel
	templateCache map[string]*template.Template
	userss        *mysql.UserModel
}

func main() {
	// The new keyword is just like this syntax &Config{}, but this more readable that initializing zero values
	cfg := new(Config)

	// Define a new command-line flag for the MySQL DSN string
	flag.StringVar(&cfg.Dsn, "dsn", "web:pass@/snippetbox?parseTime=true", "MySQL Database Connection String")

	// This is a pointer, we need to dereference the pointer before using it...
	flag.StringVar(&cfg.Addr, "addr", ":4000", "HTTP network address")

	// Define a new command-line flag for the session secret (a random key which will be used to encrypt and authenticate session cookies).
	// It should be bytes long.
	flag.StringVar(&cfg.Secret, "secret", "s6Ndh+pPbnzHbS*+9Pk8qGWhTzbpa@ge", "Session Secret")

	// Importantly, we use the flag.Parse function to parse the command line
	flag.Parse()

	// Use log.New() to create a logger for writing information messages. This
	// three parameters: the destination to write the logs to (os.Stdout), a string
	// prefix for message (INFO followed by a tab), and flags to indicate what
	// additional information to include (local date and time). Note that the value
	// are joined using bitwise OR operator |.
	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	db, err := openDB(cfg.Dsn)
	if err != nil {
		errorLog.Fatal(err)
	}

	// We also defer a call to db.Close(), so that the connection pool is close
	// before the main function exits.
	defer db.Close()

	// Load the ./ui/html/directory
	templateCache, err := newTemplateCache("./ui/html/")
	if err != nil {
		errorLog.Fatal(err)
	}

	// Use the sessions.New() function to initialize a new session manager,
	// passing in the secret key as the parameter. Then we configure it so
	/// sessions always expires after 12 hours.
	session := sessions.New([]byte(cfg.Secret))
	session.Lifetime = 12 * time.Hour

	// Initialize a new instance of application containing the dependencies...
	app := application{
		errorLog:      errorLog,
		infoLog:       infoLog,
		sessions:      session,
		snippets:      &mysql.SnippetModel{DB: db},
		templateCache: templateCache,
		userss:        &mysql.UserModel{DB: db},
	}

	tlsConfig := &tls.Config{
		PreferServerCipherSuites: true,
		CurvePreferences:         []tls.CurveID{tls.X25519, tls.CurveP256},
	}

	// Initialize a new http.Server struct. We set the Addr and Handler fields
	// that the server uses the same network address and routes as before and
	// the ErrorLog field so that the server now uses the custom errorLog
	// We used cause we wanted to inject the ErrorLog into http Server
	srv := &http.Server{
		Addr:         cfg.Addr,
		ErrorLog:     errorLog,
		Handler:      app.routes(),
		TLSConfig:    tlsConfig,
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	infoLog.Printf("Staring server on %s", cfg.Addr)
	err = srv.ListenAndServeTLS("./tls/cert.pem", "./tls/key.pem")
	// err = srv.ListenAndServe()
	errorLog.Fatal(err)
}

// The openDB function wraps sql.Open() and returns a sql.DB connection pool
// for a given DSN.
func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}
