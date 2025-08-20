package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
)

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

	a.Router.GET("/reviews/auction/:id", func(c *gin.Context) {
		a.getReviewsByAuction(c)
	})

	a.Router.GET("/reviews/of/user/:id", func(c *gin.Context) {
		a.getAllReviewsAboutUser(c)
	})

	a.Router.GET("/reviews/by/user/:id", func(c *gin.Context) {
		a.getAllReviewsByUser(c)
	})

	a.Router.GET("/reviews/user/:id", func(c *gin.Context) {
		a.getMetadataOfUser(c)
	})

	a.Router.POST("/reviews", func(c *gin.Context) {
		a.createReview(c)
	})

}