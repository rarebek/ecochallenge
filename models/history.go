package models

type History struct {
	ID        string `db:"id" json:"id"`
	UserID    int    `db:"user_id" json:"user_id"`
	EventID   string `db:"event_id" json:"event_id"`
	StartDate string `db:"start_date" json:"start_date"`
	EndDate   string `db:"end_date" json:"end_date"`
	XPEarned  int    `db:"xp_earned" json:"xp_earned"`
	CreatedAt string `db:"created_at" json:"created_at"`
	UpdatedAt string `db:"updated_at" json:"updated_at"`
}
