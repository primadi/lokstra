package model

import "time"

type Order struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	Items     []string  `json:"items"`
	Total     float64   `json:"total"`
	CreatedAt time.Time `json:"created_at"`
}
