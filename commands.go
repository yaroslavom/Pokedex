package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
)

type cacheState interface {
	Get(string) ([]byte, bool)
	Add(string, []byte)
}
type cliCommand struct {
	name        string
	description string
	callback    func(c *config, param string) error
}

type LocAreasResponse struct {
	Next     *string `json:"next"`
	Previous *string `json:"previous"`
	Results  []struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"results"`
}

type LocAreaIdResponse struct {
	PokemonEncounters []struct {
		Pokemon struct {
			Name string `json:"name"`
		} `json:"pokemon"`
	} `json:"pokemon_encounters"`
}

func GetCommand() map[string]cliCommand {
	return map[string]cliCommand{
		"exit": {
			name:        "exit",
			description: "Exit the Pokedex",
			callback:    commandExit,
		},
		"help": {
			name:        "help",
			description: "Shows available commands",
			callback:    commandHelp,
		},
		"map": {
			name:        "map",
			description: "Displays the names of 20 location areas",
			callback:    commandMap,
		},
		"mapb": {
			name:        "mapb",
			description: "Displays the names of 20 previous location areas",
			callback:    commandMapb,
		},
		"explore": {
			name:        "explore",
			description: "Displays list of Pokémon that can be encountered in this area",
			callback:    commandExplore,
		},
	}
}

func fetchData[T any](url string, c cacheState) (T, error) {
	var data T
	body, isCached := c.Get(url)

	if !isCached {
		resp, err := http.Get(url)
		if err != nil {
			return data, err
		}
		body, err = io.ReadAll(resp.Body)
		if err != nil {
			return data, err
		}
		defer resp.Body.Close()

		if resp.StatusCode > 299 {
			return data, fmt.Errorf("Response failed with status code: %d and\nbody: %s\n", resp.StatusCode, body)
		}

		c.Add(url, body)
	}
	if err := json.Unmarshal(body, &data); err != nil {
		return data, err
	}
	return data, nil
}

func moveMap(cfg *config, url *string) error {
	response, err := fetchData[LocAreasResponse](*url, cfg.Cache)
	if err != nil {
		return err
	}
	if len(response.Results) > 0 {
		for _, v := range response.Results {
			fmt.Println(v.Name)
		}
	}
	cfg.Next = response.Next
	cfg.Previous = response.Previous
	return nil
}

func commandExit(cfg *config, param string) error {
	fmt.Print("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil
}

func commandHelp(cfg *config, param string) error {
	fmt.Println("Usage:")
	commands := GetCommand()
	for _, command := range commands {
		fmt.Printf("%s: %s\n", command.name, command.description)
	}
	return nil
}

func commandMap(cfg *config, param string) error {
	if cfg.Next == nil {
		return errors.New("You are on the last page")
	}
	moveMap(cfg, cfg.Next)
	return nil
}

func commandMapb(cfg *config, param string) error {
	if cfg.Previous == nil {
		return errors.New("You are on the first page")
	}
	moveMap(cfg, cfg.Previous)
	return nil
}

func commandExplore(cfg *config, param string) error {
	if param == "" {
		return fmt.Errorf("Provide a location name")
	}
	fmt.Printf("Exploring %s...\n", param)
	url := cfg.BaseUrl + param
	response, err := fetchData[LocAreaIdResponse](url, cfg.Cache)
	if err != nil {
		return err
	}
	if len(response.PokemonEncounters) > 0 {
		fmt.Println("Found Pokémon:")
		for _, v := range response.PokemonEncounters {
			fmt.Println(v.Pokemon.Name)
		}
	} else {
		fmt.Println("No Pokémon found in this area")
	}
	return nil
}
