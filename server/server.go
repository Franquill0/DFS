package main

import (
	"bufio"
	"fmt"
	"io"
	"labo/log_init"
	"labo/utils"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
)

func put(args []string, conn net.Conn, reader *bufio.Reader) {
	filename := args[1]
	log_init.PrintAndLog("PUT request del archivo", filename, "desde", conn.RemoteAddr())
	file, err := os.Create(filename)
	log_init.PrintAndLogIfError(err)
	if err != nil {
		return
	} else {
		//defer os.Remove(filename)
	}

	start := my_time.Now()
	_, err = io.Copy(file, reader)
	log_init.PrintAndLogIfError(err)
	if err != nil {
		return
	}
	log_init.PrintAndLog("Tiempo de subida ->", my_time.GetFormattedTime(start))
}
func get(args []string, conn net.Conn) {
	filename := args[1]
	log_init.PrintAndLog("GET request del archivo", filename, "desde", conn.RemoteAddr())
}
func ls(conn net.Conn) {
	log_init.PrintAndLog("LS request desde", conn.RemoteAddr())
	for file := range metadata {
		fmt.Fprintf(conn, "%s ", file)
	}
	fmt.Fprintf(conn, "\n")
}
func info(args []string, conn net.Conn) {
	filename := args[1]
	log_init.PrintAndLog("INFO request del archivo", filename, "desde", conn.RemoteAddr())
}

func handleConnection(conn net.Conn, wg *sync.WaitGroup) {
	defer wg.Done()
	defer conn.Close()
	reader := bufio.NewReader(conn)

	line, err := reader.ReadString('\n')
	log_init.PrintAndLogIfError(err)
	if err != nil {
		return
	}
	args := strings.Fields(line)
	command := args[0]
	switch command {
	case "put":
		put(args, conn, reader)
	case "get":
		get(args, conn)
	case "ls":
		ls(conn)
	case "info":
		info(args, conn)
	default:
		log_init.PrintAndLog("Comando no reconocido: ", strings.TrimSpace(line))
	}
}

func main() {
	log_init.InitializeLog()
	defer log_init.FinalizeLog()
	port := 8080
	ln, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		log.Println(err)
	}
	log_init.PrintAndLog("Escuchando en el puerto " + strconv.Itoa(port) + "...")
	fmt.Println("Presione ENTER para salir...")

	// Rutina para salir del servidor
	go func() {
		bufio.NewReader(os.Stdin).ReadBytes('\n')
		log_init.PrintAndLog("Cerrando servidor...")
		ln.Close()
	}()

	var waitingThreads sync.WaitGroup

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println(err)
			break
		}
		waitingThreads.Add(1)

		go handleConnection(conn, &waitingThreads)
	}
	waitingThreads.Wait()
}
