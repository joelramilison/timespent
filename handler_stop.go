package main

import (
	"database/sql"
	"errors"
	"io"
	"log"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"time"
	"github.com/joelramilison/timespent/internal/database"
)



// this should return the confirmation dialog to stop the session
func (cfg *apiConfig) stopSessionHandler(w http.ResponseWriter, req *http.Request, user database.User) {

	appMode, _ := getAppMode(cfg.DB, user, req)

	returnError := func(err error) {
		if appMode == appModeRunning {
			sendComponent(w, req, stopPauseButtons(err))
			return
		} else if appMode == appModePaused {
			sendComponent(w, req, resumeStopButtons(err))
			return
		} else {
			log.Print("serious error: tried to stop session while appMode is not running nor paused\n")
			w.Header().Add("HX-Redirect", "/")
			w.WriteHeader(200)
			w.Write([]byte{})
			return
		}
	}

	urlEncodedParams, err := io.ReadAll(req.Body)
	if err != nil {
		log.Printf("error while trying to perform io.ReadAll: %v\n", err.Error())
		returnError(errors.New("internal server error, please try again"))
		return
	}
	params, err := url.ParseQuery(string(urlEncodedParams))
	if err != nil {
		log.Printf("error while trying to parse query %v\n", string(urlEncodedParams))
		returnError(errors.New("internal server error, please try again"))
		return
	}
	clientHoursString := params.Get("hours")

	if clientHoursString == "" {
		log.Printf("empty clientHoursString when trying to get parameter in stopSessionHandler\n")
		returnError(errors.New("not all fields filled out, please try again"))
		return
	}

	clientHours, err := strconv.Atoi(clientHoursString)
	if err != nil {
		log.Printf("error while trying to convert %v to integer\n", clientHours)
		returnError(errors.New("internal server error, please try again"))
		return
	}

	session, err := cfg.DB.GetNewestSession(req.Context(), user.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Printf("couldn't find newest session DB although user %v wants to stop\n", user.ID)
			returnError(errors.New("internal server error, please try again"))
			return

		} else {
			log.Printf("Database error while trying to get newest sessions for user %v: %v", user.ID, err)
			returnError(errors.New("internal server error, please try again"))
			return

		}
	}
	askToReassign := ifAskToReassignDay(clientHours, session.StartedAt, time.Now())
	if askToReassign {
		err = cfg.DB.UpdateAssignAwait(req.Context(), database.UpdateAssignAwaitParams{
				ID: user.ID, AwaitAssignDecisionUntil: sql.NullTime{
					Valid: true, Time: time.Now().Add(time.Duration(30) * time.Minute),
				},
		})
		if err != nil {
			log.Printf("Database error while trying to update assign await for user %v: %v", user.ID, err)
			returnError(errors.New("internal server error, please try again"))
			return
		}
	}
	sendComponent(w, req, stopConfirmDialog(nil, askToReassign))
		
}


