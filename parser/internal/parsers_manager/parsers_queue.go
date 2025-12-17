package parsers_manager

import (
	"context"
	"fmt"
	"parser/internal/domain/models"
	"time"
)

// Запуск воркеров для обработки очереди
func (pm *ParsersManager) startSearchWorkers() {
	for i := 0; i < pm.workers; i++ {
		pm.wg.Add(1)
		go pm.searchWorker(i)
	}
}

// метод, описывающий работу отдельного воркера. Воркер пытется забрать работу из очереди и обработать её
func (pm *ParsersManager) searchWorker(id int) {
	defer pm.wg.Done()

	for {
		select {
		case <-pm.stopWorkers: // канал для остановки всех воркеров
			// Получен сигнал остановки
			//fmt.Printf("Worker #%d: received stop signal\n", id)
			return
		default:
			job, ok := pm.jobSearchQueue.Dequeue()
			if ok {
				fmt.Printf("woker #%d - взял задачу из очереди и начал обработку\n", id)

				// проверяем тип джобы и вызываем соответствующий обработчик
				switch j := job.(type) {
				case *models.SearchVacanciesJob:
					pm.proccessSearchJob(j) // конкурентно ищем вакансии по всем доступным парсерам
				case *models.SearchVacancyDetailesJob:
					pm.proccessDetailsJob(j) // делаем запрос в конкретный сервис по конкретному ID
				}
			}
		}
	}
}

// метод обработки работы для воркера, поиск списка вакансий по всем доступным парсерам (конкурентные запросы)
func (pm *ParsersManager) proccessSearchJob(job *models.SearchVacanciesJob) {
	var results []models.SearchVacanciesResult
	var err error

	select {
	case pm.semaphore <- struct{}{}:
		// Получили слот в семафоре менеджера парсеров
		defer func() {
			<-pm.semaphore // Освобождаем слот
		}()
		// Используем глобальный Circuit Breaker
		err = pm.circuitBreaker.Execute(func() error {
			var err error
			results, err = pm.executeSearch(context.Background(), job.Params)
			return err
		})

		results, err = pm.handleSearchResult(results, err, job.Params)

	case <-time.After(pm.semaSlotGetTimeout):
		err = fmt.Errorf("❌ Таймаут ожидания свободного слота глобального семафора менеджера парсеров")
	}

	// Отправляем результат
	select {
	case job.ResultChan <- &models.JobSearchVacanciesResult{
		Results: results,
		Error:   err,
	}:
	default:
		// Получатель не ждет результата
	}
}

// метод для обработки работы для воркера, получение детальной информации по вакансии (запрос к конкретному сервису)
func (pm *ParsersManager) proccessDetailsJob(job *models.SearchVacancyDetailesJob) {
	var result models.SearchVacancyDetailesResult
	var err error

	select {
	case pm.semaphore <- struct{}{}:
		// Получили слот в семафоре менеджера парсеров
		defer func() {
			<-pm.semaphore // Освобождаем слот
		}()

		// Используем глобальный Circuit Breaker
		err = pm.circuitBreaker.Execute(func() error {
			var err error
			//	result, err = pm.executeSearchVacancyDetailes(context.Background(), job.VacancyID, job.ParserName)
			return err
		})

		//result, err = pm.handleSearchVacancyDetailesResult(result, err)
	case <-time.After(pm.semaSlotGetTimeout):
		err = fmt.Errorf("❌ Таймаут ожидания свободного слота глобального семафора менеджера парсеров")
	}

	// Отправляем результат
	select {
	case job.ResultChan <- &models.JobSearchVacancyDetailesResult{
		Result: result,
		Error:  err,
	}:
	default:
		// Получатель не ждет результата
	}
}

// метод для остановки всех воркеров
func (pm *ParsersManager) Shutdown() {
	fmt.Println("============================================================================")
	fmt.Println("Initiating shutdown...")

	// Закрываем канал - все воркеры получат сигнал
	close(pm.stopWorkers)

	// Ожидаем завершения всех воркеров
	done := make(chan struct{})

	go func() {
		pm.wg.Wait()
		// останавливаем менеджер статутос парсеров
		pm.parsersStatusManager.Stop()
		close(done)
	}()

	// Ждем с таймаутом
	select {
	case <-done:
		fmt.Println("All workers stopped gracefully")
	case <-time.After(10 * time.Second):
		fmt.Println("Warning: shutdown timeout, some workers may still be running")
	}
	// Закрываем очередь ---- нужно доработать
}
