// signing in and out

package main

import (
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"
)

func (db *Database) signIn(w http.ResponseWriter, r *http.Request) {
	username := r.PostFormValue("username")
	password := r.PostFormValue("password")
	if username != "" && password != "" {
		if db.Users.CheckPassword(username, password) {
			sid, _, ok := GetSessionCtx(r.Context())
			if ok {
				db.Sessions.InvalidateSession(sid)
			}
			sid = db.Sessions.NewSession()
			err := db.Sessions.ModifySessionData(sid, SessionData{
				Authenticated: true,
				Username:      username,
			})
			if err != nil {
				log.Printf("Modifying a new session failed: %v", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			http.SetCookie(w, &http.Cookie{
				Name:     SessionCookieName,
				Value:    sid,
				Path:     "/",
				Secure:   true,
				HttpOnly: true,
				SameSite: http.SameSiteLaxMode,
				MaxAge:   int(SessionAbsoluteTimeout.Seconds()),
			})
			w.Header().Add("Location", r.Referer()) // TODO: Should be handled in JS
			w.WriteHeader(http.StatusFound)
			return
		}
	}
	w.WriteHeader(http.StatusForbidden)
}

func (db *Database) signOut(w http.ResponseWriter, r *http.Request) {
	sid, _, ok := GetSessionCtx(r.Context())
	if !ok {
		w.WriteHeader(http.StatusNotFound) // TODO: status
		return
	}
	db.Sessions.InvalidateSession(sid)
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    "",
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})

	w.Header().Add("Location", r.Referer()) // TODO: Handle in JS
	w.WriteHeader(http.StatusFound)
}

func (db *Database) SetupAuthentication(r *chi.Mux) {
	r.Post("/session/signin", db.signIn)
	r.Post("/session/signout", db.signOut)
}
