package main

import (
	"errors"
	"io"
	"log"
	"net/http"
	"net/url"

	"github.com/google/uuid"
	"github.com/joelramilison/timespent/internal/database"
)

func (cfg *apiConfig) startSessionHandler(w http.ResponseWriter, req *http.Request, user database.User) {

	// if a session is already running or paused
	appMode, _ := getAppMode(cfg.DB, user, req)
	if appMode != appModeNothing {
	
		log.Printf("error: user started session with session already running")
		// refresh page. This error absolutely shouldn't occur
		w.Header().Add("HX-Redirect", "/")
		w.WriteHeader(200)
		w.Write([]byte{})
		return
	}

	activities, err := cfg.DB.GetUserActivities(req.Context(), user.ID)
	if err != nil {
		log.Printf("couldn't get activities list for user %v, err: %v\n", user.ID, err)
		sendComponent(w, req, startButton(errors.New("internal server error, please try again")))
		return
	}
	if len(activities) == 0 {
		// if user has no activities
		sendComponent(w, req, startButton(errors.New("first, create an activity to track")))
		return
	}
	reducedSelected, err := getSelectedActivity(req)
	if err != nil {
		sendComponent(w, req, startButton(err))
		return
	}
	var activityMatch database.Activity
	var found bool

	for _, activity := range activities {
		if reduceActivity(activity.Name) == reducedSelected {
			activityMatch = activity
			found = true
			break
		}
	}
	if !found {
		log.Printf("selected activity can't be matched to one of the user's activities")
		sendComponent(w, req, startButton(errors.New("internal server error, please try again")))
		return
	}
	
	// Create DB entry
	err = cfg.DB.StartSession(req.Context(), database.StartSessionParams{
		ID: uuid.New(), UserID: user.ID, ActivityID: uuid.NullUUID{UUID: activityMatch.ID, Valid: true},
	})
	if err != nil {
		log.Printf("error: couldn't start session, query failed: %v", err)
		w.Header().Add("HX-Redirect", "/")
		w.WriteHeader(200)
		w.Write([]byte{})
		return
	}
	
	// Replace the start button with a stop button
	w.Header().Add("HX-Retarget", "body")
	w.Header().Add("HX-Reswap", "innerHTML")
	sendComponent(w, req, appBodyInner(appModeRunning, activityMatch, nil))
	
}


func getSelectedActivity(req *http.Request) (string, error) {

	urlEncodedParams, err := io.ReadAll(req.Body)
	if err != nil {
		log.Printf("error while trying to perform io.ReadAll: %v\n", err.Error())
		return "", errors.New("internal server error, please try again")
	}
	requestValues, err := url.ParseQuery(string(urlEncodedParams))
	if err != nil {
		log.Printf("error while trying to parse query %v\n", string(urlEncodedParams))
		return  "", errors.New("internal server error, please try again")
	}

	reducedActivity := requestValues.Get("activitySelect")
	if reducedActivity == "" {
		log.Printf("user has activities but no selected activity found for start")
		return  "", errors.New("internal server error, please try again")
	}
	return reducedActivity, nil

}