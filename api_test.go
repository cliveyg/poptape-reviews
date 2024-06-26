// main_test.go

package main_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/cliveyg/poptape-reviews"
	"github.com/jarcoal/httpmock"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

var a main.App

func TestMain(m *testing.M) {

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	} else {
		log.Println("Loaded .env file")
	}

	a = main.App{}
	a.Initialize(
		os.Getenv("TESTDB_HOST"),
		os.Getenv("TESTDB_USERNAME"),
		os.Getenv("TESTDB_PASSWORD"),
		os.Getenv("TESTDB_NAME"))

	//ensureTableExists()
	runSQL(dropTable)
	runSQL(tableCreationQuery)

	code := m.Run()

	//clearTable()

	os.Exit(code)
}

func executeRequest(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	a.Router.ServeHTTP(rr, req)

	return rr
}

func checkResponseCode(t *testing.T, expected, actual int) bool {
	if expected != actual {
		t.Errorf("Expected response code %d. Got %d\n", expected, actual)
		return false
	} else {
		return true
	}
}

func clearTable() {
	if _, err := a.DB.Exec("DELETE FROM reviews"); err != nil {
		log.Fatal(err)
	}
}

func runSQL(sqltext string) {
	if _, err := a.DB.Exec(sqltext); err != nil {
		log.Fatal(err)
	}
}

func getRecCount() int {

	rows, err := a.DB.Query("SELECT COUNT(*) AS count FROM reviews")
	if err != nil {
		log.Fatal(err)
	}
	return checkCount(rows)
}

func checkCount(rows *sql.Rows) (count int) {
	for rows.Next() {
		err := rows.Scan(&count)
		if err != nil {
			log.Fatal(err)
		}
	}
	return count
}

const dropTable = `DROP TABLE IF EXISTS reviews`

const tableCreationQuery = `CREATE TABLE IF NOT EXISTS reviews
(
    review_id CHAR(36) UNIQUE NOT NULL,
    reviewed_by CHAR(36) NOT NULL,
    auction_id CHAR(36) NOT NULL,
    item_id CHAR(36) NOT NULL,
    seller CHAR(36) NOT NULL,
    review VARCHAR(2000),
    overall INT NOT NULL DEFAULT 0,
    pap_cost INT NOT NULL DEFAULT 0,
    communication INT NOT NULL DEFAULT 0,
    as_described INT NOT NULL DEFAULT 0,
    created TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
    UNIQUE (reviewed_by, item_id),
    CONSTRAINT reviews_pkey PRIMARY KEY (review_id)
)`

const insertDummyReviews = `INSERT INTO reviews 
(review_id, reviewed_by, auction_id, item_id, 
seller, review, overall, pap_cost, 
communication, as_described)
VALUES
('e8f48256-2460-418f-81b7-86dad2aa6e41',
'f38ba39a-3682-4803-a498-659f0bf05000',
'e77be9e0-bb00-49bc-9e7d-d7cc7072ab8c',
'387bfbb4-36cf-44c9-8e05-83b2ca72cdff',
'46d7d11c-fa06-4e54-8208-95433b98cfc9',
'amaze balls product',5,4,4,3),
('e8f48256-2460-418f-81b7-86dad2aa6aaa',
'f38ba39a-3682-4803-a498-659f0bf05304',
'e77be9e0-bb00-49bc-9e7d-d7cc7072ab8c',
'7d1aa876-9be8-441f-ad86-d86e5faddd81',
'46d7d11c-fa06-4e54-8208-95433b98cfc9',
'amaze balls product',5,4,4,3),
('e8f48256-2460-418f-81b7-86dad2aa6111',
'f38ba39a-3682-4803-a498-659f0bf05304',
'e77be9e0-bb00-49bc-9e7d-d7cc7072ab11',
'7d1aa876-9be8-441f-ad86-d8e5fade5441',
'46d7d11c-fa06-4e54-8208-954322222222',
'amaze balls product',4,4,4,3),
('e8f48256-2460-418f-81b7-86dad2aa6222',
'f38ba39a-3682-4803-a498-659f0bf05304',
'e77be9e0-bb00-49bc-9e7d-d7cc7072ab22',
'aabbccd6-9be8-441f-ad86-d86e5faddd81',
'46d7d11c-fa06-4e54-8208-95433b98cfc9',
'amaze balls product',4,4,4,3),
('e8f48256-2460-418f-81b7-86dad2aa6333',
'f38ba39a-3682-4803-a498-659f0bf05000',
'e77be9e0-bb00-49bc-9e7d-d7cc7072ab33',
'7d1aa876-9be8-441f-ad86-daaa51872333',
'46d7d11c-fa06-4e54-8208-aaaaaaaa8888',
'amaze balls product',4,4,4,3);`

