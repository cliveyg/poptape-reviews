package main

import (
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"net/http"
	"os"
)

// ----------------------------------------------------------------------------

type HttpBinGet struct {
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers"`
}
type HttpBinStatus struct {
	Message string `json:"message"`
}

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
	//c.JSON(http.StatusCreated, gin.H{"item": itemAll})
}

// ----------------------------------------------------------------------------

func (a *App) getReview(c *gin.Context) {

	b, st, mess := checkRequest(c)
	if !b {
		c.JSON(st, gin.H{"message": mess})
		return
	}

	rId, err := uuid.Parse(c.Param("rId"))
	if err != nil {
		a.Log.Info().Msgf("Not a uuid string: [%s]", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"message": "Bad request"})
		return
	}

	r := Review{ReviewId: rId}
	res := a.DB.Find(&r)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			a.Log.Info().Msgf("Review [%s] not found", r.ReviewId.String())
			c.JSON(http.StatusNotFound, gin.H{"message": "Review not found"})
			return
		}
		a.Log.Info().Msgf("Error finding review [%s]", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Something went pop"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"review": r})
}

// ----------------------------------------------------------------------------

func (a *App) getReviewsByItem(c *gin.Context) {

	itemId, err := uuid.Parse(c.Param("iId"))
	if err != nil {
		a.Log.Info().Msgf("Not a uuid string: [%s]", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"message": "Bad request"})
		return
	}

	rows, err := a.DB.Model(&Review{}).Where("item_id = ?", itemId).Rows()
	defer rows.Close()
	var reviews []Review

	for rows.Next() {
		var rv Review
		a.DB.ScanRows(rows, &rv)
		reviews = append(reviews, rv)
	}
	c.JSON(http.StatusOK, gin.H{"total_reviews": len(reviews), "reviews": reviews})
	/*
	r := Review{ItemId: itemId}
	res := a.DB.Find(&r)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			a.Log.Info().Msgf("Review [%s] not found", r.ReviewId.String())
			c.JSON(http.StatusNotFound, gin.H{"message": "Review not found"})
			return
		}
		a.Log.Info().Msgf("Error finding review [%s]", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Something went pop"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"review": r})

	 */
}


/*
func (a *App) getAllMyReviews(w http.ResponseWriter, r *http.Request) {
//func (a *App) getAllMyReviews(c *gin.Context){

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
		log.Println("Error parsing count: ", err)
	}
	var start int
	start, err = strconv.Atoi(r.FormValue("start"))
	if err != nil {
		// if error parsing eg; when running unit tests then default to 0
		start = 0
		log.Println("Error parsing start: ", err)
	}

	if count > 10 || count < 1 {
		count = 10
	}
	if start < 0 {
		start = 0
	}

	reviews, err := getReviewsByInput(a.oDB, "reviewed_by", publicId, start, count)
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


*/
/*
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
	res, err := rev.deleteReview(a.oDB)
	if err != nil {
		log.Print(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		mess := `{ "message": "Oopsy somthing went wrong" }`
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

*/
/*
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

	reviews, err := getReviewsByInput(a.oDB, "auction_id", auctionId, start, count)
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
	if err := rev.getReviewByItem(a.oDB); err != nil {
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
		total, err := getTotalReviews(a.oDB, "seller", publicId)
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

	reviews, err := getReviewsByInput(a.oDB, "seller", publicId, start, count)
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
		total, err := getTotalReviews(a.oDB, "reviewed_by", publicId)
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

	reviews, err := getReviewsByInput(a.oDB, "reviewed_by", publicId, start, count)
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
	totalReviewedBy, err := getTotalReviews(a.oDB, "reviewed_by", publicId)
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
	totalReviewsOf, err := getTotalReviews(a.oDB, "reviewed_by", publicId)
	if err != nil {
		log.Print(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		mess := `{ "message": "Oopsy somthing went wrong" }`
		if _, err := io.WriteString(w, mess); err != nil {
			log.Fatal(err)
		}
		return
	}

	calculatedScore, err := getScore(a.oDB, publicId)
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
