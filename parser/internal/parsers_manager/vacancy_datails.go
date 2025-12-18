package parsers_manager

import (
	"bufio"
	"fmt"
	"parser/internal/domain/models"
	"strings"
	"time"
)

// метод получения информации о вакансии из кэша с помощью кэша обратного индекса
func (pm *ParsersManager) GetVacancyDetails(scanner *bufio.Scanner) error {
	fmt.Println("\n📄 Детали вакансии (кратко):")

	// получаем ID вакансии и имя источника из ввода
	source, vacancyID, err := pm.getCompositeIDFromInput(scanner)
	if err != nil {
		return err
	}

	// создаём составной индекс, в котором будет ID вакансии и сервис, в котором этот ID нужно будет искать
	// этот составной индекс - будет ключем для кэша №2
	compositeID := fmt.Sprintf("%s_%s", source, vacancyID)

	// создаём переменную для искомой вакансии
	var targetVacancy models.VacancyDetails

	fmt.Println("⏳ Загружаем информацию...")

	// -------------------------------------------------------------------
	// пытаемся найти в кэше №2 данные по заданному ключу (составному индексу)
	searchResIndex, ok := pm.vacancyIndex.GetItem(compositeID)
	if !ok {
		return fmt.Errorf("No Vacancy with ID:%s was found in cache\n", vacancyID)
	}

	// проводим type assertion, проверяем нужный тип (так как нам функция GetItem возвращает интерфейс)
	searchResIndexChecked, ok := searchResIndex.(models.VacancyIndex)
	if !ok {
		fmt.Println("Type assertion after GetVacancyDetails ---> failed!")
		return fmt.Errorf("Type assertion after GetVacancyDetails ---> failed!\n")
	}

	// теперь из полученного из кэша индексов индекса мы можем найти нужный хэш запроса,
	// чтобы потом по этому хэшу из кэша поиска найти нужную вакансию по ID

	// пытаемся найти в кэше данные по заданному хэш ключу
	searchRes, ok := pm.searchCache.GetItem(searchResIndexChecked.SearchHash)
	if ok {
		// если можно получить данные из кэша, то получаем интерфейс.
		// проводим type assertion, проверяем нужный тип
		searchResChecked, ok := searchRes.([]models.SearchVacanciesResult)
		if !ok {
			return fmt.Errorf("Type assertion after multi-search ---> failed!\n")
		}

		for _, neededElementRes := range searchResChecked {
			if neededElementRes.ParserName == source {
				for _, vacancyRes := range neededElementRes.Vacancies {
					if vacancyRes.ID == vacancyID {
						targetVacancy.ID = vacancyRes.ID
						targetVacancy.Job = vacancyRes.Job
						targetVacancy.Salary = vacancyRes.Salary
						targetVacancy.Company = vacancyRes.Company
						targetVacancy.Area = vacancyRes.Area
						targetVacancy.URL = vacancyRes.URL
					}
				}
			}
		}
	} else {
		pm.vacancyIndex.DeleteItem(compositeID)
		return fmt.Errorf("Search data --- expired!\n")
	}

	printVacancyDetails(targetVacancy, "нужно выбрать в меню --- полное описание вакансии по ID")

	return nil
}

// метод для получения полной информации по отдельной вакансии по ID
func (pm *ParsersManager) GetFullVacancyDetails(scanner *bufio.Scanner) error {
	// получаем ID вакансии и имя источника из ввода
	_, vacancyID, err := pm.getCompositeIDFromInput(scanner)
	if err != nil {
		return err
	}

	// пытаемся найти информацию по ID вакансии в кэше для деталей отдельной вакансии
	searchResVacDet, exists := pm.vacancyDetails.GetItem(vacancyID)
	if exists {
		searchResVacDetChecked, ok := searchResVacDet.(models.VacancyDetails)
		if !ok {
			fmt.Println("Type assertion after GetVacancyDetails from cache ---> failed!")
			return fmt.Errorf("Type assertion after GetVacancyDetails from cache ---> failed!\n")
		}
		printVacancyDetails(searchResVacDetChecked, "")
	}

	// если нет данных в кэше информации по вавкансиям, то необходимо сделать новый запрос на нужный сервис с конкретным ID
	//---------------------------------------------------------------------------
	// тут необходимо создать джобу, которая будет удовлетворять интерфейсу, еперадть её в очередь, создать канал и из этого канала попытаться прочитать данные

	//---------------------------------------------------------------------------

	return fmt.Errorf("No Vacancy with ID:%s was found in vacancy details cache\n", vacancyID)
}

/*
// метод осуществляет поиск деталей вакансии в конкретном сервисе по конкретному ID
func (pm *ParsersManager) executeSearchVacancyDetailes(ctx context.Context, vacancyID, source string) (models.SearchVacancyDetailesResult, error) {
	// -----------------------------------пока в разработке----------------------------------------------
	return models.SearchVacancyDetailesResult{}, nil
}
*/

// метод получения имени источника и ID вакансии из ввода
func (pm *ParsersManager) getCompositeIDFromInput(scanner *bufio.Scanner) (string, string, error) {
	fmt.Print("Введите ID вакансии: ")
	if !scanner.Scan() {
		return "", "", fmt.Errorf("❌ Проблема со сканированием ввода\n")
	}

	// переменная, куда сохранаяется ID искомой вакансии
	vacancyID := strings.TrimSpace(scanner.Text())
	if vacancyID == "" {
		//fmt.Println("❌ ID вакансии не может быть пустым")
		return "", "", fmt.Errorf("❌ ID вакансии не может быть пустым\n")
	}

	fmt.Print("Введите источник (HH.ru/SuperJob.ru): ")
	if !scanner.Scan() {
		return "", "", fmt.Errorf("❌ ввели неверное имя сервиса\n")
	}
	// переменная, куда кладём имя сервиса, в результатах поиска которого будем искать ID вакансии
	source := strings.TrimSpace(scanner.Text())

	return source, vacancyID, nil
}

// функция вывода в консоль данных о найденой вакансии
func printVacancyDetails(vacancy models.VacancyDetails, description string) {
	defer func() {
		if rec := recover(); rec != nil {
			fmt.Println("recovered from PANIC: [ ", rec, " ]")
		}
	}()

	fmt.Println("\n" + strings.Repeat("=", 50))
	fmt.Printf("🏢 %s\n", vacancy.Job)
	fmt.Println(strings.Repeat("=", 50))
	fmt.Printf("💼 Работодатель: %s\n", vacancy.Company)
	fmt.Printf("💰 Зарплата: %s\n", *vacancy.Salary)
	fmt.Printf("📍 Местоположение: %s\n", vacancy.Area)
	//fmt.Printf("🕐 Опубликовано: %s\n", formatDate(vacancy.PublishedAt))
	fmt.Printf("🔗 Ссылка: %s\n", vacancy.URL)
	fmt.Printf("🆔 ID: %s\n", vacancy.ID)

	// Обрезаем описание для читаемости
	if len(description) > 1000 {
		description = description[:1000] + "..."
	}

	fmt.Printf("📝 Описание: %s\n", description)

	/*
		if description != "" {
			fmt.Println("\n📝 Описание:")
			//fmt.Println(cleanHTML(description))
			fmt.Println(description)
		}
	*/

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
