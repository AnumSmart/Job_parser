package manager

import (
	"bufio"
	"fmt"
	"parser/internal/model"
	"strings"
	"time"
)

func (pm *ParserManager) GetVacancyDetails(scanner *bufio.Scanner) {
	fmt.Println("\n📄 Детали вакансии")

	fmt.Print("Введите ID вакансии: ")
	if !scanner.Scan() {
		return
	}

	vacancyID := strings.TrimSpace(scanner.Text())
	if vacancyID == "" {
		fmt.Println("❌ ID вакансии не может быть пустым")
		return
	}

	fmt.Print("Введите источник (hh.ru/superjob.ru): ")
	if !scanner.Scan() {
		return
	}
	source := strings.TrimSpace(scanner.Text())

	compositeID := fmt.Sprintf("%s_%s", source, vacancyID)

	fmt.Println("⏳ Загружаем информацию...")

	// -------------------------------------------------------------------
	// тут должна быть логика поиска вакансии через составной обратный индекс
	// -------------------------------------------------------------------

	printVacancyDetails(vacancy)
}

func printVacancies(vacancies []model.HHVacancy) {
	if len(vacancies) == 0 {
		fmt.Println("😞 Вакансии не найдены")
		return
	}

	for i, vacancy := range vacancies {
		fmt.Printf("\n%d. %s\n", i+1, vacancy.Name)
		fmt.Printf("   💼 %s\n", vacancy.Employer.Name)
		fmt.Printf("   💰 %s\n", vacancy.GetSalaryString())
		fmt.Printf("   📍 %s\n", vacancy.Area.Name)
		//fmt.Printf("   🕐 %s\n", formatDate(vacancy.PublishedAt))
		fmt.Printf("   🔗 %s\n", vacancy.URL)
		fmt.Printf("   🆔 %s\n", vacancy.ID)
	}
}

func printVacancyDetails(vacancy *model.HHVacancy) {
	fmt.Println("\n" + strings.Repeat("=", 50))
	fmt.Printf("🏢 %s\n", vacancy.Name)
	fmt.Println(strings.Repeat("=", 50))
	fmt.Printf("💼 Работодатель: %s\n", vacancy.Employer.Name)
	fmt.Printf("💰 Зарплата: %s\n", vacancy.GetSalaryString())
	fmt.Printf("📍 Местоположение: %s\n", vacancy.Area.Name)
	//fmt.Printf("🕐 Опубликовано: %s\n", formatDate(vacancy.PublishedAt))
	fmt.Printf("🔗 Ссылка: %s\n", vacancy.URL)
	fmt.Printf("🆔 ID: %s\n", vacancy.ID)

	// Обрезаем описание для читаемости
	description := vacancy.Description
	if len(description) > 500 {
		description = description[:500] + "..."
	}

	if description != "" {
		fmt.Println("\n📝 Описание:")
		fmt.Println(cleanHTML(description))
	}

	fmt.Println(strings.Repeat("=", 50))
}

func formatDate(t time.Time) string {
	return t.Format("02.01.2006 15:04")
}

func cleanHTML(text string) string {
	// Простая очистка HTML тегов
	text = strings.ReplaceAll(text, "<p>", "\n")
	text = strings.ReplaceAll(text, "<br>", "\n")
	text = strings.ReplaceAll(text, "<li>", "• ")

	// Удаляем HTML теги
	var result strings.Builder
	var inTag bool

	for _, ch := range text {
		if ch == '<' {
			inTag = true
			continue
		}
		if ch == '>' {
			inTag = false
			continue
		}
		if !inTag {
			result.WriteRune(ch)
		}
	}

	return strings.TrimSpace(result.String())
}
