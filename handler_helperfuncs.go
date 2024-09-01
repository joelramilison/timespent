package main

import (
	"net/http"
	"github.com/a-h/templ"
	"bytes"
	"log"
	"github.com/joelramilison/timespent/internal/database"
	"errors"
	"database/sql"
)
func sendComponent(w http.ResponseWriter, req *http.Request, component templ.Component) {

	buf := bytes.Buffer{}
	err := component.Render(req.Context(), &buf)
	if err != nil {
		log.Printf("error rendering component: %v", err)
		w.Header().Add("HX-Redirect", "/")
		w.Write([]byte{})
		return
	}

	w.Header().Add("Content-Type", "text/html")
	w.WriteHeader(200)
	w.Write(buf.Bytes())
}

func getAppMode(db *database.Queries, user database.User, req *http.Request) int {
	session, err := db.GetNewestSession(req.Context(), user.ID)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			log.Printf("Error while retrieving newest session for userID %v: %v", user.ID, err)
		}
		return appModeNothing
	}
	if session.EndedAt.Valid {
		// session ended already
		return appModeNothing
	}
	if session.PausedAt.Valid {
		// session currently paused
		return appModePaused
	}
	return appModeRunning
}