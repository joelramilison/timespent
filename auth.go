package main

import (
	"database/sql"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
	"crypto/rand"
	"github.com/google/uuid"
	"github.com/joelramilison/timespent/internal/database"
	"golang.org/x/crypto/bcrypt"
	"errors"
)

const (
	sessionDurationString = "168h"
)

type authedHandler func(w http.ResponseWriter, req *http.Request, user database.User) 

func (cfg *apiConfig) middlewareAuth(handler authedHandler) http.HandlerFunc {

	return func(w http.ResponseWriter, req *http.Request) {

		
		toLogin := func() {
			
			// we want to redirect to the /login page.
			// the methodology depends on whether it's an HTMX request
			// or a normal one (say, typing it into the browser)
			hxReqHeader := req.Header.Get("HX-Request")

			if hxReqHeader == "" {
				http.Redirect(w, req, "/login", http.StatusFound)
			} else {
				w.Header().Add("HX-Redirect", "/login")
				w.WriteHeader(302)
				w.Write([]byte{})
			}
			
		}
		
		sessionID, userID, err := extractFromCookie(req)
		if err != nil {
			toLogin()
			return
		}

		user, err := cfg.DB.GetUser(req.Context(), userID)
		if err != nil {
			log.Printf("couldn't find user with ID %v in database, error: %v", userID.String(), err)
					
		}
		if user.SessionExpiresAt.Time.Before(time.Now()) {
			// there exists no active session, so abort even the for loop
			toLogin()
			return
		}
		err = bcrypt.CompareHashAndPassword(user.SessionIDHash, []byte(sessionID))
		if err != nil {
			// sessionID doesn't match
			toLogin()
			return
		}
		// at this point, the sessionID matches
		handler(w, req, user)
				
			}
		}


// Creates sessionID and sets the cookie
func createSession(w http.ResponseWriter, userID uuid.UUID) ([]byte, sql.NullTime, error) {

	// create sessionID
	sessionID := make([]byte, 32)
	_, err := rand.Read(sessionID)
	if err != nil {
		log.Printf("Failed to create a random session ID: %v", err)
		return []byte{}, sql.NullTime{}, err
	}

	// hash sessionID
	hashedSessionID, err := bcrypt.GenerateFromPassword(sessionID, bcrypt.DefaultCost)
	if err != nil {
		log.Printf("couldn't hash session ID, error: %v", err)
		return []byte{}, sql.NullTime{}, err
	}

	expireDuration, err := time.ParseDuration(sessionDurationString)
	if err != nil {
		log.Printf("couldn't parse session time duration from string, error: %v", err)
		return []byte{}, sql.NullTime{}, err
	}
	sessionExpiresAt := time.Now().Add(expireDuration)

	// escape sessionID to make it compatible with cookies
	escapedSessionID := url.QueryEscape(string(sessionID))
	cookieString := userID.String() + ":" + escapedSessionID

	sessionCookie := http.Cookie{
		Name: "session_id", Value: cookieString, Expires: sessionExpiresAt, Secure: true,
		HttpOnly: true,
	}
	
	http.SetCookie(w, &sessionCookie)

	return hashedSessionID, sql.NullTime{Time: sessionExpiresAt, Valid: true}, nil
}


func extractFromCookie(req *http.Request) (string, uuid.UUID, error) {
	rawCookies := req.Header.Get("Cookie")
		cookies, err := http.ParseCookie(rawCookies)
		if err != nil {
			return "", uuid.UUID{}, errors.New("couldn't parse http cookies")
		}
		var cookieString string
		for _, cookie := range cookies {
			if cookie.Name == "session_id" {
				cookieString = cookie.Value
				break
			}
		}
		if cookieString == "" {
			return "", uuid.UUID{}, errors.New("no session cookie found")
		}
		
		userIDAndSession := strings.SplitN(cookieString, ":", 2)
		if len(userIDAndSession) != 2 {
			log.Printf("Found session ID that couldn't be separated using separator ':': %v", cookieString)
			return "", uuid.UUID{}, errors.New("cookie string couldn't be parsed")
				
		}
		userID, err := uuid.Parse(userIDAndSession[0])
		if err != nil {
			log.Printf("couldn't parse UUID %v while extracting session cookie", userIDAndSession[0])
			return "", uuid.UUID{}, errors.New("UUID couldn't be parsed")
				
		}
		sessionID, err := url.QueryUnescape(userIDAndSession[1])
		if err != nil {
			log.Printf("couldn't unescape sessionID %v while extracting session cookie", userIDAndSession[1])
			return "", uuid.UUID{}, errors.New("couldn't process sessionID")
					
		}
		return sessionID, userID, nil

}