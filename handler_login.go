package main

import (
	"database/sql"
	"errors"
	"io"
	"log"
	"net/http"
	"net/url"

	"github.com/joelramilison/timespent/internal/database"
	"golang.org/x/crypto/bcrypt"
)

func (cfg *apiConfig) loginHandler(w http.ResponseWriter, req *http.Request) {
	
	writeError := func(text string) {
		msg := coloredMessage{colorString: colorRed, text: text}
		err := login(&msg).Render(req.Context(), w)
		if err != nil {
			// if can't display the error message on page, instead refresh the page to clear input fields
			log.Printf("couldn't render login page with msg.text = %v\n", msg.text)
			w.Header().Add("HX-Redirect", "/login")
			w.WriteHeader(302)
			w.Write([]byte{})
		}

	}
	
	username, password, err := extractLoginParams(req)
	if err != nil {
		writeError(err.Error())
		return
	}
	user, err := cfg.DB.GetUserbyName(req.Context(), username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError("account with username doesn't exist")
			return
		}
		// if error but not ErrNoRows:
		writeError("Internal server error, please try again")
		return
	}

	err = bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(password))
	if err != nil {
		// wrong password
		writeError("wrong password")
		return
	}
	// Crete sessionID and set it as cookie
	hashedSessionID, sessionExpiresAt, err := createSession(w, user.ID)
	if err != nil {
		writeError("Internal server error, please try again")
		return
	}
	err = cfg.DB.UpdateLoginSession(req.Context(), database.UpdateLoginSessionParams{
		ID: user.ID, SessionIDHash: hashedSessionID, SessionExpiresAt: sessionExpiresAt,
	})
	if err != nil {
		writeError("Internal server error, please try again")
		return
	}
	w.Header().Add("HX-Redirect", "/")
	w.WriteHeader(302)
	w.Write([]byte{})

}


// returns: username, password, error
func extractLoginParams(req *http.Request) (string, string, error) {

	urlEncodedParams, err := io.ReadAll(req.Body)
	if err != nil {
		log.Printf("error while trying to perform io.ReadAll: %v\n", err.Error())
		return "", "", errors.New("internal server error, please try again")
	}
	formValues, err := url.ParseQuery(string(urlEncodedParams))
	if err != nil {
		log.Printf("error while trying to parse query %v\n", string(urlEncodedParams))
		return "", "", errors.New("internal server error, please try again")
	}
	username := formValues.Get("username")
	password := formValues.Get("password")

	if username == "" || password == "" {
		return "", "", errors.New("not all fields filled out, please try again")
	}
	
	return username, password, nil
}