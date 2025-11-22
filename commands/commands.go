package commands

import (
	"fmt"
	"strings"
)

func ExecCommand(input string, functions map[string]func([]string)) {
	args := strings.Fields(input)
	command := args[0]
	function, ok := functions[command]
	if !ok {
		fmt.Println("Comando no reconocido:", input)
	} else {
		function(args)
	}
}
