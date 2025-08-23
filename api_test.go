package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/jarcoal/httpmock"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

// NewAppForTest replicates main setup but returns *App for use in tests
func NewAppForTest() *App {
	err := godotenv.Load()
	if err != nil {
		panic("Error loading .env file")
	}

	var logFile *os.File
	filePathName := os.Getenv("LOGFILE")
	logFile, err = os.OpenFile(filePathName, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		panic(err)
	}

	// format logline
	cw := zerolog.ConsoleWriter{Out: logFile, NoColor: true, TimeFormat: time.RFC3339}
	cw.FormatLevel = func(i interface{}) string {
		return strings.ToUpper(fmt.Sprintf("[ %-6s]", i))
	}
	cw.TimeFormat = "[" + time.RFC3339 + "] - "
	cw.FormatCaller = func(i interface{}) string {
		str, _ := i.(string)
		return fmt.Sprintf("['%s']", str)
	}
	cw.PartsOrder = []string{
		zerolog.LevelFieldName,
		zerolog.TimestampFieldName,
		zerolog.MessageFieldName,
		zerolog.CallerFieldName,
	}

	logger := zerolog.New(cw).With().Timestamp().Caller().Logger()
	if os.Getenv("LOGLEVEL") == "debug" {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else if os.Getenv("LOGLEVEL") == "info" {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	}

	a := &App{}
	a.Log = &logger
	a.InitialiseApp()
	return a
}

var a *App

func TestMain(m *testing.M) {
	a = NewAppForTest()
	code := m.Run()
	os.Exit(code)
}

//-----------------------------------------------------------------------------
// h e l p e r   f u n c t i o n s
//-----------------------------------------------------------------------------

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
	res := a.DB.Where("1 = 1").Delete(&Review{})
	if res.Error != nil {
		a.Log.Fatal().Msg(res.Error.Error())
	}
}

func getCountForUUIDKey(key string, id uuid.UUID) int64 {
	var tc int64
	a.DB.Model(&Review{}).Where(key + " = ?", id).Count(&tc)
	return tc
}

func getTotalRecordsInTable() int64 {
	var tc int64
	a.DB.Model(&Review{}).Count(&tc)
	return tc
}

//-----------------------------------------------------------------------------
// s t a r t   o f   t e s t s
//-----------------------------------------------------------------------------

func TestAPIStatus(t *testing.T) {
	req, _ := http.NewRequest("GET", "/reviews/status", nil)
	response := executeRequest(req)

	if checkResponseCode(t, http.StatusOK, response.Code) {
		fmt.Println("[PASS].....TestAPIStatus")
	}
}

func TestEmptyTable(t *testing.T) {

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder("GET", os.Getenv("AUTHYURL"),
		httpmock.NewStringResponder(200, `{"public_id": "f38ba39a-3682-4803-a498-659f0bf05304" }`))

	clearTable()

	req, _ := http.NewRequest("GET", "/reviews", nil)
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("X-Access-Token", "faketoken")
	response := executeRequest(req)

	if checkResponseCode(t, http.StatusNotFound, response.Code) {
		fmt.Println("[PASS].....TestEmptyTable")
	}

}

func TestNoContentTypeHeader(t *testing.T) {

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder("GET", os.Getenv("AUTHYURL"),
		httpmock.NewStringResponder(200, `{"public_id": "f38ba39a-3682-4803-a498-659f0bf05304" }`))

	clearTable()

	req, _ := http.NewRequest("GET", "/reviews", nil)
	response := executeRequest(req)

	if checkResponseCode(t, http.StatusBadRequest, response.Code) {
		fmt.Println("[PASS].....TestNoContentTypeHeader")
	}

}

func TestWrongContentTypeHeader(t *testing.T) {

	clearTable()

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder("GET", os.Getenv("AUTHYURL"),
		httpmock.NewStringResponder(200, `{"public_id": "f38ba39a-3682-4803-a498-659f0bf05304" }`))

	req, _ := http.NewRequest("GET", "/reviews", nil)
	req.Header.Set("Content-Type", "application/html; charset=UTF-8")
	response := executeRequest(req)

	if checkResponseCode(t, http.StatusBadRequest, response.Code) {
		fmt.Println("[PASS].....TestWrongContentTypeHeader")
	}

}

