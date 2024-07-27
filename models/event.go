package models

type Event struct {
	ID               string `db:"id" json:"id"`
	Name             string `db:"name" json:"name"`
	Image            string `db:"image" json:"image"`
	Description      string `db:"description" json:"description"`
	TotalXP          int    `db:"total_xp" json:"total_xp"`
	StartDate        string `db:"start_date" json:"start_date"`
	EndDate          string `db:"end_date" json:"end_date"`
	RespOfficer      string `db:"resp_officer" json:"resp_officer"`
	RespOfficerImage string `db:"resp_officer_image" json:"resp_officer_image"`
	CreatedAt        string `db:"created_at" json:"created_at"`
	UpdatedAt        string `db:"updated_at" json:"updated_at"`
}
