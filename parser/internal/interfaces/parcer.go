package interfaces

import (
	"context"
	"parser/internal/domain/models"
)

type Parser interface {
	SearchVacancies(ctx context.Context, params models.SearchParams) ([]models.Vacancy, error)
	GetName() string
}
