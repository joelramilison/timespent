package main

import (
	"log"
	"net/http"
	"github.com/google/uuid"
	"github.com/joelramilison/timespent/internal/database"
)

func (cfg *apiConfig) startSessionHandler(w http.ResponseWriter, req *http.Request, user database.User) {

	// if a session is already running or paused
	if getAppMode(cfg.DB, user, req) != appModeNothing {
	
		log.Printf("error: user started session with session already running")
		// refresh page. This error absolutely shouldn't occur
		w.Header().Add("HX-Redirect", "/")
		w.WriteHeader(200)
		w.Write([]byte{})
		return
	}

	// Create DB entry
	err := cfg.DB.StartSession(req.Context(), database.StartSessionParams{
		ID: uuid.New(), UserID: user.ID,
	})
	if err != nil {
		log.Printf("error: couldn't start session, query failed")
		w.Header().Add("HX-Redirect", "/")
		w.WriteHeader(200)
		w.Write([]byte{})
		return
	}
	
	// Replace the start button with a stop button
	sendComponent(w, req, stopPauseButtons(nil))
	
}

