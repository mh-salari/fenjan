package main

import (
	"database/sql"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gocolly/colly"

	"fenjan.ai-hue.ir/tea"
)

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
	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL)
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
	}
	return positionsURL
}

// Extract details of positions
func extractPositionDetails(positionsURL []string, visitedUrls map[string]bool) []Position {
	positions := []Position{}
	for _, url := range positionsURL {
		var title, date, description string
		c := colly.NewCollector()

		// Set request timeout timer
		c.SetRequestTimeout(60 * 5 * time.Second)

		// Check if the URL has been visited before
		if visitedUrls[url] {
			log.Println("URL has been visited before:", url)
			continue
		}

		c.OnHTML("h1", func(e *colly.HTMLElement) {
			title = strings.TrimSpace(e.Text)
		})

		c.OnHTML("span[data-careersite-propertyid=customfield3]", func(e *colly.HTMLElement) {
			date = strings.TrimSpace(e.Text)
		})
		c.OnHTML("span.jobdescription", func(e *colly.HTMLElement) {
			description = strings.TrimSpace(e.Text)
		})

		c.OnScraped(func(r *colly.Response) {

			positions = append(positions, Position{title, url, description, date})

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
	}

	return positions
}

func main() {

	// Define name of the table for the University of Helsinki positions
	tableName := "uva_nl"

	log.Println("Connecting to the 'fenjan' database 🐰.")
	db, err := sql.Open("mysql", tea.GetDbConnectionString())
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Creating the '" + tableName + "' table in the 'fenjan' database if not exists 👾.")
	tea.CreateTableIfNotExists(db, tableName)

	// Get the URLs from the database
	visitedUrls := tea.GetUrlsFromDB(db, tableName)

	log.Println("Searching the UvA University of Amsterdam  for the Ph.D. vacancies 🦉.")

	log.Println("Finding the total number of open positions advertised on the university website 🦎...")
	numPositions := findNumActivePositions()

	if numPositions == 0 {
		log.Fatal("There is an error in getting the data, The total number of open positions in equal to 0 ☠️ !")
	}
	log.Printf("Currently, there are %d open positions advertised on the website.", numPositions)

	positionsURL := getPositionsURL(numPositions)

	log.Println("Found ", len(positionsURL), "open positions.")
	positions := extractPositionDetails(positionsURL, visitedUrls)
	log.Println("Extracted details of", len(positions), "Ph.D. new open positions 🤓.")
	log.Println("Saving new positions to the database 🚀...")
	tea.SavePositionsToDB(db, positions, tableName)

	log.Println("Finished 🫡!")
}
