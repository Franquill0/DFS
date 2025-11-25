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

type Metadata map[string][]FileBlock

var metadata Metadata

const path = "metadata.json"

func printMetadata() {
	fmt.Println(metadata)
}

func getDataNodesWithFile(filename string) []string {

}

func addFile(filename string) {
	if _, ok := metadata[filename]; !ok {
		metadata[filename] = []FileBlock{}
	}
}

func loadDatanodesInfo() error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	err = json.Unmarshal(data, &metadata)
	if err != nil {
		return err
	}

	return nil
}

func existingFile(filename string) bool {
	_, ok := metadata[filename]
	return ok
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
	jsonBytes, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, jsonBytes, 0644)
}
