package main

import (
	"bufio"
	"fmt"
	"io"
	"labo/log_init"
	"net"
	"os"
	"regexp"
	"strings"
)

const blocksDirectory = "blocks/"

func read(blockID string, conn net.Conn) {

}
func store(blockID string, conn net.Conn, reader *bufio.Reader) {
	log_init.PrintAndLog("STORE request del BLOQUE", blockID)
	file, err := os.Create(blocksDirectory + blockID)
	log_init.PrintAndLogIfError(err)
	if err != nil {
		return
	}
	_, err = io.Copy(file, reader)
	log_init.PrintAndLogIfError(err)
	if err != nil {
		return
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
	log_init.PrintAndLogIfError(err)
	if err != nil {
		return
	}
	args := strings.Fields(line)
	command := args[0]
	blockID := args[1]
	switch command {
	case "read":
		read(blockID, conn)
	case "store":
		store(blockID, conn, reader)
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
	go func() {
		bufio.NewReader(os.Stdin).ReadBytes('\n')
		log_init.PrintAndLog("Cerrando datanode...")
		ln.Close()
	}()

	for {
		conn, err := ln.Accept()
		log_init.PrintAndLogIfError(err)
		if err != nil {
			break
		}
		go handleConnection(conn)
	}
}
