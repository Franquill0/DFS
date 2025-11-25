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

func sendPartsToDatanode(addr string, filename string, first, last int) {
	conn, err := net.Dial("tcp", addr)
	log_init.PrintAndLogIfError(err)
	if err != nil {
		return
	}
	defer conn.Close()

	writer := bufio.NewWriter(conn)
	fmt.Fprintf(writer, "store\n")

	for i := first; i < last; i++ {
		partName := fmt.Sprintf("%s.part%d", filename, i)
		partPath := filesDirectory + partName

		partFile, err := os.Open(partPath)
		if err != nil {
			continue
		}

		stat, _ := partFile.Stat()

		// Enviar comando store con el nombre del archivo y su tamaño
		log_init.PrintAndLog("Enviando: store", partName, stat.Size())
		fmt.Fprintf(writer, "STORE %s %d\n", partName, stat.Size())

		// Enviar contenido
		io.Copy(writer, partFile)

		writer.Flush()
		partFile.Close()
		err = os.Remove(partPath)
		if err != nil {
			log_init.PrintAndLog("Error al eliminar", filename)
		}
	}

	// Avisamos que no hay más bloques
	fmt.Fprintf(writer, "END\n")
	writer.Flush()
}

func addPartFileRange(start, end int, datanode, filename string) {
	for i := start; i < end; i++ {
		addFileBlock(filename, i, datanode)
	}
}

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

func removeFileFromDatanodes(filename string) {
	datanodes := getAvailableDatanodes()
	for _, datanode := range datanodes {
		conn, err := net.Dial("tcp", datanode)
		if err != nil {
			log_init.PrintAndLog("Error en conexión para eliminar", filename, "de", datanode)
		} else {
			writer := bufio.NewWriter(conn)
			fmt.Fprintf(writer, "remove %s\n", filename)
			log_init.PrintAndLog("REMOVE request de", filename, "hacia ", datanode)
		}
		conn.Close()
	}
}

func put(args []string, conn net.Conn, reader *bufio.Reader) {
	filename := args[1]
	log_init.PrintAndLog("PUT request del archivo", filename, "desde", conn.RemoteAddr())

	if existingFile(filename) {
		log_init.PrintAndLog("Archivo existente ", filename, "-> Sobreescribiendo contenido")
		removeFileFromDatanodes(filename)
	} else {
		addFile(filename)
	}

	file, err := downloadFile(filename, reader)
	if err != nil {
		return
	}

	// Fragmentar archivo
	parts, err := fileFragmentation(file)
	log_init.PrintAndLogIfError(err)
	if err != nil || parts == 0 {
		return
	}
	err = os.Remove(filesDirectory + filename)
	if err != nil {
		log_init.PrintAndLog("Error al eliminar", filename)
	}

	distributePartFiles(parts, filename)

}

func distributePartFiles(parts int, filename string) {
	datanodes := getAvailableDatanodes()
	datanodesAmount := len(datanodes)
	if datanodesAmount == 0 {
		log_init.PrintAndLog("No hay datanodes disponibles!")
		return
	}

	partsPerDatanode := (parts + datanodesAmount - 1) / datanodesAmount

	for index, datanode := range datanodes {
		if parts == 0 {
			break
		}
		firstPart := index * partsPerDatanode
		lastPart := 0
		if parts-partsPerDatanode >= 0 {
			lastPart = firstPart + partsPerDatanode
			parts = parts - partsPerDatanode
		} else {
			lastPart = firstPart + parts
			parts = 0
		}
		addPartFileRange(firstPart, lastPart, datanode, filename)
		go sendPartsToDatanode(datanode, filename, firstPart, lastPart)
	}
}

func get(args []string, conn net.Conn) {
	filename := args[1]
	log_init.PrintAndLog("GET request del archivo", filename, "desde", conn.RemoteAddr())

	writer := bufio.NewWriter(conn)
	datanodesWithFile := getDatanodesWithFile(filename)
	fmt.Fprintf(writer, "blocks %d ")
	for {
		fmt.Fprintf(writer, "%s ", datanode)

	}

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

	err := loadDatanodesInfo()
	if err != nil {
		log_init.PrintAndLog("Error en lectura del archivo json!")
	}

	port := 8080
	ln, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		log.Println(err)
	}
	log_init.PrintAndLog("Escuchando en el puerto " + strconv.Itoa(port) + "...")
	fmt.Println("Presione ENTER para salir...")

	// Rutina para salir del servidor
	getOutServer := false
	go func() {
		bufio.NewReader(os.Stdin).ReadBytes('\n')
		log_init.PrintAndLog("Cerrando servidor...")
		log_init.PrintAndLog("Esperando finalización de hilos...")
		getOutServer = true
		ln.Close()
	}()

	var waitingThreads sync.WaitGroup

	for {
		conn, err := ln.Accept()
		if err != nil {
			if !getOutServer {
				log_init.PrintAndLogIfError(err)
			}
			break
		}
		waitingThreads.Add(1)

		go handleConnection(conn, &waitingThreads)
	}
	waitingThreads.Wait()
	err = updateJSONMetadata()
	if err != nil {
		log_init.PrintAndLog("Error al escribir el json, error:", err)
	}
	log_init.FinalizeLog()
}
