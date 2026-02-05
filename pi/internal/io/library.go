package io

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type MusicLibrary struct {
	Dir   string               `json:"dir"`
	Files map[string][]float32 `json:"Files"`
}

func (ml *MusicLibrary) Filenames() []string {
	keys := make([]string, 0, len(ml.Files))
	for k := range ml.Files {
		keys = append(keys, k)
	}
	return keys
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func LoadMusicLibary(path string) (*MusicLibrary, error) {
	if !fileExists(path) {
		return nil, fmt.Errorf("File '%s' does not exist", path)
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	defer file.Close()

	var lib MusicLibrary
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&lib); err != nil {
		return nil, err
	}

	updatedMap := make(map[string][]float32)
	for path, embeddings := range lib.Files {
		fullPath := filepath.Join(lib.Dir, path)
		delete(lib.Files, path)

		if fileExists(fullPath) {
			updatedMap[fullPath] = embeddings
		} else {
			fmt.Printf("Removing '%s' from library as file does not exist on local drive\n", fullPath)
		}
	}

	lib.Files = updatedMap

	return &lib, nil
}