func TestReturnOnlyAuthUserReviews(t *testing.T) {

	clearTable()
	_, err := a.InsertSpecificDummyReviews()
	if err != nil {
		log.Fatal(err.Error())
	}

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder("GET", os.Getenv("AUTHYURL"),
		httpmock.NewStringResponder(200, `{"public_id": "f38ba39a-3682-4803-a498-659f0bf05304" }`))

	req, _ := http.NewRequest("GET", "/reviews", nil)
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("X-Access-Token", "faketoken")
	response := executeRequest(req)

	noError := checkResponseCode(t, http.StatusOK, response.Code)

	var revResp ReviewsResponse
	err = json.NewDecoder(response.Body).Decode(&revResp)
	if err != nil {
		log.Fatal(err.Error())
	}

	if len(revResp.Reviews) != 3 {
		t.Errorf("no of reviews returned doesn't match; should be 3 but is %d", len(revResp.Reviews))
		noError = false
	}

	u, _ := uuid.Parse("f38ba39a-3682-4803-a498-659f0bf05304")
	for _, r := range revResp.Reviews {
		if r.ReviewedBy != u {
			t.Errorf("reviewed by doesn't match")
			noError = false
		}
	}

	if noError {
		fmt.Println("[PASS].....TestReturnOnlyAuthUserReviews")
	}

}

func TestBadAuthyJson(t *testing.T) {

	clearTable()
	_, err := a.InsertSpecificDummyReviews()
	if err != nil {
		log.Fatal(err.Error())
	}

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder("GET", os.Getenv("AUTHYURL"),
		httpmock.NewStringResponder(200, `{"blah": badjson""}`))

	req, _ := http.NewRequest("GET", "/reviews", nil)
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("X-Access-Token", "faketoken")
	response := executeRequest(req)

	noError := checkResponseCode(t, http.StatusBadRequest, response.Code)
	var resp RespMessage
	err = json.NewDecoder(response.Body).Decode(&resp)
	if err != nil {
		log.Fatal(err.Error())
	}
	if resp.Message != "Unable to decode response body" {
		noError = false
		t.Errorf("bad request message [%s] doesn't match expected", resp)
	}

	if noError {
		fmt.Println("[PASS].....TestBadAuthyServiceError")
	}

}

func TestMissingXAccessToken(t *testing.T) {

	clearTable()
	_, err := a.InsertSpecificDummyReviews()
	if err != nil {
		log.Fatal(err.Error())
	}

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder("GET", os.Getenv("AUTHYURL"),
		httpmock.NewStringResponder(200, `{"public_id": "f38ba39a-3682-4803-a498-659f0bf05304" }`))

	req, _ := http.NewRequest("GET", "/reviews", nil)
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	response := executeRequest(req)

	if checkResponseCode(t, http.StatusUnauthorized, response.Code) {
		fmt.Println("[PASS].....TestMissingXAccessToken")
	}
}

func TestGetReviewsByUser(t *testing.T) {

	clearTable()
	_, err := a.InsertSpecificDummyReviews()
	if err != nil {
		log.Fatal(err.Error())
	}

	req, _ := http.NewRequest("GET", "/reviews/by/user/f38ba39a-3682-4803-a498-659f0bf05304", nil)
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	response := executeRequest(req)

	noError := checkResponseCode(t, http.StatusOK, response.Code)

	var revResp ReviewsResponse
	err = json.NewDecoder(response.Body).Decode(&revResp)
	if err != nil {
		log.Fatal(err.Error())
	}

	u, _ := uuid.Parse("f38ba39a-3682-4803-a498-659f0bf05304")
	for _, r := range revResp.Reviews {
		if r.ReviewedBy != u {
			noError = false
			t.Errorf("reviewed by doesn't match")
		}
	}

	if len(revResp.Reviews) != 3 {
		noError = false
		t.Errorf("no of reviews returned on page [%d] doesn't match expected [3]", len(revResp.Reviews))
	}

	if revResp.TotalReviews != 4 {
		noError = false
		t.Errorf("total no of reviews returned [%d] doesn't match data entered [4]", revResp.TotalReviews)
	}

	if revResp.CurrentPage != 1 {
		noError = false
		t.Errorf("current page is [%d] - should be 1", revResp.CurrentPage)
	}

	if revResp.TotalPages != 2 {
		noError = false
		t.Errorf("total pages doesn't match [%d] - should be 2", revResp.TotalPages)
	}

	if noError {
		fmt.Println("[PASS].....TestGetReviewsByUser")
	}

}

func TestBadUUID(t *testing.T) {

	clearTable()
	_, err := a.InsertSpecificDummyReviews()
	if err != nil {
		log.Fatal(err.Error())
	}

	req, _ := http.NewRequest("GET", "/reviews/f38ba39a-3682-4803-a498-659f0bf0530g", nil)
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	response := executeRequest(req)

	if checkResponseCode(t, http.StatusBadRequest, response.Code) {
		fmt.Println("[PASS].....TestBadUUID")
	}
}

