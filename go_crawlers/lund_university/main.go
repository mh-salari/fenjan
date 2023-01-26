package main

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	"fenjan.ai-hue.ir/logger"
	"fenjan.ai-hue.ir/tea"
	"github.com/gocolly/colly"
)

// Set the university name, the database table name for this university, and the url of vacant positions
var uniName string = "Lund University"
var tableName string = "lunduniversity_lu_se"
var vacantPositionsUrl string = "https://www.lunduniversity.lu.se/vacancies"

// Get Position type from tea helper package
type Position = tea.Position

// get the URL of all vacant positions
func getPositionsUrls() (urls []string) {

	c := colly.NewCollector()
	c.SetRequestTimeout(60 * time.Second)

	c.OnHTML("tbody.vacancies-list__table--body", func(e *colly.HTMLElement) {
		urls = append(urls, e.ChildAttrs("a", "href")...)
	})

	// Add the OnRequest function to log the URLs that have visited
	c.OnRequest(func(r *colly.Request) {
		log.Println("Visiting", r.URL, "🥷")
	})

	// Set error handler
	c.OnError(func(r *colly.Response, err error) {
		log.Println("Request failed ☠️!", "Error:", err)

		// Sleep if its a 429 Too Many Requests Error
		if r.StatusCode == 429 {
			rand.Seed(time.Now().UnixNano())
			n := 30 + rand.Intn(60)
			log.Printf("Sleeping %d seconds...\n", n)
			time.Sleep(time.Duration(n) * time.Second)
		}

		// Retry for 5 time
		retriesLeft := tea.RetryRequest(r, 5)
		if retriesLeft == 0 {
			logger.Error.Fatal("Reached max number of retries 🫄! ", "Error: ", err)
		}
	})

	c.Visit(vacantPositionsUrl)

	return urls
}

// Get the details of position
func getPositionDescription(url string) (position Position) {

	c := colly.NewCollector()
	c.SetRequestTimeout(60 * time.Second)

	c.OnHTML("div.content-wrap h1", func(e *colly.HTMLElement) {
		position.Title = strings.TrimSpace(e.Text)

	})

	c.OnHTML("tr.quick-info-ends td", func(e *colly.HTMLElement) {
		position.Date = strings.TrimSpace(e.Text)
	})

	c.OnHTML("div.job-desc.mb", func(e *colly.HTMLElement) {

		position.Description += fmt.Sprintln(strings.TrimSpace(strings.ToValidUTF8(e.Text, "")))

	})

	// Add the OnResponse function to randomly wait before next request
	c.OnResponse(func(r *colly.Response) {

		rand.Seed(time.Now().UnixNano())
		n := 1 + rand.Intn(5)
		log.Printf("Sleeping %d seconds...\n", n)
		time.Sleep(time.Duration(n) * time.Second)
	})

	// Add the OnRequest function to log the URLs that have visited
	c.OnRequest(func(r *colly.Request) {
		log.Println("Visiting", r.URL, "🥷")
	})

	// Set error handler
	c.OnError(func(r *colly.Response, err error) {
		log.Println("Request failed ☠️!", "Error:", err)

		// Sleep if its a 429 Too Many Requests Error
		if r.StatusCode == 429 {
			rand.Seed(time.Now().UnixNano())
			n := 30 + rand.Intn(60)
			log.Printf("Sleeping %d seconds...\n", n)
			time.Sleep(time.Duration(n) * time.Second)
		}

		// Retry for 5 time
		retriesLeft := tea.RetryRequest(r, 5)
		if retriesLeft == 0 {
			logger.Error.Fatal("Reached max number of retries 🫄! ", "Error: ", err)
		}
	})

	c.Visit(url)

	position.URL = url

	return position

}

func main() {

	// Connecting to the database and creating the university table if not exist
	log.Println("Connecting to the 'fenjan' database 🐰.")
	db, err := sql.Open("mysql", tea.GetDbConnectionString())
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Creating the '%s' table in the 'fenjan' database if not exists 👾.", tableName)
	tea.CreateTableIfNotExists(db, tableName)

	// Get the URLs from the database
	visitedUrls := tea.GetUrlsFromDB(db, tableName)

	// Getting the URL of vacant positions on the university site
	log.Printf("Searching the %s for the Ph.D. vacancies 🦉.", uniName)
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
		positions = append(positions, position)
	}
	log.Println("Extracted details of", len(positions), "open positions 🤓.")

	// Saving the positions to the database
	log.Println("Saving new positions to the database 🚀...")
	tea.SavePositionsToDB(db, positions, tableName)

	log.Println("Finished 🫡!")

}
