package department

import "time"

type Department struct {
	ID                   string    `json:"id"`
	Name                 string    `json:"name"`
	Active               bool      `json:"active"`
	VolunteerCount       int       `json:"volunteer_count"`
	ActiveVolunteerCount int       `json:"active_volunteer_count"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

type ListOptions struct {
	Query  string
	Active *bool
	Limit  int
}
