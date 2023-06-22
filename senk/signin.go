// signing in and out

package main

import (
	"log"
	"net/http"
)

// TODO: Rate limiting
func (db *Database) signIn(w http.ResponseWriter, r *http.Request) {
	username := r.PostFormValue("username")
	password := r.PostFormValue("password")
	if username != "" && password != "" {
		if db.Users.CheckPassword(username, password) {
			sid, _ := GetSessionCtx(r.Context())
			if sid != "" {
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
			w.Header().Add("Location", r.Referer())
			w.WriteHeader(http.StatusFound)
			return
		}
	}
	w.WriteHeader(http.StatusForbidden) // TODO: Show more than a blank page
}

func (db *Database) signOut(w http.ResponseWriter, r *http.Request) {
	sid, _ := GetSessionCtx(r.Context())
	if sid == "" {
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

	w.Header().Add("Location", r.Referer())
	w.WriteHeader(http.StatusFound)
}
