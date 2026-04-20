package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func cleanInput(text string) []string {
	var fields []string
	fields = strings.Fields(text)
	for i, s := range fields {
		fields[i] = strings.ToLower(strings.TrimSpace(s))
	}
	return fields
}

func startRepl() {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Welcome to the Pokedex!")
	for {
		// fmt.Print("Pokedex > ")
		if !scanner.Scan() {
			break
		}

		input := cleanInput(scanner.Text())
		if scanner.Err() != nil {
			scanner.Err()
		}
		if len(input) > 0 {
			// fmt.Printf("Your command was: %v\n", input[0])
			commands := GetCommand()
			if _, ok := commands[input[0]]; ok {
				commands[input[0]].callback()
				// if err != nil {
				// 	fmt.Println(err)
				// }
			} else {
				fmt.Print("Unknown command")
			}
		}
	}
}
