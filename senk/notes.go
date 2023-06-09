package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"
)

var (
	ErrNoAccess = errors.New("user does not have the required permission")
	ErrIdUsed   = errors.New("note with this id exists")
)

type PermissionLevel int

const (
	PermissionNone PermissionLevel = iota
	PermissionRead
	PermissionWrite
)

func (p PermissionLevel) MarshalJSON() ([]byte, error) {
	switch p {
	case PermissionRead:
		return json.Marshal("r")
	case PermissionWrite:
		return json.Marshal("w")
	default:
		return json.Marshal("0")
	}
}

func (p *PermissionLevel) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	switch str {
	default:
		*p = PermissionNone
	case "r":
		*p = PermissionRead
	case "w":
		*p = PermissionWrite
	}
	return nil
}

func (p *PermissionLevel) Limit(l PermissionLevel) PermissionLevel {
	if *p >= l {
		return l
	} else {
		return *p
	}
}

type NoteMeta struct {
	Owner        string
	Public       PermissionLevel
	Creation     time.Time // TODO: Show in client
	Modification time.Time
	Access       time.Time
	// TODO: Sharing
}

func (n *NoteMeta) GetPermissions(user string) PermissionLevel {
	if user == n.Owner {
		return PermissionWrite
	} else {
		return n.Public.Limit(PermissionRead)
	}
}

type Metadata struct {
	Notes map[string]NoteMeta
	mu    sync.RWMutex
}

func (m *Metadata) Initialize() {
	if m.Notes == nil {
		m.Notes = make(map[string]NoteMeta)
	}
}

func (m *Metadata) GetNoteMeta(user, id string) NoteMeta {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.Notes[fmt.Sprintf("%s/%s", user, id)]
}

func (m *Metadata) SetNoteMeta(user, id string, meta NoteMeta) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Notes[fmt.Sprintf("%s/%s", user, id)] = meta
}

func (m *Metadata) BumpNoteTimers(user, id string, write bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := fmt.Sprintf("%s/%s", user, id)
	meta := m.Notes[key]
	now := time.Now()
	if meta.Creation.IsZero() {
		meta.Creation = now
	}
	if write {
		meta.Modification = now
	}
	meta.Access = now
	m.Notes[key] = meta
}

type Note struct {
	Path     string
	Metadata NoteMeta
}

func (m *Metadata) GetUserNotes(user string) (notes []Note) {
	// TODO: Maybe make this more efficient
	m.mu.RLock()
	defer m.mu.RUnlock()
	for k, n := range m.Notes {
		if b, _, _ := strings.Cut(k, "/"); b == user {
			notes = append(notes, Note{k, n})
		}
	}
	return
}

type NoteWrite struct {
	user    string // user performing the action
	owner   string // note owner
	id      string // note id
	create  bool   // abort if note already exists
	delete  bool   // note is to be deleted if true (content is ignored)
	content string
	resp    chan error
}

func (w *NoteWrite) Execute(db *Database) error {
	s := db.storage.UserStores[w.user]

	if w.create {
		_, err := s.Stat(w.id, false)
		if !errors.Is(err, os.ErrNotExist) {
			return ErrIdUsed
		}
	} else if !db.Metadata.CheckPermission(w.owner, w.id, w.user, PermissionWrite) {
		return ErrNoAccess
	}

	db.Metadata.BumpNoteTimers(w.owner, w.id, true)

	if w.delete { // TODO: Trash
		err := s.Remove(w.id)
		if err != nil {
			return err
		}
		return nil
	}

	f, err := s.Overwrite(w.id)
	if err != nil {
		return err
	}
	_, err = f.WriteString(w.content)
	if err != nil {
		return err
	}

	return nil
}

type NoteReadResp struct {
	v   string
	err error
}

type NoteRead struct {
	user  string // user performing the action
	owner string // note owner
	id    string // note id
	resp  chan NoteReadResp
}

func (r *NoteRead) Execute(db *Database) (string, error) {
	if !db.Metadata.CheckPermission(r.owner, r.id, r.user, PermissionRead) {
		return "", ErrNoAccess // TODO: What if the note doesn't exist?
	}

	if r.user == r.owner {
		db.Metadata.BumpNoteTimers(r.user, r.id, false)
	}

	s := db.storage.UserStores[r.user]

	f, err := s.Open(r.id, 0)
	if err != nil {
		return "", err
	}

	bytes, err := io.ReadAll(f)
	if err != nil {
		return "", err
	}

	return string(bytes), nil
}

// CheckPermission returns true if the accessor has the permission
// to perform the operation to the user's note with the given slug.
func (m *Metadata) CheckPermission(owner, slug, accessor string, operation PermissionLevel) bool {
	n := m.GetNoteMeta(owner, slug)
	return n.GetPermissions(accessor) >= operation
}
