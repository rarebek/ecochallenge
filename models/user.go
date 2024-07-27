package models

type User struct {
	ID          int    `db:"id" json:"id"`
	FirstName   string `db:"first_name" json:"first_name"`
	LastName    string `db:"last_name" json:"last_name"`
	Avatar      string `db:"avatar" json:"avatar"`
	BirthDate   string `db:"birth_date" json:"birth_date"`
	Location    string `db:"location" json:"location"`
	PhoneNumber string `db:"phone_number" json:"phone_number"`
	XP          int    `db:"xp" json:"xp"`
}

type RankingResponse struct {
	ID       int    `json:"id"`
	Rank     int    `json:"rank"`
	UserName string `json:"user_name"`
	XP       int    `json:"xp"`
	Avatar   string `json:"avatar"`
}
