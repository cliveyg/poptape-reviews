package main_test

import (
	"fmt"
	"github.com/cliveyg/poptape-reviews"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

// NewAppForTest replicates your main setup but returns *App for use in test
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
	// Note: We don't defer logFile.Close() here because the process is not short-lived like main()

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

var app *main.App

func TestMain(m *testing.M) {
	app = NewAppForTest()
	code := m.Run()
	os.Exit(code)
}

func executeRequest(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	app.Router.ServeHTTP(rr, req)
	return rr
}

// start of tests

func TestAPIStatus(t *testing.T) {
	req, _ := http.NewRequest("GET", "/status", nil)
	response := executeRequest(req)

	if response.Code != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", response.Code)
	}
}