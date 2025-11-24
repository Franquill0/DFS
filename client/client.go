package main

import (
	"bufio"
	"fmt"
	"labo/log_init"
	"log"
	"net"
	"os"
	"strings"
)

const serverIPPort = "192.168.18.41:8080"

func put(args []string) {
	if len(args) != 2 {
		fmt.Println("Uso -> put <archivo>")
		return
	}
	log.Println("Put request:", args[1])
	conn := stablishConn()
	if conn != nil {
		defer conn.Close()
	}
	conn.Write([]byte(strings.Join(args, " ") + "\n"))
	connReader := bufio.NewReader(conn)

	response, err := connReader.ReadString('\n')
	if err != nil {
		log_init.PrintAndLogIfError(err)
		return
	}
	fmt.Print(response)

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
	conn := stablishConn()
	if conn != nil {
		defer conn.Close()
	}
	conn.Write([]byte(strings.Join(args, " ") + "\n"))
	connReader := bufio.NewReader(conn)

	response, err := connReader.ReadString('\n')
	if err != nil {
		log_init.PrintAndLogIfError(err)
		return
	}
	fmt.Print(response)

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

func stablishConn() net.Conn {
	conn, err := net.Dial("tcp", serverIPPort)
	if err != nil {
		log_init.PrintAndLogIfError(err)
		return nil
	}
	return conn
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

	commandReader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("> ")
		input, _ := commandReader.ReadString('\n')
		execCommand(input, functionMap)
	}
}
