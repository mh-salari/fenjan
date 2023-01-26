package main

import (
	"database/sql"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"fenjan.ai-hue.ir/logger"
	"fenjan.ai-hue.ir/tea"
)

// Set the university name, the database table name for this university, and the url of vacant positions
var uniName string = "University of Helsinki"
var tableName string = "helsinki_fi"
var vacantPositionsUrl string = "https://www.helsinki.fi/en/ajax_get_jobs/en/null/null/null/0"

// Get Position type from tea helper package
type Position = tea.Position

func getPositions() ([]Position, error) {

	// Get html response from the URL, retry for max of 5 times
	var resp *http.Response
	var err error
	for i := 0; i < 5; i++ {
		resp, err = http.Get(vacantPositionsUrl)
		if err == nil {
			break
		} else {
			log.Println(err)
			log.Println("Retrying 🧌!")
		}
	}
	if err != nil {
		return nil, err
	}

	// Pars the response
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

	// Connecting to the database and creating the university table if not exist
	log.Println("Connecting to the 'fenjan' database 🐰.")
	db, err := sql.Open("mysql", tea.GetDbConnectionString())
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Creating the '%s' table in the 'fenjan' database if not exists 👾.", tableName)
	tea.CreateTableIfNotExists(db, tableName)

	// Get the URLs from the database
	visitedUrls := tea.GetUrlsFromDB(db, tableName)

	// Getting vacant positions on the university site
	log.Printf("Searching the %s for the Ph.D. vacancies 🦉.", uniName)
	positions, err := getPositions()
	if err != nil {
		logger.Error.Fatal("Source: ", uniName, "🦂 ", "Source: ", uniName, "🦂 ", "No response from request ", err)
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
