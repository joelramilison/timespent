package main

import (
	"net/http"
	"github.com/joelramilison/timespent/internal/database"
)

const (
    appModeRunning = 1
    appModePaused = 2
    appModeNothing = 3
)

func (cfg *apiConfig) appHandler(w http.ResponseWriter, req *http.Request, user database.User) {

	appMode := getAppMode(cfg.DB, user, req)
	app(appMode).Render(req.Context(), w)
}


