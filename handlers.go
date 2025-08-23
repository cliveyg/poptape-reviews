package main

import (
	"database/sql"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
	"os"
	"strconv"
)

// ----------------------------------------------------------------------------

func (a *App) createReview(c *gin.Context) {

	a.Log.Debug().Msg("In createReview")

	b, st, mess := a.bouncerSaysOk(c)
	if !b {
		c.JSON(st, gin.H{"message": mess})
		return
	}
	publicId := mess
	xhdr := c.GetHeader("X-Access-Token")
	a.Log.Debug().Msgf("Public Id is [%s]", publicId)
	var rv Review
	var err error
	if err = c.ShouldBindJSON(&rv); err != nil {
		a.Log.Info().Msgf("Input data does not match review: [%s]", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"message": "Input data is incorrect"})
		return
	}

	if rv.ReviewedBy.String() != publicId {
		a.Log.Info().Msg("Supplied reviewedBy id does not match publicId")
		c.JSON(http.StatusBadRequest, gin.H{"message": "Bad request 2"})
		return
	}

	// check auction id and item id's here
	//var item Item
	//var auction Auction
	var itemAll map[string]any
	var auctionAll map[string]any
	requests := []HTTPRequest{
		{
			URL:     os.Getenv("ITEMURL")+rv.ItemId.String(),
			Headers: map[string]string{"x-access-token": xhdr,
				                       "Content-Type": "application/json"},
			Result:  &itemAll,
		},
		{
			URL:     os.Getenv("AUCTIONURL")+rv.AuctionId.String(),
			Headers: map[string]string{"x-access-token": xhdr,
				                       "Content-Type": "application/json"},
			Result:  &auctionAll,
		},
	}

	results := a.fetchAndUnmarshalRequests(requests)

	_, err = json.Marshal(results)
	if err != nil {
		a.Log.Info().Msgf("Error marshalling to json [%s]", err.Error())
	}

	//a.Log.Info().Msgf("Item is [%s]", item)
	// now we have the item and auction deets we can check them
	// TODO: business logic goes ere - need to check winner of auction matches user

	var reviewId uuid.UUID
	reviewId, err = uuid.NewRandom()
	if err != nil {
		a.Log.Info().Msgf("Create review failed: [%s]", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"message": "Bad request"})
		return
	}

	rv.ReviewId = reviewId

	res := a.DB.Create(&rv)
	if res.Error != nil {
		a.Log.Info().Msgf("Review creation failed: [%s]", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Something went bang."})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"review_id": reviewId})
}

// ----------------------------------------------------------------------------

func (a *App) fetchReviewsByUUID(c *gin.Context, rk, uuidst string) {

	b, st, mess := checkRequest(c)
	if !b {
		c.JSON(st, gin.H{"message": mess})
		return
	}

	orderby := c.DefaultQuery("orderby", "created")
	sort := c.DefaultQuery("sort", "desc")

	if orderby != "created" {
		a.Log.Info().Msgf("Not a valid orderby value: [%s]", orderby)
		c.JSON(http.StatusBadRequest, gin.H{"message": "Bad request"})
		return
	}

	//TODO: add more possible values to orderby results
	var oss string
	if sort == "asc" || sort == "desc" {
		oss = orderby + " " + sort
	} else {
		a.Log.Info().Msgf("Not a valid sort value: [%s]", sort)
		c.JSON(http.StatusBadRequest, gin.H{"message": "Bad request"})
		return
	}

	id, err := uuid.Parse(uuidst)
	if err != nil {
		a.Log.Info().Msgf("Not a uuid string: [%s]", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"message": "Bad request"})
		return
	}

	var page int
	page, err = strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil {
		a.Log.Info().Msgf("Error in query string [%s]", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"message": "Bad request"})
		return
	}
	if page <= 0 {
		page = 1
	}

	var ospsize int
	ospsize, err = strconv.Atoi(os.Getenv("PAGESIZE"))
	if err != nil {
		a.Log.Info().Msgf("Error in query string [%s]", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"message": "Bad request"})
		return
	}
	var pagesize int
	pagesize, err = strconv.Atoi(c.DefaultQuery("pagesize", os.Getenv("PAGESIZE")))
	if err != nil {
		a.Log.Info().Msgf("Error in query string [%s]", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"message": "Bad request"})
		return
	}
	if pagesize > 100 || pagesize <= 0 {
		pagesize = ospsize
	}

	// get total records that match criteria
	var tc int64
	a.DB.Model(&Review{}).Where(rk + " = ?", id).Count(&tc)

	rows, err := a.DB.Scopes(Paginate(page, pagesize)).Model(&Review{}).Where(rk + " = ?", id).Order(oss).Rows()
	if err != nil {
		a.Log.Info().Msgf("Error fetching data: [%s]", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"message": "Bad request"})
		return
	}
	defer func(rows *sql.Rows) {
		err = rows.Close()
		if err != nil {
			a.Log.Info().Msgf("Error is: [%s]", err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Something went bang"})
			return
		}
	}(rows)

	var reviews []Review

	for rows.Next() {
		var rv Review
		err = a.DB.ScanRows(rows, &rv)
		if err != nil {
			a.Log.Info().Msgf("Error is: [%s]", err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Something went bang"})
			return
		}
		reviews = append(reviews, rv)
	}
	if tc == 0 {
		c.JSON(http.StatusNotFound, gin.H{"total_reviews": tc})
		return
	}

	// add prev/next url to output
	var urls []interface{}
	var totalPages int
	if err = CreateURLS(c, &urls, &page, &pagesize, &totalPages, &tc); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if page > totalPages {
		c.JSON(http.StatusBadRequest, gin.H{"message": "page value is incorrect"})
		return
	}

	if len(urls) > 0 {
		c.JSON(http.StatusOK, gin.H{"total_reviews": tc, "total_pages": totalPages, "current_page": page, "urls": urls, "reviews": reviews})
		return
	}
	c.JSON(http.StatusOK, gin.H{"total_reviews": tc, "total_pages": totalPages, "current_page": page, "reviews": reviews})
}

// ----------------------------------------------------------------------------

func (a *App) getReview(c *gin.Context) {
	a.fetchReviewsByUUID(c, "review_id", c.Param("id"))
}

// ----------------------------------------------------------------------------

func (a *App) getAllMyReviews(c *gin.Context) {

	b, st, mess := a.bouncerSaysOk(c)
	if !b {
		c.JSON(st, gin.H{"message": mess})
		return
	}
	a.fetchReviewsByUUID(c, "reviewed_by", mess)
}

// ----------------------------------------------------------------------------

func (a *App) deleteReview(c *gin.Context) {

	b, st, mess := a.bouncerSaysOk(c)
	if !b {
		c.JSON(st, gin.H{"message": mess})
		return
	}
	pId := mess

	rId, err := uuid.Parse(c.Param("id"))
	if err != nil {
		a.Log.Info().Msgf("Not a uuid string: [%s]", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"message": "Bad request"})
		return
	}

	res := a.DB.Where("reviewed_by = ?", pId).Delete(&Review{}, rId)
	if res.Error != nil {
		a.Log.Info().Msgf("Error deleting review [%s]", res.Error.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Something went pop"})
		return
	}
	if res.RowsAffected == 1 {
		c.JSON(http.StatusOK, gin.H{"review_deleted": rId})
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Unable to delete review"})
	}

}

// ----------------------------------------------------------------------------

func (a *App) getReviewsByItem(c *gin.Context) {
	a.fetchReviewsByUUID(c, "item_id", c.Param("id"))
}

// ----------------------------------------------------------------------------

func (a *App) getReviewsByAuction(c *gin.Context) {
	a.fetchReviewsByUUID(c, "auction_id", c.Param("id"))
}

// ----------------------------------------------------------------------------

func (a *App) getAllReviewsAboutUser(c *gin.Context) {
	a.fetchReviewsByUUID(c, "seller", c.Param("id"))
}

// ----------------------------------------------------------------------------

func (a *App) getAllReviewsByUser(c *gin.Context) {
	a.fetchReviewsByUUID(c, "reviewed_by", c.Param("id"))
}

// ----------------------------------------------------------------------------

func (a *App) getMetadataOfUser(c *gin.Context) {

	b, st, mess := checkRequest(c)
	if !b {
		c.JSON(st, gin.H{"message": mess})
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		a.Log.Info().Msgf("Not a uuid string: [%s]", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"message": "Bad request"})
		return
	}

	// check user exists
	sc := 999
	err, sc = a.checkUserExists(c)
	if err != nil {
		a.Log.Info().Msg(err.Error())
		if sc == 404 {
			c.JSON(http.StatusNotFound, gin.H{"message": "User doesn't exist"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Something went wrong"})
		return
	}

	// get total records that match criteria
	var totalReviewsOf int64
	a.DB.Model(&Review{}).Where("seller = ?", id).Count(&totalReviewsOf)
	var totalReviewsBy int64
	a.DB.Model(&Review{}).Where("reviewed_by = ?", id).Count(&totalReviewsBy)
	calculatedScore, err := getScore(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Something went splat"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"public_id": id.String(), "score": calculatedScore, "total_reviews_of_user": totalReviewsOf, "total_reviews_by_user": totalReviewsBy})
}
