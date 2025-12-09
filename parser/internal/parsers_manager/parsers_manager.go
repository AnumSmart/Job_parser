// описание структуры мэнеджера парсеров и его конструктора
package parsers_manager

import (
	"parser/configs"
	"parser/internal/circuitbreaker"
	"parser/internal/inmemory_cache"
	"parser/internal/interfaces"
	"parser/internal/parsers_status_manager"
)

type ParsersManager struct {
	parsers              []interfaces.Parser                         // парсеры, которыми оперирует мэнеджер
	config               *configs.Config                             // общий конфиг
	searchCache          *inmemory_cache.InmemoryShardedCache        // поисковый кэш
	vacancyIndex         *inmemory_cache.InmemoryShardedCache        // кэш для обратного индекса
	parsersStatusManager *parsers_status_manager.ParserStatusManager // менеджер сотсояний парверов внутри менеджера
	circuitBreaker       interfaces.CBInterface                      // глобальный circut breaker
}

// Конструктор для мэнеджера парсинга из разных источников
func NewParserManager(config *configs.Config,
	searchCache *inmemory_cache.InmemoryShardedCache,
	vacancyIndex *inmemory_cache.InmemoryShardedCache,
	pStatManager *parsers_status_manager.ParserStatusManager,
	parsers ...interfaces.Parser) *ParsersManager {
	return &ParsersManager{
		parsers:              parsers,
		config:               config,
		searchCache:          searchCache,
		vacancyIndex:         vacancyIndex,
		parsersStatusManager: pStatManager,
		circuitBreaker:       circuitbreaker.NewCircutBreaker(config.Manager.CircuitBreakerCfg),
	}
}

// GetAllParsers возвращает список доступных парсеров
func (pm *ParsersManager) GetParserNames() []string {
	names := make([]string, len(pm.parsers))
	for i, parser := range pm.parsers {
		names[i] = parser.GetName()
	}
	return names
}
