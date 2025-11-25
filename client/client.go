package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"labo/log_init"
	"labo/utils"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
)

const serverIPPort = "192.168.18.41:8080"
const partsDirectory = "parts"

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

	err = getPartsFromDatanodes(filename, datanodes)
	defer removeRemainingPartsFile(filename)
	if err != nil {
		log_init.PrintAndLogIfError(err)
		return
	}

	// Reconstruyo el archivo
	err = reconstructFile(filename, "copia_"+filename, blocks)
	log_init.PrintAndLogIfError(err)

}
func reconstructFile(filename, output string, blocks int) error {
	entries, err := os.ReadDir(partsDirectory)
	if err != nil {
		return err
	}

	pattern := fmt.Sprintf(`^%s\.part([0-9]+)$`, regexp.QuoteMeta(filename))
	re := regexp.MustCompile(pattern)

	// mapa de índice -> nombre de archivo
	partMap := make(map[int]string)
	var indices []int

	for _, entry := range entries {
		name := entry.Name()
		match := re.FindStringSubmatch(name)
		if len(match) == 2 {
			idx, _ := strconv.Atoi(match[1])
			partMap[idx] = name
			indices = append(indices, idx)
		}
	}

	if len(indices) == 0 {
		log_init.PrintAndLog("No se encontraron partes para", filename)
		return errors.New("No se encontraron partes para " + filename)
	} else if len(indices) != blocks {
		log_init.PrintAndLog("Error: Se han encontrado", len(indices), "archivos donde deberían ser", blocks)
		log_init.PrintAndLog("Abortando reconstrucción de", filename)
		return errors.New("Error: No se descargaron suficientes bloques")
	}

	sort.Ints(indices)

	outFile, err := os.Create(output)
	if err != nil {
		return err
	}
	defer outFile.Close()

	for _, idx := range indices {
		partName := partMap[idx]
		partFile, err := os.Open(partsDirectory + "/" + partName)
		if err != nil {
			return err
		}

		_, err = io.Copy(outFile, partFile)
		partFile.Close()
		if err != nil {
			return err
		}
		log_init.PrintAndLog("Concatenado:", partName)
	}
	log_init.PrintAndLog("Archivo reconstruido:", output)
	return nil
}

func removeRemainingPartsFile(filename string) {
	entries, err := os.ReadDir(partsDirectory)
	if err != nil {
		log_init.PrintAndLogIfError(err)
		return
	}

	pattern := fmt.Sprintf(`^%s\.part([0-9]+)$`, regexp.QuoteMeta(filename))
	re := regexp.MustCompile(pattern)

	partMap := make(map[int]string)

	for _, entry := range entries {
		name := entry.Name()
		match := re.FindStringSubmatch(name)
		if len(match) == 2 {
			idx, _ := strconv.Atoi(match[1])
			partMap[idx] = name
		}
	}
	for _, part := range partMap {
		filepath := partsDirectory + "/" + part
		os.Remove(filepath)
		log_init.PrintAndLog("Eliminado", part)
	}
}

func handleDatanodeConnection(conn net.Conn, filename string, wg *sync.WaitGroup) {
	defer wg.Done()
	writer := bufio.NewWriter(conn)
	reader := bufio.NewReader(conn)
	fmt.Fprintf(writer, "read %s\n", filename)
	writer.Flush()

	for {
		header, err := reader.ReadString('\n')
		if err != nil {
			log_init.PrintAndLog("Error al leer comando del handleDatanodeConnection!")
		}

		header = strings.TrimSpace(header)
		if header == "end" {
			return
		}

		fullLine := strings.Fields(header)
		fileSize, _ := strconv.Atoi(fullLine[2])
		filename := fullLine[1]
		file, err := os.Create(partsDirectory + "/" + filename)
		if err != nil {
			log_init.PrintAndLog("Error al crear archivo", filename)
		}

		_, err = io.CopyN(file, reader, int64(fileSize))
		if err != nil {
			log_init.PrintAndLog("Error al copiar el contenido del archivo", filename)
		}

		file.Close()

		fmt.Fprintf(writer, "OK\n")
		writer.Flush()
	}

}

func getPartsFromDatanodes(filename string, datanodes []string) error {
	var wg sync.WaitGroup
	for _, datanode := range datanodes {
		log_init.PrintAndLog("Estableciendo conexión con datanode", datanode)
		conn := stablishConn(datanode)
		if conn != nil {
			log_init.PrintAndLog("Conexión con datanode", datanode, "exitosa")
		} else {
			errorMsg := fmt.Sprintf("Conexión con datanode %s falló, abortando operación GET del archivo %s\n", filename)
			return errors.New(errorMsg)
		}
		wg.Add(1)
		go handleDatanodeConnection(conn, filename, &wg)
	}
	wg.Wait()
	return nil
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
	conn := stablishConn(serverIPPort)
	if conn != nil {
		defer conn.Close()
	} else {
		log_init.PrintAndLog("Conexión fallida con el servidor")
		return
	}

	writer := bufio.NewWriter(conn)
	filename := args[1]
	fmt.Fprintf(writer, "info %s\n", filename)
	writer.Flush()
	reader := bufio.NewReader(conn)
	for {
		header, err := reader.ReadString('\n')
		if err != nil {
			log_init.PrintAndLog("Error en leer info:", err)
			return
		}

		header = strings.TrimSpace(header)
		if header == "end" {
			log_init.PrintAndLog("Fin de comunicación de INFO")
			return
		}

		fmt.Println(header)

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
