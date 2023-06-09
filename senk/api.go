// general api

package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func (db *Database) getIndex(w http.ResponseWriter, r *http.Request) {
	_, session, ok := GetSessionCtx(r.Context())
	if !ok || !session.Data.Authenticated {
		http.Error(w, "Not authenticated", http.StatusForbidden)
		return
	}

	// TODO: Sort by recency
	notes := db.Metadata.GetUserNotes(session.Data.Username)
	bytes, err := json.Marshal(notes)
	if err != nil {
		http.Error(w, "Couldn't marshal note index", http.StatusInternalServerError)
		log.Printf("Error marshalling note index: %v", err)
		return
	}
	w.Write(bytes)
}

// expects following chi URL params: user, id
func (db *Database) readNote(w http.ResponseWriter, r *http.Request) {
	user := strings.TrimPrefix(chi.URLParam(r, "user"), "~")
	note := chi.URLParam(r, "id")
	_, session, _ := GetSessionCtx(r.Context())
	if !session.Data.Authenticated {
		session.Data.Username = ""
	}

	respc := make(chan NoteReadResp)
	db.storage.Reads <- NoteRead{
		user: session.Data.Username,
		owner: user,
		id: note,
		resp: respc,
	}

	resp := <- respc
	if errors.Is(resp.err, os.ErrNotExist) {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	} else if errors.Is(resp.err, ErrNoAccess) {
		http.Error(w, "Insufficient permissions", http.StatusForbidden)
		return
	} else if resp.err != nil {
		http.Error(w, "Undefined error", http.StatusInternalServerError)
		log.Printf("Error serving file read request: %v", resp.err)
		return
	}
	w.Write([]byte(resp.v))
}

// expects following chi URL params: user, id
func (db *Database) writeNote(w http.ResponseWriter, r *http.Request) {
	user := strings.TrimPrefix(chi.URLParam(r, "user"), "~")
	note := chi.URLParam(r, "id")
	_, session, _ := GetSessionCtx(r.Context())
	if !session.Data.Authenticated {
		http.Error(w, "Only authenticated users can edit notes", http.StatusForbidden)
		return
	}

	bytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		log.Printf("Error serving note write request (couldn't read request body): %v", err)
		return
	}

	respc := make(chan error)
	db.storage.Writes <- NoteWrite{
		user: session.Data.Username,
		owner: user,
		id: note,
		delete: false,
		content: string(bytes),
		resp: respc,
	}

	err = <- respc

	if errors.Is(err, ErrNoAccess) {
		http.Error(w, "Insufficient permissions", http.StatusForbidden)
		return
	} else if err != nil {
		http.Error(w, "Undefined error", http.StatusInternalServerError)
		log.Printf("Error serving note write request: %v", err)
		return
	}
}

func (db *Database) createNote(w http.ResponseWriter, r* http.Request) {
	_, session, _ := GetSessionCtx(r.Context())
	if !session.Data.Authenticated {
		http.Error(w, "Only authenticated users can create notes", http.StatusForbidden)
		return
	}

	var id string
	for i := 0; i < 10; i++ { // Retry in case of id collision, at most 10 times
		id = uuid.NewString()

		respc := make(chan error)
		db.storage.Writes <- NoteWrite{
			user: session.Data.Username,
			owner: session.Data.Username,
			id: id,
			create: true,
			delete: false,
			content: "",
			resp: respc,
		}

		err := <- respc

		if errors.Is(err, ErrNoAccess) {
			http.Error(w, "Insufficient permissions", http.StatusForbidden)
			return
		} else if errors.Is(err, ErrIdUsed) {
			log.Printf("Note ID collision: ~%s/%s", session.Data.Username, id)
			id = ""
			continue
		} else if err != nil {
			http.Error(w, "Undefined error", http.StatusInternalServerError)
			log.Printf("Error serving note delete request: %v", err)
			return
		}
		break
	}

	if id == "" {
		http.Error(w, "Couldn't assign unique note ID, try again.", http.StatusInternalServerError)
		return
	}

	now := time.Now()
	db.Metadata.SetNoteMeta(session.Data.Username, id, NoteMeta{session.Data.Username, PermissionNone, now, now, now})

	w.Write([]byte(id))
}

// TODO: Enable after implementing history
/* func (db *Database) deleteNote(w http.ResponseWriter, r* http.Request) {
	user := strings.TrimPrefix(chi.URLParam(r, "user"), "~")
	note := chi.URLParam(r, "id")
	_, session, _ := GetSessionCtx(r.Context())
	if !session.Data.Authenticated {
		http.Error(w, "Only authenticated users can delete notes", http.StatusForbidden)
		return
	}

	respc := make(chan error)
	db.storage.Writes <- NoteWrite{
		user: session.Data.Username,
		owner: user,
		id: note,
		delete: true,
		content: "",
		resp: respc,
	}

	err := <- respc

	if errors.Is(err, ErrNoAccess) {
		http.Error(w, "Insufficient permissions", http.StatusForbidden)
		return
	} else if err != nil {
		http.Error(w, "Undefined error", http.StatusInternalServerError)
		log.Printf("Error serving note delete request: %v", err)
		return
	}
} */
