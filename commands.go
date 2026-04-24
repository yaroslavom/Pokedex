package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
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

type LocAreas struct {
	Next     *string `json:"next"`
	Previous *string `json:"previous"`
	Results  []struct {
		Name string `json:"name"`
	} `json:"results"`
}

type LocAreaId struct {
	PokemonEncounters []struct {
		Pokemon struct {
			Name string `json:"name"`
		} `json:"pokemon"`
	} `json:"pokemon_encounters"`
}

type Pokemon struct {
	Name           string `json:"name"`
	BaseExperience int    `json:"base_experience"`
	Height         int    `json:"height"`
	Weight         int    `json:"weight"`
	Stats          []struct {
		BaseStat int `json:"base_stat"`
		Stat     struct {
			Name string `json:"name"`
		} `json:"stat"`
	} `json:"stats"`
	Types []struct {
		Type struct {
			Name string `json:"name"`
		} `json:"type"`
	} `json:"types"`
}

func GetCommand() map[string]cliCommand {
	return map[string]cliCommand{
		"help": {
			name:        "help",
			description: "Displays a help message",
			callback:    commandHelp,
		},
		"exit": {
			name:        "exit",
			description: "Exit the Pokédex",
			callback:    commandExit,
		},
		"map": {
			name:        "map",
			description: "Get the next page of locations",
			callback:    commandMap,
		},
		"mapb": {
			name:        "mapb",
			description: "Get the previous page of locations",
			callback:    commandMapb,
		},
		"explore": {
			name:        "explore",
			description: "Explore a location area to see its Pokémon",
			callback:    commandExplore,
		},
		"catch": {
			name:        "catch",
			description: "Attempt to catch a Pokémon",
			callback:    commandCatch,
		},
		"inspect": {
			name:        "inspect",
			description: "View details about a caught Pokémon",
			callback:    commandInspect,
		},
		"pokedex": {
			name:        "pokedex",
			description: "List all Pokémon in your Pokédex",
			callback:    commandPokedex,
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

func moveMap(cfg *config, url string) error {
	areas, err := fetchData[LocAreas](url, cfg.Cache)
	if err != nil {
		return err
	}
	cfg.Nav.Next = areas.Next
	cfg.Nav.Previous = areas.Previous
	for _, v := range areas.Results {
		fmt.Println(v.Name)
	}
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
	if cfg.Nav.Next == nil {
		return errors.New("You are on the last page")
	}
	moveMap(cfg, *cfg.Nav.Next)
	return nil
}

func commandMapb(cfg *config, param string) error {
	if cfg.Nav.Previous == nil {
		return errors.New("You are on the first page")
	}
	moveMap(cfg, *cfg.Nav.Previous)
	return nil
}

func commandExplore(cfg *config, param string) error {
	if param == "" {
		return fmt.Errorf("Provide a location name")
	}
	fmt.Printf("Exploring %s...\n", param)
	url := cfg.Nav.BaseUrl + cfg.Nav.LocationAreas + param
	areaId, err := fetchData[LocAreaId](url, cfg.Cache)
	if err != nil {
		return err
	}
	if len(areaId.PokemonEncounters) > 0 {
		fmt.Println("Found Pokémon:")
		for _, v := range areaId.PokemonEncounters {
			fmt.Println(v.Pokemon.Name)
		}
	} else {
		fmt.Println("No Pokémon found in this area")
	}
	return nil
}

func commandCatch(cfg *config, param string) error {
	if param == "" {
		return fmt.Errorf("Provide a Pokémon name")
	}
	fmt.Printf("Throwing a Pokeball at %s...\n", param)
	url := cfg.Nav.BaseUrl + cfg.Nav.Pokemon + param
	pok, err := fetchData[Pokemon](url, cfg.Cache)
	if err != nil {
		return err
	}
	// BaseExperience range: 36-608xp
	// Possible to implement the logic that the more Pokémon there are, the lower the difficulty becomes
	difficulty := pok.BaseExperience / 3
	if difficulty > 95 {
		difficulty = 95
	}
	roll := rand.Intn(100)
	if roll > difficulty {
		fmt.Printf("%s was caught!\n", param)
		cfg.Pokedex[param] = pok
	} else {
		fmt.Printf("%s escaped!\n", param)
		fmt.Println("You may now inspect it with the inspect command.")
	}
	return nil
}

func commandInspect(cfg *config, param string) error {
	if param == "" {
		return fmt.Errorf("Provide a Pokémon name")
	}
	pok, ok := cfg.Pokedex[param]
	if !ok {
		fmt.Printf("You have not caught %s yet\n", param)
		return nil
	}
	fmt.Printf("Name: %s\n", pok.Name)
	fmt.Printf("Height: %.1fm\n", float64(pok.Height)/10)
	fmt.Printf("Weight: %.1fkg\n", float64(pok.Weight)/10)
	fmt.Println("Stats:")
	for _, s := range pok.Stats {
		fmt.Printf("  -%s: %v\n", s.Stat.Name, s.BaseStat)
	}
	fmt.Println("Types:")
	for _, t := range pok.Types {
		fmt.Printf("  - %s\n", t.Type.Name)
	}
	return nil
}

func commandPokedex(cfg *config, param string) error {
	if !(len(cfg.Pokedex) > 0) {
		fmt.Printf("Your Pokedex is empty.")
		return nil
	}
	fmt.Println("Your Pokedex:")
	for k := range cfg.Pokedex {
		fmt.Printf("  - %s\n", k)
	}
	return nil
}
