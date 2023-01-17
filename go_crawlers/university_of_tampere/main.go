package main

import (
	"database/sql"
	"log"

	"fenjan.ai-hue.ir/tea"
	"github.com/gocolly/colly"
)

type Position = tea.Position

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

func main() {

	// Define name of the table for the University of Tampere positions
	tableName := "tuni_fi"

	log.Println("Searching the University of Tampere for the Ph.D. vacancies 🦉.")
	// Load environment variables from .env file

	log.Println("Connecting to the 'fenjan' database 🐰.")
	db, err := sql.Open("mysql", tea.GetDbConnectionString())
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Creating the '" + tableName + "' table in the 'fenjan' database if not exists 👾.")
	tea.CreateTableIfNotExists(db, tableName)

	// Create an empty slice to store position details
	positions := []Position{}
	log.Println("Finding URLs of open positions advertised on University of Tampere website 🦒...")

	// Get all position URLs from the main page
	positionUrls := getPositionUrls()
	log.Println("Found", len(positionUrls), "open positions 😺.")
	log.Println("Extracting details of open positions 🐢...")

	// Get the URLs from the database
	visitedUrls := tea.GetUrlsFromDB(db, tableName)

	// Loop through each URL and get the title and text of each position
	for _, url := range positionUrls[:] {
		// Check if the URL has been visited before
		if visitedUrls[url] {
			log.Println("URL has been visited before:", url)
			continue
		}

		title, description := extractPositionDetails(url)
		positions = append(positions, Position{title, url, description, ""})
	}
	log.Println("Extracted details of", len(positions), "open positions 🤓.")

	log.Println("Saving positions to the database 🚀...")
	tea.SavePositionsToDB(db, positions, tableName)

	log.Println("Finished 🫡!")
}
