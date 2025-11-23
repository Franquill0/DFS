package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

type FileBlock struct {
	Block string `json:"block"`
	Node  string `json:"node"`
}

var metadata = map[string][]FileBlock{}

func PrintMetadata() {
	fmt.Println(metadata)
}

func AddFile(filename string) {
	if _, ok := metadata[filename]; !ok {
		metadata[filename] = []FileBlock{}
	}
}

func AddFileBlock(filename, block, node string) error {
	var err error
	if existingFile, ok := metadata[filename]; !ok {
		err = errors.New("Archivo no existente!")
	} else {
		existingFile = append(existingFile, FileBlock{
			Block: block,
			Node:  node,
		})
		err = nil
	}
	return err
}

func RemoveFile(filename string) error {
	var err error
	if _, ok := metadata[filename]; ok {
		delete(metadata, filename)
		err = nil
	} else {
		err = errors.New("Archivo no existente!")
	}
	return err
}

func UpdateJSONMetadata() error {
	jsonBytes, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile("metadata.json", jsonBytes, 0644)
}
