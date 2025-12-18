// описываем конструкторы для разных типов джоб
package parsers_manager

import (
	"context"
	"parser/internal/domain/models"
	"parser/internal/interfaces"
	"parser/internal/jobs"
	"parser/pkg"
	"time"
)

// NewSearchJob - создает джобу для поиска вакансий
func (pm *ParsersManager) NewSearchJob(params models.SearchParams) *jobs.SearchJob {
	return &jobs.SearchJob{
		BaseJob: jobs.BaseJob{
			ID:         pkg.QuickUUID(),
			ResultChan: make(chan *jobs.JobOutput, 1), // обязательно - буферизированный канал
			CreatedAt:  time.Now(),
		},
		Params: params,
	}
}

// NewFetchVacancyJob - создает джобу для получения деталей вакансии
func (pm *ParsersManager) NewFetchVacancyJob(source, vacancyID string) *jobs.FetchDetailsJob {
	return &jobs.FetchDetailsJob{
		BaseJob: jobs.BaseJob{
			ID:         pkg.QuickUUID(),
			ResultChan: make(chan *jobs.JobOutput, 1), // обязательно - буферизированный канал
			CreatedAt:  time.Now(),
		},
		Source:    source,
		VacancyID: vacancyID,
	}
}

func (pm *ParsersManager) tryEnqueueJob(ctx context.Context, job interfaces.Job, timeout time.Duration) bool {

	start := time.Now()

	for {
		// Пробуем добавить в очередь
		if pm.jobSearchQueue.Enqueue(job) {
			return true
		}

		// Проверяем таймаут
		if time.Since(start) > timeout {
			return false
		}

		// Проверяем отмену контекста
		select {
		case <-ctx.Done():
			return false
		default:
			// Небольшая пауза перед следующей попыткой
			time.Sleep(50 * time.Millisecond)
		}
	}
}
