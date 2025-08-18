package main

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
	"time"
)

type user struct {
	PublicId string `json:"public_id"`
}

func (a *App) bouncerSaysOk(c *gin.Context) (bool, int, string) {

	ct := c.GetHeader("Content-type")
	bm := "Ooh you are naughty"

	if !(ct == "application/json" ||
		ct == "application/json; charset=UTF-8") {
		bm = "Request must be json"
		return false, http.StatusBadRequest, bm
	}

	x := c.GetHeader("X-Access-Token")

	if x != "" {

		// call authy microservice
		req, err := http.NewRequest("GET", os.Getenv("AUTHYURL"), nil)
		if err != nil {
			a.Log.Info().Msgf("Error is [%s]", err.Error())
			return false, http.StatusUnauthorized, bm
		}

		req.Header.Set("X-Access-Token", x)
		req.Header.Set("Content-Type", "application/json; charset=UTF-8")

		client := &http.Client{Timeout: time.Second * 10}
		resp, e := client.Do(req)
		if e != nil {
			a.Log.Info().Msgf("HTTP req failed with [%s]", err.Error())
			bm = "I'm sorry Dave"
			return false, http.StatusServiceUnavailable, bm
		} else {
			defer resp.Body.Close()
			if resp.StatusCode == 200 {
				var u user
				if err := json.NewDecoder(resp.Body).Decode(&u); err != nil {
					a.Log.Info().Msgf("Error deserializing JSON [%s]", err.Error())
					return false, http.StatusBadRequest, "Unable to decode response body"
				}
				return true, http.StatusOK, u.PublicId
			}
		}
	}
	a.Log.Info().Msg("No x-access-token found")

	return false, http.StatusUnauthorized, bm
}
