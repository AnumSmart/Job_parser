package interfaces

import (
	"job_parser/internal/domain/models"
)

type Parser interface {
	SearchVacancies(params models.SearchParams) ([]models.Vacancy, error)
	GetName() string
}