func Test404ForValidUUID(t *testing.T) {

	clearTable()
	_, err := a.InsertSpecificDummyReviews()
	if err != nil {
		log.Fatal(err.Error())
	}

	req, _ := http.NewRequest("GET", "/reviews/f38ba39a-3682-4803-a498-659f0bf05311", nil)
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	response := executeRequest(req)

	if checkResponseCode(t, http.StatusNotFound, response.Code) {
		fmt.Println("[PASS].....Test404ForValidUUID")
	}
}

func Test404ForRandomURL(t *testing.T) {

	req, _ := http.NewRequest("GET", "/reviews/f38ba39a/someurl/999", nil)
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	response := executeRequest(req)

	if checkResponseCode(t, http.StatusNotFound, response.Code) {
		fmt.Println("[PASS].....Test404ForRandomURL")
	}
}

func TestGetReviewsByAuction(t *testing.T) {

	clearTable()
	_, err := a.InsertSpecificDummyReviews()
	if err != nil {
		log.Fatal(err.Error())
	}

	req, _ := http.NewRequest("GET", "/reviews/auction/e77be9e0-bb00-49bc-9e7d-d7cc7072ab8c", nil)
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	response := executeRequest(req)

	noError := checkResponseCode(t, http.StatusOK, response.Code)

	var revResp ReviewsResponse
	err = json.NewDecoder(response.Body).Decode(&revResp)
	if err != nil {
		log.Fatal(err.Error())
	}

	if len(revResp.Reviews) != 2 {
		noError = false
		t.Errorf("no of reviews returned doesn't match")
	}

	if noError {
		fmt.Println("[PASS].....TestGetReviewsByAuction")
	}
}

func TestGetReviewById(t *testing.T) {

	clearTable()
	_, err := a.InsertSpecificDummyReviews()
	if err != nil {
		log.Fatal(err.Error())
	}

	req, _ := http.NewRequest("GET", "/reviews/e8f48256-2460-418f-81b7-86dad2aa6333", nil)
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	response := executeRequest(req)

	noError := checkResponseCode(t, http.StatusOK, response.Code)
	var revResp ReviewsResponse
	err = json.NewDecoder(response.Body).Decode(&revResp)
	if err != nil {
		log.Fatal(err.Error())
	}

	if revResp.Reviews[0].ReviewedBy.String() != "f38ba39a-3682-4803-a498-659f0bf05000" {
		noError = false
		t.Errorf("reviewed by doesn't match")
	}
	if revResp.Reviews[0].AuctionId.String() != "e77be9e0-bb00-49bc-9e7d-d7cc7072ab33" {
		noError = false
		t.Errorf("auction id by doesn't match")
	}
	if revResp.Reviews[0].ItemId.String() != "7d1aa876-9be8-441f-ad86-daaa51872333" {
		noError = false
		t.Errorf("item id by doesn't match")
	}
	if revResp.Reviews[0].Seller.String() != "f38ba39a-3682-4803-a498-659f0bf05304" {
		noError = false
		t.Errorf("item id by doesn't match")
	}
	if revResp.Reviews[0].Overall != 2 {
		noError = false
		t.Errorf("overall by doesn't match")
	}
	if revResp.Reviews[0].PapCost != 2 {
		noError = false
		t.Errorf("post_and_packaging by doesn't match")
	}
	if revResp.Reviews[0].Comm != 6 {
		noError = false
		t.Errorf("communication by doesn't match")
	}
	if revResp.Reviews[0].AsDesc != 1 {
		noError = false
		t.Errorf("as_described by doesn't match")
	}
	if noError {
		fmt.Println("[PASS].....TestGetReviewById")
	}
}

func TestGetReviewByItem(t *testing.T) {

	clearTable()
	_, err := a.InsertSpecificDummyReviews()
	if err != nil {
		log.Fatal(err.Error())
	}

	req, _ := http.NewRequest("GET", "/reviews/item/7d1aa876-9be8-441f-ad86-d86e5faddd81", nil)
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	response := executeRequest(req)

	noError := checkResponseCode(t, http.StatusOK, response.Code)
	var revResp ReviewsResponse
	err = json.NewDecoder(response.Body).Decode(&revResp)
	if err != nil {
		log.Fatal(err.Error())
	}

	if revResp.Reviews[0].ReviewId.String() != "e8f48256-2460-418f-81b7-86dad2aa6aaa" {
		noError = false
		t.Errorf("review id doesn't match")
	}
	if revResp.Reviews[0].ItemId.String() != "7d1aa876-9be8-441f-ad86-d86e5faddd81" {
		noError = false
		t.Errorf("item id doesn't match")
	}
	if noError {
		fmt.Println("[PASS].....TestGetReviewByItem")
	}
}

