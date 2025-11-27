package manager

import (
	"parser/configs"
	"parser/internal/interfaces"
)

type ParserManager struct {
	parsers []interfaces.Parser
	config  *configs.Config
}

// Конструктор для мэнеджера парсинга из разных источников
func NewParserManager(config *configs.Config, parsers ...interfaces.Parser) *ParserManager {
	return &ParserManager{
		parsers: parsers,
		config:  config,
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
