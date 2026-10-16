package main

import "sync"

// worker — это функция одного рабочего
func worker(jobs <-chan File, results chan<- File, wg *sync.WaitGroup) {
	//  defer wg.Done() автоматически выполнится, когда воркер завершит работу
	defer wg.Done()

	//  Этот цикл берет файлы из канала по одному.
	// Он сам остановится, когда канал jobs будет закрыт (close(jobs))
	for file := range jobs {
		hash, err := CalculateHash(file.Path)
		if err != nil {
			continue // Если файл не прочитался — пропускаем
		}

		file.Hash = hash // Записываем посчитанный хеш
		results <- file  // Отправляем готовую структуру с хешем в канал results
	}
}
