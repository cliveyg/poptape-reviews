package main

import (
	"database/sql"
	"errors"
	//"encoding/json"
)

type review struct {
	ReviewId  string `json:"review_id"`
	Review    string `json:"review"`
	PublicId  string `json:"public_id"`
	AuctionId string `json:"auction_id"`
	Overall   int    `json:"overall"`
	PapCost   int    `json:"post_and_packaging"`
	Comm      int    `json:"communication"`
	AsDesc    int    `json:"as_described"`
	Created   string `json:"created"`
}

// ----------------------------------------------------------------------------

func (r *review) getReview(db *sql.DB) error {
	return db.QueryRow("SELECT review,"+
							   "public_id,"+
							   "auction_id,"+
							   "overall,"+
							   "pap_cost,"+
							   "communication,"+
							   "as_described,"+
							   "created FROM reviews WHERE review_id=$1",
		r.ReviewId).Scan(&r.Review,
						 &r.PublicId,
						 &r.AuctionId,
						 &r.Overall,
						 &r.PapCost,
						 &r.Comm,
						 &r.AsDesc,
						 &r.Created)
}

// ----------------------------------------------------------------------------

// not sure this should be allowed?
func (r *review) updateReview(db *sql.DB) error {
	return errors.New("Not implemented")
}

// ----------------------------------------------------------------------------

func (r *review) deleteReview(db *sql.DB) error {

	_, err := db.Exec("DELETE FROM reviews WHERE review_id=$1", r.ReviewId)

	return err
}

// ----------------------------------------------------------------------------

func (r *review) createReview(db *sql.DB) error {

	err := db.QueryRow(
		"INSERT INTO reviews("+
				"review_id,"+
                "review,"+
                "public_id,"+
                "auction_id,"+
                "overall,"+
                "pap_cost,"+
                "communication,"+
                "as_described)"+
		"VALUES($1,$2,$3,$4,$5,$6,$7,$8) RETURNING review_id",
		r.ReviewId,
		r.Review,
        r.PublicId,
        r.AuctionId,
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

func getReviews(db *sql.DB, public_id string, start, count int) ([]review, error) {

	rows, err := db.Query(
		"SELECT review_id,"+
				"review,"+
                "public_id,"+
				"auction_id,"+
                "overall,"+
                "pap_cost,"+
                "communication,"+
                "as_described,"+
				"created FROM reviews WHERE public_id=$1 "+
				"LIMIT $2 OFFSET $3",
		public_id, count, start)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	reviews := []review{}

	for rows.Next() {
		var r review
		if err := rows.Scan(&r.ReviewId, &r.Review, &r.PublicId,
							&r.AuctionId, &r.Overall, &r.PapCost,
							&r.Comm, &r.AsDesc, &r.Created); err != nil {
			return nil, err
		}
		reviews = append(reviews, r)
	}

	return reviews, nil

}

// ----------------------------------------------------------------------------

func getReviewsByAuction(db *sql.DB, auction_id string, start, count int) ([]review, error) {

    rows, err := db.Query(
        "SELECT review_id,"+
                "review,"+
                "public_id,"+
                "auction_id,"+
                "overall,"+
                "pap_cost,"+
                "communication,"+
                "as_described,"+
                "created FROM reviews WHERE auction_id=$1 "+
                "LIMIT $2 OFFSET $3",
        auction_id, count, start)

    if err != nil {
        return nil, err
    }

    defer rows.Close()

    reviews := []review{}

    for rows.Next() {
        var r review
        if err := rows.Scan(&r.ReviewId, &r.Review, &r.PublicId,
                            &r.AuctionId, &r.Overall, &r.PapCost,
                            &r.Comm, &r.AsDesc, &r.Created); err != nil {
            return nil, err
        }
        reviews = append(reviews, r)
    }

    return reviews, nil

}

// ----------------------------------------------------------------------------
