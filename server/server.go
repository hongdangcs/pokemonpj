package main

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

var pokedex []Pokemon

var LIMIT_DATA int = 10

type User struct {
	ID string
	NAME string
}

// Global slice to store connected users
var users []User
var mu sync.Mutex

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
type PokemonOfUser struct {
	Name     string
	TypeOfPokemon       string
	Selected []PokemonDetail // Map to store selected Pokémon by type
}

type Lobby struct {
	ID int
	PLAYERS []PokemonOfUser
}

var lobby Lobby

func main() {
	_, err := ioutil.ReadFile("pokedex.json")

	if err != nil {
		fmt.Println("pokedex.json file is not exist")
		fetchingData()
	} else {
		fmt.Println("pokedex.json file is exist")

		// Load Pokémon data from pokedex.json
    // loadPokedex()
	}

	// Start TCP server
	addr := ":8080"
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
	defer listener.Close()
	log.Printf("Server listening on %s\n", addr)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept connection: %v", err)
			continue
		}
		go handleClient(conn)
	}
}

func handleClient(conn net.Conn) {
	defer conn.Close()

	// Read client name
	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		log.Printf("Failed to read client name: %v", err)
		return
	}
	clientName := string(buffer[:n])

	// Get the client's IP address
	clientAddr := conn.RemoteAddr().String()
	// fmt.Printf("Client %s connected from IP: %s\n", clientName , extractPort(clientAddr))

	// Add the client's information to the global slice
	mu.Lock()
	users = append(users, User{ID: extractPort(clientAddr), NAME: clientName})
	mu.Unlock()

	// Print the client's information and the list of all connected users
	printUsers()

	trimValue := strings.TrimSpace(clientName)

	// Send random 3 Pokemon to client
	sendRandomPokemon(trimValue, conn)

	readPokemonOfUser(conn)

	// handleStartGame(conn) 
}

// func handleStartGame(conn net.Conn) {
// 	// Read client name
// 	isStart := make([]byte, 1024)
// 	n, err := conn.Read(isStart)
// 	if err != nil {
// 		log.Printf("Failed to read client name: %v", err)
// 		return
// 	}
// 	clientAccept := string(isStart[:n])
// 	fmt.Println(clientAccept)

// 	// Filter users by name "Alice"
// 	filteredUsers := filterUsersByName(users, clientAccept)
// 	fmt.Print(filteredUsers)
// 	userJoin := filteredUsers[0]
// 	var listOfPlayers []
// 	lobby.ID = randomNumber()
// 	lobby.PLAYERS = append(lobby.PLAYERS, userJoin)

// 	// Print filtered users
// 	// for _, user := range filteredUsers {
// 	// 	fmt.Printf("ID: %d, NAME: %s\n", user.ID, user.NAME)
// 	// }
// }

func randomNumber() int {
	// Seed the random number generator using the current time
	rand.Seed(time.Now().UnixNano())

	// Generate a random number in the range 100 to 999
	randomNumber := rand.Intn(900) + 100

	return randomNumber
}

// filterUsersByName filters the users by their name
func filterUsersByName(users []User, name string) []User {
	var filteredUsers []User
	for _, user := range users {
		if strings.Contains(user.NAME, name) {
			filteredUsers = append(filteredUsers, user)
		}
	}
	return filteredUsers
}

func readPokemonOfUser(conn net.Conn) {
	var length int32
	err := binary.Read(conn, binary.LittleEndian, &length)
	if err != nil {
		fmt.Printf("failed to read length: %v", err)
	}

	userSent := make([]byte, length)
	_, err = io.ReadFull(conn, userSent)
	if err != nil {
		fmt.Printf("failed to read JSON data: %v", err)
	}

	var pokemonOfUser PokemonOfUser
	err = json.Unmarshal(userSent, &pokemonOfUser)
	if err != nil {
		fmt.Printf("failed to unmarshal JSON data: %v", err)
	}
	fileName := pokemonOfUser.Name + ".json"
	// fmt.Println(fileName)

	// Call the function to save the JSON to a file
	err = saveUserPokemonFile(pokemonOfUser, removeString(fileName))
	if err != nil {
			fmt.Println("Error:", err)
			return
	}

	fmt.Println("JSON data successfully written to file:", removeString(fileName))
}

func saveUserPokemonFile(pokemon PokemonOfUser, filename string) error {
	fmt.Println(filename)
	// Marshal the struct to JSON
	jsonData, err := json.Marshal(pokemon)
	if err != nil {
			return fmt.Errorf("error marshalling JSON: %w", err)
	}

	// Write the JSON to a file
	file, err := os.Create(filename)
	if err != nil {
			return fmt.Errorf("error creating file: %w", err)
	}
	defer file.Close()

	_, err = file.Write(jsonData)
	if err != nil {
			return fmt.Errorf("error writing to file: %w", err)
	}

	return nil
}

