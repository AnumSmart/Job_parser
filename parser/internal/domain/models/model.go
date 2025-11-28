package models

import "time"

// общая структура поиска
type SearchParams struct {
	Text    string
	Area    string
	PerPage int
	Page    int
}

// Стуктура общей вакансии для всех ответов
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
	Description string
}

// Структура для определния результатов поиска
type SearchResult struct {
	ParserName string
	SearchHash string
	Vacancies  []Vacancy
	Error      error
	Duration   time.Duration
}
