package utils

import (
	"log"
	"os"
	"path/filepath"
)

func ClearDir(dir string) {
	files, err := os.ReadDir(dir)
	if err != nil {
		log.Printf("Couldn't open dir, error: %v", err)
	}

	for _, file := range files {
		name := file.Name()
		itemPath := filepath.Join(dir, name)

		err = os.RemoveAll(itemPath)
		if err != nil {
			log.Printf("Couldn't remove item: %s, error: %v", itemPath, err)
		}
	}
}
