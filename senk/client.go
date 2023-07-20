package main

import (
	"embed"
	"log"
	"net/http"
)

//go:embed html/*
var html embed.FS

func serveStatic(file, contentType string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		f, err := html.ReadFile("html/" + file)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			log.Printf("Error: serveStatic couldn't find file \"%s\" (it wasn't bundled during compilation).", file)
		} else {
			w.Header().Add("Content-Type", contentType)
			w.Write(f)
		}
	}
}

func serveLogin(w http.ResponseWriter, r *http.Request) {
	serveStatic("signin.html", "text/html")(w, r)
}

func serveApp(w http.ResponseWriter, r *http.Request) {
	serveStatic("app.html", "text/html")(w, r)
}

func (db *Database) serveMain(w http.ResponseWriter, r *http.Request) {
	sid, _ := GetSessionCtx(r.Context())
	if sid == "" {
		serveLogin(w, r)
	} else {
		serveApp(w, r)
	}
}

// TODO: Bundle data in HTML responses, to avoid additional request and make the app even barely usable without JS.
