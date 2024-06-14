package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type pokemon struct {
	Name           string   `json:"name"`
	Type           []string `json:"types"`
	ID             int      `json:"national_id"`
	Attack         int      `json:"attack"`
	Defense        int      `json:"defense"`
	SpecialAttack  int      `json:"sp_atk"`
	SpecialDefense int      `json:"sp_def"`
	Speed          int      `json:"speed"`
	HP             int      `json:"hp"`
	Exp            int      `json:"exp"`
	From           int      `json:"from"`
	FromLevel      int      `json:"from_level"`
	To             int      `json:"to"`
	ToLevel        int      `json:"to_level"`
}

type playerPokemon struct {
	Name    string  `json:"playername"`
	Pokemon pokemon `json:"pokemon"`
}

type evolution struct {
	From      int `json:"from"`
	To        int `json:"to"`
	FromLevel int `json:"level"`
	ToLevel   int `json:"level"`
}

func main() {
	/*
		linkArr := [17]string{
			"https://pokedex.org/assets/skim-monsters-1.txt",
			"https://pokedex.org/assets/descriptions-1.txt",
			"https://pokedex.org/assets/evolutions.txt",
			"https://pokedex.org/assets/types.txt",
			"https://pokedex.org/assets/monsters-supplemental-1.txt",
			"https://pokedex.org/assets/descriptions-2.txt",
			"https://pokedex.org/assets/monsters-supplemental-2.txt",
			"https://pokedex.org/assets/skim-monsters-2.txt",
			"https://pokedex.org/assets/descriptions-3.txt",
			"https://pokedex.org/assets/monsters-supplemental-3.txt",
			"https://pokedex.org/assets/skim-monsters-3.txt",
			"https://pokedex.org/assets/moves-1.txt",
			"https://pokedex.org/assets/moves-2.txt",
			"https://pokedex.org/assets/moves-3.txt",
			"https://pokedex.org/assets/monster-moves-1.txt",
			"https://pokedex.org/assets/monster-moves-2.txt",
			"https://pokedex.org/assets/monster-moves-3.txt",
		}
	*/

	//crawPokemon()

	// get all pokemon from pokedex.json
	file, err := os.Open("pokedex.json")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer file.Close()

	// Decode the JSON data into a slice of Pokemon
	var pokemons []pokemon
	err = json.NewDecoder(file).Decode(&pokemons)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Create a map to store the Pokemon data
	pokemonMap := make(map[string]pokemon)

	// Populate the map with the Pokemon data
	for _, pokemon := range pokemons {
		pokemonMap[pokemon.Name] = pokemon
	}

	// Print the Pokemon data in the map
	for _, pokemon := range pokemonMap {
		fmt.Println(pokemon)
	}

}

func crawPokemon() {

	pokemon := getPokemon("https://pokedex.org/assets/skim-monsters-1.txt")
	pokemon = append(pokemon, getPokemon("https://pokedex.org/assets/skim-monsters-2.txt")...)
	pokemon = append(pokemon, getPokemon("https://pokedex.org/assets/skim-monsters-3.txt")...)

	jsonData, err := json.MarshalIndent(pokemon, "", "  ")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	err = os.WriteFile("pokedex.json", jsonData, 0644)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println("Pokemon data saved to pokedex.json")

}

func getPokemon(link string) []pokemon {
	url := link
	method := "GET"

	pokemons := []pokemon{}
	evolutions := getEvolutions()
	exps := getExp()

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		fmt.Println(err)
		return pokemons
	}
	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return pokemons
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return pokemons
	}

	bodyArr := strings.Split(string(body), "descriptions")
	bodyArr = bodyArr[1:]
	for _, v := range bodyArr {
		pokemon := pokemon{}
		pokemonName := strings.Split(v, "male_female_ratio")[0]
		pokemonName = strings.Split(pokemonName, "\"")[len(strings.Split(pokemonName, "\""))-3]
		pokemon.Name = pokemonName
		pokemonTypes := getSubstringBetween(v, "types", "attack")
		pokemonTypesArr := strings.Split(pokemonTypes, "\"name\":\"")
		pokemonTypesArr = pokemonTypesArr[1:]
		for _, v := range pokemonTypesArr {
			pokemon.Type = append(pokemon.Type, strings.Split(v, "\"")[0])
		}
		pokemonId := getSubstringBetween(v, "national_id\":", ",")
		pokemon.ID, _ = strconv.Atoi(pokemonId)

		pokemon.Attack, _ = strconv.Atoi(getSubstringBetween(v, "attack\":", ","))
		pokemon.Defense, _ = strconv.Atoi(getSubstringBetween(v, "defense\":", ","))
		pokemon.SpecialAttack, _ = strconv.Atoi(getSubstringBetween(v, "sp_atk\":", ","))
		pokemon.SpecialDefense, _ = strconv.Atoi(getSubstringBetween(v, "sp_def\":", ","))
		pokemon.Speed, _ = strconv.Atoi(getSubstringBetween(v, "speed\":", ","))
		pokemon.HP, _ = strconv.Atoi(getSubstringBetween(v, "hp\":", ","))
		pokemon.From = evolutions[pokemon.ID].From
		pokemon.FromLevel = evolutions[pokemon.ID].FromLevel
		pokemon.To = evolutions[pokemon.ID].To
		pokemon.ToLevel = evolutions[pokemon.ID].ToLevel
		pokemon.Exp = exps[pokemon.ID]

		pokemons = append(pokemons, pokemon)
	}

	return pokemons
}

