package models

type VacancyDetails struct {
	ID          string
	Job         string
	Company     string
	Salary      *string
	Currency    string
	Area        string
	Experience  string
	Schedule    string
	URL         string
	Seeker      string // "hh", "superjob", ...
	Description string
}