const createJson = `{"auction_id":"f38ba39a-3682-4803-a498-659f0b111111",
"item_id": "f80689a6-9fba-4859-bdde-0a307c696ea8",
"reviewed_by": "46d7d11c-fa06-4e54-8208-95433b98cfc9",
"seller": "4a48341f-bcef-4362-9d80-24a4960507ea",
"review": "amazing product",
"overall": 4,
"post_and_packaging": 3,
"communication": 4,
"as_described": 4}`

type Review struct {
	ReviewId   string `json:"review_id"`
	Review     string `json:"review"`
	ReviewedBy string `json:"reviewed_by"`
	AuctionId  string `json:"auction_id"`
	ItemId     string `json:"item_id"`
	Seller     string `json:"seller"`
	Overall    int    `json:"overall"`
	PapCost    int    `json:"post_and_packaging"`
	Comm       int    `json:"communication"`
	AsDesc     int    `json:"as_described"`
	Created    string `json:"created"`
}

type CreateResp struct {
	ReviewId string `json:"review_id"`
}

// ----------------------------------------------------------------------------
// s t a r t   o f   t e s t s
// ----------------------------------------------------------------------------

func TestAPIStatus(t *testing.T) {

	req, _ := http.NewRequest("GET", "/reviews/status", nil)
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	response := executeRequest(req)

	fmt.Println(fmt.Sprintf("Resp body is %s", response.Body.String()))

	if checkResponseCode(t, http.StatusOK, response.Code) {
		fmt.Println("[PASS].....TestAPIStatus")
	}
}

// get no reviews for authed user
func TestEmptyTable(t *testing.T) {

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder("GET", "https://poptape.club/authy/checkaccess/10",
		httpmock.NewStringResponder(200, `{"public_id": "f38ba39a-3682-4803-a498-659f0bf05304" }`))

	clearTable()

	req, _ := http.NewRequest("GET", "/reviews", nil)
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("X-Access-Token", "faketoken")
	response := executeRequest(req)

	if body := response.Body.String(); body != "[]" {
		t.Errorf("Expected an empty array. Got %s", body)
	}

	if checkResponseCode(t, http.StatusNotFound, response.Code) {
		fmt.Println("[PASS].....TestEmptyTable")
	}

}

// get reviews for authed user
func TestReturnOnlyAuthUserReviews(t *testing.T) {

	clearTable()
	runSQL(insertDummyReviews)

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder("GET", "https://poptape.club/authy/checkaccess/10",
		httpmock.NewStringResponder(200, `{"public_id": "f38ba39a-3682-4803-a498-659f0bf05304" }`))

	req, _ := http.NewRequest("GET", "/reviews", nil)
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("X-Access-Token", "faketoken")
	response := executeRequest(req)

	noError := checkResponseCode(t, http.StatusOK, response.Code)

	reviews := make([]Review, 0)
	err := json.NewDecoder(response.Body).Decode(&reviews)
	if err != nil {
		t.Fatal(err)
	}

	if len(reviews) != 3 {
		t.Errorf("no of reviews returned doesn't match should be 3 but is %d", len(reviews))
		noError = false
	}

	for _, r := range reviews {
		if r.ReviewedBy != "f38ba39a-3682-4803-a498-659f0bf05304" {
			t.Errorf("reviewed by doesn't match")
			noError = false
		}
	}

	if noError {
		fmt.Println("[PASS].....TestReturnOnlyAuthUserReviews")
	}

}

// test missing access token
func TestMissingXAccessToken(t *testing.T) {

	clearTable()
	runSQL(insertDummyReviews)

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder("GET", "https://poptape.club/authy/checkaccess/10",
		httpmock.NewStringResponder(200, `{"public_id": "f38ba39a-3682-4803-a498-659f0bf05304" }`))

	req, _ := http.NewRequest("GET", "/reviews", nil)
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	response := executeRequest(req)

	if checkResponseCode(t, http.StatusUnauthorized, response.Code) {
		fmt.Println("[PASS].....TestMissingXAccessToken")
	}
}

