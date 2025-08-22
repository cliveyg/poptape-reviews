package main

import (
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
	json.NewDecoder(response.Body).Decode(&revResp)
	if err != nil {
		log.Fatal(err.Error())
	}

	if len(revResp.Reviews) != 4 {
		t.Errorf("no of reviews returned doesn't match; should be 4 but is %d", len(revResp.Reviews))
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
	json.NewDecoder(response.Body).Decode(&revResp)
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