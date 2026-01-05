// реализация консольного приложения
package cli

import (
	"bufio"
	"fmt"
	"os"
	"parser/configs"
	"parser/internal/parsers_manager"
	"strings"
)

// CLIApp содержит зависимости для CLI режима
type CLIApp struct {
	config        *configs.Config
	parserManager *parsers_manager.ParsersManager
	scanner       *bufio.Scanner
}

// NewCLIApp создает экземпляр CLI приложения
func NewCLIApp(config *configs.Config, parserManager *parsers_manager.ParsersManager) *CLIApp {
	return &CLIApp{
		config:        config,
		parserManager: parserManager,
		scanner:       bufio.NewScanner(os.Stdin),
	}
}

// Run запускает CLI приложение
func (a *CLIApp) Run() error {
	fmt.Println("🚀 Multi-Source Vacancy Parser (CLI) запущен!")
	fmt.Println("==========================")

	for {
		a.printMenu()
		fmt.Print("Выберите действие: ")

		if !a.scanner.Scan() {
			break
		}

		choice := strings.TrimSpace(a.scanner.Text())

		if err := a.handleChoice(choice); err != nil {
			if err == ErrExitRequested {
				return nil
			}
			fmt.Printf("Ошибка: %v\n", err)
		}
		fmt.Println()
	}

	return nil
}

// метод - для распечатки меню
func (a *CLIApp) printMenu() {
	fmt.Println("📋 Меню:")
	fmt.Println("1. Поиск вакансий (расширенный)")
	fmt.Println("2. Получить описание вакансии по ID")
	fmt.Println("3. Получить полное описание вакансии по ID")
	fmt.Println("4. Выход")
}

// метод - выбора действия
func (a *CLIApp) handleChoice(choice string) error {
	switch choice {
	case "1":
		return a.parserManager.MultiSearch(a.scanner)
	case "2":
		return a.parserManager.GetVacancyDetails(a.scanner)
	case "3":
		return a.parserManager.GetFullVacancyDetails(a.scanner)
	case "4":
		a.parserManager.Shutdown()
		fmt.Println("👋 До свидания!")
		return ErrExitRequested
	default:
		fmt.Println("❌ Неверный выбор. Попробуйте снова.")
	}
	return nil
}

var ErrExitRequested = fmt.Errorf("exit requested")
