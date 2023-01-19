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

func getPositionsUrls() (urls []string) {
	c := colly.NewCollector()

	c.OnHTML("td a", func(e *colly.HTMLElement) {
		urls = append(urls, e.Request.AbsoluteURL(e.Attr("href")))

	})

	// Set error handler
	c.OnError(func(r *colly.Response, err error) {
		fmt.Println("Request failed ☠️!", "Error:", err)
		r.Request.Retry()
	})

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL, "🦍")
	})

	c.Visit("https://rekry.saima.fi/certiahome/open_jobs_view_new.html?did=5600&jc=14&lang=en")
	return urls
}

func getPositionDescription(url string) Position {
	var description string
	var title string
	var date string
	c := colly.NewCollector()
	c.SetRequestTimeout(60 * time.Second)

	c.OnHTML("td.title", func(e *colly.HTMLElement) {
		title = strings.TrimSpace(e.Text)
	})

	// Application up to

	c.OnHTML("td.normal", func(e *colly.HTMLElement) {
		description += fmt.Sprintln(strings.TrimSpace(e.Text))
	})

	// Set error handler
	c.OnError(func(r *colly.Response, err error) {
		fmt.Println("Request failed ☠️!", "Error:", err)
		r.Request.Retry()
	})

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL, "🦍")
	})

	c.Visit(url)

	return Position{Title: title, URL: url, Description: description, Date: date}

}

func main() {

	// Define name of the table for the University of Turku
	tableName := "utu_fi"

	log.Println("Connecting to the 'fenjan' database 🐰.")
	db, err := sql.Open("mysql", tea.GetDbConnectionString())
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Creating to the " + tableName + " table in the 'fenjan' database if not exists 👾.")
	tea.CreateTableIfNotExists(db, tableName)

	// Get the URLs from the database
	visitedUrls := tea.GetUrlsFromDB(db, tableName)

	log.Println("Finding URLs of open positions in University of Turku 🦉.")
	positionsUrl := getPositionsUrls()
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
