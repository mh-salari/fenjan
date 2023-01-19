package main

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"fenjan.ai-hue.ir/tea"
	"github.com/gocolly/colly"
)

type Position = tea.Position

// func getPositionsUrlsFromRSS() (urls []string) {
// 	fp := gofeed.NewParser()
// 	feed, err := fp.ParseURL("https://portal.mytum.de/jobs/asRss")
// 	if err != nil {
// 		log.Panicln("Error in parsing the RSS file", err)
// 	}
// 	for _, item := range feed.Items {
// 		urls = append(urls, item.Link)
// 	}
// 	return urls
// }

func getPositionsUrls() (urls []string) {

	c := colly.NewCollector(
	// colly.MaxDepth(1),
	)
	c.SetRequestTimeout(60 * time.Second)

	// Find and visit all links
	c.OnHTML("span.next a", func(e *colly.HTMLElement) {
		nextPageUrl := e.Request.AbsoluteURL(e.Attr("href"))
		c.Visit(nextPageUrl)
	})

	c.OnHTML("span", func(e *colly.HTMLElement) {

		link, _ := e.DOM.Find("h6 a").Attr("href")
		dateString := e.DOM.Find("h5").Text()

		if link != "" && dateString != "" {
			dateList := strings.Split(dateString, ".")

			intYear, _ := strconv.Atoi(dateList[2])
			intMonth, _ := strconv.Atoi(dateList[1])
			intDay, _ := strconv.Atoi(dateList[0])
			date := time.Date(intYear, time.Month(intMonth), intDay, 0, 0, 0, 0, time.UTC)

			pastMonth := time.Now().AddDate(0, -1, 0)

			if date.After(pastMonth) {

				urls = append(urls, link)
			}
		}
	})

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL, "🐗")
	})
	// Set error handler
	c.OnError(func(r *colly.Response, err error) {
		fmt.Println("Request failed ☠️!", "Error:", err)
		r.Request.Retry()
	})

	c.Visit("https://portal.mytum.de/jobs/wissenschaftler/newsboard_view?b_start:int=0&-C=")
	return urls
}

func getPositionDescription(url string) Position {
	var paragraphs []string
	var title string
	c := colly.NewCollector()
	c.SetRequestTimeout(60 * time.Second)

	c.OnHTML("h1", func(e *colly.HTMLElement) {
		title = strings.TrimSpace(e.Text)

	})
	c.OnHTML("p", func(e *colly.HTMLElement) {
		paragraphs = append(paragraphs, strings.TrimSpace(e.Text))
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

	date := strings.Split(paragraphs[0], ",")[0]
	description := strings.Join(paragraphs[1:], "\n")

	return Position{Title: title, URL: url, Description: description, Date: date}

}

func main() {

	// Define name of the table for the Technical University of Munich (TUM)
	tableName := "tum_de"
	log.Println("Connecting to the 'fenjan' database 🐰.")
	db, err := sql.Open("mysql", tea.GetDbConnectionString())
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Creating to the " + tableName + " table in the 'fenjan' database if not exists 👾.")
	tea.CreateTableIfNotExists(db, tableName)

	// Get the URLs from the database
	visitedUrls := tea.GetUrlsFromDB(db, tableName)

	log.Println("Finding URLs of open positions in Technical University of Munich (TUM)🦉.")
	positionsUrl := getPositionsUrls()
	// positionsUrl := getPositionsUrlsFromRSS()
	log.Printf("Found %d open positions", len(positionsUrl))

	// Loop through each position
	positions := []Position{}
	for _, url := range positionsUrl {

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
