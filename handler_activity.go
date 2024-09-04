package main

import (
	"log"
	"net/http"
	"strings"

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

	activities, err := cfg.DB.GetUserActivities(req.Context(), user.ID)
	if err != nil {
		log.Printf("couldn't get activities for user %v\n", user.ID)
		cfg.appHandler(w, req, user)
		return

	}
	activitiesPage(activities).Render(req.Context(), w)

}

func (cfg *apiConfig) createActivityHandler(w http.ResponseWriter, req *http.Request, user database.User) {

	writeError := func(errMsg string) {
		w.Header().Add("HX-Retarget", "#activitiesAndErrMsgDiv")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<p style="color: red;">` + errMsg + "</p>"))
	}

	// process activity name input field
	params, err := extractAndVerifyParams(req, []string{"createActivityName"})
	inputName := params["createActivityName"]
	if err != nil {
		writeError(err.Error())
		return
	} else if hasForbiddenCharacters(inputName) {
		writeError("only letters and numbers")
		return
	}

	oldActivities, err := cfg.DB.GetUserActivities(req.Context(), user.ID)
	if err != nil {
		log.Printf("couldn't get activities of user %v: %v", user.ID, err)
		cfg.appHandler(w, req, user)
		return
	}

	// check if the reduced name form conflicts with an existing one
	for _, oldActivity := range oldActivities {
		if reduceActivity(oldActivity.Name) == reduceActivity(inputName) {
			writeError("too similar to existing activity")
			return
		}
	}

	activity, err := cfg.DB.CreateActivity(req.Context(), database.CreateActivityParams{
		ID: uuid.New(), Name: inputName, UserID: user.ID,
	})
	if err != nil {
		log.Printf("couldn't create activity with name %v for user %v: %v", params["createActivityName"], user.ID, err)
		writeError(err.Error())
		return
	}
	
	sendComponent(w, req, activitiesPage(append([]database.Activity{activity}, oldActivities...)))
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

func (cfg *apiConfig) activityEditMenuHandler(w http.ResponseWriter, req *http.Request, user database.User) {

	if req.Header.Get("HX-Request") == "" {
		cfg.appHandler(w, req, user)
		return
	}

	activities, err := cfg.DB.GetUserActivities(req.Context(), user.ID)
	if err != nil {
		log.Printf("couldn't get activities of user %v: %v", user.ID, err)
		w.Header().Add("HX-Retarget", "#activitiesAndErrMsgDiv")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<p style="color: red;">` + "internal server error, please try again" + "</p>"))
		return
	}

	sendComponent(w, req, editList(activities))
	
}

// when user initially presses "Delete" on an activity, make him confirm
func (cfg *apiConfig) deleteActivityHandler(w http.ResponseWriter, req *http.Request, user database.User) {

	idString := req.PathValue("id")
	if idString == "" {
		log.Printf("user %v tried to delete activity but didn't receive id string\n", user.ID)
		w.WriteHeader(200)
		w.Write([]byte(`<p style="color: red;">internal error, please try again</p>`))
		return
	}

	sendComponent(w, req, confirmDeleteFormAndInfo(idString))

}

func (cfg *apiConfig) confirmDeleteActivityHandler(w http.ResponseWriter, req *http.Request, user database.User) {

	writeError := func(text string) {
		w.Header().Add("HX-Retarget", "#activitiesAndErrMsgDiv")
		w.Header().Add("HX-Reswap", "inner")
		w.WriteHeader(200)
		w.Write([]byte(`<p style="color: red;">`+ text + `</p>`))
	}

	// get activity ID from URL
	idString := req.PathValue("id")
	if idString == "" {
		log.Printf("user %v tried to delete activity but didn't receive id string\n", user.ID)
		writeError("internal error, please try again")
		return
	}

	activityID, err := uuid.Parse(idString)
	if err != nil {
		log.Printf("user %v tried to delete activity but couldn't parse activity UUID: %v\n", user.ID, err)
		writeError("internal error, please try again")
		return
	}

	// fetch activity from DB
	activity, err := cfg.DB.GetActivity(req.Context(), activityID)
	if err != nil {
		log.Printf("user %v tried to delete activity but database error: %v\n", user.ID, err)
		writeError("internal error, please try again")
		return
	}

	// check if user has typed the name of the activity correctly to confirm the action
	params, err := extractAndVerifyParams(req, []string{"confirmDeleteInput"})
	if err != nil {
		writeError(err.Error())
		return
	}
	if activity.Name != params["confirmDeleteInput"] {
		writeError("names don't match, try again")
		return
	}

	err = cfg.DB.DeleteActivity(req.Context(), activityID)
	if err != nil {
		log.Printf("user %v tried to delete activity but database error: %v\n", user.ID, err)
		writeError("internal error, please try again")
		return
	}
	
	// get new activities list to display
	activities, err := cfg.DB.GetUserActivities(req.Context(), user.ID)
	if err != nil {
		log.Printf("user %v deleted activity, then got error trying to retrieve activities list from DB: %v\n", user.ID, err)
		cfg.appHandler(w, req, user)
		return
	}

	sendComponent(w, req, activitiesPage(activities))



}