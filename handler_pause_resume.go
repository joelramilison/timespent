package main

import (
	"net/http"
	"log"
	"database/sql"
	"errors"
	"time"
	"math"
	"github.com/joelramilison/timespent/internal/database"
)

func (cfg *apiConfig) pauseSessionHandler(w http.ResponseWriter, req *http.Request, user database.User) {

	session, err := cfg.DB.GetNewestSession(req.Context(), user.ID)
	if err != nil {
		log.Printf("error trying to get newest session to pause: %v", err)
		sendComponent(w, req, stopPauseButtons(errors.New("internal server error, please try again")))
		return
	}

	err = cfg.DB.PauseSession(req.Context(), database.PauseSessionParams{
			PausedAt: sql.NullTime{Time: time.Now(), Valid: true}, ID: session.ID,
	})
	if err != nil {
		log.Printf("Error executing the PauseSession query: %v", err)
		sendComponent(w, req, stopPauseButtons(errors.New("internal server error, please try again")))
		return
	}
	sendComponent(w, req, resumeStopButtons(nil))

}

func (cfg *apiConfig) resumeSessionHandler(w http.ResponseWriter, req *http.Request, user database.User) {

	session, err := cfg.DB.GetNewestSession(req.Context(), user.ID)
	if err != nil {
		log.Printf("error trying to get newest session to resume: %v", err)
		sendComponent(w, req, resumeStopButtons(errors.New("internal server error, please try again")))
		return
	}
	if !session.PausedAt.Valid {
		log.Printf("user %v wanted to resume session but the most recent session isn't paused", user.ID)
		sendComponent(w, req, resumeStopButtons(errors.New("internal server error, please try again")))
		return
	}
	pauseDuration := time.Since(session.PausedAt.Time)
	pauseDurationSeconds := int32(math.Round(pauseDuration.Seconds()))
	
	err = cfg.DB.ResumeSession(req.Context(), database.ResumeSessionParams{
		PauseSeconds: pauseDurationSeconds + session.PauseSeconds, ID: session.ID,
	})
	if err != nil {
		log.Printf("user %v wanted to resume session but resumeSession query failed: %v", user.ID, err)
		sendComponent(w, req, resumeStopButtons(errors.New("internal server error, please try again")))
		return
	}
	sendComponent(w, req, stopPauseButtons(nil))

}