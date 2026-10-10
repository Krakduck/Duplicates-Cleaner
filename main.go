package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

func main() {
	path := flag.String("path", "", "папка для поиска дубликатов")
	flag.Parse()

	interactive := false
	correct := false
	list := []File{}
	for i := 0; correct != true; i++ {
		if *path == "" {
			interactive = true
			reader := bufio.NewReader(os.Stdin)
			fmt.Print("Введите путь к папке без кавычек: ")
			input, err := reader.ReadString('\n')
			if err != nil {
				fmt.Fprintln(os.Stderr, "Ошибка чтения:", err)
				//os.Exit(1) //тоже самое, что return, но безопаснее для скриптов
			}
			*path = strings.TrimSpace(input) //обрезание переноса строки
		} else {
			correct = true
		}

		if *path == "" {
			fmt.Fprintln(os.Stderr, "Путь не задан. Использование: dupefinder -path <папка>")
			bufio.NewReader(os.Stdin).ReadString('\n') //пауза перед закрытием. новый ридер для надежности
			//os.Exit(1)
		} else {
			correct = true
		}
		var err error
		list, err = scan(*path)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Ошибка:", err)
			bufio.NewReader(os.Stdin).ReadString('\n') //пауза перед закрытием. новый ридер для надежности
			os.Exit(1)
		} else {
			correct = true
		}
	}

	start1 := time.Now()

	FileSortToWeight := PrepareToHash(list)

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

	duplicates := []DuplicateGroup{}
	for hash, files := range hashMap {
		if len(files) > 1 {
			duplicate := DuplicateGroup{
				Hash:  hash,
				Files: files,
			}
			duplicates = append(duplicates, duplicate)
		}
	}

	for _, duplicate := range duplicates {
		a := duplicate.Files
		for _, file := range a {
			fmt.Println(file.Path)
		}
	}
	fmt.Println("total amount fileP:", len(list))
	fmt.Println("total amount duplicate:", len(duplicates))
	fmt.Println("duration:", time.Since(start1))

	if interactive {
		fmt.Print("\nНажмите Enter для выхода...")
		bufio.NewReader(os.Stdin).ReadString('\n') //пауза перед закрытием. новый ридер для надежности
	}
}
