package main

import (
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"
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

	// get parameters from HTTP request
	params, err := extractAndVerifyParams(req, []string{"dayOfMonth", "month", "year", "activitySelect"})
	if err != nil {
		log.Printf("user %v wanted to start session but coulnd't extract params: %v", user.ID, err)
		sendComponent(w, req, startButton(errors.New("internal server error, please try again")))
		return
	}

	// compile the time object for the local (client-side) date when the session started
	localDateTime, err := dataParamsToTime(params["year"], params["dayOfMonth"], params["month"])
	if err != nil {

		log.Printf("user %v wanted to start session but coulnd't process date: %v", user.ID, err)
		sendComponent(w, req, startButton(errors.New("internal server error, please try again")))
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
	
	// reduced: whitespace removed from the name to make it viable for HTML
	reducedSelected := params["activitySelect"]
	var activityMatch database.Activity
	var found bool

	// find out which activity the user selected for this session
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
		StartedAtLocalDate: localDateTime,
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

// takes string parameters to compile a final time.Time date object
func dataParamsToTime(year, dayOfMonth, month string) (time.Time, error) {

	// final string: YYYY-MM-DD
	
	if dayOfMonthInt, err := strconv.Atoi(dayOfMonth); err != nil {
		return time.Time{}, err
	} else if dayOfMonthInt < 10 {
		dayOfMonth = "0" + dayOfMonth
	}	
	if monthInt, err := strconv.Atoi(month); err != nil {
		return time.Time{}, err
	} else if monthInt < 10 {
		month = "0" + month
	}
	dateString := year + "-" + month + "-" + dayOfMonth
	localDateTime, err := time.Parse(time.DateOnly, dateString)
	if err != nil {
		return time.Time{}, err
	}
	return localDateTime, nil
}