func TestGetReviewsOfUser(t *testing.T) {

	clearTable()
	_, err := a.InsertSpecificDummyReviews()
	if err != nil {
		log.Fatal(err.Error())
	}

	req, _ := http.NewRequest("GET", "/reviews/of/user/46d7d11c-fa06-4e54-8208-95433b98cfc9", nil)
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	response := executeRequest(req)

	noError := checkResponseCode(t, http.StatusOK, response.Code)
	var revResp ReviewsResponse
	err = json.NewDecoder(response.Body).Decode(&revResp)
	if err != nil {
		log.Fatal(err.Error())
	}

	for _, r := range revResp.Reviews {
		if r.Seller.String() != "46d7d11c-fa06-4e54-8208-95433b98cfc9" {
			noError = false
			t.Errorf("reviewed by doesn't match")
		}
	}

	if len(revResp.Reviews) != 3 {
		noError = false
		t.Errorf("no of reviews returned doesn't match: expected 3 and got %d", len(revResp.Reviews))
	}
	if noError {
		fmt.Println("[PASS].....TestGetReviewsOfUser")
	}
}

func TestGetReviewsByAnotherUser(t *testing.T) {

	clearTable()
	_, err := a.InsertSpecificDummyReviews()
	if err != nil {
		log.Fatal(err.Error())
	}

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder("GET", os.Getenv("AUTHYURL"),
		httpmock.NewStringResponder(200, `{"public_id": "f38ba39a-3682-4803-a498-659f0bf05000" }`))

	req, _ := http.NewRequest("GET", "/reviews/by/user/f38ba39a-3682-4803-a498-659f0bf05304", nil)
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("X-Access-Token", "somefaketoken")
	response := executeRequest(req)

	noError := checkResponseCode(t, http.StatusOK, response.Code)

	var revResp ReviewsResponse
	err = json.NewDecoder(response.Body).Decode(&revResp)
	if err != nil {
		log.Fatal(err.Error())
	}

	for _, r := range revResp.Reviews {
		if r.ReviewedBy.String() != "f38ba39a-3682-4803-a498-659f0bf05304" {
			noError = false
			t.Errorf("reviewed by doesn't match")
		}
	}

	if len(revResp.Reviews) != 3 {
		noError = false
		t.Errorf("no of reviews returned doesn't match")
	}
	if noError {
		fmt.Println("[PASS].....TestGetReviewsByAnotherUser")
	}
}

func TestDeleteReviewOk(t *testing.T) {

	clearTable()
	_, err := a.InsertSpecificDummyReviews()
	if err != nil {
		log.Fatal(err.Error())
	}

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder("GET", os.Getenv("AUTHYURL"),
		httpmock.NewStringResponder(200, `{"public_id": "f38ba39a-3682-4803-a498-659f0bf05304" }`))

	req, _ := http.NewRequest("DELETE", "/reviews/e8f48256-2460-418f-81b7-86dad2aa6222", nil)
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("X-Access-Token", "faketoken")
	response := executeRequest(req)

	noError := checkResponseCode(t, http.StatusOK, response.Code)

	req, _ = http.NewRequest("GET", "/reviews/e8f48256-2460-418f-81b7-86dad2aa6222", nil)
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("X-Access-Token", "faketoken")
	response = executeRequest(req)

	noError = checkResponseCode(t, http.StatusNotFound, response.Code)
	if noError {
		fmt.Println("[PASS].....TestDeleteReviewOk")
	}
}

func TestDeleteFail(t *testing.T) {

	clearTable()
	_, err := a.InsertSpecificDummyReviews()
	if err != nil {
		log.Fatal(err.Error())
	}

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder("GET", os.Getenv("AUTHYURL"),
		httpmock.NewStringResponder(200, `{"public_id": "f38ba39a-3682-4803-a498-659f0bf05304" }`))

	req, _ := http.NewRequest("DELETE", "/reviews/e8f48256-2460-418f-81b7-86dad2aa6333", nil)
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("X-Access-Token", "faketoken")
	response := executeRequest(req)

	if checkResponseCode(t, http.StatusBadRequest, response.Code) {
		fmt.Println("[PASS].....TestDeleteFail")
	}
}

func TestDeleteNotAuthedFail(t *testing.T) {

	clearTable()
	_, err := a.InsertSpecificDummyReviews()
	if err != nil {
		log.Fatal(err.Error())
	}

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder("GET", os.Getenv("AUTHYURL"),
		httpmock.NewStringResponder(401, `{"public_id": "f38ba39a-3682-4803-a498-659f0bf05304" }`))

	req, _ := http.NewRequest("DELETE", "/reviews/e8f48256-2460-418f-81b7-86dad2aa6222", nil)
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("X-Access-Token", "faketoken")
	response := executeRequest(req)

	if checkResponseCode(t, http.StatusUnauthorized, response.Code) {
		fmt.Println("[PASS].....TestDeleteNotAuthedFail")
	}
}