func sendRandomPokemon(clientName string, conn net.Conn) {
	// Read the JSON file
	data, err := ioutil.ReadFile("pokedex.json")
	if err != nil {
			fmt.Println("Error:", err)
			return
	}

	// Define a variable to hold the decoded data
	var pokemonData []Pokemon

	// Unmarshal the JSON data into the struct
	if err := json.Unmarshal(data, &pokemonData); err != nil {
			fmt.Println("Error:", err)
			return
	}

	// Encode Pokemon data into JSON bytes
	jsonData, err := json.Marshal(pokemonData)
	if err != nil {
			fmt.Println("Error:", err)
			return
	}

	/*  */
	// Write the length of the JSON output first
	jsonLength := int32(len(jsonData))
	err = binary.Write(conn, binary.LittleEndian, jsonLength)
	if err != nil {
		log.Fatalf("Failed to write length: %v", err)
	}
	/*  */

	// Write Pokemon data to client
	_, err = conn.Write(/* jsonOutput */jsonData)
	if err != nil {
		log.Printf("Failed to send Pokemon data to client: %v", err)
		return
	}
}

// Function to print the list of users beautifully
func printUsers() {
	mu.Lock()
	defer mu.Unlock()

	fmt.Println("Current connected users:")
	for i, user := range users {
		fmt.Printf("%d. Name: %s, IP: %s\n", i+1, user.NAME, user.ID)
	}
}

func extractPort(address string) string {
	// Split the string by the colon
	parts := strings.Split(address, ":")
	// The last part is the port
	port := parts[len(parts)-1]
	return port
}

func loadPokedex() {
    file, err := os.Open("pokedex.json")
    if err != nil {
        log.Fatalf("Failed to open pokedex.json: %v", err)
    }
    defer file.Close()

    data, err := ioutil.ReadAll(file)
    if err != nil {
        log.Fatalf("Failed to read pokedex.json: %v", err)
    }

    if err := json.Unmarshal(data, &pokedex); err != nil {
        log.Fatalf("Failed to unmarshal pokedex.json: %v", err)
    }
}

func fetchGetPokemons(url string) Pokemon {
    // Make the GET request
    resp, err := http.Get(url)

    if err != nil {
        log.Fatalf("Failed to make request: %v", err)
    }
    defer resp.Body.Close()

    // Read the response body
    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        log.Fatalf("Failed to read response body: %v", err)
    }

    // Parse the JSON response
    var typeResponse Pokemon
    err = json.Unmarshal(body, &typeResponse)
    if err != nil {
        log.Fatalf("Failed to unmarshal JSON: %v", err)
    }

    // Limit the number of Pokemon items to the first three
    if len(typeResponse.POKEMON) > LIMIT_DATA {
        typeResponse.POKEMON = typeResponse.POKEMON[:LIMIT_DATA]
    }

    return typeResponse
}

func fetchGetStatPokemon(pokemon PokemonItem) PokemonDetail {
    // Make a GET request to the Pokemon URL
    pokemonResp, err := http.Get(pokemon.POKEMON_DETAIL.URL)
    if err != nil {
        log.Fatalf("Failed to make request to Pokemon URL: %v", err)
    }
    defer pokemonResp.Body.Close()

    // Read the response body
    pokemonBody, err := ioutil.ReadAll(pokemonResp.Body)
    if err != nil {
        log.Fatalf("Failed to read response body: %v", err)
    }

    // Parse the JSON response to get stats
    var pokemonDetail PokemonDetail
    err = json.Unmarshal(pokemonBody, &pokemonDetail)
    if err != nil {
        log.Fatalf("Failed to unmarshal Pokemon detail JSON: %v", err)
    }

    return pokemonDetail
}

func fetchingData() {
	types := []string{"grass", "fire", "water"}

	// Create a slice to store all the Pokemon data
	var allPokemon []Pokemon
    
	for _, item := range types {
		url := fmt.Sprintf("https://pokeapi.co/api/v2/type/%s", item)
        typeResponse := fetchGetPokemons(url)
        
		// Loop through each Pokemon item
		for index, pokemon := range typeResponse.POKEMON {
            pokemonDetail := fetchGetStatPokemon(pokemon)

			// Append the stats to the Pokemon item
			typeResponse.POKEMON[index].POKEMON_DETAIL.STATS = pokemonDetail.STATS
            typeResponse.POKEMON[index].POKEMON_DETAIL.TYPES = pokemonDetail.TYPES
		}

		// Append the Pokemon data to the slice
		allPokemon = append(allPokemon, typeResponse)
	}

	// Marshal all the Pokemon data into JSON
	allPokemonJSON, err := json.MarshalIndent(allPokemon, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal Pokemon to JSON: %v", err)
	}

	// Write JSON data to a file
	err = ioutil.WriteFile("pokedex.json", allPokemonJSON, 0644)
	if err != nil {
		log.Fatalf("Failed to write JSON to file: %v", err)
	}

	log.Println("All Pokemon data saved to pokedex.json")

	loadPokedex()
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

func removeString(str string) string {
	str = strings.ReplaceAll(str, "\r", "")
  str = strings.ReplaceAll(str, "\n", "")

	return str
}