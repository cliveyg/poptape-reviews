package main

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"io"
	"math"
	"net/http"
	"os"
	"strconv"
	"sync"
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

func Paginate(page, pagesize int) func(db *gorm.DB) *gorm.DB {
	return func (db *gorm.DB) *gorm.DB {

		offset := (page - 1) * pagesize
		return db.Offset(offset).Limit(pagesize)
	}
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

