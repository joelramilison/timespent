package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"github.com/google/uuid"
	"github.com/joelramilison/timespent/internal/database"
	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)



func (cfg *apiConfig) registerUserHandler(w http.ResponseWriter, req *http.Request) {
	
	
	writeError := func(text string) {
		w.Header().Add("Content-Type", "text/html")
		w.WriteHeader(200)
		w.Write([]byte(fmt.Sprintf(`<p style="color: red;">%v</p>`, text)))
	}

	username, password, err := extractRegisterParams(req)
	if err != nil {
		writeError(err.Error())
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		writeError("internal error, please try again")
		log.Printf("couldn't hash refresh token, error: %v\n", err)
		return
	}

	userID := uuid.New()

	// create sessionID and set token
	hashedSessionID, sessionExpiresAt, err := createSession(w, userID)
	if err != nil {
		writeError("Internal server error, please try again")
		return
	}

	// Create user row in database
	err = cfg.DB.CreateUser(req.Context(), database.CreateUserParams{
		ID: userID, PasswordHash: hashedPassword, Username: username,
		SessionIDHash: hashedSessionID, SessionExpiresAt: sessionExpiresAt,
	})
	if err != nil {
		if pgerr, ok := err.(*pq.Error); ok {
			if pgerr.Code.Name() == "unique_violation" {
				writeError("Username exists already, please try another one")
				return
			}
		}
		// if error but not "unique_violation":
		writeError("Internal server error, please try again")
		log.Printf("internal error while trying to add new user to DB: %v", err)
		return
	}
	
	w.Header().Add("HX-Redirect", "/")
	w.WriteHeader(302)
	w.Write([]byte{})

}

// returns: username, password, timeZone, error
func extractRegisterParams(req *http.Request) (string, string, error) {

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
	confirmPassword := formValues.Get("confirmPassword")

	if username == "" || password == "" || confirmPassword == ""  {
		
		return "", "",  errors.New("not all fields filled out, please try again")
	}
	if password != confirmPassword {
		return "", "", errors.New("password and confirm password don't match, please try again")
	}
	return username, password, nil
}