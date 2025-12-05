// status_manager.go
package manager

import "parser/internal/interfaces"

func (pm *ParsersManager) updateParserStatus(name string, success bool, err error) {
	if pm.parsersStatusManager != nil {
		pm.parsersStatusManager.UpdateStatus(name, success, err)
	}
}

func (pm *ParsersManager) updateAllParsersStatus(success bool) {
	if pm.parsersStatusManager != nil {
		for _, name := range pm.getAllParsersNames() {
			pm.parsersStatusManager.UpdateStatus(name, success, nil)
		}
	}
}

func (pm *ParsersManager) getHealthyParsers() []string {
	if pm.parsersStatusManager != nil {
		return pm.parsersStatusManager.GetHealthyParsers()
	}
	return pm.getAllParsersNames()
}

func (pm *ParsersManager) getAllParsersNames() []string {
	names := make([]string, len(pm.parsers))
	for i, parser := range pm.parsers {
		names[i] = parser.GetName()
	}
	return names
}

func (pm *ParsersManager) findParserByName(name string) interfaces.Parser {
	for _, parser := range pm.parsers {
		if parser.GetName() == name {
			return parser
		}
	}
	return nil
}
