package main

import (
	"net/http"
	"github.com/a-h/templ"
	"bytes"
	"log"
	"github.com/joelramilison/timespent/internal/database"
	"errors"
	"database/sql"
	"io"
	"net/url"
)

// Renders and sends an HTML component to the user as an HTMX swap response.
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

// returns appMode and the most recent session for further use
func getAppMode(db *database.Queries, user database.User, req *http.Request) (int, database.Session) {
	session, err := db.GetNewestSession(req.Context(), user.ID)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			log.Printf("Error while retrieving newest session for userID %v: %v", user.ID, err)
		}
		return appModeNothing, session
	}
	if session.EndedAt.Valid {
		// session ended already
		return appModeNothing, session
	}
	if session.PausedAt.Valid {
		// session currently paused
		return appModePaused, session
	}
	return appModeRunning, session
}


// returns error if an error occured or if the value for at least one key is "" 
func extractAndVerifyParams(req *http.Request, keys []string, includeNotAllFieldsFileldOuterror bool) (map[string]string, error) {

	result := map[string]string{}

	urlEncodedParams, err := io.ReadAll(req.Body)
	if err != nil {
		log.Printf("error while trying to perform io.ReadAll: %v\n", err.Error())
		return result, errors.New("internal server error, please try again")
	}
	params, err := url.ParseQuery(string(urlEncodedParams))
	if err != nil {
		log.Printf("error while trying to parse query %v\n", string(urlEncodedParams))
		return result, errors.New("internal server error, please try again")
	}
	
	for _, key := range keys {
		if result[key] = params.Get(key); result[key] == "" && includeNotAllFieldsFileldOuterror {
			return map[string]string{}, errors.New("not all fields were filled out")
		}
		
	}

	return result, nil
}


// Send an HTTP response with a red-colored HTML error message, performing an HTMX inner swap 
func writeError(w http.ResponseWriter, errMsg, cssIdentifier string) {

	w.Header().Add("HX-Retarget", cssIdentifier)
	w.Header().Add("HX-Reswap", "innerHTML")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`<p style="color: red;">` + errMsg + "</p>"))

}