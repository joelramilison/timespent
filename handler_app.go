package main

import (
	"net/http"

	"github.com/joelramilison/timespent/internal/database"
)

func appHandler(w http.ResponseWriter, req *http.Request, user database.User) {

	app().Render(req.Context(), w)
}