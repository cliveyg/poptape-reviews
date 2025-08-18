package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
)
/*
func (a *App) initializeRoutes() {

	// endpoints
	a.oRouter.HandleFunc("/reviews/status", a.getStatus).Methods("GET")
	//a.oRouter.HandleFunc("/reviews", a.getAllMyReviews).Methods("GET")
	a.oRouter.HandleFunc("/reviews", a.createReview).Methods("POST")
	a.oRouter.HandleFunc("/reviews/{reviewId}", a.getReview).Methods("GET")
	a.oRouter.HandleFunc("/reviews/{reviewId}", a.deleteReview).Methods("DELETE")
	a.oRouter.HandleFunc("/reviews/auction/{auctionId}",
		a.getAllReviewsByAuction).Methods("GET")
	a.oRouter.HandleFunc("/reviews/item/{itemId}",
		a.getReviewByItem).Methods("GET")
	a.oRouter.HandleFunc("/reviews/of/user/{publicId}",
		a.getAllReviewsAboutUser).Methods("GET")
	a.oRouter.HandleFunc("/reviews/by/user/{publicId}",
		a.getAllReviewsByUser).Methods("GET")
	a.oRouter.HandleFunc("/reviews/user/{publicId}",
		a.getMetadataOfUser).Methods("GET")

}

 */

func (a *App) initialiseRoutes() {

	a.Log.Info().Msg("Initialising routes")

	a.Router.GET("/reviews/status", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "System running...", "version": os.Getenv("VERSION")})
	})

	a.Router.GET("/reviews/:rId", func(c *gin.Context) {
		a.getReview(c)
	})

	a.Router.GET("/reviews/item/:iId", func(c *gin.Context) {
		a.getReviewsByItem(c)
	})

	a.Router.POST("/reviews", func(c *gin.Context) {
		a.createReview(c)
	})

}