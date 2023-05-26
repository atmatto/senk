package main

import (
	"fmt"
	"github.com/atmatto/atylar"
	"path/filepath"
)

type Storage struct {
	Root       string
	UserStores map[string]atylar.Store

	Reads  chan NoteRead
	Writes chan NoteWrite
}

func (s *Storage) LoadAll(usernames []string) error {
	var err error
	for _, u := range usernames {
		s.UserStores[u], err = atylar.New(filepath.Join(s.Root, u))
		if err != nil {
			return fmt.Errorf("failed to initialize note storage for user \"%s\": %v", u, err)
		}
	}
	return nil
}

func InitStorage(path string) Storage {
	return Storage{Root: path, UserStores: make(map[string]atylar.Store)}
}

func (db *Database) StartStorageWorker() {
	s := &db.storage
	s.Reads = make(chan NoteRead)
	s.Writes = make(chan NoteWrite)

	go func() {
		for {
			select {
			case read := <-s.Reads:
				str, err := read.Execute(db)
				read.resp <- NoteReadResp{str, err}
			case write := <-s.Writes:
				err := write.Execute(db)
				write.resp <- err
			}
		}
	}()
}
