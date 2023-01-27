package main

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"

	"fenjan.ai-hue.ir/logger"
	"fenjan.ai-hue.ir/tea"
	"github.com/gocolly/colly"
)

// Set the university name, the database table name for this university
var uniName string = "UvA University of Amsterdam"
var tableName string = "uva_nl"

// Get Position type from tea helper package
type Position = tea.Position

// Find total number of active position current;y advertised on UvA University of Amsterdam
func findNumActivePositions() int {
	c := colly.NewCollector(
		colly.AllowedDomains("vacatures.uva.nl"),
	)

	var numPositions int
	c.OnHTML("span#tile-search-results-label", func(e *colly.HTMLElement) {
		re := regexp.MustCompile(`\d+`)
		results := re.FindAllString(e.Text, -1)
		numPositions, _ = strconv.Atoi(results[len(results)-1])
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
			logger.Error.Fatal("Source: ", uniName, "🦂 ", "Reached max number of retries 🫄! ", "Error: ", err)
		}
	})

	c.Visit("https://vacatures.uva.nl/UvA/search/?locale=en_GB")

	return numPositions
}

// Get url pf positions advertised
func getPositionsURL(numPositions int) []string {
	var positionsURL []string
	for startRow := 0; startRow < numPositions; startRow += 10 {
		url := fmt.Sprintf("https://vacatures.uva.nl/UvA/tile-search-results/?q=phd&startrow=%d", startRow)
		c := colly.NewCollector()

		// Set request timeout timer
		c.SetRequestTimeout(60 * 5 * time.Second)

		c.OnHTML("a[href]", func(e *colly.HTMLElement) {
			href := "https://vacatures.uva.nl" + e.Attr("href")

			if strings.Contains(strings.ToLower(e.Text), "phd") && !tea.Contains(positionsURL, href) {
				positionsURL = append(positionsURL, href)
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
				logger.Error.Fatal("Source: ", uniName, "🦂 ", "Reached max number of retries 🫄! ", "Error: ", err)
			}
		})

		c.Visit(url)
	}
	return positionsURL
}

// Extract details of positions
func extractPositionDetails(positionsURL []string, visitedUrls map[string]bool) []Position {
	positions := []Position{}
	for _, url := range positionsURL {
		position := Position{}
		c := colly.NewCollector()

		// Set request timeout timer
		c.SetRequestTimeout(60 * 5 * time.Second)

		// Check if the URL has been visited before
		if visitedUrls[url] {
			log.Println("URL has been visited before:", url)
			continue
		}

		c.OnHTML("h1", func(e *colly.HTMLElement) {
			position.Title = strings.TrimSpace(e.Text)
		})

		c.OnHTML("span[data-careersite-propertyid=customfield3]", func(e *colly.HTMLElement) {
			position.Date = strings.TrimSpace(e.Text)
		})
		c.OnHTML("span.jobdescription", func(e *colly.HTMLElement) {
			position.Description = strings.TrimSpace(e.Text)
		})

		c.OnScraped(func(r *colly.Response) {
			position.URL = url
			positions = append(positions, position)
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
				logger.Error.Fatal("Source: ", uniName, "🦂 ", "Reached max number of retries 🫄! ", "Error: ", err)
			}
		})

		c.Visit(url)
	}

	return positions
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
	log.Println("Finding the total number of open positions advertised on the university website 🦎...")
	numPositions := findNumActivePositions()

	if numPositions == 0 {
		logger.Error.Fatal("Source: ", uniName, "🦂 ", "There is an error in getting the data, The total number of open positions in equal to 0 ☠️ !")
	}

	log.Printf("Currently, there are %d open positions advertised on the website.", numPositions)
	positionsURL := getPositionsURL(numPositions)

	log.Println("Found ", len(positionsURL), "open positions.")

	// Extract details of the positions
	positions := extractPositionDetails(positionsURL, visitedUrls)
	log.Println("Extracted details of", len(positions), "Ph.D. new open positions 🤓.")

	// Saving the positions to the database
	log.Println("Saving new positions to the database 🚀...")
	tea.SavePositionsToDB(db, positions, tableName)

	log.Println("Finished 🫡!")
}
