package main

import (
	"fmt"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"log"
	"os"
	"strings"
	"time"
)

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// setup logging - for some reason we have to do this here
	// when abstracted into another method it doesn't seem to work
	var logFile *os.File

	filePathName := os.Getenv("LOGFILE")
	logFile, err = os.OpenFile(filePathName, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer func(logFile *os.File) {
		err := logFile.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(logFile)

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

	// create log
	logger := zerolog.New(cw).With().Timestamp().Caller().Logger()
	logger.Info().Msg("-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-")
	logger.Info().Msg("Logging setup successfully")
	if os.Getenv("LOGLEVEL") == "debug" {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else if os.Getenv("LOGLEVEL") == "info" {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	}

	a := App{}
	//a.Initialize(
	//	os.Getenv("DB_HOST"),
	//	os.Getenv("DB_USERNAME"),
	//	os.Getenv("DB_PASSWORD"),
	//	os.Getenv("DB_NAME"))

	a.Log = &logger
	a.InitialiseApp()
	a.Run(":" + os.Getenv("PORT"))

}
