package main

import (
	"database/sql"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"fenjan.ai-hue.ir/logger"
	"fenjan.ai-hue.ir/tea"
	"github.com/gocolly/colly"
)

// Set the university name, the database table name for this university, and the url of vacant positions
var uniName string = "Technical University of Munich (TUM)"
var tableName string = "tum_de"
var vacantPositionsUrl string = "https://portal.mytum.de/jobs/wissenschaftler/newsboard_view?b_start:int=0&-C="

// Get Position type from tea helper package
type Position = tea.Position

// get the URL of all vacant positions
func getPositionsUrlsAndDates() (urls []string, dates []string) {
	var NumVisitedPages int
	c := colly.NewCollector()
	c.SetRequestTimeout(60 * time.Second)

	// Find and visit all links
	c.OnHTML("span.next a", func(e *colly.HTMLElement) {
		nextPageUrl := e.Request.AbsoluteURL(e.Attr("href"))
		if NumVisitedPages < 10 {
			NumVisitedPages += 1
			c.Visit(nextPageUrl)
		}
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
				dates = append(dates, dateString)
			}
		}
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

	return urls, dates
}

// Get the details of position
func getPositionDescription(url string) (position Position) {

	c := colly.NewCollector()
	c.SetRequestTimeout(60 * time.Second)

	c.OnHTML("h1", func(e *colly.HTMLElement) {
		position.Title = strings.TrimSpace(e.Text)

	})
	c.OnHTML("p", func(e *colly.HTMLElement) {
		position.Description += strings.TrimSpace(e.Text) + "\n"
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
