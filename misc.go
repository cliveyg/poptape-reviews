package main

import (
	"crypto/tls"
	"fmt"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"
	"time"
)

// ----------------------------------------------------------------------------
// h e l p e r   f u n c t i o n s
// ----------------------------------------------------------------------------

func IsValidUUID(u string) bool {
	_, err := uuid.Parse(u)
	return err == nil
}

func ValidAuction(auctionId, publicId, x string) bool {

	fullURL := GetURL("AUCTIONURL") + auctionId
	log.Println("FULL URL is "+fullURL)
	log.Println("publicId is "+publicId)
	valid := ValidThing(fullURL, x, "auction", publicId)
	log.Printf("Is valid? %t",valid)
	return ValidThing(fullURL, x, "auction", publicId)
}

func ValidItem(itemId, x string) bool {

	fullURL := GetURL("ITEMURL") + itemId
	return ValidThing(fullURL, x, "item", "")
}

func GetURL(t string) string {

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	return os.Getenv(t)

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

func ValidThing(URL, x, thingType, UUID string) bool {

	log.Printf("URL in ValidThing is [%s]",URL)
	req, err := http.NewRequest("GET", URL, nil)
	if err != nil {
		log.Print(err)
		return false
	}
	req.Header.Set("X-Access-Token", x)
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")

	skipVerify := false
	if os.Getenv("ENVIRONMENT") == "DEV" {
		log.Println("skipVerify set to true")
		skipVerify = true
	}
	// skip verify to avoid x509 cert check if in dev env
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: skipVerify},
	}

	client := &http.Client{Timeout: time.Second * 10, Transport: tr}
	resp, e := client.Do(req)
	if e != nil {
		log.Print(fmt.Sprintf("The HTTP request failed with error %s", e))
		return false
	} else {
		log.Println("WOOOOOOOO!")
		defer resp.Body.Close()
		//TODO: check if auction finished and user won
		// when thingType is 'auction'
		if thingType == "auction" {
			log.Printf("Input UUID is [%s]",UUID)
		}
		log.Printf("Response status code is [%d]",resp.StatusCode)
		if resp.StatusCode == 200 {
			return true
		}
	}
	log.Println("We is losst")
	return false

}
