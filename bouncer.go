package main

import (
    "net/http"
    "log"
    "fmt"
    "os"
    "github.com/joho/godotenv"
)

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
		client := &http.Client{}
        req, err := http.NewRequest("GET", getAuthyURL(), nil)
        if err != nil {
            log.Fatal(err)
			return false, http.StatusUnauthorized, badmess
        }
        req.Header.Set("X-Access-Token", x)
		req.Header.Set("Content-Type", "application/json; charset=UTF-8")
		resp, e := client.Do(req)
		if e != nil {
			log.Fatal(fmt.Sprintf("The HTTP request failed with error %s", e))
			badmess = `{"message": "I'm sorry Dave"}`
			return false, http.StatusServiceUnavailable, badmess
		} else {
			if resp.StatusCode == 200 {
				//mess := `{"message": "System running..."}`
                return true, http.StatusOK, ""
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
