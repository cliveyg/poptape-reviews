package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"io"
	"log"
	"net/http"
	"strconv"
)

// ----------------------------------------------------------------------------

func (a *App) getStatus(w http.ResponseWriter, _ *http.Request) {

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	mess := `{"message": "System running..."}`
	if _, err := io.WriteString(w, mess); err != nil {
		log.Fatal(err)
	}

}

// ----------------------------------------------------------------------------

func (a *App) getAllMyReviews(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	b, st, mess := bouncerSaysOk(r)
	if !b {
		w.WriteHeader(st)
		if _, err := io.WriteString(w, mess); err != nil {
			log.Fatal(err)
		}
		return
	}
	// successfully authenticated which means mess is the public_id
	publicId := mess

	count, err := strconv.Atoi(r.FormValue("count"))
	if err != nil {
		// if error parsing eg; when running unit tests then default to 10
		count = 10
		log.Println("Error parsing count: ",err)
	}
	var start int
	start, err = strconv.Atoi(r.FormValue("start"))
	if err != nil {
		// if error parsing eg; when running unit tests then default to 0
		start = 0
		log.Println("Error parsing start: ",err)
	}

	if count > 10 || count < 1 {
		count = 10
	}
	if start < 0 {
		start = 0
	}

	reviews, err := getReviewsByInput(a.DB, "reviewed_by", publicId, start, count)
	if err != nil {
		log.Print(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		if _, err := io.WriteString(w, `{ "message": "Oopsy somthing went wrong" }`); err != nil {
			log.Fatal(err)
		}
		return
	}

	jsonData, _ := json.Marshal(reviews)
	if len(reviews) == 0 {
		w.WriteHeader(http.StatusNotFound)
	} else {
		w.WriteHeader(http.StatusOK)
	}
	if _, err := w.Write(jsonData); err != nil {
		log.Fatal(err)
	}

}

// ----------------------------------------------------------------------------

func (a *App) getReview(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	b, st, mess := CheckRequest(r)
	if !b {
		w.WriteHeader(st)
		if _, err := io.WriteString(w, mess); err != nil {
			log.Fatal(err)
		}
		return
	}

	vars := mux.Vars(r)
	reviewId := vars["reviewId"]

	if !IsValidUUID(reviewId) {
		w.WriteHeader(http.StatusBadRequest)
		if _, err := io.WriteString(w, `{ "message": "Not a valid review ID" }`); err != nil {
			log.Fatal(err)
		}
		return
	}

	rev := Review{ReviewId: reviewId}
	if err := rev.getReview(a.DB); err != nil {
		switch err {
		case sql.ErrNoRows:
			w.WriteHeader(http.StatusNotFound)
		default:
			log.Println("***********************************")
			log.Print("Error is" + err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			if _, err := io.WriteString(w, `{ "message": "Oopsy somthing went wrong" }`); err != nil {
				log.Fatal(err)
			}
		}
		return
	}

	jsonData, _ := json.Marshal(rev)
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(jsonData); err != nil {
		log.Fatal(err)
	}

}

// ----------------------------------------------------------------------------

func (a *App) deleteReview(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	b, st, mess := bouncerSaysOk(r)
	if !b {
		w.WriteHeader(st)
		if _, err := io.WriteString(w, mess); err != nil {
			log.Fatal(err)
		}
		return
	}
	publicId := mess

	vars := mux.Vars(r)
	reviewId := vars["reviewId"]

	if !IsValidUUID(reviewId) {
		w.WriteHeader(http.StatusBadRequest)
		mess := `{ "message": "Not a valid review ID" }`
		if _, err := io.WriteString(w, mess); err != nil {
			log.Fatal(err)
		}
		return
	}

	rev := Review{ReviewId: reviewId, ReviewedBy: publicId}
	res, err := rev.deleteReview(a.DB)
	if err != nil {
		log.Print(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		mess :=`{ "message": "Oopsy somthing went wrong" }`
		if _, err := io.WriteString(w, mess); err != nil {
			log.Fatal(err)
		}
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
		if _, err := io.WriteString(w, mess); err != nil {
			log.Fatal(err)
		}
		return
	}
	publicId := mess

	var rev Review
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&rev); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		mess = fmt.Sprintf("{ \"error\": \"%s\" }", err)
		if _, err := io.WriteString(w, mess); err != nil {
			log.Fatal(err)
		}
		return
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(r.Body)

	reviewId, err := uuid.NewRandom()
	if err != nil {
		log.Fatal(err)
	}
	rev.ReviewId = reviewId.String()
	rev.ReviewedBy = publicId

	// need to run these calls in parallel
	x := r.Header.Get("X-Access-Token")
	if !ValidAuction(rev.AuctionId, publicId, x) {
		w.WriteHeader(http.StatusBadRequest)
		mess := `{ "message": "Auction does not exist" }`
		if _, err := io.WriteString(w, mess); err != nil {
			log.Fatal(err)
		}
		return
	}
	/*
	   if !ValidItem(rev.ItemId, x) {
	       w.WriteHeader(http.StatusBadRequest)
	       io.WriteString(w, `{ "message": "Item does not exist" }`)
	       return
	   }
	*/
	//TODO: Check that user actually won the auction

	if err := rev.createReview(a.DB); err != nil {
		log.Print(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		mess := `{ "message": "Oopsy somthing went wrong" }`
		if _, err := io.WriteString(w, mess); err != nil {
			log.Fatal(err)
		}
		return
	}
	w.WriteHeader(http.StatusCreated)
	mess = fmt.Sprintf("{ \"review_id\": \"%s\" }", rev.ReviewId)
	if _, err := io.WriteString(w, mess); err != nil {
		log.Fatal(err)
	}
}

// ----------------------------------------------------------------------------

func (a *App) getAllReviewsByAuction(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	b, st, mess := CheckRequest(r)
	if !b {
		w.WriteHeader(st)
		if _, err := io.WriteString(w, mess); err != nil {
			log.Fatal(err)
		}
		return
	}
	vars := mux.Vars(r)
	auctionId := vars["auctionId"]

	if !IsValidUUID(auctionId) {
		w.WriteHeader(http.StatusBadRequest)
		mess := `{ "message": "Not a valid auction ID" }`
		if _, err := io.WriteString(w, mess); err != nil {
			log.Fatal(err)
		}
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
		mess := `{ "message": "Oopsy somthing went wrong" }`
		if _, err := io.WriteString(w, mess); err != nil {
			log.Fatal(err)
		}
		return
	}

	jsonData, _ := json.Marshal(reviews)
	if len(reviews) == 0 {
		w.WriteHeader(http.StatusNotFound)
	} else {
		w.WriteHeader(http.StatusOK)
	}
	if _, err := w.Write(jsonData); err != nil {
		log.Fatal(err)
	}

}

// ----------------------------------------------------------------------------

func (a *App) getReviewByItem(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	b, st, mess := CheckRequest(r)
	if !b {
		w.WriteHeader(st)
		if _, err := io.WriteString(w, mess); err != nil {
			log.Fatal(err)
		}
		return
	}
	vars := mux.Vars(r)
	itemId := vars["itemId"]

	if !IsValidUUID(itemId) {
		w.WriteHeader(http.StatusBadRequest)
		mess := `{ "message": "Not a valid item ID" }`
		if _, err := io.WriteString(w, mess); err != nil {
			log.Fatal(err)
		}
		return
	}

	rev := Review{ItemId: itemId}
	if err := rev.getReviewByItem(a.DB); err != nil {
		switch err {
		case sql.ErrNoRows:
			w.WriteHeader(http.StatusNotFound)
		default:
			log.Print(err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			mess := `{ "message": "Oopsy somthing went wrong" }`
			if _, err := io.WriteString(w, mess); err != nil {
				log.Fatal(err)
			}
		}
		return
	}

	jsonData, _ := json.Marshal(rev)
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(jsonData); err != nil {
		log.Fatal(err)
	}

}

// ----------------------------------------------------------------------------

func (a *App) getAllReviewsAboutUser(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	b, st, mess := CheckRequest(r)
	if !b {
		w.WriteHeader(st)
		if _, err := io.WriteString(w, mess); err != nil {
			log.Fatal(err)
		}
		return
	}
	vars := mux.Vars(r)
	publicId := vars["publicId"]

	if !IsValidUUID(publicId) {
		w.WriteHeader(http.StatusBadRequest)
		mess := `{ "message": "Not a valid public ID" }`
		if _, err := io.WriteString(w, mess); err != nil {
			log.Fatal(err)
		}
		return
	}

	if r.FormValue("totalonly") != "" {
		// just return the total count
		total, err := getTotalReviews(a.DB, "seller", publicId)
		if err != nil {
			log.Print(err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			mess := `{ "message": "Oopsy somthing went wrong" }`
			if _, err := io.WriteString(w, mess); err != nil {
				log.Fatal(err)
			}
			return
		}
		w.WriteHeader(http.StatusOK)
		mess := fmt.Sprintf("{ \"total_reviews\": \"%d\" }", total)
		if _, err := io.WriteString(w, mess); err != nil {
			log.Fatal(err)
		}
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

	reviews, err := getReviewsByInput(a.DB, "seller", publicId, start, count)
	if err != nil {
		log.Print(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		mess := `{ "message": "Oopsy somthing went wrong" }`
		if _, err := io.WriteString(w, mess); err != nil {
			log.Fatal(err)
		}
		return
	}

	jsonData, _ := json.Marshal(reviews)
	if len(reviews) == 0 {
		w.WriteHeader(http.StatusNotFound)
	} else {
		w.WriteHeader(http.StatusOK)
	}
	if _, err := w.Write(jsonData); err != nil {
		log.Fatal(err)
	}

}

// ----------------------------------------------------------------------------

func (a *App) getAllReviewsByUser(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	b, st, mess := CheckRequest(r)
	if !b {
		w.WriteHeader(st)
		if _, err := io.WriteString(w, mess); err != nil {
			log.Fatal(err)
		}
		return
	}
	vars := mux.Vars(r)
	publicId := vars["publicId"]

	if !IsValidUUID(publicId) {
		w.WriteHeader(http.StatusBadRequest)
		mess := `{ "message": "Not a valid public ID" }`
		if _, err := io.WriteString(w, mess); err != nil {
			log.Fatal(err)
		}
		return
	}

	if r.FormValue("totalonly") != "" {
		// just return the total count
		total, err := getTotalReviews(a.DB, "reviewed_by", publicId)
		if err != nil {
			log.Print(err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			mess :=`{ "message": "Oopsy somthing went wrong" }`
			if _, err := io.WriteString(w, mess); err != nil {
				log.Fatal(err)
			}
			return
		}
		w.WriteHeader(http.StatusOK)
		mess := fmt.Sprintf("{ \"total_reviews\": \"%d\" }", total)
		if _, err := io.WriteString(w, mess); err != nil {
			log.Fatal(err)
		}
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

	reviews, err := getReviewsByInput(a.DB, "reviewed_by", publicId, start, count)
	if err != nil {
		log.Print(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		mess := `{ "message": "Oopsy somthing went wrong" }`
		if _, err := io.WriteString(w, mess); err != nil {
			log.Fatal(err)
		}
		return
	}

	jsonData, _ := json.Marshal(reviews)
	if len(reviews) == 0 {
		w.WriteHeader(http.StatusNotFound)
	} else {
		w.WriteHeader(http.StatusOK)
	}
	if _, err := w.Write(jsonData); err != nil {
		log.Fatal(err)
	}

}

// ----------------------------------------------------------------------------

func (a *App) getMetadataOfUser(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	b, st, mess := CheckRequest(r)
	if !b {
		w.WriteHeader(st)
		if _, err := io.WriteString(w, mess); err != nil {
			log.Fatal(err)
		}
		return
	}
	vars := mux.Vars(r)
	publicId := vars["publicId"]

	if !IsValidUUID(publicId) {
		w.WriteHeader(http.StatusBadRequest)
		mess := `{ "message": "Not a valid public ID" }`
		if _, err := io.WriteString(w, mess); err != nil {
			log.Fatal(err)
		}
		return
	}

	// get the total count of reviews by
	totalReviewedBy, err := getTotalReviews(a.DB, "reviewed_by", publicId)
	if err != nil {
		log.Print(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		mess := `{ "message": "Oopsy somthing went wrong" }`
		if _, err := io.WriteString(w, mess); err != nil {
			log.Fatal(err)
		}
		return
	}

	// get the total count of reviews of
	totalReviewsOf, err := getTotalReviews(a.DB, "reviewed_by", publicId)
	if err != nil {
		log.Print(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		mess := `{ "message": "Oopsy somthing went wrong" }`
		if _, err := io.WriteString(w, mess); err != nil {
			log.Fatal(err)
		}
		return
	}

	calculatedScore, err := getScore(a.DB, publicId)
	if err != nil {
		log.Print(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		mess := `{ "message": "Oopsy somthing went wrong" }`
		if _, err := io.WriteString(w, mess); err != nil {
			log.Fatal(err)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	//io.WriteString(w, `{ "reviews_by": `+totalReviewedBy+`,
	//                 }`)
	multiline := "{ \"total_reviews_of\": " + strconv.Itoa(totalReviewsOf) + " ,\n" +
		" \"total_reviews_by\": " + strconv.Itoa(totalReviewedBy) + " ,\n" +
		" \"calculated_score\": " + strconv.Itoa(calculatedScore) + " }"
	if _, err := io.WriteString(w, multiline); err != nil {
		log.Fatal(err)
	}
	return
}
