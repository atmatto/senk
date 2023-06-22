// session management

package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"log"
	"net/http"
	"sync"
	"time"
)

const (
	SessionIdleTimeout     = time.Hour * 24 * 90  // remember session for 90 days
	SessionAbsoluteTimeout = time.Hour * 24 * 365 // require the user to reauthenticate every 365 days
	SessionCookieName      = "id"
)

var (
	ErrSessionInvalid = errors.New("session does not exist")
)

type SessionData struct {
	Authenticated bool
	Username      string // only makes sense if Authenticated == true
}

type Session struct {
	Created    time.Time
	LastActive time.Time
	Data       SessionData
}

func (s *Session) IsExpired() bool {
	return time.Now().Sub(s.LastActive) >= SessionIdleTimeout || time.Now().Sub(s.Created) >= SessionAbsoluteTimeout
}

type Sessions struct {
	Map map[string]Session // indexed by the id
	mu  sync.RWMutex
}

func (sessions *Sessions) Initialize() {
	if sessions.Map == nil {
		sessions.Map = make(map[string]Session)
	}
}

// NewSession returns the session's id. In case of failure, it returns an empty string.
func (sessions *Sessions) NewSession() string {
	sessions.mu.Lock()
	defer sessions.mu.Unlock()

	id := make([]byte, 64)
	if _, err := io.ReadFull(rand.Reader, id); err != nil {
		log.Printf("Error reading from rand.Reader: %v", err)
		return ""
	}
	strid := base64.URLEncoding.EncodeToString(id)

	if _, ok := sessions.Map[strid]; ok {
		log.Printf("Generated an already used session id!")
		return ""
	}

	sessions.Map[strid] = Session{
		Created:    time.Now(),
		LastActive: time.Now(),
	}

	return strid
}

// TODO: Invalidate all sessions of a given user

func (sessions *Sessions) InvalidateSession(id string) {
	sessions.mu.Lock()
	defer sessions.mu.Unlock()
	delete(sessions.Map, id)
}

func (sessions *Sessions) ModifySessionData(id string, data SessionData) error {
	sessions.mu.Lock()
	defer sessions.mu.Unlock()
	session, ok := sessions.Map[id]
	if ok {
		session.Data = data
		sessions.Map[id] = session
		return nil
	} else {
		return ErrSessionInvalid
	}
}

type ContextKey string

// SessionRetrievalMiddleware saves session data to the request's context, based on the cookie.
func (sessions *Sessions) SessionRetrievalMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(SessionCookieName)
		if err == nil {
			sessions.mu.Lock()
			session, ok := sessions.Map[cookie.Value]
			if ok {
				if session.IsExpired() {
					delete(sessions.Map, cookie.Value)
				} else {
					session.LastActive = time.Now()
					sessions.Map[cookie.Value] = session
					r = r.WithContext(context.WithValue(r.Context(), ContextKey("session"), session))
					r = r.WithContext(context.WithValue(r.Context(), ContextKey("sessionId"), cookie.Value))
				}
			}
			sessions.mu.Unlock()
		}
		next.ServeHTTP(w, r)
	})
}

// GetSessionCtx retrieves the session data from the given context
// and returns the id and session struct.
func GetSessionCtx(ctx context.Context) (string, Session) {
	s, ok1 := ctx.Value(ContextKey("session")).(Session)
	id, ok2 := ctx.Value(ContextKey("sessionId")).(string)
	if !ok1 || !ok2 {
		return "", Session{}
	}
	return id, s
}

// TODO: Is this even needed?
//func (sessions *Sessions) SetSessionDataCtx(r *http.Request, data SessionData) {
//	id, s, ok := GetSessionCtx(r.Context())
//	if ok {
//
//	} else {
//
//	}
//}
