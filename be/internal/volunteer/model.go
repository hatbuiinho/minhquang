package volunteer

import "time"

type Status string

const (
	StatusActive   Status = "active"
	StatusDeparted Status = "departed"
)

type Volunteer struct {
	ID               string     `json:"id"`
	FullName         string     `json:"full_name"`
	DharmaName       string     `json:"dharma_name"`
	BirthDate        string     `json:"birth_date"`
	CultivationPlace string     `json:"cultivation_place"`
	Phone            string     `json:"phone"`
	DepartmentID     string     `json:"department_id,omitempty"`
	Department       string     `json:"department"`
	Notes            string     `json:"notes"`
	AvatarURL        string     `json:"avatar_url"`
	ArrivalDate      time.Time  `json:"arrival_date"`
	DepartureDate    *time.Time `json:"departure_date,omitempty"`
	Status           Status     `json:"status"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

type Input struct {
	FullName         string
	DharmaName       string
	BirthDate        string
	CultivationPlace string
	Phone            string
	Department       string
	Notes            string
	AvatarURL        string
	ArrivalDate      time.Time
	DepartureDate    *time.Time
}

type ListOptions struct {
	Query         string
	Status        string
	DepartmentID  string
	Today         time.Time
	Limit         int
	Offset        int
	SortBy        string
	SortDirection string
}

type BulkPatch struct {
	Field        string
	TextValue    string
	DateValue    *time.Time
	DepartmentID string
	Department   string
	UpdatedAt    time.Time
}
