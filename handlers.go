package main

import (
    //"encoding/json"
    "net/http"
    "log"
    "io"
    "github.com/gorilla/mux"
    //"github.com/google/uuid"
)

// ----------------------------------------------------------------------------

func getStatus(w http.ResponseWriter, r *http.Request) {

    w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	mess := `{"message": "System running..."}`
    io.WriteString(w, mess)
}

// ----------------------------------------------------------------------------

func getAllMyReviews(w http.ResponseWriter, r *http.Request) {

    w.Header().Set("Content-Type", "application/json; charset=UTF-8")

    b, st, mess := bouncerSaysOk(r)
    if !b {
        w.WriteHeader(st)
		io.WriteString(w, mess)
		return
    }

    log.Print("Meep")

}

// ----------------------------------------------------------------------------

func getMyReview(w http.ResponseWriter, r *http.Request) {

    w.Header().Set("Content-Type", "application/json; charset=UTF-8")

    b, st, mess := bouncerSaysOk(r)
    if !b {
        w.WriteHeader(st)
        io.WriteString(w, mess)
		return
    }

    vars := mux.Vars(r)
    reviewId := vars["reviewId"]

	if !IsValidUUID(reviewId) {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, `{ "message": "Not a valid review ID" }`)
		return
	}

    log.Print(reviewId)

}

// ----------------------------------------------------------------------------


