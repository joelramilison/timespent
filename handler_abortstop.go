package main

import (
	"net/http"
	"github.com/a-h/templ"
	"log"
	"github.com/joelramilison/timespent/internal/database"
)

func (cfg *apiConfig) abortStopHandler(w http.ResponseWriter, req *http.Request, user database.User) {

	appMode := getAppMode(cfg.DB, user, req)
	if appMode == appModeRunning {
		templ.Handler(stopPauseButtons(nil)).ServeHTTP(w, req)
	} else if appMode == appModePaused {
		templ.Handler(resumeStopButtons(nil)).ServeHTTP(w, req)
	} else {
		log.Printf("appMode was neither running nor paused when aborting stop")
		w.Header().Add("HX-Redirect", "/")
		w.WriteHeader(302)
		w.Write([]byte{})
	}

}