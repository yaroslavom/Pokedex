package main

import (
	"bufio"
	"fmt"
	"os"
	"pokedexcli/internal/pokecache"
	"strings"
	"time"
)

type config struct {
	Next     *string
	Previous *string
	Cache    *pokecache.Cache
}

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
	baseUrl := "https://pokeapi.co/api/v2/location-area"
	cache := pokecache.NewCache(25 * time.Second)
	cfg := config{Next: &baseUrl, Previous: nil, Cache: &cache}
	commands := GetCommand()

	fmt.Println("Welcome to the Pokedex!")
	for {
		fmt.Print("Pokedex > ")
		if !scanner.Scan() {
			break
		}

		if scanner.Err() != nil {
			scanner.Err()
		}

		input := cleanInput(scanner.Text())
		if len(input) > 0 {
			// fmt.Printf("Your command was: %v\n", input[0])
			if _, ok := commands[input[0]]; ok {
				err := commands[input[0]].callback(&cfg)
				if err != nil {
					fmt.Println(err)
				}
			} else {
				fmt.Println("Unknown command")
			}
		}
	}
}
