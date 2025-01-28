package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/alldo-dev/pokedexcli/internal"
	"io"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"
)

const BaseURL = "https://pokeapi.co/api/v2/location-area/"

var pokedex = make(map[string]PokemonData)

func fetchLocationAreas(url string, cache *pokecache.Cache) (*LocationAreaResponse, error) {
	if url == "" {
		url = BaseURL
	}

	// Check the cache
	if cachedData, found := cache.Get(url); found {
		fmt.Println("Cache hit!")
		var data LocationAreaResponse
		if err := json.Unmarshal(cachedData, &data); err != nil {
			return nil, err
		}
		return &data, nil
	} else {
		fmt.Println("Cache miss. Fetching from API...")
	}

	// Fetch from PokeAPI
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("failed to fetch data")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Save response to cache
	cache.Add(url, body)

	var data LocationAreaResponse
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, err
	}

	// if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
	// 	return nil, err
	// }

	return &data, nil
}

type Config struct {
	Next     *string
	Previous *string
}

type LocationAreaResponse struct {
	Count    int     `json:"count"`
	Next     *string `json:"next"`
	Previous *string `json:"previous"`
	Results  []struct {
		Name string `json:"name"`
	} `json:"results"`
}

type cliCommand struct {
	name        string
	description string
	callback    func(args []string) error
}

// Command registry
var commands = map[string]cliCommand{}

func main() {
	config := &Config{}
	cache := pokecache.NewCache(10 * time.Second)

	// Register the commands
	commands["exit"] = cliCommand{
		name:        "exit",
		description: "Exit the Pokedex",
		callback:    commandExit,
	}

	commands["help"] = cliCommand{
		name:        "help",
		description: "Displays a help message",
		callback:    commandHelp,
	}

	commands["map"] = cliCommand{
		name:        "map",
		description: "Displays the next 20 location areas",
		callback: func(args []string) error {
			return commandMap(config, cache)
		},
	}

	commands["mapb"] = cliCommand{
		name:        "mapb",
		description: "Displays the previous 20 location areas",
		callback: func(args []string) error {
			return commandMapBack(config, cache)
		},
	}

	commands["explore"] = cliCommand{
		name:        "explore",
		description: "Explore a location area to find Pokemon",
		callback: func(args []string) error {
			return commandExplore(args, cache)
		},
	}

	commands["catch"] = cliCommand{
		name:        "catch",
		description: "Attempts to catch a Pokemon by name",
		callback: func(args []string) error {
			return commandCatch(args, cache)
		},
	}

	commands["inspect"] = cliCommand{
		name:        "inspect",
		description: "Displays details about a caught Pokemon",
		callback: func(args []string) error {
			return commandInspect(args)
		},
	}

	commands["pokedex"] = cliCommand{
		name:        "pokedex",
		description: "Lists all Pokemon you have caught",
		callback: func(args []string) error {
			if len(pokedex) == 0 {
				fmt.Println("Your Pokedex is empty. Go catch some Pokemon!")
				return nil
			}

			fmt.Println("Your Pokedex:")
			for name := range pokedex {
				fmt.Printf(" - %s\n", name)
			}
			return nil
		},
	}

	// Start REPL
	fmt.Println("---- POKEDEXCLI  ----")
	scanner := bufio.NewScanner(os.Stdin)

	for {
		// Print the prompt without a newline
		fmt.Print("Pokedex > ")

		// Wait for user input
		scanner.Scan()
		input := scanner.Text()

		// Clean and split the input into words
		words := cleanInput(input)

		// If there are no words, prompt again
		if len(words) == 0 {
			continue
		}

		// Get the command and arguments
		command := words[0]
		args := words[1:]

		// Look up the command in the registry
		if cmd, exists := commands[command]; exists {
			err := cmd.callback(args)
			if err != nil {
				fmt.Printf("Error %s\n", err.Error())
			} else {
				fmt.Println("Unknown command")
			}
		}

	}
}

func cleanInput(text string) []string {
	// Trim leading/trailing whitespace, lowercase, and split by spaces
	text = strings.TrimSpace(text)
	text = strings.ToLower(text)
	return strings.Fields(text)
}

