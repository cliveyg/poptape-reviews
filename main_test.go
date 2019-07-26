// main_test.go

package main_test

import (
	"os"
	"testing"
	"github.com/joho/godotenv"
	"."
	"net/http"
	"net/http/httptest"
	"log"
	"github.com/jarcoal/httpmock"
)

var a main.App

func TestMain(m *testing.M) {

    err := godotenv.Load()
    if err != nil {
      log.Fatal("Error loading .env file")
    }

	a = main.App{}
	a.Initialize(
		os.Getenv("TESTDB_USERNAME"),
		os.Getenv("TESTDB_PASSWORD"),
		os.Getenv("TESTDB_NAME"))

	ensureTableExists()

	code := m.Run()

	clearTable()

	os.Exit(code)
}

func ensureTableExists() {
	if _, err := a.DB.Exec(tableCreationQuery); err != nil {
		log.Fatal(err)
	}
}

func executeRequest(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	a.Router.ServeHTTP(rr, req)

	return rr
}

func checkResponseCode(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("Expected response code %d. Got %d\n", expected, actual)
	}
}

func clearTable() {
	a.DB.Exec("DELETE FROM reviews")
}

const tableCreationQuery = `CREATE TABLE IF NOT EXISTS reviews
(
    review_id CHAR(36) UNIQUE NOT NULL,
    public_id CHAR(36) NOT NULL,
    auction_id CHAR(36) NOT NULL,
    review VARCHAR(2000),
    overall INT NOT NULL DEFAULT 0,
    pap_cost INT NOT NULL DEFAULT 0,
    communication INT NOT NULL DEFAULT 0,
    as_described INT NOT NULL DEFAULT 0,
    created TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
    UNIQUE (public_id, auction_id),
    CONSTRAINT reviews_pkey PRIMARY KEY (review_id)
)`

// ----------------------------------------------------------------------------
// s t a r t   o f   t e s t s
// ----------------------------------------------------------------------------

func TestAPIStatus(t *testing.T) {

    req, _ := http.NewRequest("GET", "/reviews/status", nil)
    req.Header.Set("Content-Type", "application/json; charset=UTF-8")
    response := executeRequest(req)

    checkResponseCode(t, http.StatusOK, response.Code)

}

func TestEmptyTable(t *testing.T) {

//    httpmock.Activate()
//    defer httpmock.DeactivateAndReset()
//    httpmock.RegisterResponder("GET", "https://poptape.club/authy/checkaccess/10",
//    httpmock.NewStringResponder(200, `{"public_id": "f38ba39a-3682-4803-a498-659f0bf05304" }`))

    clearTable()

    req, _ := http.NewRequest("GET", "/reviews", nil)
    req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("X-Access-Token", "faketoken")
    response := executeRequest(req)

    checkResponseCode(t, http.StatusOK, response.Code)

    if body := response.Body.String(); body != "[]" {
        t.Errorf("Expected an empty array. Got %s", body)
    }
}



//func TestEmptyTable(t *testing.T) {
//	clearTable()
//
//	req, _ := http.NewRequest("GET", "/reviews", nil)
//	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
//	response := executeRequest(req)
//
//	checkResponseCode(t, http.StatusOK, response.Code)
//
//	if body := response.Body.String(); body != "[]" {
//		t.Errorf("Expected an empty array. Got %s", body)
//	}
//}

