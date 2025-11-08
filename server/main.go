package main

import (
	"fmt"
	"log"
	"net"
)

func handleConnection(conn net.Conn) {
	defer conn.Close()
	// Read from the client
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		log.Println(err)
	}

	fmt.Println(string(buf[:n]))

	fmt.Fprintf(conn, "Echo - "+string(buf[:n]))
}

func main() {
	ln, err := net.Listen("tcp", "localhost:8080")
	if err != nil {
		log.Fatalln(err)
	}

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
