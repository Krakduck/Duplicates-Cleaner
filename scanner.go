package main

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"
)

type File struct {
	Name string
	Path string
	Size int64
	Type string
	Date time.Time
	Hash string
}

func scan(path string, excl string) ([]File, error) {
	var ListFile []File
	exclude := strings.Split(excl, ",")
	err := filepath.WalkDir(path, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if slices.Contains(exclude, path) {
			if d.IsDir() {
				return filepath.SkipDir // пропустить всю папку и её содержимое
			}
			return nil // пропустить файл, идти дальше
		}

		info, err := d.Info()
		if err != nil {
			return err
		}

		if !d.IsDir() {
			file := File{
				Name: d.Name(),
				Path: path,
				Size: info.Size(),
				Type: filepath.Ext(d.Name()),
				Date: info.ModTime(),
				Hash: "",
			}
			ListFile = append(ListFile, file)
		}
		return nil
	})
	if err != nil {
		return []File{}, err
	}
	return ListFile, nil
}

func PrepareToHash(AllFile []File) map[int64][]File {
	x := make(map[int64][]File)

	for i := range AllFile { // можно написать _,file, а можно просто i(index), но тогда нужно будет обращаться value[i], а не file
		key := AllFile[i].Size
		x[key] = append(x[key], AllFile[i])
	}
	return x
}
