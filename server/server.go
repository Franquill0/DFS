package main

import (
	"bufio"
	"fmt"
	"labo/log_init"
	"log"
	"net"
	"strconv"
)

var commandsMap = map[string]func([]string, net.Conn){
	"put":  put,
	"get":  get,
	"ls":   ls,
	"info": info,
}

func put(args []string, conn net.Conn) {
	log_init.PrintAndLog("Put request: " + args[1])
	conn.Write([]byte("Put request\n"))
}
func get(args []string, conn net.Conn) {
	log.Println("Get request:", args[1])
}
func ls(args []string, conn net.Conn) {
	log.Println("Ls request:")
}
func info(args []string, conn net.Conn) {
	log.Println("Info request:", args[1])
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)

	line, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Error leyendo comando:", err)
		conn.Write([]byte("ERR al leer el comando!\n"))
		return
	}
	execCommand(line, commandsMap, conn)
}

func main() {
	log_init.InitializeLog()
	defer log_init.FinalizeLog()
	port := 8080
	ln, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		log.Println(err)
	}
	fmt.Println("Escuchando en el puerto", port, "...")

	defer ln.Close()

	for {
		// Accept an incoming connection
		conn, err := ln.Accept()
		if err != nil {
			log.Println(err)
		}

		// Handle the connection
		go handleConnection(conn)
	}
}
