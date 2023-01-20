package main

import (
	"database/sql"
	"log"
	"strings"
	"time"

	"fenjan.ai-hue.ir/tea"
	"github.com/gocolly/colly"
	"github.com/mmcdole/gofeed"
)

type Position = tea.Position

func getPositionsUrlsFromRSS() (urls []string) {
	fp := gofeed.NewParser()
	feed, err := fp.ParseURL("https://www.fu-berlin.de/universitaet/beruf-karriere/jobs/english/index.rss")
	if err != nil {
		log.Panicln("Error in parsing the RSS file", err)
	}

	for _, item := range feed.Items {
		urls = append(urls, item.Link)
	}
	return urls
}

func getPositionDescription(url string) Position {
	var description string
	var title string
	var date string
	c := colly.NewCollector()
	c.SetRequestTimeout(60 * time.Second)

	c.OnHTML("div.box-job-offer-header h2", func(e *colly.HTMLElement) {
		title = strings.TrimSpace(e.Text)

	})

	c.OnHTML("h3", func(e *colly.HTMLElement) {

		if strings.Contains(e.Text, "Bewerbungsende") {

			date = strings.TrimSpace(strings.ReplaceAll(e.Text, "Bewerbungsende: ", ""))
		}
	})

	c.OnHTML("div.editor-content p", func(e *colly.HTMLElement) {
		description += strings.TrimSpace(e.Text)

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

	return Position{Title: title, URL: url, Description: description, Date: date}

}

func main() {

	// Define name of the table for the Freie Universität Berlin
	tableName := "fu_berlin_de"

	log.Println("Connecting to the 'fenjan' database 🐰.")
	db, err := sql.Open("mysql", tea.GetDbConnectionString())
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Creating to the " + tableName + " table in the 'fenjan' database if not exists 👾.")
	tea.CreateTableIfNotExists(db, tableName)

	// Get the URLs from the database
	visitedUrls := tea.GetUrlsFromDB(db, tableName)

	log.Println("Finding URLs of open positions in Freie Universität Berlin 🦉.")
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
