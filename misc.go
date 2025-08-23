package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"io"
	"math"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

// ----------------------------------------------------------------------------
// h e l p e r   f u n c t i o n s
// ----------------------------------------------------------------------------

func CreateURLS(c *gin.Context, urls *[]interface{}, page, pagesize, totalPages *int, tc *int64) error {
	var prev string
	var next string
	*totalPages = int(math.Ceil(float64(*tc)/float64(*pagesize)))
	prev = `{ "prev_url": "`+os.Getenv("PREVNEXTURL")+c.Request.URL.Path+`?page=`+strconv.Itoa(*page-1)+`" }`
	next = `{ "next_url": "`+os.Getenv("PREVNEXTURL")+c.Request.URL.Path+`?page=`+strconv.Itoa(*page+1)+`" }`

	var prevobj map[string]interface{}
	err := json.Unmarshal([]byte(prev), &prevobj)
	if err != nil {
		return err
	}
	var nextobj map[string]interface{}
	err = json.Unmarshal([]byte(next), &nextobj)
	if err != nil {
		return err
	}

	if *page > 1 && *page < *totalPages {
		// we have prev and next
		*urls = append(*urls, prevobj)
		*urls = append(*urls, nextobj)
	} else if *page > 1 {
		// only prev
		*urls = append(*urls, prevobj)
	} else if *page == *totalPages {
		// no prev or next
	} else if *page < *totalPages {
		// only next
		*urls = append(*urls, nextobj)
	}
	return nil
}

// ----------------------------------------------------------------------------

func Paginate(page, pagesize int) func(db *gorm.DB) *gorm.DB {
	return func (db *gorm.DB) *gorm.DB {

		offset := (page - 1) * pagesize
		return db.Offset(offset).Limit(pagesize)
	}
}

// ----------------------------------------------------------------------------

func checkRequest(c *gin.Context) (bool, int, string) {

	ct := c.GetHeader("Content-type")

	if !(ct == "application/json" ||
		ct == "application/json; charset=UTF-8") {
		return false, http.StatusBadRequest, "Request must be json"
	}
	return true, http.StatusOK, ""
}

// ----------------------------------------------------------------------------

func (a *App) checkUserExists(c *gin.Context) (error, int) {

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		a.Log.Info().Msgf("Not a uuid string: [%s]", err.Error())
		return err, 400
	}

	var req *http.Request
	req, err = http.NewRequest("GET", os.Getenv("AUTHYUSER")+id.String(), nil)
	if err != nil {
		a.Log.Info().Msgf("Error is [%s]", err.Error())
		return err, 400
	}
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	client := &http.Client{Timeout: time.Second * 10}
	resp, e := client.Do(req)
	if e != nil {
		a.Log.Info().Msgf("HTTP req failed with [%s]", e.Error())
		return e, 400
	}
	if resp.StatusCode == 200 {
		return nil, resp.StatusCode
	}

	mess := fmt.Sprintf("Error fetching username. Status code is [%d]", resp.StatusCode)
	return errors.New(mess), resp.StatusCode
}

// ----------------------------------------------------------------------------

// HTTPRequest describes a single HTTP request with headers and a destination object for the response.
type HTTPRequest struct {
	URL     string
	Headers map[string]string
	Result  interface{} // pointer to the struct to unmarshal into
}

// HTTPResponse contains the HTTP status code and any error.
type HTTPResponse struct {
	StatusCode int
	Err        error
}

// ----------------------------------------------------------------------------

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
			defer func(Body io.ReadCloser) {
				err = Body.Close()
				if err != nil {
					responses[idx] = HTTPResponse{Err: err}
					return
				}
			}(resp.Body)
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

// ----------------------------------------------------------------------------

func getScore(publicId uuid.UUID) (count int, err error) {

	res := fmt.Sprintf("public id is [%s]", publicId.String())
	print(res)

	return 89, nil
}

