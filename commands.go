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
	callback    func(*config) error
}

type LocationResponse struct {
	Count    int     `json:"count"`
	Next     *string `json:"next"`
	Previous *string `json:"previous"`
	Results  []struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"results"`
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
			fmt.Printf("Response failed with status code: %d and\nbody: %s\n", resp.StatusCode, body)
			return data, err
		}

		c.Add(url, body)
	}
	if err := json.Unmarshal(body, &data); err != nil {
		return data, err
	}
	return data, nil
}

func moveMap(cfg *config, url *string) error {
	response, err := fetchData[LocationResponse](*url, cfg.Cache)
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

func commandExit(cfg *config) error {
	fmt.Print("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil
}

func commandHelp(cfg *config) error {
	fmt.Println("Usage:")
	commands := GetCommand()
	for _, command := range commands {
		fmt.Printf("%s: %s\n", command.name, command.description)
	}
	return nil
}

func commandMap(cfg *config) error {
	if cfg.Next == nil {
		return errors.New("You are on the last page")
	}
	moveMap(cfg, cfg.Next)
	return nil
}

func commandMapb(cfg *config) error {
	if cfg.Previous == nil {
		return errors.New("You are on the first page")
	}
	moveMap(cfg, cfg.Previous)
	return nil
}