// get reviews by user - no auth needed
func TestGetReviewsByUser(t *testing.T) {

	clearTable()
	runSQL(insertDummyReviews)

	req, _ := http.NewRequest("GET", "/reviews/by/user/f38ba39a-3682-4803-a498-659f0bf05304", nil)
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("X-Access-Token", "faketoken")
	response := executeRequest(req)

	noError := checkResponseCode(t, http.StatusOK, response.Code)

	reviews := make([]Review, 0)
	err := json.NewDecoder(response.Body).Decode(&reviews)
	if err != nil {
		t.Fatal(err)
	}

	for _, r := range reviews {
		if r.ReviewedBy != "f38ba39a-3682-4803-a498-659f0bf05304" {
			noError = false
			t.Errorf("reviewed by doesn't match")
		}
	}

	if len(reviews) != 3 {
		noError = false
		t.Errorf("no of reviews returned doesn't match")
	}
	if noError {
		fmt.Println("[PASS].....TestGetReviewsByUser")
	}

}

// test bad uuid
func TestBadUUID(t *testing.T) {

	clearTable()
	runSQL(insertDummyReviews)

	req, _ := http.NewRequest("GET", "/reviews/f38ba39a-3682-4803-a498-659f0bf0530g", nil)
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("X-Access-Token", "faketoken")
	response := executeRequest(req)

	if checkResponseCode(t, http.StatusBadRequest, response.Code) {
		fmt.Println("[PASS].....TestBadUUID")
	}
}

// test 404
func Test404ForValidUUID(t *testing.T) {

	clearTable()
	runSQL(insertDummyReviews)

	req, _ := http.NewRequest("GET", "/reviews/f38ba39a-3682-4803-a498-659f0bf05311", nil)
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("X-Access-Token", "faketoken")
	response := executeRequest(req)

	if checkResponseCode(t, http.StatusNotFound, response.Code) {
		fmt.Println("[PASS].....Test404ForValidUUID")
	}
}

// test 404
func Test404ForRandomURL(t *testing.T) {

	clearTable()
	runSQL(insertDummyReviews)

	req, _ := http.NewRequest("GET", "/reviews/f38ba39a/someurl/999", nil)
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("X-Access-Token", "faketoken")
	response := executeRequest(req)

	if checkResponseCode(t, http.StatusNotFound, response.Code) {
		fmt.Println("[PASS].....Test404ForRandomURL")
	}
}

// get reviews by auction - no auth needed
func TestGetReviewsByAuction(t *testing.T) {

	clearTable()
	runSQL(insertDummyReviews)

	req, _ := http.NewRequest("GET", "/reviews/auction/e77be9e0-bb00-49bc-9e7d-d7cc7072ab8c", nil)
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("X-Access-Token", "faketoken")
	response := executeRequest(req)

	noError := checkResponseCode(t, http.StatusOK, response.Code)

	reviews := make([]Review, 0)
	err := json.NewDecoder(response.Body).Decode(&reviews)
	if err != nil {
		noError = false
		t.Errorf("Error decoding JSON: " + err.Error())
	}

	if len(reviews) != 2 {
		noError = false
		t.Errorf("no of reviews returned doesn't match")
	}

	if noError {
		fmt.Println("[PASS].....TestGetReviewsByAuction")
	}
}

// get review by id - no auth needed
func TestGetReviewById(t *testing.T) {

	clearTable()
	runSQL(insertDummyReviews)

	req, _ := http.NewRequest("GET", "/reviews/e8f48256-2460-418f-81b7-86dad2aa6333", nil)
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	response := executeRequest(req)

	noError := checkResponseCode(t, http.StatusOK, response.Code)
	var rev Review
	err := json.NewDecoder(response.Body).Decode(&rev)
	if err != nil {
		noError = false
		t.Errorf("Error decoding JSON: " + err.Error())
	}

	if rev.ReviewedBy != "f38ba39a-3682-4803-a498-659f0bf05000" {
		noError = false
		t.Errorf("reviewed by doesn't match")
	}
	if noError {
		fmt.Println("[PASS].....TestGetReviewById")
	}
}

