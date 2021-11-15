package main

import (
	"database/sql"
	"log"
	"net/http"
	"html/template"
	"flag"
	"os"
	"time"
	"maxkavun.ml/snippetbox/pkg/models/mysql"

	_ "github.com/go-sql-driver/mysql"

	"github.com/golangcollege/sessions"
)

type application struct {
	errorLog *log.Logger
	infoLog *log.Logger
	session *sessions.Session
	snippets *mysql.SnippetModel
	templateCache map[string]*template.Template
	users *mysql.UserModel
}

func main() {

	addr := flag.String("addr", ":4000", "HTTP Network address")
	dsn := flag.String("dsn", "root:qwerty@/snippetbox?parseTime=true", "MySQL data source name")
	secret := flag.String("secret", "+MbQeThWmZq4t7w!z%C*F-JaNcRfUjXn", "Secret key")

	flag.Parse()

	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	db, err := openDB(*dsn)
	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(5)
	if err != nil {
		errorLog.Fatal(err)
	}

	defer db.Close()

	templateCache, err := newTemplateCache("./ui/html")
	if err != nil {
		errorLog.Fatal(err)
	}

	session := sessions.New([]byte(*secret))
	session.Lifetime = 12 * time.Hour
	session.Secure = true
	session.SameSite = http.SameSiteStrictMode

	app := &application{
		errorLog: errorLog,
		infoLog: infoLog,
		session: session,
		snippets: &mysql.SnippetModel{DB: db},
		templateCache: templateCache,
		users: &mysql.UserModel{DB: db},
	}

	srv := &http.Server{
		Addr: *addr,
		ErrorLog: errorLog,
		Handler: app.routes(),
	}

	infoLog.Printf("Starting server on %s", *addr)
	err = srv.ListenAndServe()
	errorLog.Fatal(err)
}

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