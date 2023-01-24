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

	c.OnHTML("div#mainjoblist tr", func(e *colly.HTMLElement) {
		date := e.ChildText("span:last-child")
		url := e.ChildAttr("a", "href")

		if date != "" && url != "" {
			dates = append(dates, date)
			urls = append(urls, url)
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

	c.Visit("https://web103.reachmee.com/ext/I003/304/main?site=5&lang=UK&validator=a72aeedd63ec10de71e46f8d91d0d57c")
	return urls, dates
}

func getPositionDescription(url string) Position {
	var description string
	var title string

	c := colly.NewCollector()
	c.SetRequestTimeout(60 * time.Second)

	c.OnHTML("h1#jobad-heading", func(e *colly.HTMLElement) {
		title = strings.TrimSpace(e.Text)

	})

	c.OnHTML("div.jobad-body p", func(e *colly.HTMLElement) {

		description += fmt.Sprintln(" ", strings.TrimSpace(strings.ToValidUTF8(e.Text, "")), " ")

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

	return Position{Title: title, URL: url, Description: description, Date: ""}

}

func main() {

	// Define name of the table for the Chalmers University of Technology
	tableName := "chalmers_se"

	log.Println("Connecting to the 'fenjan' database 🐰.")
	db, err := sql.Open("mysql", tea.GetDbConnectionString())
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Creating to the " + tableName + " table in the 'fenjan' database if not exists 👾.")
	tea.CreateTableIfNotExists(db, tableName)

	// Get the URLs from the database
	visitedUrls := tea.GetUrlsFromDB(db, tableName)

	log.Println("Finding URLs of open positions in Chalmers University of Technology 🦉.")
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
