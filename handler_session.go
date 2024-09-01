package main

import (
	"bytes"
	"errors"
	"io"
	"log"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"time"
	"github.com/a-h/templ"
	"github.com/google/uuid"
	"github.com/joelramilison/timespent/internal/database"
)

func (cfg *apiConfig) startSessionHandler(w http.ResponseWriter, req *http.Request, user database.User) {

	if hasRunningSession(cfg.DB, user, req) {
	
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
	sendComponent(w, req, stopButton())
	
}


func (cfg *apiConfig) stopSessionHandler(w http.ResponseWriter, req *http.Request, user database.User) {

	if !hasRunningSession(cfg.DB, user, req) {
	
		log.Printf("error: user stopped session with no session running")
		w.Header().Add("HX-Redirect", "/")
		w.WriteHeader(200)
		w.Write([]byte{})
		return
	}

	// Retrieve from DB
	session, err := cfg.DB.GetNewestSession(req.Context(), user.ID)
	if err != nil {
		log.Printf("error fetching newest session when trying to stop session: %v", err)
		sendComponent(w, req, stopConfirmDialog("internal server error, please try again"))
		return
	}

	// Get the pause duration to subtract from the session time
	inputPauseSeconds, err := extractPauseSeconds(req, session)
	if err != nil {
		sendComponent(w, req, stopConfirmDialog(err.Error()))
		return
	}

	// Update changes in DB
	err = cfg.DB.StopSession(req.Context(), database.StopSessionParams{ID: session.ID, PauseSeconds: int32(inputPauseSeconds)})
	if err != nil {
		sendComponent(w, req, stopConfirmDialog("internal server error, please try again"))
		log.Printf("SQL error (func StopSession) while trying to stop session: %v", err)
		return
	}

	sendComponent(w, req, startButton())

}


func extractPauseSeconds(req *http.Request, session database.Session) (int, error) {

	urlEncodedParams, err := io.ReadAll(req.Body)
	if err != nil {
		log.Printf("error while trying to perform io.ReadAll: %v\n", err.Error())
		return 0,  errors.New("internal server error, please try again")
	}

	formValues, err := url.ParseQuery(string(urlEncodedParams))
	if err != nil {
		log.Printf("error while trying to parse query %v\n", string(urlEncodedParams))
		return 0,  errors.New("internal server error, please try again")
	}

	pauseString := formValues.Get("pauseMinutes")
	if pauseString == "" {
		log.Printf("couldn't get pauseMinutes from stop Confirmation dialog request")
		return 0, errors.New("couldn't process value for pause")
	}
	
	pauseMinutesFloat, err := strconv.ParseFloat(pauseString, 64)
	if err != nil {
		return 0, errors.New("pause minutes input needs to be a number")
	}
	pauseSeconds := int(math.Round(pauseMinutesFloat * 60))

	durationSinceStart := time.Since(session.StartedAt)
	if durationSinceStart < time.Duration(pauseSeconds) * time.Second {
		return 0, errors.New("error: pause longer than session")
	}

	return pauseSeconds, nil

}


func sendComponent(w http.ResponseWriter, req *http.Request, component templ.Component) {

	buf := bytes.Buffer{}
	err := component.Render(req.Context(), &buf)
	if err != nil {
		log.Printf("error rendering component: %v", err)
		w.Header().Add("HX-Redirect", "/")
		w.Write([]byte{})
		return
	}

	w.Header().Add("Content-Type", "text/html")
	w.WriteHeader(200)
	w.Write(buf.Bytes())
}