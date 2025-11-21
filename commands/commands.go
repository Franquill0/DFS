package commands

import (
	"fmt"
	"strings"
)

func ExecCommand(input string, functions map[string]func([]string)) {
	trimmedInput := strings.TrimSpace(input)
	args := strings.Split(trimmedInput, " ")
	command := strings.ToLower(args[0])
	function, ok := functions[command]
	if !ok {
		fmt.Println("Comando no reconocido:", trimmedInput)
	} else {
		function(args)
	}
}
