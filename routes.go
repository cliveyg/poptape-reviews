package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
)
/*
func (a *App) initializeRoutes() {

	// endpoints
	a.ORouter.HandleFunc("/reviews/status", a.getStatus).Methods("GET")
	//a.ORouter.HandleFunc("/reviews", a.getAllMyReviews).Methods("GET")
	a.ORouter.HandleFunc("/reviews", a.createReview).Methods("POST")
	a.ORouter.HandleFunc("/reviews/{reviewId}", a.getReview).Methods("GET")
	a.ORouter.HandleFunc("/reviews/{reviewId}", a.deleteReview).Methods("DELETE")
	a.ORouter.HandleFunc("/reviews/auction/{auctionId}",
		a.getAllReviewsByAuction).Methods("GET")
	a.ORouter.HandleFunc("/reviews/item/{itemId}",
		a.getReviewByItem).Methods("GET")
	a.ORouter.HandleFunc("/reviews/of/user/{publicId}",
		a.getAllReviewsAboutUser).Methods("GET")
	a.ORouter.HandleFunc("/reviews/by/user/{publicId}",
		a.getAllReviewsByUser).Methods("GET")
	a.ORouter.HandleFunc("/reviews/user/{publicId}",
		a.getMetadataOfUser).Methods("GET")

}

 */

func (a *App) initialiseRoutes() {

	a.Log.Info().Msg("Initialising routes")

	a.Router.GET("/reviews/status", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "System running...", "version": os.Getenv("VERSION")})
	})

	a.Router.GET("/reviews", func(c *gin.Context) {
		a.getAllMyReviews(c)
	})

	a.Router.GET("/reviews/:id", func(c *gin.Context) {
		a.getReview(c)
	})

	a.Router.DELETE("/reviews/:id", func(c *gin.Context) {
		a.deleteReview(c)
	})

	a.Router.GET("/reviews/item/:id", func(c *gin.Context) {
		a.getReviewsByItem(c)
	})

	a.Router.GET("/reviews/of/user/:id", func(c *gin.Context) {
		a.getAllReviewsAboutUser(c)
	})

	a.Router.GET("/reviews/by/user/:id", func(c *gin.Context) {
		a.getAllReviewsByUser(c)
	})

	a.Router.POST("/reviews", func(c *gin.Context) {
		a.createReview(c)
	})

}