package main

import (
	"database/sql"
	"errors"
)

type review struct {
	reviewId string json:"review_id"
	review   string json:"review"
	publicId string json:"public_id"
	overall  int    json:"overall"
	papCost  int    json:"post_and_packaging"
	comm     int    json:"communication"
}

func (r *review) getReview(db *sql.DB) error {
  return errors.New("Not implemented")
}

// not sure this should be allowed?
func (r *review) updateReview(db *sql.DB) error {
  return errors.New("Not implemented")
}

func (r *review) deleteReview(db *sql.DB) error {
  return errors.New("Not implemented")
}

func (r *review) createReview(db *sql.DB) error {
  return errors.New("Not implemented")
}

func getReviews(db *sql.DB, start, count int) ([]review, error) {
  return nil, errors.New("Not implemented")
}
