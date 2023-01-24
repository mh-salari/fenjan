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
	c.SetRequestTimeout(60 * time.Second)

	c.OnHTML("a.aalto-listing__link", func(e *colly.HTMLElement) {
		urls = append(urls, e.Request.AbsoluteURL(e.Attr("href")))
	})

	c.OnHTML("a.aalto-pager__item-link", func(e *colly.HTMLElement) {
		c.Visit(e.Request.AbsoluteURL(e.Attr("href")))
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

	c.Visit("https://www.aalto.fi/en/open-positions?sort_by=field_application_end_value&field_unit_target_id=All&field_category_target_id%5B13336%5D=13336&page")

	return urls
}

func getPositionDescription(url string) (position Position) {

	c := colly.NewCollector()
	c.SetRequestTimeout(60 * time.Second)

	c.OnHTML("div.article-container.aalto-article__top", func(e *colly.HTMLElement) {
		position.Title = strings.TrimSpace(e.ChildText("h1"))

	})

	c.OnHTML("div.aalto-article__info-text time", func(e *colly.HTMLElement) {
		position.Date = strings.TrimSpace(e.Text)
	})

	c.OnHTML("div.aalto-article div.aalto-user-generated-content", func(e *colly.HTMLElement) {

		position.Description += fmt.Sprintln(strings.TrimSpace(strings.ToValidUTF8(e.Text, "")))

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

	position.URL = url

	return position

}

func main() {

	// Set the name of table for the Aalto University positions
	tableName := "aalto_fi"

	log.Println("Connecting to the 'fenjan' database 🐰.")

	db, err := sql.Open("mysql", tea.GetDbConnectionString())
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Creating the 'aalto_fi' table in the 'fenjan' database if not exists 👾.")
	tea.CreateTableIfNotExists(db, tableName)

	// Get the URLs from the database
	visitedUrls := tea.GetUrlsFromDB(db, tableName)

	log.Println("Searching the Aalto University  for the Ph.D. vacancies 🦉.")
	positions := []Position{}

	positionsUrls := getPositionsUrls()

	log.Println("Found ", len(positionsUrls), " open positions 🐝")
	// Extract description of the positions
	for _, url := range positionsUrls {
		// Check if the URL has been visited before
		if visitedUrls[url] {
			log.Println("URL has been visited before:", url)
			continue
		}

		position := getPositionDescription(url)
		// if position.Date != "" {
		positions = append(positions, position)
		// }
	}

	log.Println("Extracted details of", len(positions), "open positions 🤓.")

	log.Println("Saving new positions to the database 🚀...")
	tea.SavePositionsToDB(db, positions, tableName)

	log.Println("Finished 🫡!")

}
