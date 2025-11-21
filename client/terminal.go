package main

import (
	"bufio"
	"fmt"
	"labo/commands"
	"labo/log_init"
	"log"
	"os"
)

func put(args []string) {
	if len(args) != 2 {
		fmt.Println("Uso -> put <archivo>")
		return
	}
	log.Println("Put request:", args[1])
}
func get(args []string) {
	if len(args) != 2 {
		fmt.Println("Uso -> get <archivo>")
		return
	}
}
func ls(args []string) {
	if len(args) != 1 {
		fmt.Println("Uso -> ls")
		return
	}
}
func info(args []string) {
	if len(args) != 2 {
		fmt.Println("Uso -> info <archivo>")
		return
	}
}
func exit(args []string) {
	log_init.FinalizeLog()
	os.Exit(0)
}

func main() {
	log_init.InitializeLog()

	functionMap := map[string]func([]string){
		"get":  get,
		"put":  put,
		"ls":   ls,
		"info": info,
		"exit": exit,
	}

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("> ")
		input, _ := reader.ReadString('\n')
		commands.ExecCommand(input, functionMap)
	}
}
