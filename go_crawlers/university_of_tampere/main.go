package main

import (
	"database/sql"
	"log"
	"os"

	"github.com/joho/godotenv"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gocolly/colly"
)

type Position struct {
	Title       string
	Description string
	URL         string
}

func main() {
	log.Println("Searching the University of Tampere for the Ph.D. vacancies 🦉.")
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

	log.Println("Creating the 'tuni_fi' table in the 'fenjan' database if not exists 👾.")
	createTableIfNotExists(db)

	// Create an empty slice to store position details
	positions := []Position{}
	log.Println("Finding URLs of open positions advertised on University of Tampere website 🦒...")
	// Get all position URLs from the main page
	positionUrls := getPositionUrls()
	log.Println("Found", len(positionUrls), "open positions 😺.")
	log.Println("Extracting details of open positions 🐢...")

	// Get the URLs from the database
	visitedUrls := getUrlsFromDB(db)

	// Loop through each URL and get the title and text of each position
	for _, url := range positionUrls[:] {
		// Check if the URL has been visited before
		if visitedUrls[url] {
			log.Println("URL has been visited before:", url)
			continue
		}

		title, description := extractPositionDetails(url)
		positions = append(positions, Position{Title: title, Description: description, URL: url})
	}
	log.Println("Extracted details of", len(positions), "open positions 🤓.")

	log.Println("Saving positions to the database 🚀...")
	savePositionsToDB(db, positions)
	log.Println("Finished 🫡!")
}

// getPositionUrls function uses colly to scrape the main page and finds all the URLs of the open positions
func getPositionUrls() []string {
	// Create an empty slice to store URLs
	positionUrls := []string{}
	// Create a new collector
	c := colly.NewCollector()
	// Find all URLs in the div element with class name "RSSFeedLiftup__StyledWrapper-sc-bi69hc-0 fbpMqB"
	c.OnHTML("div.RSSFeedLiftup__StyledWrapper-sc-bi69hc-0.fbpMqB a[href]", func(e *colly.HTMLElement) {
		positionUrls = append(positionUrls, e.Attr("href"))
	})
	// Add the OnRequest function to log the URLs that are visited
	c.OnRequest(func(r *colly.Request) {
		log.Println("Visiting", r.URL)
	})
	c.Visit("https://www.tuni.fi/en/about-us/working-at-tampere-universities/open-positions-at-tampere-university")
	return positionUrls
}

// extractPositionDetails function uses colly to scrape a position page and finds the title and text of the position
func extractPositionDetails(url string) (string, string) {
	// Create a new collector
	c := colly.NewCollector()
	var title, description string
	// Find the title in h1 element
	c.OnHTML("h1", func(e *colly.HTMLElement) {
		title = e.Text
	})
	// Find the text in p element
	c.OnHTML("p", func(e *colly.HTMLElement) {
		description += e.Text + " "
	})
	// Add the OnRequest function to log the URLs that are visited
	c.OnRequest(func(r *colly.Request) {
		log.Println("Visiting", r.URL)
	})
	c.Visit(url)
	return title, description
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

// createTableIfNotExists function creates the "tuni_fi" table if it doesn't already exist
func createTableIfNotExists(db *sql.DB) {
	// SQL statement to create the table
	query := `CREATE TABLE IF NOT EXISTS tuni_fi (
		id INT AUTO_INCREMENT PRIMARY KEY,
		title VARCHAR(255) NOT NULL,
		description TEXT NOT NULL,
		url VARCHAR(255) NOT NULL
	)`

	// Execute the SQL statement
	_, err := db.Exec(query)
	if err != nil {
		log.Fatal(err)
	}
}

// savePositionsToDB function saves the scraped positions to a MySQL database
func savePositionsToDB(db *sql.DB, positions []Position) {
	// Prepare the SQL statement
	stmt, err := db.Prepare("INSERT INTO tuni_fi (title, description, url) VALUES (?, ?, ?)")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	// Loop through each position and execute the SQL statement
	for _, position := range positions {
		_, err := stmt.Exec(position.Title, position.Description, position.URL)
		if err != nil {
			log.Fatal(err)
		}
	}
}

// getUrlsFromDB function return a map of all urls in the "positions" table
func getUrlsFromDB(db *sql.DB) map[string]bool {
	// SELECT statement to retrieve URLs from the table
	rows, err := db.Query("SELECT url FROM tuni_fi")
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
