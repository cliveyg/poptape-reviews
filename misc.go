package main

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"io"
	"log"
	"net/http"
	"sync"
	"time"
)

// ----------------------------------------------------------------------------
// h e l p e r   f u n c t i o n s
// ----------------------------------------------------------------------------

func IsValidUUID(u string) bool {
	_, err := uuid.Parse(u)
	return err == nil
}

func checkRequest(c *gin.Context) (bool, int, string) {

	ct := c.GetHeader("Content-type")

	if !(ct == "application/json" ||
		ct == "application/json; charset=UTF-8") {
		return false, http.StatusBadRequest, `{"message": "Request must be json"}`
	}
	return true, http.StatusOK, ""
}

// HTTPRequest describes a single HTTP request with headers and a destination object for the response.
type HTTPRequest struct {
	URL     string
	Headers map[string]string
	Result  interface{} // Pointer to the struct to unmarshal into
}

// HTTPResponse contains the HTTP status code and any error.
type HTTPResponse struct {
	StatusCode int
	Err        error
}

func (a *App) fetchAndUnmarshalRequests(requests []HTTPRequest) []HTTPResponse {
	var wg sync.WaitGroup
	responses := make([]HTTPResponse, len(requests))

	for i, req := range requests {
		wg.Add(1)
		go func(idx int, r HTTPRequest) {
			defer wg.Done()
			httpReq, err := http.NewRequest("GET", r.URL, nil)
			if err != nil {
				responses[idx] = HTTPResponse{Err: err}
				return
			}
			for k, v := range r.Headers {
				httpReq.Header.Set(k, v)
			}
			resp, err := http.DefaultClient.Do(httpReq)
			if err != nil {
				responses[idx] = HTTPResponse{Err: err}
				return
			}
			defer resp.Body.Close()
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				responses[idx] = HTTPResponse{StatusCode: resp.StatusCode, Err: err}
				return
			}
			if r.Result != nil {
				a.Log.Info().Msgf("Body is [%s]", body)
				if err := json.Unmarshal(body, r.Result); err != nil {
					responses[idx] = HTTPResponse{StatusCode: resp.StatusCode, Err: err}
					return
				}
			}
			responses[idx] = HTTPResponse{StatusCode: resp.StatusCode}
		}(i, req)
	}
	wg.Wait()
	return responses
}

func ValidThing(URL, x, thingType, UUID string) bool {

	req, err := http.NewRequest("GET", URL, nil)
	if err != nil {
		log.Print(err)
		return false
	}

	req.Header.Set("X-Access-Token", x)
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")

	client := &http.Client{Timeout: time.Second * 10}
	resp, e := client.Do(req)

	//NB Going to leave this commented code here for the moment
	// removed so as to pass tests using httpmock
	//
	//req, err := http.NewRequest("GET", URL, nil)
	//if err != nil {
	//	log.Print(err)
	//	return false
	//}
	//req.Header.Set("X-Access-Token", x)
	//req.Header.Set("Content-Type", "application/json; charset=UTF-8")

	//skipVerify := false
	//if os.Getenv("ENVIRONMENT") == "DEV" {
	//	//log.Println("skipVerify set to true")
	//	skipVerify = true
	//}
	// skip verify to avoid x509 cert check if in dev env
	//tr := &http.Transport{
	//	TLSClientConfig: &tls.Config{InsecureSkipVerify: skipVerify},
	//}

	//client := &http.Client{Timeout: time.Second * 10, Transport: tr}
	//resp, e := client.Do(req)
	if e != nil {
		log.Print(fmt.Sprintf("The HTTP request failed with error %s", e))
		return false
	} else {
		defer resp.Body.Close()
		//TODO: check if auction finished and user won
		// when thingType is 'auction'
		if thingType == "auction" {
			log.Printf("Input UUID is [%s]", UUID)
		}
		//log.Printf("Response status code is [%d]",resp.StatusCode)
		if resp.StatusCode == 200 {
			return true
		}
	}
	return false

}
