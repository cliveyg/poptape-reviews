package main

import (
	"database/sql"
    "encoding/json"
    "net/http"
    "log"
    "io"
    "github.com/gorilla/mux"
	"strconv"
    //"github.com/google/uuid"
)

// ----------------------------------------------------------------------------

func (a *App) getStatus(w http.ResponseWriter, r *http.Request) {

    w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	mess := `{"message": "System running..."}`
    io.WriteString(w, mess)
}

// ----------------------------------------------------------------------------

func (a *App) getAllMyReviews(w http.ResponseWriter, r *http.Request) {

    w.Header().Set("Content-Type", "application/json; charset=UTF-8")

    b, st, mess := bouncerSaysOk(r)
    if !b {
        w.WriteHeader(st)
		io.WriteString(w, mess)
		return
    }
	// successfully authenticated which means mess is the public_id
	publicId := mess

	count, _ := strconv.Atoi(r.FormValue("count"))
	start, _ := strconv.Atoi(r.FormValue("start"))

	if count > 10 || count < 1 {
		count = 10
	}
	if start < 0 {
		start = 0
	}

	reviews, err := getReviews(a.DB, publicId, start, count)
	if err != nil {
		log.Print(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, `{ "message": "Oopsy somthing went wrong" }`)
		return
	}

    jsonData, _ := json.Marshal(reviews)
    w.WriteHeader(http.StatusOK)
    w.Write(jsonData)

}

// ----------------------------------------------------------------------------

func (a *App) getMyReview(w http.ResponseWriter, r *http.Request) {

    w.Header().Set("Content-Type", "application/json; charset=UTF-8")

    b, st, mess := bouncerSaysOk(r)
    if !b {
        w.WriteHeader(st)
        io.WriteString(w, mess)
		return
    }
	//publicId = mess

    vars := mux.Vars(r)
    reviewId := vars["reviewId"]

	if !IsValidUUID(reviewId) {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, `{ "message": "Not a valid review ID" }`)
		return
	}

	rev := review{ReviewId: reviewId}
	if err := rev.getReview(a.DB); err != nil {
		switch err {
		case sql.ErrNoRows:
			w.WriteHeader(http.StatusNotFound)
		default:
			log.Print(err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			io.WriteString(w, `{ "message": "Oopsy somthing went wrong" }`)
		}
		return
	}

	jsonData, _ := json.Marshal(rev)
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)

}

// ----------------------------------------------------------------------------

func (a *App) deleteReview(w http.ResponseWriter, r *http.Request) {

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

    rev := review{ReviewId: reviewId}
    if err := rev.deleteReview(a.DB); err != nil {
		log.Print(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, `{ "message": "Oopsy somthing went wrong" }`)
        return
    }

    w.WriteHeader(http.StatusGone)

}

// ----------------------------------------------------------------------------
