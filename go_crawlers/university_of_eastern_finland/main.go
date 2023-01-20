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

func getPositionsUrlsAndDates() (urls []string, dates []string) {
	c := colly.NewCollector()

	c.OnHTML("article.rss-feed-item", func(e *colly.HTMLElement) {

		date := strings.Split(strings.TrimSpace(e.ChildText("div.rss-feed-item__content")), "-")[0]
		date = strings.ReplaceAll(date, "Apply by: ", "")

		url := e.ChildAttr("a", "href")

		dates = append(dates, date)
		urls = append(urls, url)

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

	c.Visit("https://www.uef.fi/en/open-positions")
	return urls, dates
}

func getPositionDescription(url string) Position {
	var description string
	var title string
	c := colly.NewCollector()
	c.SetRequestTimeout(60 * time.Second)

	c.OnHTML("div.title", func(e *colly.HTMLElement) {
		title = strings.TrimSpace(e.Text)

	})

	c.OnHTML("div.normal", func(e *colly.HTMLElement) {

		description += fmt.Sprintln(strings.TrimSpace(e.Text))

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

	return Position{Title: title, URL: url, Description: description, Date: ""}

}

func main() {

	// Define name of the table for the University of Eastern Finland
	tableName := "uef_fi"

	log.Println("Connecting to the 'fenjan' database 🐰.")
	db, err := sql.Open("mysql", tea.GetDbConnectionString())
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Creating to the " + tableName + " table in the 'fenjan' database if not exists 👾.")
	tea.CreateTableIfNotExists(db, tableName)

	// Get the URLs from the database
	visitedUrls := tea.GetUrlsFromDB(db, tableName)

	log.Println("Finding URLs of open positions in University of Eastern Finland 🦉.")
	positionsUrl, positionsDate := getPositionsUrlsAndDates()
	log.Printf("Found %d open positions", len(positionsUrl))

	// Loop through each position
	positions := []Position{}
	for idx, url := range positionsUrl[:] {

		// Check if the URL has been visited before
		if visitedUrls[url] {
			log.Println("URL has been visited before:", url)
			continue
		}
		position := getPositionDescription(url)
		position.Date = positionsDate[idx]
		positions = append(positions, position)
	}
	log.Println("Extracted details of", len(positions), "open positions 🤓.")

	log.Println("Saving new positions to the database 🚀...")
	tea.SavePositionsToDB(db, positions, tableName)
	log.Println("Finished 🫡!")

}
