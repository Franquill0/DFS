package main

import (
	"fmt"
	"net"
	"strings"
)

func execCommand(input string, functions map[string]func([]string, net.Conn), conn net.Conn) {
	args := strings.Fields(input)
	command := args[0]
	function, ok := functions[command]
	if !ok {
		fmt.Println("Comando no reconocido:", input)
	} else {
		function(args, conn)
	}
}
