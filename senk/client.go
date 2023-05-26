package main

import (
	"embed"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
)

//go:embed html/*
var html embed.FS

func serveStatic(file string) http.HandlerFunc {
	// TODO: Check content type header
	return func(w http.ResponseWriter, r *http.Request) {
		f, err := html.ReadFile("html/" + file)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
		} else {
			w.Write(f)
		}
	}
}

func (db *Database) serveMain(w http.ResponseWriter, r *http.Request) {
	sid, session, ok := GetSessionCtx(r.Context())
	if !ok {
		w.Write([]byte("<!DOCTYPE html>Not logged in <form method=\"POST\" action=\"session/signin\"><input type=\"text\" name=\"username\"><input type=\"text\" name=\"password\"><input type=\"submit\"></form>"))
	} else {
		w.Write([]byte(fmt.Sprint("<!DOCTYPE html>Logged in <form method=\"POST\" action=\"session/signout\"><input type=\"submit\"></form>", sid, session)))
	}
}

func (db *Database) SetupClient(r *chi.Mux) {
	r.Get("/", db.serveMain)
}