func (cfg *apiConfig) confirmSessionStopHandler(w http.ResponseWriter, req *http.Request, user database.User) {

	// if there is no running session
	appMode, _ := getAppMode(cfg.DB, user, req)
	if appMode == appModeNothing {
	
		log.Printf("error: user stopped session with no session running")
		w.Header().Add("HX-Redirect", "/")
		w.WriteHeader(200)
		w.Write([]byte{})
		return
	}

	askForAssignDecision := user.AwaitAssignDecisionUntil.Valid && user.AwaitAssignDecisionUntil.Time.After(time.Now()) 

	// Retrieve from DB
	session, err := cfg.DB.GetNewestSession(req.Context(), user.ID)
	if err != nil {
		log.Printf("error fetching newest session when trying to stop session: %v", err)
		sendComponent(w, req, stopConfirmDialog(errors.New("internal server error, please try again"), askForAssignDecision))
		return
	}


	// Get the pause duration to subtract from the session time
	inputPauseSeconds, assignYesterdayChoice, err := extractStopParams(req, session, askForAssignDecision)
	if err != nil {
		sendComponent(w, req, stopConfirmDialog(err, askForAssignDecision))
		return
	}
	assignYesterday := sql.NullBool{Valid: false}
	if askForAssignDecision {
		if assignYesterdayChoice == "yesterday" {
			assignYesterday.Valid = true
			assignYesterday.Bool = true
		}
	}

	// If user is in pause right now, also add the current pause duration to the database
	var fromCurrentPause int32
	if session.PausedAt.Valid {
		fromCurrentPause = int32(time.Since(session.PausedAt.Time).Seconds())
	}

	// Update changes in DB
	err = cfg.DB.StopSession(req.Context(), database.StopSessionParams{
		ID: session.ID,
		PauseSeconds: inputPauseSeconds + session.PauseSeconds + fromCurrentPause,
		AssignToDayBeforeStart: assignYesterday})
	if err != nil {
		sendComponent(w, req, stopConfirmDialog(errors.New("internal server error, please try again"), askForAssignDecision))
		log.Printf("SQL error (func StopSession) while trying to stop session: %v", err)
		return
	}

	activities, err := cfg.DB.GetUserActivities(req.Context(), user.ID)
	if err != nil {
			sendComponent(w, req, startButton(errors.New("stopped session but couldn't load activity list")))
			return

	}

	w.Header().Add("HX-Retarget", "body")
	w.Header().Add("HX-Reswap", "innerHTML")
	sendComponent(w, req, appBodyInner(appModeNothing, database.Activity{}, activities))
	
	

}


func extractStopParams(req *http.Request, session database.Session, askForAssignDecision bool) (int32, string, error) {

	urlEncodedParams, err := io.ReadAll(req.Body)
	if err != nil {
		log.Printf("error while trying to perform io.ReadAll: %v\n", err.Error())
		return 0, "", errors.New("internal server error, please try again")
	}

	params, err := url.ParseQuery(string(urlEncodedParams))
	if err != nil {
		log.Printf("error while trying to parse query %v\n", string(urlEncodedParams))
		return 0, "", errors.New("internal server error, please try again")
	}

	pauseString := params.Get("pauseMinutes")
	if pauseString == "" {
		log.Printf("couldn't get pauseMinutes from stop Confirmation dialog request")
		return 0, "", errors.New("couldn't process value for pause")
	}
	assignYesterdayChoice := params.Get("assignYesterdayGroup")
	if askForAssignDecision && assignYesterdayChoice == "" {
		return 0, "", errors.New("you need to click one of the radio buttons")
	}

	pauseMinutesFloat, err := strconv.ParseFloat(pauseString, 64)
	if err != nil {
		return 0, "", errors.New("pause minutes input needs to be a number")
	}
	pauseSeconds := int32(math.Round(pauseMinutesFloat * 60))

	durationSinceStart := time.Since(session.StartedAt)
	if durationSinceStart < time.Duration(pauseSeconds) * time.Second {
		return 0, "", errors.New("error: pause longer than session")
	}

	return pauseSeconds, assignYesterdayChoice, nil

}



// the assignment is only important for the statistics.
// if the session started at night (after midnight), it might be a good idea
// to ask the user if they want to count it as the day before.
func ifAskToReassignDay(clientHoursNow int, sessionStartTime, timeNow time.Time) bool {
	
	sinceStart := timeNow.Sub(sessionStartTime)
	hoursSinceStart := int(sinceStart.Hours())

	var clientStartHour int

	// Divide the factor 24 out because time of day is cyclical
	hoursShortened := hoursSinceStart % 24
	if clientHoursNow >= hoursShortened {
		clientStartHour = clientHoursNow - hoursShortened
	} else {
		clientStartHour = 24 - (hoursShortened - clientHoursNow)
	}
	
	// if the session was started before roughly 6:00 
	// AND it's before roughly 12:00 on the same day
	if clientStartHour < 6 && hoursSinceStart < 12 - clientStartHour {
		return true
	}
	return false
	
	
}