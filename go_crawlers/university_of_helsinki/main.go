package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

type Position []struct {
	Title         string `json:"title"`
	URL           string `json:"url"`
	Department    string `json:"department"`
	Date          string `json:"date"`
	DateTimestamp int    `json:"date_timestamp"`
	Description   string `json:"description"`
	Key           int    `json:"key"`
}

func main() {

	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	log.Println("Connecting to the 'fenjan' database 🐰.")
	db, err := sql.Open("mysql", getDbConnectionString())
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Creating the 'helsinki_fi' table in the 'fenjan' database if not exists 👾.")
	createTableIfNotExists(db)

	// Get the URLs from the database
	visitedUrls := getUrlsFromDB(db)

	log.Println("Searching the University of Helsinki  for the Ph.D. vacancies 🦉.")
	positions, err := getAndParseData()
	if err != nil {
		fmt.Println(err)
		return
	}

	// Loop through each position
	newPositions := Position{}
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
	savePositionsToDB(db, newPositions)
	log.Println("Finished 🫡!")

}

func getAndParseData() (Position, error) {
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

	var positions Position
	if err := json.Unmarshal([]byte(result[0].Data), &positions); err != nil { // Parse []byte to the go struct pointer
		return nil, err
	}
	return positions, nil
}

// getDbConnectionString function reads the database connection details from environment variables
func getDbConnectionString() string {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	username := os.Getenv("DB_USERNAME")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")
	return username + ":" + password + "@tcp(" + host + ":" + port + ")/" + dbname
}

// createTableIfNotExists function creates the "helsinki_fi" table if it doesn't already exist
func createTableIfNotExists(db *sql.DB) {
	// SQL statement to create the table
	query := `CREATE TABLE IF NOT EXISTS helsinki_fi (
		id INT AUTO_INCREMENT PRIMARY KEY,
		title VARCHAR(255) NOT NULL,
		description TEXT NOT NULL,
		url VARCHAR(255) NOT NULL,
		date VARCHAR(255) NOT NULL
	)`

	// Execute the SQL statement
	_, err := db.Exec(query)
	if err != nil {
		log.Fatal(err)
	}
}

// savePositionsToDB function saves the scraped positions to a MySQL database
func savePositionsToDB(db *sql.DB, positions Position) {
	// Prepare the SQL statement
	stmt, err := db.Prepare("INSERT INTO helsinki_fi (title, description, url, date) VALUES (?, ?, ?, ?)")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	// Loop through each position and execute the SQL statement
	for _, position := range positions {
		_, err := stmt.Exec(position.Title, position.Description, position.URL, position.Date)
		if err != nil {
			log.Fatal(err)
		}
	}
}

// getUrlsFromDB function return a map of all urls in the "positions" table
func getUrlsFromDB(db *sql.DB) map[string]bool {
	// SELECT statement to retrieve URLs from the table
	rows, err := db.Query("SELECT url FROM helsinki_fi")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	// Create a map to store the URLs
	urls := make(map[string]bool)
	for rows.Next() {
		var url string
		if err := rows.Scan(&url); err != nil {
			log.Fatal(err)
		}
		urls[url] = true
	}
	return urls
}
