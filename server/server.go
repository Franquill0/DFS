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
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

const filesDirectory = "files/"

func downloadFile(filename string, reader *bufio.Reader) (*os.File, error) {
	file, err := os.OpenFile(filesDirectory+filename, os.O_CREATE|os.O_RDWR, 0644)
	log_init.PrintAndLogIfError(err)
	if err != nil {
		return file, err
	}

	start := my_time.Now()
	_, err = io.Copy(file, reader)
	log_init.PrintAndLogIfError(err)
	if err != nil {
		return file, err
	}
	log_init.PrintAndLog("Tiempo de subida al servidor ->", my_time.GetFormattedTime(start))

	file.Seek(0, 0)
	return file, nil
}

func fileFragmentation(file *os.File) (int, error) {
	defer file.Close()
	filename := filepath.Base(file.Name())
	const blockSize = 1024
	buffer := make([]byte, blockSize)
	part := 0
	for {
		n, err := file.Read(buffer)
		if err == io.EOF {
			break
		} else if err != nil {
			return 0, nil
		}

		// Crear archivo fragmento
		partFileName := fmt.Sprintf("%s.part%d", filename, part)
		partFile, err := os.Create(filesDirectory + partFileName)
		if err != nil {
			return 0, nil
		}
		defer partFile.Close()

		// Guardar el fragmento (solo n bytes)
		_, err = partFile.Write(buffer[:n])
		if err != nil {
			return 0, nil
		}

		part++
	}
	log_init.PrintAndLog("Archivo", filename, "dividido en", part, "partes\n")
	return part, nil
}

func put(args []string, conn net.Conn, reader *bufio.Reader) {
	filename := args[1]
	log_init.PrintAndLog("PUT request del archivo", filename, "desde", conn.RemoteAddr())

	file, err := downloadFile(filename, reader)
	if err != nil {
		return
	}

	// Fragmentar archivo
	parts, err := fileFragmentation(file)
	log_init.PrintAndLogIfError(err)
	if err != nil {
		return
	}

	datanodes := getDatanodes()
	datanodesAmount := len(datanodes)
	if datanodesAmount == 0 {
		log_init.PrintAndLog("No hay datanodes disponibles!")
		return
	} else if datanodesAmount == 1 {
		go sendPartsToDatanode(datanodes[0], 0, parts, filename)
	}

	partsPerDatanode := int(parts / datanodesAmount)

	for index := 0; index < datanodesAmount-1; index++ {
		firstPart := index * partsPerDatanode
		lastPart := (index + 1) * partsPerDatanode
		go sendPartsToDatanode(datanodes[index], firstPart, lastPart, filename)
	}
	go sendPartsToDatanode(datanodes[datanodesAmount-1])

	// Enviar fragmentos a datanodes
	/*
		datanodeConn, err := net.Dial("tcp", "192.168.18.41:40249")
		fmt.Fprintf(datanodeConn, "store %s.part0\n", filename)
		part0, err := os.Open(filesDirectory + filename + ".part0")
		_, err = io.Copy(datanodeConn, part0)
		part0.Close()
		datanodeConn.Close()
	*/

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
		log_init.PrintAndLog("Esperando finalización de hilos...")
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
