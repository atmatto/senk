// simple database for data about users

package main

import (
	"encoding/json"
	"errors"
	"log"
	"os"
	"path/filepath"
)

type Database struct {
	file     string // path to database file
	Users    Users
	Sessions Sessions
	Metadata Metadata
	storage  Storage
}

func (db *Database) Save() error {
	db.Users.mu.RLock()
	defer db.Users.mu.RUnlock()
	db.Sessions.mu.RLock()
	defer db.Sessions.mu.RUnlock()
	db.Metadata.mu.RLock()
	defer db.Metadata.mu.RUnlock()

	bytes, err := json.Marshal(db)
	if err != nil {
		log.Printf("Error marshalling database: %v", err)
		return err
	}
	err = os.WriteFile(db.file, bytes, 0600)
	if err != nil {
		log.Printf("Error saving database: %v", err)
	}
	return err
}

func LoadDatabase(path string) (*Database, error) {
	var db Database
	db.file = filepath.Join(path, "_db")

	bytes, err := os.ReadFile(db.file)
	if errors.Is(err, os.ErrNotExist) {
		log.Printf("Database file does not exist, will create.")
	} else if err != nil {
		log.Printf("Error reading database: %v", err)
		return nil, err
	} else {
		err = json.Unmarshal(bytes, &db)
		if err != nil {
			log.Printf("Error unmarshalling database: %v", err)
			return nil, err
		}
	}

	db.storage = InitStorage(path)
	err = db.storage.LoadAll(db.Users.GetAllUsernames())
	if err != nil {
		log.Printf("Error initializing storage: %v", err)
		return nil, err
	}

	return &db, nil
}
