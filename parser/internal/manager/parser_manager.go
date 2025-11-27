package manager

import (
	"parser/configs"
	"parser/internal/inmemory_cache"
	"parser/internal/interfaces"
)

type ParserManager struct {
	parsers      []interfaces.Parser
	config       *configs.Config
	searchCache  *inmemory_cache.InmemoryShardedCache
	vacancyIndex *inmemory_cache.InmemoryShardedCache
}

// Конструктор для мэнеджера парсинга из разных источников
func NewParserManager(config *configs.Config,
	searchCache *inmemory_cache.InmemoryShardedCache,
	vacancyIndex *inmemory_cache.InmemoryShardedCache,
	parsers ...interfaces.Parser) *ParserManager {
	return &ParserManager{
		parsers:      parsers,
		config:       config,
		searchCache:  searchCache,
		vacancyIndex: vacancyIndex,
	}
}

// GetAllParsers возвращает список доступных парсеров
func (pm *ParserManager) GetParserNames() []string {
	names := make([]string, len(pm.parsers))
	for i, parser := range pm.parsers {
		names[i] = parser.GetName()
	}
	return names
}
