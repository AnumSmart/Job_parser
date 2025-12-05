package manager

import (
	"context"
	"fmt"
	"parser/internal/domain/models"
)

func (pm *ParsersManager) selectParsersForSearch() []string {
	// Сначала берем здоровые парсеры
	healthyParsers := pm.getHealthyParsers()
	if len(healthyParsers) > 0 {
		return healthyParsers
	}

	// Если все парсеры нездоровы, берем все
	fmt.Println("⚠️  Все парсеры в нездоровом состоянии, пробуем перезапуск...")
	return pm.getAllParsersNames()
}

func (pm *ParsersManager) searchWithParsers(ctx context.Context, params models.SearchParams, parserNames []string) ([]models.SearchResult, []error) {
	searchCtx, cancel := context.WithTimeout(ctx, pm.config.API.ConcSearchTimeout)
	defer cancel()

	executor := NewParserSearchExecutor(pm, parserNames, pm.config.Manager.MaxConcurrentParsers)
	return executor.Execute(searchCtx, params)
}
