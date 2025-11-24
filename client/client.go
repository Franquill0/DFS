package main

import (
	"bufio"
	"fmt"
	"io"
	"labo/log_init"
	"net"
	"os"
)

const serverIPPort = "192.168.18.41:8080"

func put(args []string) {
	if len(args) != 2 {
		fmt.Println("Uso -> put <archivo>")
		return
	}
	filename := args[1]
	file, err := os.Open(filename)
	log_init.PrintAndLogIfError(err)
	if err != nil {
		return
	} else {
		defer file.Close()
	}
	log_init.PrintAndLog("PUT request del archivo", filename)
	conn := stablishConn()
	if conn != nil {
		defer conn.Close()
	}
	// Envío el comando con el nombre del archivo
	fmt.Fprintf(conn, "put %s\n", filename)

	_, err = io.Copy(conn, file)
	log_init.PrintAndLogIfError(err)
	if err != nil {
		return
	}
	log_init.PrintAndLog("Archivo cargado", filename)

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
	} else {
		log_init.PrintAndLog("Conexión fallida con el servidor")
		return
	}
	log_init.PrintAndLog("LS request")
	fmt.Fprintf(conn, "ls\n")
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
