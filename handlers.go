package main

import (
	"database/sql"
    "encoding/json"
    "net/http"
    "log"
    "io"
    "github.com/gorilla/mux"
	"strconv"
    "github.com/google/uuid"
	"fmt"
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

	reviews, err := getReviewsByInput(a.DB, "public_id", publicId, start, count)
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

func (a *App) getReview(w http.ResponseWriter, r *http.Request) {

    w.Header().Set("Content-Type", "application/json; charset=UTF-8")

    b, st, mess := CheckRequest(r)
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
	publicId := mess

    vars := mux.Vars(r)
    reviewId := vars["reviewId"]

    if !IsValidUUID(reviewId) {
        w.WriteHeader(http.StatusBadRequest)
        io.WriteString(w, `{ "message": "Not a valid review ID" }`)
        return
    }

    rev := review{ReviewId: reviewId, PublicId: publicId}
	res, err := rev.deleteReview(a.DB)
    if err != nil {
		log.Print(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, `{ "message": "Oopsy somthing went wrong" }`)
        return
    }

	if res == 1 {
		w.WriteHeader(http.StatusGone)
	} else {
		w.WriteHeader(http.StatusNotAcceptable)
	}

}

// ----------------------------------------------------------------------------

func (a *App) createReview(w http.ResponseWriter, r *http.Request) {

    w.Header().Set("Content-Type", "application/json; charset=UTF-8")

    b, st, mess := bouncerSaysOk(r)
    if !b {
        w.WriteHeader(st)
        io.WriteString(w, mess)
        return
    }
	publicId := mess

	var rev review
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&rev); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		mess = fmt.Sprintf("{ \"error\": \"%s\" }",err)
		io.WriteString(w, mess)
		return
	}
	defer r.Body.Close()

	reviewId, err := uuid.NewRandom()
    if err !=nil {
		log.Fatal(err)
    }
	rev.ReviewId = reviewId.String()
	rev.PublicId = publicId

	x := r.Header.Get("X-Access-Token")
	if !ValidAuction(rev.AuctionId, x) {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, `{ "message": "Auction does not exist" }`)
		return
	}

	//TODO: Check that user actually won the auction

	if err := rev.createReview(a.DB); err != nil {
        log.Print(err.Error())
        w.WriteHeader(http.StatusInternalServerError)
        io.WriteString(w, `{ "message": "Oopsy somthing went wrong" }`)
        return
	}
	w.WriteHeader(http.StatusCreated)
	mess = fmt.Sprintf("{ \"review_id\": \"%s\" }",rev.ReviewId)
	io.WriteString(w, mess)
}

// ----------------------------------------------------------------------------

func (a *App) getAllReviewsByAuction(w http.ResponseWriter, r *http.Request) {

    w.Header().Set("Content-Type", "application/json; charset=UTF-8")

    b, st, mess := CheckRequest(r)
    if !b {
        w.WriteHeader(st)
        io.WriteString(w, mess)
        return
    }
    vars := mux.Vars(r)
    auctionId := vars["auctionId"]

    if !IsValidUUID(auctionId) {
        w.WriteHeader(http.StatusBadRequest)
        io.WriteString(w, `{ "message": "Not a valid auction ID" }`)
        return
    }

    count, _ := strconv.Atoi(r.FormValue("count"))
    start, _ := strconv.Atoi(r.FormValue("start"))

    if count > 10 || count < 1 {
        count = 10
    }
    if start < 0 {
        start = 0
    }

    reviews, err := getReviewsByInput(a.DB, "auction_id", auctionId, start, count)
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

func (a *App) getAllReviewsByUser(w http.ResponseWriter, r *http.Request) {

    w.Header().Set("Content-Type", "application/json; charset=UTF-8")

    b, st, mess := CheckRequest(r)
    if !b {
        w.WriteHeader(st)
        io.WriteString(w, mess)
        return
    }
    vars := mux.Vars(r)
    publicId := vars["publicId"]

    if !IsValidUUID(publicId) {
        w.WriteHeader(http.StatusBadRequest)
        io.WriteString(w, `{ "message": "Not a valid public ID" }`)
        return
    }

    count, _ := strconv.Atoi(r.FormValue("count"))
    start, _ := strconv.Atoi(r.FormValue("start"))

    if count > 10 || count < 1 {
        count = 10
    }
    if start < 0 {
        start = 0
    }

    reviews, err := getReviewsByInput(a.DB, "public_id", publicId, start, count)
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
