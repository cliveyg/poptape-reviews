package main

import (
	"database/sql"
	"errors"
	//"encoding/json"
)

type review struct {
	ReviewId string `json:"review_id"`
	Review   string `json:"review"`
	PublicId string `json:"public_id"`
	Overall  int    `json:"overall"`
	PapCost  int    `json:"post_and_packaging"`
	Comm     int    `json:"communication"`
	AsDesc   int    `json:"as_described"`
}

// ----------------------------------------------------------------------------

func (r *review) getReview(db *sql.DB) error {
	return db.QueryRow("SELECT review,"+
							   "public_id,"+
							   "overall,"+
							   "pap_cost,"+
							   "communication,"+
							   "as_described FROM reviews WHERE review_id=$1",
		r.ReviewId).Scan(&r.Review,
						 &r.PublicId,
						 &r.Overall,
						 &r.PapCost,
						 &r.Comm,
						 &r.AsDesc)
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
	return errors.New("Not implemented")
}

// ----------------------------------------------------------------------------

func getReviews(db *sql.DB, public_id string, start, count int) ([]review, error) {

	rows, err := db.Query(
		"SELECT review_id,"+
				"review,"+
                "public_id,"+
                "overall,"+
                "pap_cost,"+
                "communication,"+
                "as_described FROM reviews WHERE public_id=$1 "+
				"LIMIT $2 OFFSET $3",
		public_id, count, start)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	reviews := []review{}

	for rows.Next() {
		var r review
		if err := rows.Scan(&r.ReviewId, &r.Review, &r.PublicId, &r.Overall,
                            &r.PapCost, &r.Comm, &r.AsDesc); err != nil {
			return nil, err
		}
		reviews = append(reviews, r)
	}

	return reviews, nil

}
