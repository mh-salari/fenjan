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

	c.OnHTML("section.listing-page-jobs div.grid__item", func(e *colly.HTMLElement) {

		date := e.ChildText("div.teaser__date")
		url := e.Request.AbsoluteURL(e.ChildAttr("a", "href"))
		dates = append(dates, date)
		urls = append(urls, url)

	})

	c.OnHTML("a.pager__link--next", func(e *colly.HTMLElement) {

		c.Visit(e.Request.AbsoluteURL(e.Attr("href")))

	})

	// Add the OnRequest function to log the URLs that have visited
	c.OnRequest(func(r *colly.Request) {
		log.Println("Visiting", r.URL)
	})
	// Set error handler
	c.OnError(func(r *colly.Response, err error) {
		log.Println("Request failed ☠️!", "Error:", err)
		r.Request.Retry()
	})

	c.Visit("https://www.oulu.fi/en/university/jobs")
	return urls, dates
}

func getPositionDescription(url string) Position {
	var description string
	var title string
	c := colly.NewCollector()
	c.SetRequestTimeout(60 * time.Second)

	c.OnHTML("td.title", func(e *colly.HTMLElement) {
		title = strings.TrimSpace(e.Text)

	})

	c.OnHTML("td.normal", func(e *colly.HTMLElement) {

		description += fmt.Sprintln(strings.TrimSpace(e.Text))

	})

	// Add the OnRequest function to log the URLs that have visited
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

	// Define name of the table for the University of Oulu
	tableName := "oulu_fi"

	log.Println("Connecting to the 'fenjan' database 🐰.")
	db, err := sql.Open("mysql", tea.GetDbConnectionString())
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Creating to the " + tableName + " table in the 'fenjan' database if not exists 👾.")
	tea.CreateTableIfNotExists(db, tableName)

	// Get the URLs from the database
	visitedUrls := tea.GetUrlsFromDB(db, tableName)

	log.Println("Finding URLs of open positions in University of Oulu 🦉.")
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
