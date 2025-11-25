package main

import (
	"bufio"
	"fmt"
	"io"
	"labo/log_init"
	"labo/utils"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
)

const serverIPPort = "192.168.18.41:8080"
const partsDirectory = "parts/"

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
	conn := stablishConn(serverIPPort)
	if conn != nil {
		defer conn.Close()
	}
	// Envío el comando con el nombre del archivo
	stat, _ := file.Stat()
	fmt.Fprintf(conn, "put %s %d\n", filename, stat.Size())

	log_init.PrintAndLog("Enviando archivo:", filename)
	startTime := my_time.Now()
	_, err = io.CopyN(conn, file, stat.Size())
	log_init.PrintAndLogIfError(err)
	if err != nil {
		return
	}
	log_init.PrintAndLog("Archivo enviado:", filename)
	log_init.PrintAndLog("Tiempo de subida ->", my_time.GetFormattedTime(startTime))
}
func get(args []string) {
	if len(args) != 2 {
		fmt.Println("Uso -> get <archivo>")
		return
	}
	log_init.PrintAndLog("Iniciando conexión con el servidor...")
	conn := stablishConn(serverIPPort)
	if conn == nil {
		log_init.PrintAndLog("Error en la conexión con", serverIPPort)
		return
	}
	filename := args[1]
	fmt.Fprintf(conn, "get %s\n", filename)
	log_init.PrintAndLog("Conexión exitosa")
	reader := bufio.NewReader(conn)
	line, err := reader.ReadString('\n')
	if err != nil {
		log_init.PrintAndLog("Error en leer los datanodes y bloques!")
		return
	}
	fullLine := strings.Fields(line)
	blocks, err := strconv.Atoi(fullLine[1])
	if err != nil {
		log_init.PrintAndLog("Error en conversión a entero de", fullLine[0])
		return
	}
	var datanodes []string
	for index := 2; index < len(fullLine); index++ {
		datanodes = append(datanodes, fullLine[index])
	}
	log_init.PrintAndLog("Cantidad de bloques:", blocks)
	log_init.PrintAndLog("Datanodes con los archivos:", datanodes)

	getPartsFromDatanodes(datanodes, blocks)

}

func handleDatanodeConnection(conn net.Conn, wg *sync.WaitGroup) {
	defer wg.Done()
	writer := bufio.NewWriter(conn)
	fmt.Fprintf(writer, "read\n")
	writer.Flush()

}

func getPartsFromDatanodes(datanodes []string, blocks int) {
	var wg sync.WaitGroup
	for _, datanode := range datanodes {
		log_init.PrintAndLog("Estableciendo conexión con datanode", datanode)
		conn := stablishConn(datanode)
		if conn != nil {
			log_init.PrintAndLog("Conexión con datanode", datanode, "exitosa")
		} else {
			log_init.PrintAndLog("Conexión con datanode", datanode, "falló!")
			break
		}
		wg.Add(1)
		go handleDatanodeConnection(conn, &wg)
	}
	wg.Wait()

}

func ls(args []string) {
	if len(args) != 1 {
		fmt.Println("Uso -> ls")
		return
	}
	conn := stablishConn(serverIPPort)
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

func stablishConn(ipPort string) net.Conn {
	conn, err := net.Dial("tcp", ipPort)
	if err != nil {
		log_init.PrintAndLog("Error en la conexión con", ipPort, "->", err)
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
