package interfaces

import "parser/internal/model"

type Parser interface {
	SearchVacancies(params SearchParams) ([]model.Vacancy, error)
	GetVacancyByID(vacancyID string) (*model.Vacancy, error)
	SimpleSearch(query string, limit int) ([]model.Vacancy, error)
	GetName() string
}

type SearchParams struct {
	Text    string
	Area    string
	PerPage int
	Page    int
}
