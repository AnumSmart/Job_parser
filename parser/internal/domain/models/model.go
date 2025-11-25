package models

import "time"

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

// Структура для определния результатов поиска
type SearchResult struct {
	ParserName string
	Vacancies  []Vacancy
	Error      error
	Duration   time.Duration
}
