package utils

import (
	"fmt"
	"io/ioutil"
	"os"
)

// func RemoveFilesWithKeyword(keyword string) error {
// 	files, err := ioutil.ReadDir("../files")
// 	if err != nil {
// 		return err
// 	}

// 	for _, file := range files {
// 		if !file.IsDir() && strings.Contains(file.Name(), keyword) {
// 			err := os.Remove("../files/" + file.Name())
// 			if err != nil {
// 				return err
// 			} else {
// 				return err
// 			}
// 		}
// 	}
// 	return nil
// }

// func FindFileWithKeyword(keyword string) (string, error) {
// 	files, err := ioutil.ReadDir("../files")
// 	if err != nil {
// 		return "", err
// 	}

// 	for _, file := range files {
// 		if !file.IsDir() && strings.Contains(file.Name(), keyword) {
// 			return file.Name(), nil
// 		}
// 	}

// 	return "", fmt.Errorf("no file found with keyword: %s", keyword)
// }

func RemoveAllFiles() error {
	files, err := ioutil.ReadDir("../files")
	if err != nil {
		return err
	}

	for _, file := range files {
		if !file.IsDir() {
			err := os.Remove("../files/" + file.Name())
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func FindFirstFile() (string, error) {
	files, err := ioutil.ReadDir("../files")
	if err != nil {
		return "", err
	}

	for _, file := range files {
		if !file.IsDir() {
			return file.Name(), nil
		}
	}

	return "", fmt.Errorf("no files found in the directory")
}
