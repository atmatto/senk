package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
)

var (
	ErrNoAccess = errors.New("user does not have the required permission")
)

type PermissionLevel int

const (
	PermissionNone PermissionLevel = iota
	PermissionRead
	PermissionWrite
)

func (p *PermissionLevel) MarshalJson() ([]byte, error) {
	switch *p {
	case PermissionRead:
		return json.Marshal("r")
	case PermissionWrite:
		return json.Marshal("w")
	default:
		return json.Marshal("0")
	}
}

func (p *PermissionLevel) UnmarshalJson(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	switch str {
	case "r":
		*p = PermissionRead
	case "w":
		*p = PermissionWrite
	default:
		*p = PermissionNone
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
	Owner  string
	Public PermissionLevel
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

func (m *Metadata) GetNoteMeta(user, slug string) NoteMeta {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.Notes[fmt.Sprintf("%s/%s", user, slug)]
}

func (m *Metadata) SetNoteMeta(user, slug string, meta NoteMeta) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Notes[fmt.Sprintf("%s/%s", user, slug)] = meta
}

func (m *Metadata) GetUserNotes(user string) (notes []NoteMeta) {
	// TODO: Maybe make this more efficient
	m.mu.RLock()
	defer m.mu.RUnlock()
	for k, n := range m.Notes {
		if b, _, _ := strings.Cut(k, "/"); b == user {
			notes = append(notes, n)
		}
	}
	return
}

type NoteWrite struct {
	user    string // user performing the action
	owner   string // note owner
	slug    string
	delete  bool // note is to be deleted if true
	content string
	resp    chan error
}

func (w *NoteWrite) Execute(db *Database) error {
	if !db.Metadata.CheckPermission(w.owner, w.slug, w.user, PermissionWrite) {
		return ErrNoAccess
	}

	s := db.storage.UserStores[w.user]

	if w.delete {
		err := s.Remove(w.slug)
		if err != nil {
			return err
		}
		return nil
	}

	f, err := s.Overwrite(w.slug)
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
	slug  string
	resp  chan NoteReadResp
}

func (w *NoteRead) Execute(db *Database) (string, error) {
	if !db.Metadata.CheckPermission(w.owner, w.slug, w.user, PermissionRead) {
		return "", ErrNoAccess
	}

	s := db.storage.UserStores[w.user]

	f, err := s.Open(w.slug, 0)
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