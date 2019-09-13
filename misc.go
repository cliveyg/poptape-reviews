package main

import (
	"github.com/google/uuid"
    "net/http"
    "crypto/tls"
    "log"
    "os"
	"fmt"
    "github.com/joho/godotenv"
	"time"
)

// ----------------------------------------------------------------------------
// h e l p e r   f u n c t i o n s
// ----------------------------------------------------------------------------

func IsValidUUID(u string) bool {
    _, err := uuid.Parse(u)
    return err == nil
}


func ValidAuction(auctionId, x string) (bool) {

	fullURL := GetAuctionURL() + auctionId
	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		log.Print(err)
		return false
	}
	req.Header.Set("X-Access-Token", x)
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")

    // skip verify to avoid x509 cert check - not sure if this is a good idea
    tr := &http.Transport{
        TLSClientConfig: &tls.Config{InsecureSkipVerify : true},
    }

	client := &http.Client{Timeout: time.Second * 10, Transport: tr}
	resp, e := client.Do(req)
	if e != nil {
		log.Print(fmt.Sprintf("The HTTP request failed with error %s", e))
		return false
	} else {
		defer resp.Body.Close()
		if resp.StatusCode == 200 {
			return true
		}
	}
	return false
}


func GetAuctionURL() string {

    err := godotenv.Load()
    if err != nil {
      log.Fatal("Error loading .env file")
    }
    return os.Getenv("AUCTIONURL")

}

func CheckRequest(r *http.Request) (bool, int, string) {

	contype := r.Header.Get("Content-type")

    if !(contype == "application/json" ||
        contype == "application/json; charset=UTF-8") {
        badmess := `{"message": "Request must be json"}`
        return false, http.StatusBadRequest, badmess
    }
	return true, http.StatusOK, ""
}