// get review by item - no auth needed
func TestGetReviewByItem(t *testing.T) {

	clearTable()
	runSQL(insertDummyReviews)

	req, _ := http.NewRequest("GET", "/reviews/item/7d1aa876-9be8-441f-ad86-d86e5faddd81", nil)
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	response := executeRequest(req)

	noError := checkResponseCode(t, http.StatusOK, response.Code)
	var rev Review
	err := json.NewDecoder(response.Body).Decode(&rev)
	if err != nil {
		noError = false
		t.Errorf("Error decoding JSON: " + err.Error())
	}

	if rev.ReviewId != "e8f48256-2460-418f-81b7-86dad2aa6aaa" {
		noError = false
		t.Errorf("review id doesn't match")
	}
	if rev.ItemId != "7d1aa876-9be8-441f-ad86-d86e5faddd81" {
		noError = false
		t.Errorf("item id doesn't match")
	}
	if noError {
		fmt.Println("[PASS].....TestGetReviewByItem")
	}
}

// get reviews of the users selling  - no auth needed
func TestGetReviewsOfUser(t *testing.T) {

	clearTable()
	runSQL(insertDummyReviews)

	req, _ := http.NewRequest("GET", "/reviews/of/user/46d7d11c-fa06-4e54-8208-95433b98cfc9", nil)
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	response := executeRequest(req)

	noError := checkResponseCode(t, http.StatusOK, response.Code)
	reviews := make([]Review, 0)

	err := json.NewDecoder(response.Body).Decode(&reviews)
	if err != nil {
		noError = false
		t.Errorf("Error decoding JSON: " + err.Error())
	}
	for _, r := range reviews {
		if r.Seller != "46d7d11c-fa06-4e54-8208-95433b98cfc9" {
			noError = false
			t.Errorf("reviewed by doesn't match")
		}
	}

	if len(reviews) != 3 {
		noError = false
		t.Errorf("no of reviews returned doesn't match: expected 3 and got %d", len(reviews))
	}
	if noError {
		fmt.Println("[PASS].....TestGetReviewsOfUser")
	}
}

// get reviews written by the user  - no auth needed
func TestGetReviewsByAnotherUser(t *testing.T) {

	clearTable()
	runSQL(insertDummyReviews)

	req, _ := http.NewRequest("GET", "/reviews/by/user/f38ba39a-3682-4803-a498-659f0bf05304", nil)
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	response := executeRequest(req)

	noError := checkResponseCode(t, http.StatusOK, response.Code)
	reviews := make([]Review, 0)

	err := json.NewDecoder(response.Body).Decode(&reviews)
	if err != nil {
		noError = false
		t.Errorf("Error decoding JSON: " + err.Error())
	}
	for _, r := range reviews {
		if r.ReviewedBy != "f38ba39a-3682-4803-a498-659f0bf05304" {
			noError = false
			t.Errorf("reviewed by doesn't match")
		}
	}

	if len(reviews) != 3 {
		noError = false
		t.Errorf("no of reviews returned doesn't match")
	}
	if noError {
		fmt.Println("[PASS].....TestGetReviewsByAnotherUser")
	}
}

// get delete review for authed user
func TestDeleteReviewOk(t *testing.T) {

	clearTable()
	runSQL(insertDummyReviews)

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder("GET", "https://poptape.club/authy/checkaccess/10",
		httpmock.NewStringResponder(200, `{"public_id": "f38ba39a-3682-4803-a498-659f0bf05304" }`))

	req, _ := http.NewRequest("DELETE", "/reviews/e8f48256-2460-418f-81b7-86dad2aa6222", nil)
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("X-Access-Token", "faketoken")
	response := executeRequest(req)

	noError := checkResponseCode(t, http.StatusGone, response.Code)

	req, _ = http.NewRequest("GET", "/reviews/e8f48256-2460-418f-81b7-86dad2aa6222", nil)
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("X-Access-Token", "faketoken")
	response = executeRequest(req)

	noError = checkResponseCode(t, http.StatusNotFound, response.Code)
	if noError {
		fmt.Println("[PASS].....TestDeleteReviewOk")
	}
}