func TestCreateReviewOk(t *testing.T) {

	clearTable()
	_, err := a.InsertSpecificDummyReviews()
	if err != nil {
		log.Fatal(err.Error())
	}
	oldRecCnt := getTotalRecordsInTable()

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("GET", os.Getenv("AUTHYURL"),
		httpmock.NewStringResponder(200, `{"public_id": "f38ba39a-3682-4803-a498-659f0bf05304" }`))

	httpmock.RegisterResponder("GET", "=~^https://poptape.club/auctionhouse/auction/.",
		httpmock.NewStringResponder(200, `{"public_id": "f38ba39a-3682-4803-a498-659f0bf05304" }`))

	httpmock.RegisterResponder("GET", "=~^https://poptape.club/items/.",
		httpmock.NewStringResponder(200, `{"public_id": "f38ba39a-3682-4803-a498-659f0bf05304" }`))

	//auction_id, review, overall, pap_cost, communication, as_described)
	payload := []byte(createJson)

	req, _ := http.NewRequest("POST", "/reviews", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("X-Access-Token", "faketoken")
	response := executeRequest(req)

	noError := checkResponseCode(t, http.StatusCreated, response.Code)
	var crep CreateReviewResp
	err = json.NewDecoder(response.Body).Decode(&crep)
	if err != nil {
		noError = false
		t.Errorf("Error decoding returned JSON: " + err.Error())
	}

	if getTotalRecordsInTable() != oldRecCnt+1 {
		noError = false
		t.Errorf("Before and after record counts out by more than +1")
	}

	if noError {
		fmt.Println("[PASS].....TestCreateReviewOk")
	}
	log.Printf("Total call count is %d", httpmock.GetTotalCallCount())

}

func TestCreateReviewFailBouncer(t *testing.T) {

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("GET", os.Getenv("AUTHYURL"),
		httpmock.NewStringResponder(401, `{}`))

	//auction_id, review, overall, pap_cost, communication, as_described)
	payload := []byte(createJson)

	req, _ := http.NewRequest("POST", "/reviews", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("X-Access-Token", "faketoken")
	response := executeRequest(req)
	noError := checkResponseCode(t, http.StatusUnauthorized, response.Code)

	if noError {
		fmt.Println("[PASS].....TestCreateReviewFailBouncer")
	}

}

func TestCreateReviewFailBadInput1(t *testing.T) {

	clearTable()

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("GET", os.Getenv("AUTHYURL"),
		httpmock.NewStringResponder(200, `{"public_id": "f38ba39a-3682-4803-a498-659f0bf05304" }`))

	httpmock.RegisterResponder("GET", "=~^https://poptape.club/auctionhouse/auction/.",
		httpmock.NewStringResponder(200, `{"public_id": "f38ba39a-3682-4803-a498-659f0bf05304" }`))

	httpmock.RegisterResponder("GET", "=~^https://poptape.club/items/.",
		httpmock.NewStringResponder(200, `{"public_id": "f38ba39a-3682-4803-a498-659f0bf05304" }`))

	payload := []byte(createJsonMissingReviewedBy)
	req, _ := http.NewRequest("POST", "/reviews", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("X-Access-Token", "faketoken")
	response := executeRequest(req)

	noError := checkResponseCode(t, http.StatusBadRequest, response.Code)
	var resp RespMessage
	err := json.NewDecoder(response.Body).Decode(&resp)
	if err != nil {
		noError = false
		t.Errorf("Error decoding returned JSON: " + err.Error())
	}
	if resp.Message != "Input data is incorrect" {
		noError = false
		t.Errorf("bad request message [%s] doesn't match expected", resp)
	}

	if noError {
		fmt.Println("[PASS].....TestCreateReviewFailBadInput1")
	}

}

func TestCreateReviewFailBadInput2(t *testing.T) {

	clearTable()

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("GET", os.Getenv("AUTHYURL"),
		httpmock.NewStringResponder(200, `{"public_id": "f38ba39a-3682-4803-a498-659f0bf05304" }`))

	httpmock.RegisterResponder("GET", "=~^https://poptape.club/auctionhouse/auction/.",
		httpmock.NewStringResponder(200, `{"public_id": "f38ba39a-3682-4803-a498-659f0bf05304" }`))

	httpmock.RegisterResponder("GET", "=~^https://poptape.club/items/.",
		httpmock.NewStringResponder(200, `{"public_id": "f38ba39a-3682-4803-a498-659f0bf05304" }`))

	payload := []byte(createJsonReviewedByIncorrect)
	req, _ := http.NewRequest("POST", "/reviews", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("X-Access-Token", "faketoken")
	response := executeRequest(req)

	noError := checkResponseCode(t, http.StatusBadRequest, response.Code)
	var resp RespMessage
	err := json.NewDecoder(response.Body).Decode(&resp)
	if err != nil {
		noError = false
		t.Errorf("Error decoding returned JSON: " + err.Error())
	}
	if resp.Message != "Reviewer doesn't match logged in user" {
		noError = false
		t.Errorf("bad request message [%s] doesn't match expected", resp)
	}

	if noError {
		fmt.Println("[PASS].....TestCreateReviewFailBadInput2")
	}

}

func TestGetReviewsByUserFailNoContentHeader(t *testing.T) {

	clearTable()
	_, err := a.InsertSpecificDummyReviews()
	if err != nil {
		log.Fatal(err.Error())
	}

	req, _ := http.NewRequest("GET", "/reviews/by/user/f38ba39a-3682-4803-a498-659f0bf05304", nil)
	response := executeRequest(req)

	noError := checkResponseCode(t, http.StatusBadRequest, response.Code)
	var resp RespMessage
	err = json.NewDecoder(response.Body).Decode(&resp)
	if err != nil {
		noError = false
		t.Errorf("Error decoding returned JSON: " + err.Error())
	}
	if resp.Message != "Request must be json" {
		noError = false
		t.Errorf("bad request message [%s] doesn't match expected", resp.Message)
	}

	if noError {
		fmt.Println("[PASS].....TestGetReviewsByUserFailNoContentHeader")
	}

}

func TestInvalidOrderBy(t *testing.T) {

	clearTable()
	_, err := a.InsertSpecificDummyReviews()
	if err != nil {
		log.Fatal(err.Error())
	}

	req, _ := http.NewRequest("GET", "/reviews/by/user/f38ba39a-3682-4803-a498-659f0bf05304?orderby=blah", nil)
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	response := executeRequest(req)

	noError := checkResponseCode(t, http.StatusBadRequest, response.Code)
	var resp RespMessage
	err = json.NewDecoder(response.Body).Decode(&resp)
	if err != nil {
		noError = false
		t.Errorf("Error decoding returned JSON: " + err.Error())
	}
	if resp.Message != "Not a valid orderby value" {
		noError = false
		t.Errorf("bad request message [%s] doesn't match expected", resp.Message)
	}

	if noError {
		fmt.Println("[PASS].....TestInvalidOrderBy")
	}

}

func TestInvalidSortValue(t *testing.T) {

	clearTable()
	_, err := a.InsertSpecificDummyReviews()
	if err != nil {
		log.Fatal(err.Error())
	}

	req, _ := http.NewRequest("GET", "/reviews/by/user/f38ba39a-3682-4803-a498-659f0bf05304?sort=blah", nil)
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	response := executeRequest(req)

	noError := checkResponseCode(t, http.StatusBadRequest, response.Code)
	var resp RespMessage
	err = json.NewDecoder(response.Body).Decode(&resp)
	if err != nil {
		noError = false
		t.Errorf("Error decoding returned JSON: " + err.Error())
	}
	if resp.Message != "Not a valid sort value" {
		noError = false
		t.Errorf("bad request message [%s] doesn't match expected", resp.Message)
	}

	if noError {
		fmt.Println("[PASS].....TestInvalidSortValue")
	}

}

func TestInvalidPageValue(t *testing.T) {

	clearTable()
	_, err := a.InsertSpecificDummyReviews()
	if err != nil {
		log.Fatal(err.Error())
	}

	req, _ := http.NewRequest("GET", "/reviews/by/user/f38ba39a-3682-4803-a498-659f0bf05304?page=a", nil)
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	response := executeRequest(req)

	noError := checkResponseCode(t, http.StatusBadRequest, response.Code)
	var resp RespMessage
	err = json.NewDecoder(response.Body).Decode(&resp)
	if err != nil {
		noError = false
		t.Errorf("Error decoding returned JSON: " + err.Error())
	}
	if resp.Message != "Not a valid page value" {
		noError = false
		t.Errorf("bad request message [%s] doesn't match expected", resp.Message)
	}

	if noError {
		fmt.Println("[PASS].....TestInvalidPageValue")
	}

}

func TestNegativePageValue(t *testing.T) {

	clearTable()
	_, err := a.InsertSpecificDummyReviews()
	if err != nil {
		log.Fatal(err.Error())
	}

	req, _ := http.NewRequest("GET", "/reviews/by/user/f38ba39a-3682-4803-a498-659f0bf05304?page=-1", nil)
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	response := executeRequest(req)

	noError := checkResponseCode(t, http.StatusOK, response.Code)
	var revResp ReviewsResponse
	err = json.NewDecoder(response.Body).Decode(&revResp)
	if err != nil {
		log.Fatal(err.Error())
	}

	if revResp.CurrentPage != 1 {
		noError = false
		t.Errorf("Current page value [%d] incorrect. Should be 1", revResp.CurrentPage)
	}

	if noError {
		fmt.Println("[PASS].....TestNegativePageValue")
	}

}

func TestInvalidPageSizeEnvVar(t *testing.T) {

	clearTable()
	_, err := a.InsertSpecificDummyReviews()
	if err != nil {
		log.Fatal(err.Error())
	}

	t.Setenv("PAGESIZE", "a")
	req, _ := http.NewRequest("GET", "/reviews/by/user/f38ba39a-3682-4803-a498-659f0bf05304?page=1", nil)
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	response := executeRequest(req)

	noError := checkResponseCode(t, http.StatusInternalServerError, response.Code)
	var resp RespMessage
	err = json.NewDecoder(response.Body).Decode(&resp)
	if err != nil {
		noError = false
		t.Errorf("Error decoding returned JSON: " + err.Error())
	}
	if resp.Message != "Error in pagesize env var" {
		noError = false
		t.Errorf("bad request message [%s] doesn't match expected", resp.Message)
	}

	if noError {
		fmt.Println("[PASS].....TestInvalidPageSizeEnvVar")
	}

}

func TestInvalidPageSizeQueryString(t *testing.T) {

	clearTable()
	_, err := a.InsertSpecificDummyReviews()
	if err != nil {
		log.Fatal(err.Error())
	}

	req, _ := http.NewRequest("GET", "/reviews/by/user/f38ba39a-3682-4803-a498-659f0bf05304?pagesize=a", nil)
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	response := executeRequest(req)

	noError := checkResponseCode(t, http.StatusBadRequest, response.Code)
	var resp RespMessage
	err = json.NewDecoder(response.Body).Decode(&resp)
	if err != nil {
		noError = false
		t.Errorf("Error decoding returned JSON: " + err.Error())
	}
	if resp.Message != "Error in pagesize querystring" {
		noError = false
		t.Errorf("bad request message [%s] doesn't match expected", resp.Message)
	}

	if noError {
		fmt.Println("[PASS].....TestInvalidPageSizeQueryString")
	}

}


func TestInvalidPageSize(t *testing.T) {

	clearTable()
	_, err := a.InsertSpecificDummyReviews()
	if err != nil {
		log.Fatal(err.Error())
	}

	req, _ := http.NewRequest("GET", "/reviews/by/user/f38ba39a-3682-4803-a498-659f0bf05304?pagesize=200", nil)
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	response := executeRequest(req)

	noError := checkResponseCode(t, http.StatusOK, response.Code)

	if noError {
		fmt.Println("[PASS].....TestInvalidPageSize")
	}

}

func TestPageValueTooBig(t *testing.T) {

	clearTable()
	_, err := a.InsertSpecificDummyReviews()
	if err != nil {
		log.Fatal(err.Error())
	}

	req, _ := http.NewRequest("GET", "/reviews/by/user/f38ba39a-3682-4803-a498-659f0bf05304?page=10", nil)
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	response := executeRequest(req)

	noError := checkResponseCode(t, http.StatusBadRequest, response.Code)
	var resp RespMessage
	err = json.NewDecoder(response.Body).Decode(&resp)
	if err != nil {
		noError = false
		t.Errorf("Error decoding returned JSON: " + err.Error())
	}
	if resp.Message != "Page value is incorrect" {
		noError = false
		t.Errorf("bad request message [%s] doesn't match expected", resp.Message)
	}

	if noError {
		fmt.Println("[PASS].....TestPageValueTooBig")
	}

}

func TestDeleteReviewFailBadUUID(t *testing.T) {

	clearTable()
	_, err := a.InsertSpecificDummyReviews()
	if err != nil {
		log.Fatal(err.Error())
	}

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder("GET", os.Getenv("AUTHYURL"),
		httpmock.NewStringResponder(200, `{"public_id": "f38ba39a-3682-4803-a498-659f0bf05304" }`))

	req, _ := http.NewRequest("DELETE", "/reviews/e8f48256-2460-418f-81b7-86dad2aa622z", nil)
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("X-Access-Token", "faketoken")
	response := executeRequest(req)

	noError := checkResponseCode(t, http.StatusBadRequest, response.Code)
	var resp RespMessage
	err = json.NewDecoder(response.Body).Decode(&resp)
	if err != nil {
		noError = false
		t.Errorf("Error decoding returned JSON: " + err.Error())
	}
	if resp.Message != "Not a uuid string" {
		noError = false
		t.Errorf("bad request message [%s] doesn't match expected", resp.Message)
	}

	if noError {
		fmt.Println("[PASS].....TestDeleteReviewFailBadUUID")
	}
}

func TestGetMetadataOK(t *testing.T) {

	clearTable()
	_, err := a.InsertSpecificDummyReviews()
	if err != nil {
		log.Fatal(err.Error())
	}

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder("GET", "=~username",
		httpmock.NewStringResponder(200, `{"foo": "bar"}`))

	req, _ := http.NewRequest("GET", "/reviews/user/f38ba39a-3682-4803-a498-659f0bf05304", nil)
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	response := executeRequest(req)

	noError := checkResponseCode(t, http.StatusOK, response.Code)

	var mResp MetadataResp
	err = json.NewDecoder(response.Body).Decode(&mResp)
	if err != nil {
		log.Fatal(err.Error())
	}

	if mResp.PublicId != "f38ba39a-3682-4803-a498-659f0bf05304" {
		noError = false
		t.Errorf("returned public id doesn't match sent id")
	}
	if mResp.Score != 89 {
		noError = false
		t.Errorf("returned score [%d] doesn't match expected [89]", mResp.Score)
	}
	if mResp.TotalReviewsByUser != 4 {
		noError = false
		t.Errorf("returned reviews by user [%d] doesn't match expected [89]", mResp.TotalReviewsByUser)
	}
	if mResp.TotalReviewsOfUser != 1 {
		noError = false
		t.Errorf("returned reviews of user [%d] doesn't match expected [89]", mResp.TotalReviewsOfUser)
	}

	if noError {
		fmt.Println("[PASS].....TestGetMetadataOK")
	}
}

func TestGetMetadataFailNoContentTypeHdr(t *testing.T) {

	clearTable()
	_, err := a.InsertSpecificDummyReviews()
	if err != nil {
		log.Fatal(err.Error())
	}
	httpmock.RegisterResponder("GET", "=~username",
		httpmock.NewStringResponder(200, `{}`))

	req, _ := http.NewRequest("GET", "/reviews/user/f38ba39a-3682-4803-a498-659f0bf05304", nil)
	response := executeRequest(req)

	noError := checkResponseCode(t, http.StatusBadRequest, response.Code)

	if noError {
		fmt.Println("[PASS].....TestGetMetadataFailNoContentTypeHdr")
	}
}

func TestPaginationOK(t *testing.T) {

	clearTable()
	_, err := a.InsertSpecificDummyReviews()
	if err != nil {
		log.Fatal(err.Error())
	}

	t.Setenv("PAGESIZE", "1")

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder("GET", os.Getenv("AUTHYURL"),
		httpmock.NewStringResponder(200, `{"public_id": "f38ba39a-3682-4803-a498-659f0bf05000" }`))

	req, _ := http.NewRequest("GET", "/reviews/by/user/f38ba39a-3682-4803-a498-659f0bf05304?page=2", nil)
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("X-Access-Token", "somefaketoken")
	response := executeRequest(req)

	noError := checkResponseCode(t, http.StatusOK, response.Code)

	var revResp ReviewsResponse
	err = json.NewDecoder(response.Body).Decode(&revResp)
	if err != nil {
		log.Fatal(err.Error())
	}

	for _, r := range revResp.Reviews {
		if r.ReviewedBy.String() != "f38ba39a-3682-4803-a498-659f0bf05304" {
			noError = false
			t.Errorf("reviewed by doesn't match")
		}
	}

	if len(revResp.Reviews) != 1 {
		noError = false
		t.Errorf("no of reviews returned doesn't match")
	}

	if revResp.TotalReviews != 4 {
		noError = false
		t.Errorf("total no of reviews [%d] returned doesn't match expected [4]", revResp.TotalReviews)
	}

	if len(revResp.URLS) != 2 {
		noError = false
		t.Errorf("total no of reviews [%d] returned doesn't match expected [2]", len(revResp.URLS))
	}

	pu := URL{
		PrevURL: "https://prevnext.com/reviews/by/user/f38ba39a-3682-4803-a498-659f0bf05304?page=1",
		NextURL: "https://prevnext.com/reviews/by/user/f38ba39a-3682-4803-a498-659f0bf05304?page=3",
	}
	if revResp.URLS[0].PrevURL != pu.PrevURL {
		noError = false
		t.Errorf("Prev URL [%s] doesn't match expected", revResp.URLS[0].PrevURL)
	}

	if revResp.URLS[1].NextURL != pu.NextURL  {
		noError = false
		t.Errorf("Prev URL [%s] doesn't match expected", revResp.URLS[1].NextURL)
	}

	if revResp.TotalPages != 4 {
		noError = false
		t.Errorf("Total pages [%d] doesn't match expected [4]", revResp.TotalPages)
	}

	if noError {
		fmt.Println("[PASS].....TestPaginationOK")
	}
}