package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"sort"
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

	fmt.Println("-----")
	fmt.Println("HEROES INCLUDED")
	fmt.Println("-----")
	for key, value := range heroesMap {
		fmt.Println(fmt.Sprintf("%s appears %d times", key, value))
	}

	// Get all the cards
	url = "https://marvelcdb.com/api/public/cards/?_format=json"

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
	var cards []cardStruct
	err = json.Unmarshal(body, &cards)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Get different arrays so we order and print them separately
	var basicCards []cardStruct
	var justiceCards []cardStruct
	var aggressionCards []cardStruct
	var leadershipCards []cardStruct
	var protectionCards []cardStruct

	// For each card add it to corresponding array with the number of usages
	for _, card := range cards {

		// Skip duplicate cards, like Energy (that is tagged with different id for different packs)
		if card.DuplicateOf != "" {
			continue
		}

		// Wasp's Swarm Tactics is not marked as a duplicate.
		// Decks are using AntMan's Swarm Tactics instead.
		if card.Code == "13020" {
			continue
		}

		// Assign number of usages
		card.TotalUses = cardsMap[card.Code]

		switch card.Faction {
		case "Basic":
			basicCards = append(basicCards, card)
		case "Justice":
			justiceCards = append(justiceCards, card)
		case "Aggression":
			aggressionCards = append(aggressionCards, card)
		case "Protection":
			protectionCards = append(protectionCards, card)
		case "Leadership":
			leadershipCards = append(leadershipCards, card)
		}
	}

	// Order all the arrays
	sort.SliceStable(basicCards, func(i, j int) bool {
		return basicCards[i].TotalUses > basicCards[j].TotalUses
	})

	sort.SliceStable(justiceCards, func(i, j int) bool {
		return justiceCards[i].TotalUses > justiceCards[j].TotalUses
	})

	sort.SliceStable(aggressionCards, func(i, j int) bool {
		return aggressionCards[i].TotalUses > aggressionCards[j].TotalUses
	})

	sort.SliceStable(protectionCards, func(i, j int) bool {
		return protectionCards[i].TotalUses > protectionCards[j].TotalUses
	})

	sort.SliceStable(leadershipCards, func(i, j int) bool {
		return leadershipCards[i].TotalUses > leadershipCards[j].TotalUses
	})

	printAspect("BASIC", basicCards)
	printAspect("JUSTICE", justiceCards)
	printAspect("AGGRESSION", aggressionCards)
	printAspect("PROTECTION", protectionCards)
	printAspect("LEADERSHIP", leadershipCards)
}

func printAspect(aspectName string, cards []cardStruct) {
	fmt.Println("-----")
	fmt.Println(strings.ToUpper(aspectName) + " ASPECT")
	fmt.Println("-----")
	for _, card := range cards {
		fmt.Println(fmt.Sprintf("%s - %s - %s - %s appears %d times", card.Code, card.Faction, card.Type, card.Name, card.TotalUses))
	}
}

type deckStruct struct {
	Hero  string         `json:"investigator_name"`
	Cards map[string]int `json:"slots"`
}

type cardStruct struct {
	Code        string `json:"code"`
	Name        string `json:"real_name"`
	Type        string `json:"type_name"`
	Faction     string `json:"faction_name"`
	TotalUses   int    `json:"-"`
	DuplicateOf string `json:"duplicate_of_code"`
}
