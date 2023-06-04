package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	// Load configuration and database

	dbPath := os.Getenv("SENK_DIR")
	if dbPath == "" {
		log.Fatalf("Environment variable SENK_DIR must point to the data directory.")
	}
	err := os.MkdirAll(dbPath, 0700)
	if err != nil {
		log.Fatalf("Failed to initialize data directory: %v", err)
	}

	db, err := LoadDatabase(dbPath)
	if err != nil {
		log.Fatalf("Failed to load database: %v", err)
	}

	addr := os.Getenv("SENK_ADDR")
	if addr == "" {
		addr = ":3000"
		log.Printf("Use the SENK_ADDR environment variable to set the address for the server to listen on. Using the default value: %s", addr)
	}

	// Background tasks

	ticker := time.NewTicker(time.Minute)
	go func() {
		for range ticker.C {
			err := db.Save()
			if err != nil {
				log.Printf("Failed to periodically save database.")
			}
		}
	}()

	db.StartStorageWorker()

	// TODO: for testing:
	log.Println(db.Users.AddUser("teus2", "VeryStrongP@ssword123Wtf#"))

	// Routes

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(db.Sessions.SessionRetrievalMiddleware)

	r.Post("/session/signin", db.signIn)
	r.Post("/session/signout", db.signOut)
	
	r.Get("/", db.serveMain)
	r.Get("/app.js", serveStatic("app.js", "text/javascript"))
	r.Get("/style.css", serveStatic("style.css", "text/css"))

	r.Route("/api", func (r chi.Router) {
		r.Get("/index", db.getIndex)
		r.Post("/new", db.createNote)
	})
	
	r.Route("/{user:~[a-z][a-z0-9_-]+}", func (r chi.Router) {
		r.Get("/", db.serveMain) // TODO: for anonymous users
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/raw", db.readNote)
			r.Get("/", db.serveMain) // TODO: for anonymous users
			r.Put("/", db.writeNote)
			// r.Delete("/", db.deleteNote)
		})
	})

	// Server

	server := http.Server{
		Addr:    addr,
		Handler: r,
	}

	// Cleanup

	cleanup := func() {
		log.Printf("Cleaning up...")
		ticker.Stop()
		_ = db.Save()
	}

	closed := make(chan struct{})
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		<-sigint

		log.Printf("Process interrupted, shutting down the server.")

		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			log.Printf("Error when shutting down the server: %v", err)
		}

		cleanup()

		close(closed)
	}()

	// Serving

	log.Printf("Starting the server.")

	if err = server.ListenAndServe(); err != http.ErrServerClosed {
		cleanup()
		log.Fatalf("Failed to ListenAndServer: %v", err)
	}

	<-closed

	log.Printf("Finished cleaning up.")
}
