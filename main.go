package main

import (
	"bufio"
	"fmt"
	"os"
	"sync"
	"time"
)

func main() {

	//C:\Users\Роман\Desktop\егэ\русский
	//C:\Users\Роман\Desktop\егэ\русский\РУ-11-ЕГЭ-2027-ДЕМО.pdf,C:\Users\Роман\Desktop\егэ\русский\РУ-11-ЕГЭ-2027-ДЕМО — копия.pdf
	//-path C:\Users\Роман\Desktop\егэ\русский -exclude "C:\Users\Роман\Desktop\егэ\русский\РУ-11-ЕГЭ-2027-ДЕМО.pdf,C:\Users\Роман\Desktop\егэ\русский\РУ-11-ЕГЭ-2027-ДЕМО — копия.pdf"

	opts, err := parseOptions()
	if err != nil {
		fmt.Fprintln(os.Stderr, "\nОшибка:", err)
		os.Exit(1)
	}

	start1 := time.Now()

	FileSortToWeight := PrepareToHash(opts.Files)

	//кандидаты на хеширование
	var candidates []File
	for _, files := range FileSortToWeight {
		if len(files) > 1 {
			candidates = append(candidates, files...) //... распаковывает один срез в другой. добавляет элементы массива по одному, ибо список не может принять сразу весь список файлов
		}
	}

	// 2. Создаем каналы связи. (Буфер len(candidates) нужен, чтобы отправка не блокировала главный поток)
	jobs := make(chan File, len(candidates)) // chan - специальны тип данных для каналов. len(candidates) - размер буфера
	results := make(chan File, len(candidates))

	var wg sync.WaitGroup // 3. Создаем счетчик для контроля горутин

	for i := 0; i < 8; i++ {
		wg.Add(1) // Увеличиваем счетчик ждущих горутин
		go worker(jobs, results, &wg)
	}

	// 5. Отправляем все файлы в канал задач
	for _, file := range candidates {
		jobs <- file // Закидываем файл в канал
	}
	close(jobs) // Сигнализируем воркерам: задачи закончились!

	// 6. Запускаем отдельный фоновый "наблюдатель", который закроет results
	go func() {
		wg.Wait()      // Ждем, пока ВСЕ 8 воркеров сделают wg.Done().
		close(results) // Закрываем канал результатов
	}()

	// 7. Сбор результатов обратно в hashMap
	hashMap := make(map[string][]File)
	for file := range results { //берет элементы из канала
		hashMap[file.Hash] = append(hashMap[file.Hash], file)
	}

	var duplicates []DuplicateGroup
	for hash, files := range hashMap {
		if len(files) > 1 {
			duplicate := DuplicateGroup{
				Hash:  hash,
				Files: files, //срез
			}
			duplicates = append(duplicates, duplicate)
		}
	}

	for index, duplicate := range duplicates {
		var totalSize int64
		a := duplicate.Files
		for _, file := range a {
			fmt.Println(index, file.Path)
			totalSize += file.Size
		}
		fmt.Println("эта группа занимает", totalSize/1024/1024/1024, "GB = ", totalSize/1024/1024, "MB = ", totalSize/1024, "KB")
		fmt.Println("\n")
	}
	fmt.Println("total amount fileP:", len(opts.Files))
	fmt.Println("total amount duplicate:", len(duplicates), "файла/-ов имеют дубликаты")
	fmt.Println("duration:", time.Since(start1))

	// удаление по флагу
	if opts.AutoDelete {
		removeAll(duplicates)
	}

	// интерактивное подтверждение
	if opts.Interactive {
		askAndDelete(duplicates)
		fmt.Print("\nНажмите Enter для выхода...")
		bufio.NewReader(os.Stdin).ReadString('\n')
	}

}
