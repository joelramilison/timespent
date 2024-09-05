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

	// generalize a function to return an error
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

	// Get the pause duration to subtract from the session time
	inputPauseSeconds, assignYesterdayChoice, askedForAssignDecision, err := extractStopParams(req)
	if err != nil {
		sendComponent(w, req, stopConfirmDialog(err, askedForAssignDecision))
		return
	}

	// Retrieve from DB
	session, err := cfg.DB.GetNewestSession(req.Context(), user.ID)
	if err != nil {
		log.Printf("error fetching newest session when trying to stop session: %v", err)
		sendComponent(w, req, stopConfirmDialog(errors.New("internal server error, please try again"), askedForAssignDecision))
		return
	}

	// If user is in pause right now, also add the current pause duration to the database
	var fromCurrentPause int32
	if session.PausedAt.Valid {
		fromCurrentPause = int32(time.Since(session.PausedAt.Time).Seconds())
	}

	// throw an error if the input pause + the current tracked pause
	// is longer than the tracked working time
	durationSinceStart := time.Since(session.StartedAt)
	totalPause := inputPauseSeconds + session.PauseSeconds + fromCurrentPause

	if durationSinceStart.Seconds() <  float64(totalPause) {
		sendComponent(w, req, stopConfirmDialog(errors.New("entered pause duration is too high"), askedForAssignDecision))
		return
	}


	var assignToLocalDate sql.NullTime 

	// set the date depending on whether the user chose to assign the session to yesterday
	if assignYesterdayChoice == "yesterday" {
		assignToLocalDate = sql.NullTime{Valid: true, Time: session.StartedAtLocalDate.AddDate(0, 0, -1)}
	} else {
		// else: either user chose to assign to yesterday OR they weren't even asked
		assignToLocalDate = sql.NullTime{Valid: true, Time: session.StartedAtLocalDate}
	}

	

	// Update changes in DB
	err = cfg.DB.StopSession(req.Context(), database.StopSessionParams{
		ID: session.ID,
		PauseSeconds: inputPauseSeconds + session.PauseSeconds + fromCurrentPause,
		CorrespondingDate: assignToLocalDate,})
	if err != nil {
		sendComponent(w, req, stopConfirmDialog(errors.New("internal server error, please try again"), askedForAssignDecision))
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

// this function returns askedForAssignchoice even if the user
// hasn't clicked any radio buttons (instead of returning zero values only)
func extractStopParams(req *http.Request) (int32, string, bool, error) {

	urlEncodedParams, err := io.ReadAll(req.Body)
	if err != nil {
		log.Printf("error while trying to perform io.ReadAll: %v\n", err.Error())
		return 0, "", false, errors.New("internal server error, please try again")
	}

	params, err := url.ParseQuery(string(urlEncodedParams))
	if err != nil {
		log.Printf("error while trying to parse query %v\n", string(urlEncodedParams))
		return 0, "", false, errors.New("internal server error, please try again")
	}

	pauseString := params.Get("pauseMinutes")
	if pauseString == "" {
		log.Printf("couldn't get pauseMinutes from stop Confirmation dialog request")
		return 0, "", false, errors.New("couldn't process value for pause")
	}

	askedForAssignChoiceString := params.Get("askedForAssignChoice")
	if askedForAssignChoiceString == "" {
		log.Printf("couldn't get askedForAssignChoice from stop Confirmation dialog request")
		return 0, "", false, errors.New("internal server error, please try again")
	}
	askedForAssignChoice, err := strconv.ParseBool(askedForAssignChoiceString)
	if err != nil {
		log.Printf("couldn't prse askedForAssignChoice from stop Confirmation dialog request into bool")
		return 0, "", false, errors.New("internal server error, please try again")	
	}
	
	assignYesterdayChoice := params.Get("assignYesterdayGroup")
	if askedForAssignChoice && assignYesterdayChoice == "" {
		return 0, "", askedForAssignChoice, errors.New("you need to click one of the radio buttons")
	}

	pauseMinutesFloat, err := strconv.ParseFloat(pauseString, 64)
	if err != nil {
		return 0, "", askedForAssignChoice, errors.New("pause minutes input needs to be a number")
	}
	pauseSeconds := int32(math.Round(pauseMinutesFloat * 60))
	return pauseSeconds, assignYesterdayChoice, askedForAssignChoice, nil

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