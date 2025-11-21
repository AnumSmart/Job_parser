package parser

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"parser/internal/interfaces"
	"parser/internal/model"
	"strconv"
	"time"
)

type SuperJobParser struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

func NewSuperJobParser(apiKey string) *SuperJobParser {
	return &SuperJobParser{
		baseURL: "https://api.superjob.ru/2.0/vacancies/",
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (p *SuperJobParser) GetName() string {
	return "SuperJob"
}

func (p *SuperJobParser) SearchVacancies(params interfaces.SearchParams) ([]model.HHVacancy, error) {
	apiURL, err := p.buildURL(params)
	if err != nil {
		return nil, fmt.Errorf("build URL failed: %w", err)
	}

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, err
	}

	// Добавляем заголовки для SuperJob API
	req.Header.Add("X-Api-App-Id", p.apiKey)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response failed: %w", err)
	}

	var sjResponse model.SuperJobResponse
	if err := json.Unmarshal(body, &sjResponse); err != nil {
		return nil, fmt.Errorf("parse JSON failed: %w", err)
	}

	return p.convertToVacancies(sjResponse.Items), nil
}

func (p *SuperJobParser) buildURL(params interfaces.SearchParams) (string, error) {
	u, err := url.Parse(p.baseURL)
	if err != nil {
		return "", err
	}

	query := u.Query()

	if params.Text != "" {
		query.Set("keyword", params.Text)
	}
	if params.Area != "" {
		query.Set("town", p.convertArea(params.Area))
	}
	if params.PerPage > 0 {
		query.Set("count", strconv.Itoa(params.PerPage))
	}
	if params.Page > 0 {
		query.Set("page", strconv.Itoa(params.Page-1)) // SuperJob использует 0-based
	}

	u.RawQuery = query.Encode()
	return u.String(), nil
}

func (p *SuperJobParser) convertArea(area string) string {
	// Конвертируем коды регионов HH.ru в названия SuperJob
	areas := map[string]string{
		"1": "Москва",
		"2": "Санкт-Петербург",
	}
	if name, ok := areas[area]; ok {
		return name
	}
	return ""
}

func (p *SuperJobParser) convertToVacancies(sjVacancies []model.SuperJobVacancy) []model.HHVacancy {
	vacancies := make([]model.HHVacancy, len(sjVacancies))
	for i, sjv := range sjVacancies {
		vacancies[i] = model.HHVacancy{
			ID:   strconv.Itoa(sjv.ID),
			Name: sjv.Profession,
			Salary: model.Salary{
				From:     sjv.PaymentFrom,
				To:       sjv.PaymentTo,
				Currency: sjv.Currency,
			},
			Employer: model.Employer{
				Name: sjv.FirmName,
			},
			Area: model.Area{
				Name: sjv.Town.Title,
			},
			URL:         sjv.Link,
			Description: sjv.VacancyRichText,
		}
	}
	return vacancies
}

func (p *SuperJobParser) GetVacancyByID(vacancyID string) (*model.HHVacancy, error) {
	// Реализация получения деталей вакансии по ID
	// Аналогично HH Parser
	return nil, nil
}

func (p *SuperJobParser) SimpleSearch(query string, limit int) ([]model.HHVacancy, error) {
	params := interfaces.SearchParams{
		Text:    query,
		PerPage: limit,
	}
	return p.SearchVacancies(params)
}
