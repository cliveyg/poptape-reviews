package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	//"io/ioutil"
	"github.com/joho/godotenv"
	"os"
	"time"
)

type user struct {
	PublicId string `json:"public_id"`
}

func bouncerSaysOk(r *http.Request) (bool, int, string) {

	contype := r.Header.Get("Content-type")
	badmess := `{"message": "Ooh you are naughty"}`

	if !(contype == "application/json" ||
		contype == "application/json; charset=UTF-8") {
		badmess = `{"message": "Request must be json"}`
		return false, http.StatusBadRequest, badmess
	}

	x := r.Header.Get("X-Access-Token")

	if x != "" {
		// call authy microservice
		req, err := http.NewRequest("GET", getAuthyURL(), nil)
		if err != nil {
			log.Print(err)
			return false, http.StatusUnauthorized, badmess
		}

		log.Print(fmt.Sprintf("X-Access-Token [%s]", x))
		req.Header.Set("X-Access-Token", x)
		req.Header.Set("Content-Type", "application/json; charset=UTF-8")

		client := &http.Client{Timeout: time.Second * 10}
		resp, e := client.Do(req)
		if e != nil {
			log.Print(fmt.Sprintf("The HTTP request failed with error %s", e))
			badmess = `{"message": "I'm sorry Dave"}`
			return false, http.StatusServiceUnavailable, badmess
		} else {
			defer resp.Body.Close()
			//bodyBytes, _ := ioutil.ReadAll(resp.Body)
			//log.Print(string(bodyBytes))
			if resp.StatusCode == 200 {
				var u user
				json.NewDecoder(resp.Body).Decode(&u)
				return true, http.StatusOK, u.PublicId
			}
		}
	}
	return false, http.StatusUnauthorized, badmess
}

func getAuthyURL() string {

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	return os.Getenv("AUTHYURL")

}
