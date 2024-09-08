package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/joelramilison/timespent/internal/database"
)

func (cfg *apiConfig) activitiesPageHandler(w http.ResponseWriter, req *http.Request, user database.User) {

	/* if no HTMX request, refer back to the main page.
	It needs to be an HTMX request because we will (later) pass the current client date
	to display daily statistics next to each activity item */
	if req.Header.Get("HX-Request") == "" {
		cfg.appHandler(w, req, user)
		return
	}

	// get parameters from HTTP request
	params, err := extractAndVerifyParams(req, []string{"dayOfMonth", "month", "year"}, false)
	if err != nil {
		log.Printf("user %v - coulnd't extract params: %v", user.ID, err)
		cfg.appHandler(w, req, user)
		return
	}
	sendActivitiesPage(cfg, w, req, user, params)

}



func (cfg *apiConfig) createActivityHandler(w http.ResponseWriter, req *http.Request, user database.User) {
	
	// process activity name input field
	params, err := extractAndVerifyParams(req, []string{"createActivityName", "dayOfMonth", "month", "year"}, true)
	inputName := params["createActivityName"]
	if err != nil {
		writeError(w, err.Error(), "#activitiesAndErrMsgDiv")
		return
	} else if hasForbiddenCharacters(inputName) {
		writeError(w, "only letters and numbers", "#activitiesAndErrMsgDiv")
		return
	}
	// Fetch current list of activities
	oldActivities, err := cfg.DB.GetUserActivities(req.Context(), user.ID)
	if err != nil {
		log.Printf("couldn't get activities of user %v: %v", user.ID, err)
		writeError(w, "internal server error, please try again", "#activitiesAndErrMsgDiv")
		return
	}

	// check if the reduced name form conflicts with an existing one
	for _, oldActivity := range oldActivities {
		if reduceActivity(oldActivity.Name) == reduceActivity(inputName) {
			writeError(w, "too similar to existing activity", "#activitiesAndErrMsgDiv")
			return
		}
	}

	_, err = cfg.DB.CreateActivity(req.Context(), database.CreateActivityParams{
		ID: uuid.New(), Name: inputName, UserID: user.ID,
	})
	if err != nil {
		log.Printf("couldn't create activity with name %v for user %v: %v", params["createActivityName"], user.ID, err)
		writeError(w, "internal server error, please try again", "#activitiesAndErrMsgDiv")
		return
	}
	sendActivitiesPage(cfg, w, req, user, params)
}




func hasForbiddenCharacters(input string) bool {

	allowed := "abcdefghiklmnopqrstuvwxyz1234567890 -"
	for _, r := range strings.ToLower(input) {

		if !strings.ContainsRune(allowed, r) {
			return true
		}
	}
	return false
}



// When user clicks on "Edit": they want to edit their list of available activities
func (cfg *apiConfig) activityEditMenuHandler(w http.ResponseWriter, req *http.Request, user database.User) {

	// redirect to main page if it's not an HTMX request (made by button clicks)
	if req.Header.Get("HX-Request") == "" {
		cfg.appHandler(w, req, user)
		return
	}

	activities, err := cfg.DB.GetUserActivities(req.Context(), user.ID)
	if err != nil {
		log.Printf("couldn't get activities of user %v: %v", user.ID, err)
		writeError(w, "internal server error, please try again", "#activitiesAndErrMsgDiv")
		return
	} 

	// get time parameters from HTTP request
	params, err := extractAndVerifyParams(req, []string{"dayOfMonth", "month", "year"}, false)
	if err != nil {

		log.Printf("user %v - coulnd't extract params: %v", user.ID, err)
		writeError(w, "internal server error, please try again", "#activitiesAndErrMsgDiv")
		return
	}

	// compile the time object for the local (client-side) date 
	date, err := dataParamsToTime(params["year"], params["dayOfMonth"], params["month"])
	if err != nil {
		log.Printf("user %v - coulnd't process date: %v", user.ID, err)
		writeError(w, "internal server error, please try again", "#activitiesAndErrMsgDiv")
		return
	}

	year, month, day := date.Date()
	sendComponent(w, req, editList(activities, year, int(month), day))
	
}

// when user initially presses "Delete" on an activity, make him confirm
func (cfg *apiConfig) deleteActivityHandler(w http.ResponseWriter, req *http.Request, user database.User) {

	// Get the id of the activity to delete
	idString := req.PathValue("id")
	if idString == "" {
		log.Printf("user %v tried to delete activity but didn't receive id string\n", user.ID)
		writeError(w, "internal server error, please try again", "#activitiesAndErrMsgDiv")
		return
	}
	// get date parameters from HTTP request
	params, err := extractAndVerifyParams(req, []string{"dayOfMonth", "month", "year"}, false)
	if err != nil {
		log.Printf("user %v - coulnd't extract params: %v", user.ID, err)
		writeError(w, "internal server error, please try again", "#activitiesAndErrMsgDiv")
		return
	}

	// compile the time object for the local (client-side) date 
	date, err := dataParamsToTime(params["year"], params["dayOfMonth"], params["month"])
	if err != nil {
		log.Printf("user %v - coulnd't process date: %v", user.ID, err)
		writeError(w, "internal server error, please try again", "#activitiesAndErrMsgDiv")
		return
	}
	year, month, day := date.Date()
	sendComponent(w, req, confirmDeleteFormAndInfo(idString, year, int(month), day))

}

