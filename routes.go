package main

import (
    //"net/http"
    "github.com/gorilla/mux"
)

func NewRouter() *mux.Router {

    router := mux.NewRouter()

    // endpoints
    router.HandleFunc("/reviews/status", getStatus).Methods("GET")
    router.HandleFunc("/reviews", getAllMyReviews).Methods("GET")
	router.HandleFunc("/reviews/{reviewId}", getMyReview).Methods("GET")

    return router
}
