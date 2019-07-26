package main

import (
)

func (a *App) initializeRoutes() {

    // endpoints
    a.Router.HandleFunc("/reviews/status", a.getStatus).Methods("GET")
    a.Router.HandleFunc("/reviews", a.getAllMyReviews).Methods("GET")
	a.Router.HandleFunc("/reviews", a.createReview).Methods("POST")
    a.Router.HandleFunc("/reviews/{reviewId}", a.getReview).Methods("GET")
	a.Router.HandleFunc("/reviews/{reviewId}", a.deleteReview).Methods("DELETE")
	a.Router.HandleFunc("/reviews/auction/{auctionId}",
						a.getAllReviewsByAuction).Methods("GET")
    a.Router.HandleFunc("/reviews/user/{publicId}",
                        a.getAllReviewsByUser).Methods("GET")

}
