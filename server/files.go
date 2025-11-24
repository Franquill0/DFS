package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
)

type FileBlock struct {
	Block string `json:"block"`
	Node  string `json:"node"`
}

var metadata = map[string][]FileBlock{}

func printMetadata() {
	fmt.Println(metadata)
}

func addFile(filename string) {
	if _, ok := metadata[filename]; !ok {
		metadata[filename] = []FileBlock{}
	}
}

func addFileBlock(filename string, block int, node string) error {
	var err error
	if _, ok := metadata[filename]; !ok {
		err = errors.New("File not found: " + filename)
	} else {
		metadata[filename] = append(metadata[filename], FileBlock{
			Block: "b" + strconv.Itoa(block),
			Node:  node,
		})
		err = nil
	}
	return err
}

func removeFile(filename string) error {
	var err error
	if _, ok := metadata[filename]; ok {
		delete(metadata, filename)
		err = nil
	} else {
		err = errors.New("File not found: " + filename)
	}
	return err
}

func updateJSONMetadata() error {
	path := "metadata.json"
	jsonBytes, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, jsonBytes, 0644)
}
