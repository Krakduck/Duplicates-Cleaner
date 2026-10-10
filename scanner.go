package main

import (
	"os"
	"path/filepath"
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

func scan(path string) ([]File, error) {
	ListFile := []File{} //"C:\Users\Роман\Desktop\егэ\русский"
	//  "Z:\\айфоша\\DCIM"  "C:\\Users\\Роман\\Desktop\\егэ\\русский"  "C:\\Users\\Роман\\Desktop\\Volume_1 (192.168.1.77) - Ярлык.lnk"   "Z:\\1С+Ключи Пользователей"

	err := filepath.WalkDir(path, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
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
