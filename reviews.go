package main

import (
    "log"
    "os"
    "github.com/joho/godotenv"
)

func main() {

    err := godotenv.Load()
    if err != nil {
      log.Fatal("Error loading .env file")
    }

	a := App{}
	a.Initialize(
        os.Getenv("DB_HOST"),
		os.Getenv("DB_USERNAME"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"))

	a.Run(os.Getenv("PORT"))

}

