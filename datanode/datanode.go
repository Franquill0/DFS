package main

import (
	"bufio"
	"fmt"
	"io"
	"labo/log_init"
	"net"
	"os"
	"regexp"
	"strconv"
	"strings"
)

const blocksDirectory = "blocks/"

func read(conn net.Conn, reader *bufio.Reader) {

}

func store(reader *bufio.Reader) {
	for {
		header, err := reader.ReadString('\n')
		if err != nil {
			log_init.PrintAndLog("Error en leer store: ", err)
			return
		}
		fmt.Println("header -> ", header)

		header = strings.TrimSpace(header)
		if header == "END" {
			return
		}

		parts := strings.Fields(header)
		if parts[0] != "STORE" || len(parts) != 3 {
			return
		}

		blockID := parts[1]                           // Nombre del archivo
		size, _ := strconv.ParseInt(parts[2], 10, 64) // Tamaño del archivo
		fmt.Println("Archivo de tamaño", size)

		file, err := os.Create(blocksDirectory + blockID)
		if err != nil {
			log_init.PrintAndLog("Error al crear ", blockID, "->", err)
			return
		}

		// Copiar exactamente size bytes
		io.CopyN(file, reader, size)

		log_init.PrintAndLog("Guardado bloque:", blockID)
		file.Close()
	}
}

func removeFile(filename string) {
	pattern := fmt.Sprintf(`^%s\.part[0-9]+$`, regexp.QuoteMeta(filename))
	re := regexp.MustCompile(pattern)

	entries, _ := os.ReadDir(blocksDirectory)

	for _, entry := range entries {
		name := entry.Name()
		if re.MatchString(name) {
			os.Remove(name)
			log_init.PrintAndLog("Bloque eliminado:", name)
		}
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)

	line, err := reader.ReadString('\n')
	fmt.Println(line)
	if err != nil {
		log_init.PrintAndLog("Error en lectura del comando!", err)
		return
	}
	args := strings.Fields(line)
	command := args[0]
	switch command {
	case "read":
		read(conn, reader)
	case "store":
		store(reader)
	case "ping":
		log_init.PrintAndLog("Ping del servidor.")
	default:
		log_init.PrintAndLog("Comando no reconocido: ", strings.TrimSpace(line))
	}
}
func getLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}
func main() {
	log_init.InitializeLog()
	defer log_init.FinalizeLog()
	ln, err := net.Listen("tcp", ":0")
	addr := ln.Addr().(*net.TCPAddr)
	log_init.PrintAndLogIfError(err)
	listeningLog := fmt.Sprintf("Escuchando en %s:%d...", getLocalIP(), addr.Port)
	log_init.PrintAndLog(listeningLog)
	fmt.Println("Presione ENTER para salir...")

	// Rutina para salir del servidor
	getOut := false
	go func() {
		bufio.NewReader(os.Stdin).ReadBytes('\n')
		log_init.PrintAndLog("Cerrando datanode...")
		getOut = true
		ln.Close()
	}()

	for {
		conn, err := ln.Accept()
		if err != nil {
			if !getOut {
				log_init.PrintAndLogIfError(err)
			}
			break
		}
		go handleConnection(conn)
	}
}
