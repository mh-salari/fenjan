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
var uniName string = "Lappeenranta University of Technology"
var tableName string = "lut_fi"
var vacantPositionsUrl string = "https://lut.rekrytointi.com/paikat/index.php?o=A_LOJ&list=2"

// Get Position type from tea helper package
type Position = tea.Position

// get the URL of all vacant positions
func getPositionsUrlsAndDates() (urls []string, dates []string) {

	c := colly.NewCollector()
	c.SetRequestTimeout(60 * time.Second)

	c.OnHTML("div.auto_list.auto_list_open_jobs tr", func(e *colly.HTMLElement) {
		date := e.ChildText("td:last-child")
		url := e.Request.AbsoluteURL(e.ChildAttr("td:last-child a", "href"))

		if date != "" && url != "" {
			dates = append(dates, date)
			urls = append(urls, url[:strings.Index(url, "&rspvt=")])
		}
	})
	// Add the OnRequest function to log the URLs that are visited
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

	return urls, dates
}

// Get the details of position
func getPositionDescription(url string) (position Position) {

	c := colly.NewCollector()
	c.SetRequestTimeout(60 * time.Second)

	c.OnHTML("div.job_page h1", func(e *colly.HTMLElement) {
		position.Title = strings.TrimSpace(e.Text)
	})

	c.OnHTML("div.job_description", func(e *colly.HTMLElement) {
		position.Description += fmt.Sprintln(strings.TrimSpace(e.Text))
	})

	// Add the OnRequest function to log the URLs that are visited
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
	log.Printf("Creating the '%s' table in thpositionse 'fenjan' database if not exists 👾.", tableName)
	tea.CreateTableIfNotExists(db, tableName)

	// Get the URLs from the database
	visitedUrls := tea.GetUrlsFromDB(db, tableName)

	// Getting the URL of vacant positions on the university site
	log.Printf("Searching the %s for the Ph.D. vacancies 🦉.", uniName)
	positions := []Position{}
	positionsUrls, positionsDates := getPositionsUrlsAndDates()
	log.Println("Found ", len(positionsUrls), " open positions 🐝")

	// Extract description of the positions
	for idx, url := range positionsUrls {
		// Check if the URL has been visited before
		if visitedUrls[url] {
			log.Println("URL has been visited before:", url)
			continue
		}
		position := getPositionDescription(url)
		position.Date = positionsDates[idx]
		positions = append(positions, position)
	}
	log.Println("Extracted details of", len(positions), "open positions 🤓.")

	// Saving the positions to the database
	log.Println("Saving new positions to the database 🚀...")
	tea.SavePositionsToDB(db, positions, tableName)

	log.Println("Finished 🫡!")

}
