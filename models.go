package main

import (
	"database/sql"
	"errors"
	"fmt"
	"net/url"
)

type Review struct {
	ReviewId   string `json:"review_id"`
	Review     string `json:"review"`
	ReviewedBy string `json:"reviewed_by"`
	AuctionId  string `json:"auction_id"`
	ItemId     string `json:"item_id"`
	Seller     string `json:"seller"`
	Overall    int    `json:"overall"`
	PapCost    int    `json:"post_and_packaging"`
	Comm       int    `json:"communication"`
	AsDesc     int    `json:"as_described"`
	Created    string `json:"created"`
}

// ----------------------------------------------------------------------------

func (r *Review) validate() url.Values {

	errs := url.Values{}

	if r.Review == "" {
		errs.Add("Review", "The review field is required")
	}
	return errs
}

// ----------------------------------------------------------------------------

func (r *Review) getReview(db *sql.DB) error {
	return db.QueryRow("SELECT review,"+
		"reviewed_by,"+
		"auction_id,"+
		"item_id,"+
		"seller,"+
		"overall,"+
		"pap_cost,"+
		"communication,"+
		"as_described,"+
		"created FROM reviews WHERE review_id=$1",
		r.ReviewId).Scan(&r.Review,
		&r.ReviewedBy,
		&r.AuctionId,
		&r.ItemId,
		&r.Overall,
		&r.PapCost,
		&r.Comm,
		&r.AsDesc,
		&r.Created)
}

// ----------------------------------------------------------------------------

// not sure this should be allowed?
func (r *Review) updateReview(db *sql.DB) error {
	return errors.New("Not implemented")
}

// ----------------------------------------------------------------------------

func (r *Review) deleteReview(db *sql.DB) (int64, error) {

	res, err := db.Exec("DELETE FROM reviews WHERE "+
		"review_id=$1 AND reviewed_by=$2", r.ReviewId, r.ReviewedBy)
	rows, _ := res.RowsAffected()
	return rows, err
}

// ----------------------------------------------------------------------------

func (r *Review) createReview(db *sql.DB) error {

	err := db.QueryRow(
		"INSERT INTO reviews("+
			"review_id,"+
			"review,"+
			"reviewed_by,"+
			"auction_id,"+
			"item_id,"+
			"seller,"+
			"overall,"+
			"pap_cost,"+
			"communication,"+
			"as_described)"+
			"VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10) RETURNING review_id",
		r.ReviewId,
		r.Review,
		r.ReviewedBy,
		r.AuctionId,
		r.ItemId,
		r.Seller,
		r.Overall,
		r.PapCost,
		r.Comm,
		r.AsDesc).Scan(&r.ReviewId)

	if err != nil {
		return err
	}

	return nil
}

// ----------------------------------------------------------------------------

func getTotalReviews(db *sql.DB, input_type, input_id string) (count int, err error) {

	rows, err := db.Query(
		"SELECT COUNT(*) FROM reviews WHERE "+input_type+"=$1",
		input_id)

	if err != nil {
		return 0, err
	}

	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(&count); err != nil {
			return 0, err
		}
	}
	return count, nil

}

// ----------------------------------------------------------------------------

func getReviewsByInput(db *sql.DB, input_type, input_id string,
	start, count int) ([]Review, error) {

	rows, err := db.Query(
		"SELECT review_id,"+
			"review,"+
			"reviewed_by,"+
			"auction_id,"+
			"item_id,"+
			"seller,"+
			"overall,"+
			"pap_cost,"+
			"communication,"+
			"as_described,"+
			"created FROM reviews WHERE "+input_type+"=$1 "+
			"LIMIT $2 OFFSET $3",
		input_id, count, start)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	reviews := []Review{}

	for rows.Next() {
		var r Review
		if err := rows.Scan(&r.ReviewId, &r.Review, &r.ReviewedBy,
			&r.AuctionId, &r.ItemId, &r.Seller,
			&r.Overall, &r.PapCost, &r.Comm,
			&r.AsDesc, &r.Created); err != nil {
			return nil, err
		}
		reviews = append(reviews, r)
	}

	return reviews, nil

}

// ----------------------------------------------------------------------------
//TODO: refactor this into a more generic fetch

func (r *Review) getReviewByItem(db *sql.DB) error {

	return db.QueryRow("SELECT review,"+
		"review_id,"+
		"reviewed_by,"+
		"auction_id,"+
		"seller,"+
		"overall,"+
		"pap_cost,"+
		"communication,"+
		"as_described,"+
		"created FROM reviews WHERE item_id=$1",
		r.ItemId).Scan(&r.Review,
		&r.ReviewId,
		&r.ReviewedBy,
		&r.AuctionId,
		&r.Seller,
		&r.Overall,
		&r.PapCost,
		&r.Comm,
		&r.AsDesc,
		&r.Created)
}

// ----------------------------------------------------------------------------

func getScore(db *sql.DB, public_id string) (count int, err error) {

	fmt.Sprintf("public id is [%s]", public_id)

	return 89, nil
}
