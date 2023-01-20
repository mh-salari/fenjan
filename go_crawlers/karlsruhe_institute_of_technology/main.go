package main

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"fenjan.ai-hue.ir/tea"
	"github.com/gocolly/colly"
)

type Position = tea.Position

func getPositionsUrlsFromRSS() (urls []string) {
	c := colly.NewCollector()

	c.OnHTML("div#job_DOKTORANDEN a", func(e *colly.HTMLElement) {
		//  strings.TrimSpace(e.Text)
		urls = append(urls, e.Request.AbsoluteURL(e.Attr("href")))

	})

	// Add the OnRequest function to log the URLs that are visited
	c.OnRequest(func(r *colly.Request) {
		log.Println("Visiting", r.URL)
	})
	// Set error handler
	c.OnError(func(r *colly.Response, err error) {
		log.Println("Request failed ☠️!", "Error:", err)
		r.Request.Retry()
	})

	c.Visit("https://www.pse.kit.edu/english/karriere/121.php")

	return urls
}

func getPositionDescription(url string) Position {
	var description string
	var title string
	var date string
	c := colly.NewCollector()
	c.SetRequestTimeout(60 * time.Second)

	c.OnHTML("div.text h1", func(e *colly.HTMLElement) {
		title = strings.TrimSpace(e.Text)

	})

	c.OnHTML("div.text p", func(e *colly.HTMLElement) {

		per := e.DOM.Prev().Text()
		if strings.Contains(per, "Application up to") {
			date = strings.TrimSpace(e.Text)
		} else {
			description += fmt.Sprintln(strings.TrimSpace(e.Text))
		}

	})

	// Add the OnRequest function to log the URLs that are visited
	c.OnRequest(func(r *colly.Request) {
		log.Println("Visiting", r.URL)
	})
	// Set error handler
	c.OnError(func(r *colly.Response, err error) {
		log.Println("Request failed ☠️!", "Error:", err)
		r.Request.Retry()
	})

	c.Visit(url)

	// Remove num English positions, But we keep them in our DB so we won't visit them again!
	if date == "" {
		description = ""
		title = ""
	}
	return Position{Title: title, URL: url, Description: description, Date: date}

}

func main() {

	// Define name of the table for the Karlsruhe Institute of Technology (KIT)
	tableName := "kit_edu"

	log.Println("Connecting to the 'fenjan' database 🐰.")
	db, err := sql.Open("mysql", tea.GetDbConnectionString())
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Creating to the " + tableName + " table in the 'fenjan' database if not exists 👾.")
	tea.CreateTableIfNotExists(db, tableName)

	// Get the URLs from the database
	visitedUrls := tea.GetUrlsFromDB(db, tableName)

	log.Println("Finding URLs of open positions in Karlsruhe Institute of Technology (KIT) 🦉.")
	positionsUrl := getPositionsUrlsFromRSS()
	log.Printf("Found %d open positions", len(positionsUrl))

	// Loop through each position
	positions := []Position{}
	for _, url := range positionsUrl[:] {

		// Check if the URL has been visited before
		if visitedUrls[url] {
			log.Println("URL has been visited before:", url)
			continue
		}
		position := getPositionDescription(url)

		positions = append(positions, position)

	}
	log.Println("Extracted details of", len(positions), "open positions 🤓.")

	log.Println("Saving new positions to the database 🚀...")
	tea.SavePositionsToDB(db, positions, tableName)
	log.Println("Finished 🫡!")

}
