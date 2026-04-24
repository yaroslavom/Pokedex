package main

import (
	"bufio"
	"fmt"
	"os"
	"pokedexcli/internal/pokecache"
	"strings"
	"time"
)

type navigation struct {
	BaseUrl       string
	LocationAreas string
	Pokemon       string
	Next          *string
	Previous      *string
}

type config struct {
	Cache   *pokecache.Cache
	Nav     navigation
	Pokedex map[string]Pokemon
}

func ptr[T any](v T) *T {
	return &v
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
	baseUrl := "https://pokeapi.co/api/v2/"
	cache := pokecache.NewCache(25 * time.Second)
	nav := navigation{BaseUrl: baseUrl, LocationAreas: "location-area/", Pokemon: "pokemon/", Next: ptr(baseUrl + "location-area/"), Previous: nil}
	cfg := config{Nav: nav, Cache: &cache, Pokedex: make(map[string]Pokemon)}
	commands := GetCommand()

	fmt.Println("Welcome to the Pokedex!")
	for {
		fmt.Print("Pokedex > ")
		if !scanner.Scan() {
			break
		}

		if scanner.Err() != nil {
			fmt.Println(scanner.Err())
		}

		input := cleanInput(scanner.Text())
		if len(input) > 0 {
			if _, ok := commands[input[0]]; ok {
				var param string
				if len(input) > 1 {
					param = input[1]
				}
				err := commands[input[0]].callback(&cfg, param)
				if err != nil {
					fmt.Println(err)
				}
			} else {
				fmt.Println("Unknown command")
			}
		}
	}
}
