package interfaces

import "parser/internal/model"

type Parser interface {
	SearchVacancies(params SearchParams) ([]model.HHVacancy, error)
	GetVacancyByID(vacancyID string) (*model.HHVacancy, error)
	SimpleSearch(query string, limit int) ([]model.HHVacancy, error)
	GetName() string
}

type SearchParams struct {
	Text    string
	Area    string
	PerPage int
	Page    int
}
