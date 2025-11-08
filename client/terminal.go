package main

import (
	"bufio"
	"fmt"
	"labo/log"
	"os"
	"strings"
)

func put(filename string) {
	fmt.Println("Put", filename)
	log.Log("Put", filename)
}
func get(filename string) {
	fmt.Println("Get", filename)
}
func ls() {
	fmt.Println("ls")
}
func info(filename string) {
	fmt.Println("Info", filename)
}

func main() {
	log.InitializeLog()

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("> ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		args := strings.Split(input, " ")
		command := strings.ToLower(args[0])

		switch command {
		case "put":
			if len(args) != 2 {
				fmt.Println("Uso -> put <archivo>")
			} else {
				put(args[1])
			}
		case "get":
			if len(args) != 2 {
				fmt.Println("Uso -> get <archivo>")
			} else {
				get(args[1])
			}
		case "ls":
			if len(args) != 1 {
				fmt.Println("Uso -> ls")
			} else {
				ls()
			}
		case "info":
			if len(args) != 2 {
				fmt.Println("Uso -> info <archivo>")
			} else {
				info(args[1])
			}
		case "exit":
			fmt.Println("Saliendo...")
			return
		default:
			fmt.Println("Comando desconocido, encontrado: ", input)
		}
	}
}
