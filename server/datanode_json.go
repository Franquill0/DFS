package main

import (
	"encoding/json"
	"labo/log_init"
	"os"
)

type DatanodeConfig struct {
	Datanodes []string `json:"datanodes"`
}

func getDatanodes() []string {
	file, err := os.Open("datanodes_config.json")
	log_init.PrintAndLogIfError(err)
	defer file.Close()

	var cfg DatanodeConfig

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&cfg)
	log_init.PrintAndLogIfError(err)

	return cfg.Datanodes
}
