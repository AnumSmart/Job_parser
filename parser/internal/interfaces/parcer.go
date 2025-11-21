package interfaces

import (
	"time"
)

type Parser interface {
	SearchVacancies(params SearchParams) ([]Vacancy, error)
	GetName() string
}

type SearchParams struct {
	Text    string
	Area    string
	PerPage int
	Page    int
}

type Vacancy struct {
	ID          string
	Job         string
	Company     string
	Salary      *string
	Currency    string
	Area        string
	Experience  string
	Schedule    string
	URL         string
	PublishedAt time.Time
	Seeker      string // "hh", "superjob"
}
