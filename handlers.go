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

	a.Log.Info().Msg("In createReview")

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
		c.JSON(http.StatusBadRequest, gin.H{"message": "Bad request"})
		return
	}

	if rv.ReviewedBy.String() != publicId {
		a.Log.Info().Msg("Supplied reviewedBy id does not match publicId")
		c.JSON(http.StatusBadRequest, gin.H{"message": "Bad request"})
		return
	}

	// check auction id and item id's here
	var item Item
	var auction Auction
	var itemAll map[string]any
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
			Result:  &auction,
		},
	}

	results := a.fetchAndUnmarshalRequests(requests)

	_, err = json.Marshal(results)
	if err != nil {
		a.Log.Info().Msgf("Error marshalling to json [%s]", err.Error())
	}

	a.Log.Info().Msgf("Item is [%s]", item)
	// now we have the item and auction deets we can check them
	// TODO: business logic goes ere

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

	//TODO: add more possible values to order results
	if sort != "asc" && sort != "desc" {
		a.Log.Info().Msgf("Not a valid sort value: [%s]", sort)
		c.JSON(http.StatusBadRequest, gin.H{"message": "Bad request"})
		return
	}
	oss := orderby + " " + sort

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

	if totalPages < page {
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
		a.Log.Info().Msgf("Error deleting review [%s]", err.Error())
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
	c.JSON(http.StatusOK, gin.H{"user_review_metadata": "blah"})
}

/*

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
	totalReviewedBy, err := getTotalReviews(a.ODB, "reviewed_by", publicId)
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
	totalReviewsOf, err := getTotalReviews(a.ODB, "reviewed_by", publicId)
	if err != nil {
		log.Print(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		mess := `{ "message": "Oopsy somthing went wrong" }`
		if _, err := io.WriteString(w, mess); err != nil {
			log.Fatal(err)
		}
		return
	}

	calculatedScore, err := getScore(publicId)
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
	multiline := "{ \"total_reviews_of\": " + strconv.Itoa(totalReviewsOf) + " ,\n" +
		" \"total_reviews_by\": " + strconv.Itoa(totalReviewedBy) + " ,\n" +
		" \"calculated_score\": " + strconv.Itoa(calculatedScore) + " }"
	if _, err := io.WriteString(w, multiline); err != nil {
		log.Fatal(err)
	}
	return
}

 */
