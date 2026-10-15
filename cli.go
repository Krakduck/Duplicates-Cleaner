package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Options — всё, что программа получает на входе
type Options struct {
	Path        string // путь к папке
	Exclude     string // исключения через запятую
	AutoDelete  bool   // -delete
	Interactive bool   // был ли интерактивный режим
	Files       []File // результат scan
}

// parseOptions разбирает флаги, при необходимости спрашивает пользователя
// и сразу сканирует папку. Возвращает готовые опции или ошибку.
func parseOptions() (*Options, error) {
	// можно написать
	//var path string
	//flag.StringVar(&path,"path", "", "папка для поиска дубликатов")
	//но тогда path не будет указателем и его не нужно будет разыменовывать через * и можно указать несколько флагов для одной переменной

	path := flag.String("path", "", "папка для поиска дубликатов")
	exclude := flag.String("exclude", "", "исключение") // вводить через запятую В КАВЫЧКАХ
	flagDelete := flag.String("delete", "", "автоматический перенос в корзину")
	flag.Parse()

	opts := &Options{
		Path:       *path,
		Exclude:    *exclude,
		AutoDelete: *flagDelete == "true" || *flagDelete == "1",
	}

	reader := bufio.NewReader(os.Stdin)

	if opts.Path == "" {
		// Интерактивный режим: спрашиваем, пока не получим валидный путь
		opts.Interactive = true
		for {
			fmt.Print("\nВведите путь к папке без кавычек: ")
			input, err := reader.ReadString('\n')
			if err != nil {
				fmt.Fprintln(os.Stderr, "\nОшибка чтения:", err)
				continue
			}
			opts.Path = strings.TrimSpace(input)

			fmt.Print("\nВведите пути, которые надо исключить через запятую без кавычек: ")
			input, err = reader.ReadString('\n')
			if err != nil {
				fmt.Fprintln(os.Stderr, "\nОшибка чтения:", err)
				continue
			}
			opts.Exclude = strings.TrimSpace(input)

			list, err := scan(opts.Path, opts.Exclude)
			if err == nil {
				opts.Files = list
				break
			}
			fmt.Fprintln(os.Stderr, "\nОшибка:", err)
		}
	} else {
		// Неинтерактивный режим: сразу сканируем
		list, err := scan(opts.Path, opts.Exclude)
		if err != nil {
			return nil, err
		}
		opts.Files = list
	}

	return opts, nil
}

// removeAll переносит все дубликаты в корзину
func removeAll(duplicates map[string][]File) {
	for _, value := range duplicates { //value - []File
		duplicate := value
		for _, file := range duplicate[1:] {
			err := deleteFile(file.Path)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
			}
		}
	}
}

// askAndDelete спрашивает подтверждение и удаляет
func askAndDelete(duplicates map[string][]File) {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Println("\nЖелаете удалить дубликаты? (yes/no)(y/n)")
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Fprintln(os.Stderr, "\nОшибка чтения:", err)
			continue
		}
		input = strings.TrimSpace(input)
		if input == "yes" || input == "y" {
			removeAll(duplicates)
		}
		return
	}
}

func deleteFile(path string) error {
	// Безопасно экранируем путь для PowerShell
	psCommand := fmt.Sprintf(`
        try {
            Add-Type -AssemblyName Microsoft.VisualBasic;
            [Microsoft.VisualBasic.FileIO.FileSystem]::DeleteFile('%s', 'OnlyErrorDialogs', 'SendToRecycleBin');
            exit 0;
        } catch {
            Write-Error $_.Exception.Message;
            exit 1;
        }
    `, path)

	cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", psCommand)

	// Получаем подробный вывод ошибки, если она случится
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ошибка удаления (%v): %s", err, strings.TrimSpace(string(output)))
	}
	return nil
}
