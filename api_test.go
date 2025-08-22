package main_test

import (
	"fmt"
	"github.com/cliveyg/poptape-reviews"
	"github.com/jarcoal/httpmock"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

// NewAppForTest replicates main setup but returns *App for use in tests
func NewAppForTest() *main.App {
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

	a := &main.App{}
	a.Log = &logger
	a.InitialiseApp()
	return a
}

var a *main.App

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
	res := a.DB.Where("1 = 1").Delete(&main.Review{})
	if res.Error != nil {
		a.Log.Fatal().Msg(res.Error.Error())
	}
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