func getEvolutions() map[int]evolution {

	evolutions := map[int]evolution{}

	url := "https://pokedex.org/assets/evolutions.txt"
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		fmt.Println(err)
		return nil
	}
	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	to := strings.Split(string(body), "\"to\"")
	to = to[1:]
	for _, v := range to {
		evolution := evolution{}
		evolution.To, _ = strconv.Atoi(getSubstringBetween(v, "nationalId\":", ","))
		evolution.ToLevel, _ = strconv.Atoi(getSubstringBetween(v, "level\":", "}"))
		pokemonId, _ := strconv.Atoi(getSubstringBetween(v, "_id\":\"", "\""))
		evolutions[pokemonId] = evolution
	}

	from := strings.Split(string(body), "\"from\"")
	from = from[1:]
	for _, v := range from {
		evolution := evolution{}
		evolution.From, _ = strconv.Atoi(getSubstringBetween(v, "nationalId\":", ","))
		evolution.FromLevel, _ = strconv.Atoi(getSubstringBetween(v, "level\":", "}"))
		pokemonId, _ := strconv.Atoi(getSubstringBetween(v, "_id\":\"", "\""))

		if _, ok := evolutions[pokemonId]; ok {
			evolution.To = evolutions[pokemonId].To
			evolution.ToLevel = evolutions[pokemonId].ToLevel
			evolutions[pokemonId] = evolution
		} else {
			evolutions[pokemonId] = evolution
		}
	}

	return evolutions
}

func getExp() map[int]int {

	exps := map[int]int{}

	url := "https://bulbapedia.bulbagarden.net/wiki/List_of_Pok%C3%A9mon_by_effort_value_yield_(Generation_IX)"
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		fmt.Println(err)
		return nil
	}
	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if err != nil {
		fmt.Println(err)
		return nil

	}
	doc.Find("table.sortable tbody tr").Each(func(i int, s *goquery.Selection) {
		name := s.Find("td.r").Text()
		exp := s.Find("td:nth-child(4)").Text()

		name = strings.TrimSpace(name)
		exp = strings.TrimSpace(exp)
		nameInt, _ := strconv.Atoi(name)
		expInt, _ := strconv.Atoi(exp)

		exps[nameInt] = expInt
	})

	return exps
}

func getSubstringBetween(str, start, end string) string {
	startIndex := strings.Index(str, start)
	if startIndex == -1 {
		return ""
	}

	startIndex += len(start)
	endIndex := strings.Index(str[startIndex:], end)
	if endIndex == -1 {
		return ""
	}

	return str[startIndex : startIndex+endIndex]
}

func capturePokemon(playerName string, pokemon pokemon) string {
	// open playerpokemon.json file
	file, err := ioutil.ReadFile("playerpokemon.json")
	// if file not found, create a new file
	if err != nil {
		file, err := os.Create("playerpokemon.json")
		if err != nil {
			return "Error creating file: " + err.Error()
		}
		defer file.Close()

		// Write the data to the file
		_, err = file.Write([]byte("your data here"))
		if err != nil {
			return "Error writing to file: " + err.Error()
		}

		return "File created and data written successfully"
	}

	// read the file into a struct
	players := []playerPokemon{}
	err = json.Unmarshal(file, &players)
	// if there is an error, return the error
	if err != nil {
		return "Error unmarshalling file: " + err.Error()
	}

	// return if that player already has a pokemon with the same name
	for _, v := range players {
		if v.Name == playerName {
			if v.Pokemon.Name == pokemon.Name {
				return "You already have this pokemon"
			}
		}
	}

	// recalculate pokemon attribute
	rand.Seed(time.Now().UnixNano())

	// Generate a random float64 between 0.5 and 1
	ev := 0.5 + rand.Float64()*0.5
	pokemon.Exp = int(float64(pokemon.Exp) * ev)
	pokemon.SpecialDefense = int(float64(pokemon.SpecialDefense) * ev)
	pokemon.SpecialAttack = int(float64(pokemon.SpecialAttack) * ev)
	pokemon.Defense = int(float64(pokemon.Defense) * ev)
	pokemon.Attack = int(float64(pokemon.Attack) * ev)
	pokemon.HP = int(float64(pokemon.HP) * ev)

	players = append(players, playerPokemon{
		Name:    playerName,
		Pokemon: pokemon,
	})

	// write the updated struct to the file
	data, err := json.MarshalIndent(players, "", "  ")
	// if there is an error, return the error
	if err != nil {
		return "Error marshalling data: " + err.Error()
	}
	// write the updated struct to the file
	err = ioutil.WriteFile("playerpokemon.json", data, 0644)
	// if there is an error, return the error
	if err != nil {
		return "Error writing to file: " + err.Error()

	}
	return "Pokemon captured successfully"
}
