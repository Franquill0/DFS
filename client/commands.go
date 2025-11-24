package main

import (
	"fmt"
	"strings"
)

func execCommand(input string, functions map[string]func([]string)) {
	args := strings.Fields(input)
	command := args[0]
	function, ok := functions[command]
	if !ok {
		fmt.Print("Comando no reconocido:", input)
	} else {
		function(args)
	}
}
