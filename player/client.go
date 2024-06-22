package main

import (
	"bufio"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
)

type Pokemon struct {
	ID               int              `json:"id"`
	Name             string           `json:"name"`
	DAMAGE_RELATIONS DamageRelation   `json:"damage_relations"`
	POKEMON          []PokemonItem    `json:"pokemon"`
}

type DamageRelation struct {
	DOUBLE_DAMAGE_FROM []DamageRelationItem `json:"double_damage_from"`
	DOUBLE_DAMAGE_TO   []DamageRelationItem `json:"double_damage_to"`
	HALF_DAMAGE_FROM   []DamageRelationItem `json:"half_damage_from"`
	HALF_DAMAGE_TO     []DamageRelationItem `json:"half_damage_to"`
}

type DamageRelationItem struct {
	NAME string `json:"name"`
	URL  string `json:"url"`
}

type PokemonItem struct {
	POKEMON_DETAIL PokemonDetail `json:"pokemon"`
}

type PokemonDetail struct {
	NAME  string         `json:"name"`
	URL   string         `json:"url"`
	STATS []PokemonStats `json:"stats"`
  TYPES []PokemonType `json:"types"`
}

type PokemonStats struct {
	BaseStat int `json:"base_stat"`
	Effort   int `json:"effort"`
	Stat     struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"stat"`
}

type PokemonType struct {
    TYPE struct {
        Name string `json:"name"`
		URL  string `json:"url"`
    } `json:"type"`
}

// User struct to store user information and their selected Pokémon
type User struct {
	Name     string
	TypeOfPokemon       string
	Selected []PokemonDetail // Map to store selected Pokémon by type
}

func main() {
	// Connect to server
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		log.Fatalf("Failed to connect to server: %v", err)
	}
	defer conn.Close()

	// Get client name
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter your name: ")
	clientName, _ := reader.ReadString('\n')

	// Send client name to server
	_, err = conn.Write([]byte(clientName))
	if err != nil {
		log.Fatalf("Failed to send client name: %v", err)
	}

	/*  */
	// Read the available Pokémon data from the server
	pokemonData, err := readPokemonData(conn)
	if err != nil {
		log.Fatalf("Failed to read Pokémon data: %v", err)
	}

	// Display available types for the user to choose from
	fmt.Println("Available types:")
	types := make(map[int]string)
	for i, p := range pokemonData {
		fmt.Printf("%d. %s\n", i+1, p.Name)
		types[i+1] = p.Name
	}

	// Get the user's choice
	fmt.Print("Enter the number of your chosen type: ")
	choiceStr, _ := reader.ReadString('\n')
	choiceStr = strings.TrimSpace(choiceStr)
	choice, err := strconv.Atoi(choiceStr)
	if err != nil || choice < 1 || choice > len(pokemonData) {
		log.Fatalf("Invalid choice: %v", choiceStr)
	}

	// Store the chosen type
	chosenType := types[choice]
	fmt.Printf("You have chosen: %s\n", chosenType)

	// Get the Pokémon associated with the chosen type
	chosenPokemon := getPokemonByType(pokemonData, chosenType)
	// fmt.Println(chosenPokemon)

	// Print the available Pokémon for the user to choose from
	fmt.Println("Available Pokémon:")
	for i, p := range chosenPokemon {
		fmt.Printf("%d. %s\n", i+1, p.NAME)
	}

	// Select 3 Pokémon
	var selectedPokemon []PokemonDetail

	for i := 0; i < 3; i++ {
		fmt.Printf("Enter the number of your %dth chosen Pokémon: ", i+1)
		pokemonChoiceStr, _ := reader.ReadString('\n')
		pokemonChoiceStr = strings.TrimSpace(pokemonChoiceStr)
		pokemonChoice, err := strconv.Atoi(pokemonChoiceStr)
		if err != nil || pokemonChoice < 1 || pokemonChoice > len(chosenPokemon) {
			log.Fatalf("Invalid choice: %v", pokemonChoiceStr)
		}

		selectedPokemon = append(selectedPokemon, chosenPokemon[pokemonChoice-1])
	}

	// Store the selected Pokémon in the user's data
	user := User{
		Name: clientName, 
		Selected: selectedPokemon, 
		TypeOfPokemon: chosenType,
	}

	// Print selected Pokémon for each user
	// fmt.Printf("NAME: %s", user.Name)
	// fmt.Printf("TYPE: %s", user.TypeOfPokemon)
	// fmt.Println("LIST OF POKEMON:", formatDataReadable(user.Selected))

	// Send the selected Pokémon to the server
	userSelectedPokemonJSON, err := json.Marshal(user)
	if err != nil {
		log.Fatalf("Failed to marshal selected Pokémon: %v", err)
	}

	/*  */
	// Write the length of the JSON output first
	jsonLength := int32(len(userSelectedPokemonJSON))
	err = binary.Write(conn, binary.LittleEndian, jsonLength)
	if err != nil {
		log.Fatalf("Failed to write length: %v", err)
	}
	/*  */

	_, err = conn.Write(userSelectedPokemonJSON)
	if err != nil {
		log.Fatalf("Failed to send selected Pokémon to server: %v", err)
	}

	/* Start game */
	fmt.Print("Do you want to start your game: ")
	isStart, _ := reader.ReadString('\n')

	if strings.Contains(isStart, "yes") {
		fmt.Print("Please enter your name: ")
		name, _ := reader.ReadString('\n')
		_, err = conn.Write([]byte(name))
		if err != nil {
			log.Fatalf("Failed to send client name: %v", err)
		}
	}
	/* End */
}

// Function to read the Pokémon data from the server
func readPokemonData(conn net.Conn) ([]Pokemon, error) {
	var length int32
	err := binary.Read(conn, binary.LittleEndian, &length)
	if err != nil {
		return nil, fmt.Errorf("failed to read length: %v", err)
	}

	jsonData := make([]byte, length)
	_, err = io.ReadFull(conn, jsonData)
	if err != nil {
		return nil, fmt.Errorf("failed to read JSON data: %v", err)
	}

	var pokemonData []Pokemon
	err = json.Unmarshal(jsonData, &pokemonData)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON data: %v", err)
	}

	// fmt.Println(pokemonData)

	// // Marshal the JSON object to pretty-print it
	// jsonOutput, err := json.MarshalIndent(pokemonData, "", "  ")
	// if err != nil {
	// 	fmt.Println("could not marshal JSON")
	// }

	// // Print the JSON output to the terminal
	// fmt.Println(string(jsonOutput))

	return pokemonData, nil
}

// Function to get Pokémon by type
func getPokemonByType(pokemonData []Pokemon, typeName string) []PokemonDetail {
	var pokemon []PokemonDetail
	for _, p := range pokemonData {
		if p.Name == typeName {
			for _, pd := range p.POKEMON {
				pokemon = append(pokemon, pd.POKEMON_DETAIL)
			}
			break
		}
	}
	return pokemon
}

func formatDataReadable(data any) any {
	// Marshal the JSON object to pretty-print it
	jsonOutput, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		fmt.Println("could not marshal JSON")
	}

	// Print the JSON output to the terminal
	// fmt.Println(string(jsonOutput))
	return string(jsonOutput)
}