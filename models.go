package main

import (
	"github.com/google/uuid"
	"time"
)

type Review struct {
	ReviewId   uuid.UUID `gorm:"type:uuid;primaryKey" json:"review_id"`
	Review     string    `gorm:"type:varchar(2000)" json:"review"`
	ReviewedBy uuid.UUID `gorm:"type:uuid;index" json:"reviewed_by" binding:"required"` // PublicId of reviewer
	AuctionId  uuid.UUID `gorm:"type:uuid;index" json:"auction_id" binding:"required"`
	ItemId     uuid.UUID `gorm:"type:uuid;index" json:"item_id" binding:"required"`
	Seller     uuid.UUID `gorm:"type:uuid;index" json:"seller" binding:"required"` // PublicId of seller
	Overall    int       `json:"overall" binding:"required"`
	PapCost    int       `json:"post_and_packaging" binding:"required"`
	Comm       int       `json:"communication" binding:"required"`
	AsDesc     int       `json:"as_described" binding:"required"`
	Created    time.Time `gorm:"autoCreateTime" json:"created"`
}

type ReviewsResponse struct {
	CurrentPage 	int 		`json:"current_page"`
	Reviews 		[]Review 	`json:"reviews"`
	TotalPages  	int			`json:"total_pages"`
	TotalReviews	int 		`json:"total_reviews"`
	URLS			[]URL		`json:"urls"`
}

type CreateReviewResp struct {
	ReviewId string `json:"review_id"`
}

type RespMessage struct {
	Message string `json:"message"`
}

type URL struct {
	PrevURL	string `json:"prev_url,omitempty"`
	NextURL string `json:"next_url,omitempty"`
}

type Item struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Category    string `json:"category"`
	//Yaf         string `json:"yaf"`
	PublicId    string `json:"public_id"`
	ItemId      string `json:"item_id"`
	Created     string `json:"created"`
	Modified    string `json:"modified"`
}

type Auction struct {
	AuctionId string `json:"auction_id"`
	PublicId  string `json:"public_id"`
	Lots      []string `json:"lots"`
	Type      string `json:"type"`
	Name 	  string `json:"name"`
	Multiple  bool   `json:"multiple"`
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
	Status    string `json:"status"`
	Active    bool	 `json:"active"`
	Created   string `json:"created"`
	Modified  string `json:"modified"`
	Currency  string `json:"currency"`
}

// ----------------------------------------------------------------------------

