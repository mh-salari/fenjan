package main

import (
	"database/sql"
	"log"
	"time"

	"fenjan.ai-hue.ir/tea"
	"github.com/gocolly/colly"
)

type Position = tea.Position

func getPositionsUrls() []string {
	var urls []string

	c := colly.NewCollector()
	// Set request timeout timer
	c.SetRequestTimeout(60 * 5 * time.Second)

	// Extract URL of positions
	c.OnHTML("table#jobListingsTable tr", func(e *colly.HTMLElement) {
		url := e.ChildText("td:last-child")
		if url != "" {
			urls = append(urls, url)
		}
	})

	// Add the OnRequest function to log the URLs that are visited
	c.OnRequest(func(r *colly.Request) {
		log.Println("Visiting", r.URL)
	})
	// Set error handler
	c.OnError(func(r *colly.Response, err error) {
		log.Fatal("Request URL:", r.Request.URL, "failed with response:", r, "\nError:", err)
	})

	c.Visit("https://liu.se/en/work-at-liu/vacancies")

	return urls
}

func extractPositionsDetails(url string) Position {
	var position Position
	c := colly.NewCollector()

	// Set request timeout timer
	c.SetRequestTimeout(60 * 5 * time.Second)

	// Extract title of the position
	c.OnHTML("h1", func(e *colly.HTMLElement) {
		position.Title = e.Text
	})

	// Extract the description of the position
	c.OnHTML("div.job", func(e *colly.HTMLElement) {
		position.Description = e.Text
	})

	// Add the OnRequest function to log the URLs that are visited
	c.OnRequest(func(r *colly.Request) {
		log.Println("Visiting", r.URL)
	})
	// Set error handler
	c.OnError(func(r *colly.Response, err error) {
		log.Fatal("Request URL:", r.Request.URL, "failed with response:", r, "\nError:", err)
	})
	c.Visit(url)
	position.URL = url
	return position
}

func main() {

	// Set the name of table for the Linköping University positions
	tableName := "liu_se"

	log.Println("Connecting to the 'fenjan' database 🐰.")

	db, err := sql.Open("mysql", tea.GetDbConnectionString())
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Creating the 'liu_se' table in the 'fenjan' database if not exists 👾.")
	tea.CreateTableIfNotExists(db, tableName)

	// Get the URLs from the database
	visitedUrls := tea.GetUrlsFromDB(db, tableName)

	log.Println("Searching the Linköping University  for the Ph.D. vacancies 🦉.")
	positions := []Position{}

	positionsUrls := getPositionsUrls()

	// Extract description of the positions
	for _, url := range positionsUrls {
		// Check if the URL has been visited before
		if visitedUrls[url] {
			log.Println("URL has been visited before:", url)
			continue
		}

		position := extractPositionsDetails(url)
		positions = append(positions, position)
	}

	log.Println("Extracted details of", len(positions), "open positions 🤓.")

	log.Println("Saving new positions to the database 🚀...")
	tea.SavePositionsToDB(db, positions, tableName)

	log.Println("Finished 🫡!")

}