// failed delete review - cannot delete someone else's review
func TestDeleteFail(t *testing.T) {

	clearTable()
	runSQL(insertDummyReviews)

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder("GET", "https://poptape.club/authy/checkaccess/10",
		httpmock.NewStringResponder(200, `{"public_id": "f38ba39a-3682-4803-a498-659f0bf05304" }`))

	req, _ := http.NewRequest("DELETE", "/reviews/e8f48256-2460-418f-81b7-86dad2aa6333", nil)
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("X-Access-Token", "faketoken")
	response := executeRequest(req)

	if checkResponseCode(t, http.StatusNotAcceptable, response.Code) {
		fmt.Println("[PASS].....TestDeleteFail")
	}
}

// failed delete review when unauthorised
func TestDeleteNotAuthedFail(t *testing.T) {

	clearTable()
	runSQL(insertDummyReviews)

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder("GET", "https://poptape.club/authy/checkaccess/10",
		httpmock.NewStringResponder(401, `{"public_id": "f38ba39a-3682-4803-a498-659f0bf05304" }`))

	req, _ := http.NewRequest("DELETE", "/reviews/e8f48256-2460-418f-81b7-86dad2aa6222", nil)
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("X-Access-Token", "faketoken")
	response := executeRequest(req)

	if checkResponseCode(t, http.StatusUnauthorized, response.Code) {
		fmt.Println("[PASS].....TestDeleteNotAuthedFail")
	}
}

