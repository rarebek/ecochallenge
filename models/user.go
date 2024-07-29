package models

import (
	"database/sql"
	"time"
)

type User struct {
	ID          int            `db:"id" json:"id"`
	FirstName   string         `db:"first_name" json:"first_name"`
	LastName    string         `db:"last_name" json:"last_name"`
	Avatar      string         `db:"avatar" json:"avatar"`
	BirthDate   sql.NullString `db:"birth_date" json:"birth_date"`
	Location    string         `db:"location" json:"location"`
	PhoneNumber string         `db:"phone_number" json:"phone_number"`
	XP          int            `db:"xp" json:"xp"`
}

type RankingResponse struct {
	ID       int    `json:"id"`
	Rank     int    `json:"rank"`
	UserName string `json:"user_name"`
	XP       int    `json:"xp"`
	Avatar   string `json:"avatar"`
	Location string `json:"location"`
}

type Market struct {
	ID           int64     `json:"id" db:"id"`
	Name         string    `json:"name" db:"name"`
	Description  string    `json:"description" db:"description"`
	Count        int64     `json:"count" db:"count"`
	XP           int64     `json:"xp" db:"xp"`
	CategoryName string    `json:"category_name" db:"category_name"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
	ImageUrl     string    `json:"image_url" db:"image_url"`
}

type Order struct {
	ID          int       `db:"id" json:"id"`
	UserID      int       `db:"user_id" json:"user_id"`
	ItemID      int       `db:"item_id" json:"item_id"`
	OrderNumber int       `db:"order_number" json:"order_number"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
}

type Message struct {
	Message string `json:"message"`
}

type EarnXP struct {
	Id           int    `json:"id"`
	Difficulty   string `json:"difficulty"`
	CorrectCount int    `json:"correct_count"`
}