func (cfg *apiConfig) confirmDeleteActivityHandler(w http.ResponseWriter, req *http.Request, user database.User) {

	// get activity ID from URL
	idString := req.PathValue("id")
	if idString == "" {
		log.Printf("user %v tried to delete activity but didn't receive id string\n", user.ID)
		writeError(w, "internal server error, please try again", "#activitiesAndErrMsgDiv")
		return
	}

	// parse id into a UUID object
	activityID, err := uuid.Parse(idString)
	if err != nil {
		log.Printf("user %v tried to delete activity but couldn't parse activity UUID: %v\n", user.ID, err)
		writeError(w, "internal server error, please try again", "#activitiesAndErrMsgDiv")
		return
	}

	// fetch activity from DB by ID
	activity, err := cfg.DB.GetActivity(req.Context(), activityID)
	if err != nil {
		log.Printf("user %v tried to delete activity but database error: %v\n", user.ID, err)
		writeError(w, "internal server error, please try again", "#activitiesAndErrMsgDiv")
		return
	}

	// check if user has typed the name of the activity correctly to confirm the action
	params, err := extractAndVerifyParams(req, []string{"confirmDeleteInput"}, true)
	if err != nil {
		writeError(w, "internal server error, please try again", "#activitiesAndErrMsgDiv")
		return
	}
	if activity.Name != params["confirmDeleteInput"] {
		writeError(w, "names don't match, try again", "#activitiesAndErrMsgDiv")
		return
	}

	err = cfg.DB.DeleteActivity(req.Context(), activityID)
	if err != nil {
		log.Printf("user %v tried to delete activity but database error: %v\n", user.ID, err)
		writeError(w, "internal server error, please try again", "#activitiesAndErrMsgDiv")
		return
	}
	
	sendActivitiesPage(cfg, w, req, user, params)

}



func durationToString(duration time.Duration) string {

	hoursFloat := duration.Hours()
	minutes := int((hoursFloat - float64(int(hoursFloat))) * 60)
	hours := int(hoursFloat)


	return fmt.Sprintf("%vh%vm", hours, minutes)
	
}




 func sendActivitiesPage(cfg *apiConfig, w http.ResponseWriter, req *http.Request, user database.User, httpParams map[string]string) {
	
	// compile the time object for the local (client-side) date 
	date, err := dataParamsToTime(httpParams["year"], httpParams["dayOfMonth"], httpParams["month"])
	if err != nil {
		log.Printf("user %v - coulnd't process date: %v", user.ID, err)
		cfg.appHandler(w, req, user)
		return
	}

	// get list of all activities belonging to the user
	activities, err := cfg.DB.GetUserActivities(req.Context(), user.ID)
	if err != nil {
		log.Printf("couldn't get activities for user %v\n", user.ID)
		cfg.appHandler(w, req, user)
		return

	}

	// get a list of all sessions the user tracked on the day specified by the params
	sessions, err := cfg.DB.GetSessionsOnDay(req.Context(), database.GetSessionsOnDayParams{
		UserID: user.ID, CorrespondingDate: sql.NullTime{Valid: true, Time: date},
	})
	if err != nil {
		log.Printf("couldn't get sessions for date %v for user %v", date, user.ID)
		cfg.appHandler(w, req, user)
		return
	}

	// Now, we will compile a slice of data points: activityName & tracked time on that day for that activity.
	activityStatisticsMap := map[string]time.Duration{}
	idToActivityName := map[uuid.UUID]string{}

	
	for _, activity := range activities {
		activityStatisticsMap[activity.Name] = time.Duration(0)
		idToActivityName[activity.ID] = activity.Name
	}
	for _, session := range sessions {
		sessionDuration := (session.EndedAt.Time.Sub(session.StartedAt) -
			 time.Duration(session.PauseSeconds) * time.Second)
		activityStatisticsMap[idToActivityName[session.ActivityID.UUID]] += sessionDuration 
	}

	activityStatisticsSlice := make([]activityStatistics, 0, len(activities))
	for activityName, duration := range activityStatisticsMap {
		activityStatisticsSlice = append(activityStatisticsSlice, activityStatistics{name: activityName, duration: duration})
	}

	

	sort.Slice(activityStatisticsSlice, func(i, j int) bool {
		return strings.Compare(activityStatisticsSlice[i].name, activityStatisticsSlice[j].name) == -1 
	})
	year, month, day := date.Date()

	w.Header().Add("HX-Retarget", "body")
	w.Header().Add("HX-Reswap", "innerHTML")
	activitiesPage(activityStatisticsSlice, year, int(month), day).Render(req.Context(), w)

}