// commandExit exits the program
func commandExit(args []string) error {
	fmt.Println("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil
}

// CommandHelp displays the help message
func commandHelp(args []string) error {
	fmt.Println("Welcome to the Pokedex!")
	fmt.Println("Usage:")
	for _, cmd := range commands {
		fmt.Printf("  %s: %s\n", cmd.name, cmd.description)
	}
	return nil
}

// commandMap fetches and displays the next 20 location areas
func commandMap(config *Config, cache *pokecache.Cache) error {
	data, err := fetchLocationAreas(pointerToString(config.Next), cache)
	if err != nil {
		return err
	}

	// Print the location names
	for _, location := range data.Results {
		fmt.Println(location.Name)
	}

	// Update config with pagination URLs
	config.Next = data.Next
	config.Previous = data.Previous

	return nil
}

func commandMapBack(config *Config, cache *pokecache.Cache) error {
	if config.Previous == nil {
		fmt.Println("You're on the first page")
		return nil
	}

	data, err := fetchLocationAreas(pointerToString(config.Previous), cache)
	if err != nil {
		return err
	}

	// Print the location names
	for _, location := range data.Results {
		fmt.Println(location.Name)
	}

	// Update config with pagination URLs
	config.Next = data.Next
	config.Previous = data.Previous

	return nil
}

func pointerToString(ptr *string) string {
	if ptr == nil {
		return ""
	}
	return *ptr
}

func fetchPokemonFromLocation(location string, cache *pokecache.Cache) ([]string, error) {
	url := fmt.Sprintf("%s%s", BaseURL, location)

	// Check the cache
	if cachedData, found := cache.Get(url); found {
		fmt.Println("Cache hit!")
		var data PokemonEncounterResponse
		if err := json.Unmarshal(cachedData, &data); err != nil {
			return nil, err
		}
		return extractPokemonNames(data), nil
	} else {
		fmt.Println("Cache miss. Fetching from API...")
	}

	// Fetch from PokeAPI
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch data for location: %s", location)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Save response to cache
	cache.Add(url, body)

	// Parse response
	var data struct {
		PokemonEncounters []struct {
			Pokemon struct {
				Name string `json:"name"`
			} `json:"pokemon"`
		} `json:"pokemon_encounters"`
	}
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, err
	}

	return extractPokemonNames(data), nil
}

// Helper function to extract Pokémon names
func extractPokemonNames(data PokemonEncounterResponse) []string {
	var names []string
	for _, encounter := range data.PokemonEncounters {
		names = append(names, encounter.Pokemon.Name)
	}
	return names
}

func commandExplore(args []string, cache *pokecache.Cache) error {
	if len(args) == 0 {
		return errors.New("please provide a location area name")
	}

	location := args[0]
	fmt.Printf("Exploring %s...\n", location)

	pokemon, err := fetchPokemonFromLocation(location, cache)
	if err != nil {
		return err
	}

	if len(pokemon) == 0 {
		fmt.Println("No Pokémon found in this area.")
		return nil
	}

	fmt.Println("Found Pokémon:")
	for _, name := range pokemon {
		fmt.Printf(" - %s\n", name)
	}

	return nil
}

type PokemonEncounterResponse struct {
	PokemonEncounters []struct {
		Pokemon struct {
			Name string `json:"name"`
		} `json:"pokemon"`
	} `json:"pokemon_encounters"`
}

type PokemonData struct {
	BaseExperience int    `json:"base_experience"`
	Name           string `json:"name"`
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

func fetchPokemon(pokemonName string, cache *pokecache.Cache) (*PokemonData, error) {
	url := fmt.Sprintf("https://pokeapi.co/api/v2/pokemon/%s", pokemonName)

	// Check the cache
	if cachedData, found := cache.Get(url); found {
		fmt.Println("Cache hit!")
		var data PokemonData
		if err := json.Unmarshal(cachedData, &data); err != nil {
			return nil, err
		}
		return &data, nil
	} else {
		fmt.Println("Cache miss. Fetching from API...")
	}

	// Fetch from PokeAPI
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch data for Pokémon: %s", pokemonName)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Save response to cache
	cache.Add(url, body)

	// Parse response
	var data PokemonData
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, err
	}

	return &data, nil
}

func commandCatch(args []string, cache *pokecache.Cache) error {
	if len(args) == 0 {
		return fmt.Errorf("please provide the name of a Pokémon to catch")
	}

	pokemonName := strings.ToLower(args[0])
	fmt.Printf("Throwing a Pokeball at %s...\n", pokemonName)

	// Check if already caught
	if _, exists := pokedex[pokemonName]; exists {
		fmt.Printf("%s is already in your Pokedex!\n", pokemonName)
		return nil
	}

	// Fetch Pokémon data
	pokemon, err := fetchPokemon(pokemonName, cache)
	if err != nil {
		return fmt.Errorf("failed to fetch data for Pokémon %s: %v", pokemonName, err)
	}

	// Simulate catching the Pokémon
	catchChance := rand.Intn(100)
	requiredChance := 50 + pokemon.BaseExperience/10

	if catchChance < requiredChance {
		fmt.Printf("%s escaped!\n", pokemonName)
		return nil
	}

	// Add Pokémon to the Pokedex
	pokedex[pokemonName] = *pokemon
	fmt.Printf("%s was caught!\n", pokemonName)
	return nil
}

func commandInspect(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("please provide the name of a Pokémon to inspect")
	}

	pokemonName := strings.ToLower(args[0])

	// Check if the Pokémon is in the Pokedex
	pokemon, caught := pokedex[pokemonName]
	if !caught {
		fmt.Printf("You have not caught %s\n", pokemonName)
		return nil
	}

	// Print Pokémon details
	fmt.Printf("Name: %s\n", pokemon.Name)
	fmt.Printf("Height: %d\n", pokemon.Height)
	fmt.Printf("Weight: %d\n", pokemon.Weight)

	// Print stats
	fmt.Println("Stats:")
	for _, stat := range pokemon.Stats {
		fmt.Printf("  - %s: %d\n", stat.Stat.Name, stat.BaseStat)
	}

	// Print types
	fmt.Println("Types:")
	for _, t := range pokemon.Types {
		fmt.Printf("  - %s\n", t.Type.Name)
	}

	return nil
}
