package main

import (
	"encoding/json"
	"github.com/go-faker/faker/v4"
)

func (a *App) InsertFakedDummyReviews(numRevs int) ([]Review, error) {

	reviews := make([]Review, numRevs)
	err := faker.FakeData(&reviews)
	if err != nil {
		a.Log.Info().Msgf("Error faking data [%s]", err.Error())
		return nil, err
	}
	res := a.DB.Create(&reviews)
	if res.Error != nil {
		a.Log.Info().Msgf("Reviews creation failed: [%s]", err.Error())
		return nil, res.Error
	}

	return reviews, nil
}

func (a *App) InsertSpecificDummyReviews() ([]Review, error) {

	jsonData := `[
		{
			"review_id": "e8f48256-2460-418f-81b7-86dad2aa6e41",
			"review": "amaze balls product",
			"reviewed_by": "f38ba39a-3682-4803-a498-659f0bf05000",
			"auction_id": "e77be9e0-bb00-49bc-9e7d-d7cc7072ab8c",
			"item_id": "387bfbb4-36cf-44c9-8e05-83b2ca72cdff",
			"seller": "46d7d11c-fa06-4e54-8208-95433b98cfc9",
			"overall": 5,
			"post_and_packaging": 4,
			"communication": 4,
			"as_described": 3
		},
		{
			"review_id": "e8f48256-2460-418f-81b7-86dad2aa6aaa",
			"review": "awesome balls product",
			"reviewed_by": "f38ba39a-3682-4803-a498-659f0bf05304",
			"auction_id": "e77be9e0-bb00-49bc-9e7d-d7cc7072ab8c",
			"item_id": "7d1aa876-9be8-441f-ad86-d86e5faddd81",
			"seller": "46d7d11c-fa06-4e54-8208-95433b98cfc9",
			"overall": 5,
			"post_and_packaging": 4,
			"communication": 4,
			"as_described": 3
		},
		{
			"review_id": "e8f48256-2460-418f-81b7-86dad2aa6111",
			"review": "superduper ting",
			"reviewed_by": "f38ba39a-3682-4803-a498-659f0bf05304",
			"auction_id": "e77be9e0-bb00-49bc-9e7d-d7cc7072ab11",
			"item_id": "7d1aa876-9be8-441f-ad86-d8e5fade5441",
			"seller": "46d7d11c-fa06-4e54-8208-954322222222",
			"overall": 4,
			"post_and_packaging": 4,
			"communication": 7,
			"as_described": 8
		},
		{
			"review_id": "e8f48256-2460-418f-81b7-86dad2aa6222",
			"review": "changed my life",
			"reviewed_by": "f38ba39a-3682-4803-a498-659f0bf05304",
			"auction_id": "e77be9e0-bb00-49bc-9e7d-d7cc7072ab22",
			"item_id": "aabbccd6-9be8-441f-ad86-d86e5faddd81",
			"seller": "46d7d11c-fa06-4e54-8208-95433b98cfc9",
			"overall": 10,
			"post_and_packaging": 9,
			"communication": 8,
			"as_described": 9
		},
		{
			"review_id": "e8f48256-2460-418f-81b7-86dad2aa6333",
			"review": "Lorem ipsum dolor sit amet consectetur adipiscing elit. Quisque faucibus ex sapien vitae pellentesque sem placerat.tellus duis convallis. Tempus leo eu aenean sed diam urna tempor.",
			"reviewed_by": "f38ba39a-3682-4803-a498-659f0bf05000",
			"auction_id": "e77be9e0-bb00-49bc-9e7d-d7cc7072ab33",
			"item_id": "7d1aa876-9be8-441f-ad86-daaa51872333",
			"seller": "46d7d11c-fa06-4e54-8208-aaaaaaaa8888",
			"overall": 2,
			"post_and_packaging": 2,
			"communication": 6,
			"as_described": 1
		},
		{
			"review_id": "e8f48256-2460-418f-81b7-86dad2aa6002",
			"review": "changed my life - for the worse",
			"reviewed_by": "f38ba39a-3682-4803-a498-659f0bf05304",
			"auction_id": "e77be9e0-bb00-49bc-9e7d-d7cc7072cc22",
			"item_id": "aabbccd6-9be8-441f-ad86-d86e5fad7878",
			"seller": "aaaaaaaa-fa06-4e54-8208-95433b98cfc9",
			"overall": 2,
			"post_and_packaging": 8,
			"communication": 8,
			"as_described": 1
		}
	]`

	var reviews []Review
	err := json.Unmarshal([]byte(jsonData), &reviews)
	if err != nil {
		return nil, err
	}

	res := a.DB.Create(&reviews)
	if res.Error != nil {
		a.Log.Info().Msgf("Reviews creation failed: [%s]", err.Error())
		return nil, res.Error
	}

	return reviews, nil
}

const createJson = `{"auction_id":"f38ba39a-3682-4803-a498-659f0b111111",
"item_id": "f80689a6-9fba-4859-bdde-0a307c696ea8",
"reviewed_by": "f38ba39a-3682-4803-a498-659f0bf05304",
"seller": "4a48341f-bcef-4362-9d80-24a4960507ea",
"review": "amazing product",
"overall": 4,
"post_and_packaging": 3,
"communication": 4,
"as_described": 4}`