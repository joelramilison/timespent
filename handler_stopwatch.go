package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/joelramilison/timespent/internal/database"
)

func (cfg *apiConfig) stopWatchHandler(w http.ResponseWriter, req *http.Request, user database.User) {
	
	sendZero := func() {
		w.Header().Add("Content-Type", "text/html")
		w.WriteHeader(200)
		w.Write([]byte("00:00:00"))
	}
	/*_, userID, err := extractFromCookie(req)
	if err != nil {
		log.Printf("user sent stopwatch refresh request but couldn't extract cookies: %v", err)
		sendZero()
		return
	}*/


	session, err := cfg.DB.GetNewestSession(req.Context(), user.ID)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			log.Printf("Error while retrieving newest session for userID %v: %v", user.ID, err)
		}
		sendZero()
		return
	}
	if session.EndedAt.Valid {
		// session ended already
		sendZero()
		return
	}
	
	responseString, err := processStopwatchTime(session)
	if err != nil {
		log.Printf("Error while processing stopwatch time: %v", err)
		sendZero()
		return
	}
	
	w.Header().Add("Content-Type", "text/html")
	w.WriteHeader(200)
	w.Write([]byte(responseString))

}



func processStopwatchTime(session database.Session) (string, error) {

	if session.PauseSeconds < 0 {
		return "", fmt.Errorf("pause time %v is smaller than 0", session.PauseSeconds)
	}
	pauseDuration := time.Duration(session.PauseSeconds) * time.Second
	
	// time since session start (and subtract the pause duration)
	timeElapsed := time.Since(session.StartedAt.Add(pauseDuration))
	if timeElapsed <= 0 {
		return "", fmt.Errorf("session at uuid %v: %v = timeElapsed (timeNow - pauseDuration - startTime < 0)",
		session.ID, timeElapsed)
	}

	hoursTruncated := int(timeElapsed.Hours())
	minutesTruncated := int(timeElapsed.Minutes()) - 60 * hoursTruncated
	secondsRaw := timeElapsed.Seconds() - 3600.0 * float64(hoursTruncated) - 60.0 * float64(minutesTruncated)
	secondsRounded := int(math.Round(secondsRaw))

	// Fix the time formatting
	if secondsRounded == 60 {
		minutesTruncated += 1
		secondsRounded = 0
	}
	if minutesTruncated == 60 {
		hoursTruncated += 1
		minutesTruncated = 0
	}
	
	hoursString := ""
	if hoursTruncated < 10 {
		hoursString += "0"
	}
	hoursString += strconv.Itoa(hoursTruncated)

	minutesString := ""
	if minutesTruncated < 10 {
		minutesString += "0"
	}
	minutesString += strconv.Itoa(minutesTruncated)

	secondsString := ""
	if secondsRounded < 10 {
		secondsString += "0"
	}
	secondsString += strconv.Itoa(secondsRounded)
	
	responseString := hoursString + ":" + minutesString + ":" + secondsString
	return responseString, nil
}