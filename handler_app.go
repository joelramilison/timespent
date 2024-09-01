package main

import (
	"net/http"
	"errors"
	"database/sql"
	"log"
	"github.com/joelramilison/timespent/internal/database"
)

func (cfg *apiConfig) appHandler(w http.ResponseWriter, req *http.Request, user database.User) {

	runningSession := hasRunningSession(cfg.DB, user, req)
	app(runningSession).Render(req.Context(), w)
}


func hasRunningSession(db *database.Queries, user database.User, req *http.Request) bool {
	session, err := db.GetNewestSession(req.Context(), user.ID)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			log.Printf("Error while retrieving newest session for userID %v: %v", user.ID, err)
		}
		return false
	}
	if session.EndedAt.Valid {
		// session ended already
		return false
	}
	return true
}