// test review creation
func TestCreateReviewOk(t *testing.T) {

	clearTable()
	runSQL(insertDummyReviews)
	oldRecCnt := getRecCount()

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("GET", "https://poptape.club/authy/checkaccess/10",
		httpmock.NewStringResponder(200, `{"public_id": "f38ba39a-3682-4803-a498-659f0bf05304" }`))

	httpmock.RegisterResponder("GET", "=~^https://poptape.club/auctionhouse/auction/.",
		httpmock.NewStringResponder(200, `{"public_id": "f38ba39a-3682-4803-a498-659f0bf05304" }`))

	//auction_id, review, overall, pap_cost, communication, as_described)
	payload := []byte(createJson)

	req, _ := http.NewRequest("POST", "/reviews", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("X-Access-Token", "faketoken")
	response := executeRequest(req)

	noError := checkResponseCode(t, http.StatusCreated, response.Code)
	var crep CreateResp
	err := json.NewDecoder(response.Body).Decode(&crep)
	if err != nil {
		noError = false
		t.Errorf("Error decoding JSON: " + err.Error())
	}

	if getRecCount() != oldRecCnt+1 {
		noError = false
		t.Errorf("Before and after record counts out by more than +1")
	}

	if noError {
		fmt.Println("[PASS].....TestCreateReviewOk")
	}
	log.Printf("Total call count is %d", httpmock.GetTotalCallCount())

}

// test review creation fails if duplicate attempted
func TestCreateReviewDuplicateReviewFail(t *testing.T) {

	clearTable()

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder("GET", "https://poptape.club/authy/checkaccess/10",
		httpmock.NewStringResponder(200, `{"public_id": "f38ba39a-3682-4803-a498-659f0bf05304" }`))
	url := "https://poptape.club/auctionhouse/auction/f38ba39a-3682-4803-a498-659f0b111111"
	httpmock.RegisterResponder("GET", url,
		httpmock.NewStringResponder(200, `{"message": "whatevs"}`))

	payload := []byte(createJson)

	req, _ := http.NewRequest("POST", "/reviews", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("X-Access-Token", "faketoken")
	response := executeRequest(req)

	noError := checkResponseCode(t, http.StatusCreated, response.Code)

	req, _ = http.NewRequest("POST", "/reviews", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("X-Access-Token", "faketoken")
	response = executeRequest(req)

	noError = checkResponseCode(t, http.StatusInternalServerError, response.Code)
	if noError {
		fmt.Println("[PASS].....TestCreateReviewDuplicateReviewFail")
	}
}

// test review creation fails if 'overall' field is not numeric
func TestCreateReviewFailOnOverall(t *testing.T) {

	clearTable()

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder("GET", "https://poptape.club/authy/checkaccess/10",
		httpmock.NewStringResponder(200, `{"public_id": "f38ba39a-3682-4803-a498-659f0bf05304" }`))
	url := "https://poptape.club/auctionhouse/auction/f38ba39a-3682-4803-a498-659f0b111111"
	httpmock.RegisterResponder("GET", url,
		httpmock.NewStringResponder(200, `{"message": "whatevs"}`))

	var badOverall = `{"auction_id":"f38ba39a-3682-4803-a498-659f0b111111",
"item_id":"f80689a6-9fba-4859-bdde-0a307c696ea8",
"review": "amazing product",
"overall": "a",
"post_and_packaging": 3,
"communication": 4,
"as_described": 4}`

	payload := []byte(badOverall)

	req, _ := http.NewRequest("POST", "/reviews", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("X-Access-Token", "faketoken")
	response := executeRequest(req)

	if checkResponseCode(t, http.StatusBadRequest, response.Code) {
		fmt.Println("[PASS].....TestCreateReviewFailOnOverall")
	}
}

// test review creation fails if 'overall' field is not an integer
func TestCreateReviewFailOnOverallFloat(t *testing.T) {

	clearTable()

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder("GET", "https://poptape.club/authy/checkaccess/10",
		httpmock.NewStringResponder(200, `{"public_id": "f38ba39a-3682-4803-a498-659f0bf05304" }`))
	url := "https://poptape.club/auctionhouse/auction/f38ba39a-3682-4803-a498-659f0b111111"
	httpmock.RegisterResponder("GET", url,
		httpmock.NewStringResponder(200, `{"message": "whatevs"}`))

	var badOverall = `{"auction_id":"f38ba39a-3682-4803-a498-659f0b111111",
"item_id":"f80689a6-9fba-4859-bdde-0a307c696ea8",
"review": "amazing product",
"overall": 3.4,
"post_and_packaging": 3,
"communication": 4,
"as_described": 4}`

	payload := []byte(badOverall)

	req, _ := http.NewRequest("POST", "/reviews", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("X-Access-Token", "faketoken")
	response := executeRequest(req)

	if checkResponseCode(t, http.StatusBadRequest, response.Code) {
		fmt.Println("[PASS].....TestCreateReviewFailOnOverallFloat")
	}
}

// test review creation fails if 'post and packaging' field is not integer
func TestCreateReviewFailOnPostAndPackaging(t *testing.T) {

	clearTable()

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder("GET", "https://poptape.club/authy/checkaccess/10",
		httpmock.NewStringResponder(200, `{"public_id": "f38ba39a-3682-4803-a498-659f0bf05304" }`))
	url := "https://poptape.club/auctionhouse/auction/f38ba39a-3682-4803-a498-659f0b111111"
	httpmock.RegisterResponder("GET", url,
		httpmock.NewStringResponder(200, `{"message": "whatevs"}`))

	var badOverall = `{"auction_id":"f38ba39a-3682-4803-a498-659f0b111111",
"item_id":"f80689a6-9fba-4859-bdde-0a307c696ea8",
"review": "amazing product",
"overall": 4,
"post_and_packaging": "blah",
"communication": 4,
"as_described": 4}`

	payload := []byte(badOverall)

	req, _ := http.NewRequest("POST", "/reviews", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("X-Access-Token", "faketoken")
	response := executeRequest(req)

	if checkResponseCode(t, http.StatusBadRequest, response.Code) {
		fmt.Println("[PASS].....TestCreateReviewFailOnPostAndPackaging")
	}
}

// get reviews written by the user  - no auth needed
func TestGetMetadataOfUserPublicIDFail(t *testing.T) {

	clearTable()
	runSQL(insertDummyReviews)

	req, _ := http.NewRequest("GET", "/reviews/user/blahblah", nil)
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	response := executeRequest(req)

	if checkResponseCode(t, http.StatusBadRequest, response.Code) {
		fmt.Println("[PASS].....TestGetMetadataOfUserPublicIDFail")
	}
}

// get reviews written of the user  - no auth needed
func TestGetMetadataOfUserOK(t *testing.T) {

	clearTable()
	runSQL(insertDummyReviews)

	req, _ := http.NewRequest("GET", "/reviews/user/f38ba39a-3682-4803-a498-659f0bf05304", nil)
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	response := executeRequest(req)

	noError := checkResponseCode(t, http.StatusOK, response.Code)
	// TODO: Add tests for totals returned
	if noError {
		fmt.Println("[PASS].....TestGetMetadataOfUserOK")
	}
}
