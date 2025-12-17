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
	Seeker      string // "hh", "superjob", ...
	Description string
}

// Структура для определния результатов поиска списка вакансий по всем доступным парсерам
type SearchVacanciesResult struct {
	Vacancies  []Vacancy
	ParserName string
	SearchHash string
	Error      error
	Duration   time.Duration
}

type SearchVacancyDetailesResult struct {
	ParserName string
	VacancyID  string
	Error      error
	Duration   time.Duration
	Vacancy    Vacancy
}
