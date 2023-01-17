package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"fenjan.ai-hue.ir/tea"
	_ "github.com/go-sql-driver/mysql"
)

type Position = tea.Position

func getPositions() ([]Position, error) {
	// Get request
	resp, err := http.Get("https://www.helsinki.fi/en/ajax_get_jobs/en/null/null/null/0")
	if err != nil {
		fmt.Println("No response from request")
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body) // response body is []byte
	if err != nil {
		return nil, err
	}
	var result []struct {
		Command  string      `json:"command"`
		Method   string      `json:"method"`
		Selector string      `json:"selector"`
		Data     string      `json:"data"`
		Settings interface{} `json:"settings"`
	}
	if err := json.Unmarshal(body, &result); err != nil { // Parse []byte to the go struct pointer
		return nil, err
	}
	var positionsAjax []struct {
		Title         string `json:"title"`
		URL           string `json:"url"`
		Department    string `json:"department"`
		Date          string `json:"date"`
		DateTimestamp int    `json:"date_timestamp"`
		Description   string `json:"description"`
		Key           int    `json:"key"`
	}
	if err := json.Unmarshal([]byte(result[0].Data), &positionsAjax); err != nil { // Parse []byte to the go struct pointer
		return nil, err
	}

	positions := []Position{}

	for _, position := range positionsAjax {
		positions = append(positions, Position{position.Title, position.URL, position.Description, position.Date})
	}
	return positions, nil
}

func main() {

	// Define name of the table for the University of Helsinki positions
	tableName := "helsinki_fi"
	log.Println("Connecting to the 'fenjan' database 🐰.")
	db, err := sql.Open("mysql", tea.GetDbConnectionString())
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Creating to the " + tableName + " table in the 'fenjan' database if not exists 👾.")
	tea.CreateTableIfNotExists(db, tableName)

	// Get the URLs from the database
	visitedUrls := tea.GetUrlsFromDB(db, tableName)

	log.Println("Searching the University of Helsinki  for the Ph.D. vacancies 🦉.")
	positions, err := getPositions()
	if err != nil {
		fmt.Println(err)
		return
	}

	// Loop through each position
	newPositions := []Position{}
	for _, position := range positions {

		// Check if the URL has been visited before
		if visitedUrls[position.URL] {
			log.Println("URL has been visited before:", position.URL)
			continue
		}
		newPositions = append(newPositions, position)
	}
	log.Println("Extracted details of", len(newPositions), "open positions 🤓.")

	log.Println("Saving new positions to the database 🚀...")
	tea.SavePositionsToDB(db, newPositions, tableName)
	log.Println("Finished 🫡!")

}
