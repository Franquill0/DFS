package main

import (
	"encoding/json"
	"labo/log_init"
	"net"
	"os"
	"time"
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

func getAvailableDatanodes() []string {
	totalDatanodes := getDatanodes()
	var datanodesAvailable []string

	for _, datanode := range totalDatanodes {
		if isDatanodeUp(datanode) {
			datanodesAvailable = append(datanodesAvailable, datanode)
		}
	}
	return datanodesAvailable
}

func isDatanodeUp(address string) bool {
	conn, err := net.DialTimeout("tcp", address, time.Second)
	if err != nil {
		return false
	}
	log_init.PrintAndLog("PING a ", address)
	conn.Write([]byte("ping\n"))
	defer conn.Close()
	return true
}
