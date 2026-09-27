package main

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
)

type DuplicateGroup struct {
	Hash  string
	Files []File
}

func CalculateHash(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	// 2. Создаем хешер (он реализует io.Writer)
	hasher := sha256.New()

	// 3. Запускаем потоковое копирование
	// Данные будут качаться по 32 КБ из файла прямиком в хешер
	_, err = io.Copy(hasher, f)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}
