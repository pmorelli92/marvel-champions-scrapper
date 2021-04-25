package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
)

func main() {

	// Get the first 3 pages of the hall of fame (90 most liked decks)
	url := "https://marvelcdb.com/decklists/halloffame"

	// Create an array to store the deck IDs
	deckIDs := make([]string, 0)

	for i := 0; i < 3; i++ {
		if i != 0 {
			url = "https://marvelcdb.com/decklists/halloffame/" + strconv.Itoa(i+1)
		}

		rs, err := http.Get(url)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		body, err := io.ReadAll(rs.Body)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		strBody := string(body)
		regex, err := regexp.Compile(`/decklist/view/.*">`)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Get the 30 decks in this page via regex in format of
		// in format of /decklist/view/1771/doctor-strange-tough-enough-heroic-ally-swarm-1.0
		matches := regex.FindAllStringSubmatch(strBody, -1)

		for _, match := range matches {
			split := strings.Split(match[0], "/")
			deckIDs = append(deckIDs, split[3])
		}
	}

	// Create a hero map to store the hero's name, this act as a safety list
	// so I know that the most popular decks cover all heroes
	heroesMap := make(map[string]int, 0)

	// Create a card map (key: id) to store the cards those decks are using
	cardsMap := make(map[string]int, 0)

	// Get each deck by ID
	for _, deckID := range deckIDs {
		url = "https://marvelcdb.com/api/public/decklist/" + deckID + ".json"

		rs, err := http.Get(url)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		body, err := io.ReadAll(rs.Body)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Unmarshall the result into the deck struct
		deck := deckStruct{}
		err = json.Unmarshal(body, &deck)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Add the hero and increment the apparitions rate by one
		heroesMap[deck.Hero] = heroesMap[deck.Hero] + 1

		// Add each deck and the quantity of apparitions
		for card, uses := range deck.Cards {
			cardsMap[card] = cardsMap[card] + uses
		}
	}

	// Create a card map (key: faction) to store the cards those decks are using
	cardsFactionMap := make(map[string][]cardStruct, 0)

	// For each card ID get the card name
	for cardID, value := range cardsMap {
		url = "https://marvelcdb.com/api/public/card/" + cardID + ".json"

		rs, err := http.Get(url)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		body, err := io.ReadAll(rs.Body)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Unmarshall the result into the deck struct
		card := cardStruct{}
		err = json.Unmarshal(body, &card)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		if card.Faction == "Hero" {
			continue
		}

		// Add the quantity
		card.TotalUses = value
		cardsFactionMap[card.Faction] = append(cardsFactionMap[card.Faction], card)
	}

	fmt.Println("-----")
	fmt.Println("HEROES INCLUDED")
	fmt.Println("-----")
	for key, value := range heroesMap {
		fmt.Println(fmt.Sprintf("%s appears %d times", key, value))
	}

	for key, value := range cardsFactionMap {
		fmt.Println("-----")
		fmt.Println(strings.ToUpper(key))
		fmt.Println("-----")
		for _, card := range value {
			fmt.Println(fmt.Sprintf("%s - %s - %s appears %d times", card.Faction, card.Type, card.Name, card.TotalUses))
		}
	}
}

type deckStruct struct {
	Hero  string         `json:"investigator_name"`
	Cards map[string]int `json:"slots"`
}

type cardStruct struct {
	Name      string `json:"real_name"`
	Type      string `json:"type_name"`
	Faction   string `json:"faction_name"`
	TotalUses int    `json:"-"`
}
