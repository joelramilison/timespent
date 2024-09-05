package main

import (
	"net/http"
	"database/sql"
	"errors"
	"log"
	"github.com/joelramilison/timespent/internal/database"
)

const (
    appModeRunning = 1
    appModePaused = 2
    appModeNothing = 3
)

func (cfg *apiConfig) appHandler(w http.ResponseWriter, req *http.Request, user database.User) {

	/* depending on HTMX vs. normal browser request,
	return either the whole page or just the body for a hx-swap */
	showApp := func(appMode int, activity database.Activity, activities []database.Activity) {
		if req.Header.Get("HX-Request") == "" {
			app(appMode, activity, activities).Render(req.Context(), w)
		} else {
			appBodyInner(appMode, activity, activities).Render(req.Context(), w)
		}
	}

	appMode, session := getAppMode(cfg.DB, user, req)

	// if no session currently running
	if appMode == appModeNothing {
		activities, err := cfg.DB.GetUserActivities(req.Context(), user.ID)
		if err != nil {
			if !errors.Is(err, sql.ErrNoRows) {
				log.Printf("couldn't get user activities from DB for user %v, error: %v\n", user.ID, err)
			}
			activities = []database.Activity{}
		}
		// 'activities' will populate the activity selector for starting a new session
		showApp(appMode, database.Activity{}, activities)

	// if appMode == appModePause or appModeRunning
	} else {
		activity, err := cfg.DB.GetActivity(req.Context(), session.ActivityID.UUID)
		if err != nil {
			log.Printf("couldn't get activity with ID %v, error: %v\n", session.ActivityID.UUID, err)
			showApp(appMode, database.Activity{
				Name: "couldn't load activity", ColorCode: "000000"}, nil)

		} else {
			showApp(appMode, activity, nil)

		}
		
	}
	
}


