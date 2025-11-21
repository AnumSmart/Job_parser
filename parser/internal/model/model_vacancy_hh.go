package model

import (
	"parser/pkg"
)

// HHVacancy представляет структуру вакансии с HH.ru
type HHVacancy struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Salary      Salary   `json:"salary"`
	Employer    Employer `json:"employer"`
	Area        Area     `json:"area"`
	URL         string   `json:"url"`
	Description string   `json:"description"`
}

// GetSalaryString возвращает форматированную строку зарплаты
func (v HHVacancy) GetSalaryString() string {
	if v.Salary.From == 0 && v.Salary.To == 0 {
		return "не указана"
	}

	if v.Salary.From > 0 && v.Salary.To > 0 {
		return pkg.FormatSalary(v.Salary.From, v.Salary.To, v.Salary.Currency)
	} else if v.Salary.From > 0 {
		return pkg.FormatSalary(v.Salary.From, 0, v.Salary.Currency)
	} else {
		return pkg.FormatSalary(0, v.Salary.To, v.Salary.Currency)
	}
}
