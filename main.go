package main

import "fmt"

func main() {
	list, err := scan()
	if err != nil {
		fmt.Println(err)
	}

	hashMap := make(map[string][]File)
	FileSortToWeight := PrepareToHash(list)

	for size, value := range FileSortToWeight { // value здесь лишь КОПИЯ среза из карты
		if len(value) > 1 {
			for i, file := range value { // можно написать _,file, а можно просто i(index), но тогда нужно будет обращаться value[i], а не file
				hash, err := CalculateHash(file.Path)
				if err != nil {
					fmt.Println(err)
					continue
				}
				FileSortToWeight[size][i].Hash = hash // если бы мы написали file.Hash = hash, то м ыбы поменяли поле у КОПИИ среза, а так мы обращаемся к нему напрямую
				hashMap[hash] = append(hashMap[hash], file)
			}
		}
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
}
