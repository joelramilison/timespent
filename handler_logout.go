package main

import (
	"net/http"
	"log"
	"github.com/joelramilison/timespent/internal/database"
)

func (cfg *apiConfig) logoutHandler(w http.ResponseWriter, req *http.Request, user database.User) {
	
	err := cfg.DB.LogUserOut(req.Context(), user.ID)
	if err != nil {
		log.Printf("couldn't log user out: %v", err)
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte{})
}