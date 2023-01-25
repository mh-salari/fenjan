package main

import (
	"database/sql"
	"log"
	"math/rand"
	"strings"
	"time"

	"fenjan.ai-hue.ir/logger"
	"fenjan.ai-hue.ir/tea"
	"github.com/gocolly/colly"
)

// Set the university name, the database table name for this university, and the url of vacant positions
var uniName string = "KTH Royal Institute of Technology"
var tableName string = "kth_se"
var vacantPositionsUrl string = "https://www.kth.se/en/om/work-at-kth/doktorander-1.572201"

// Get Position type from tea helper package
type Position = tea.Position

// get the URL, Title and Date of all vacant positions
func getPositionsUrlsAndTitleAndDate() (positions []Position) {

	c := colly.NewCollector()
	c.SetRequestTimeout(60 * time.Second)

	c.OnHTML("tr", func(e *colly.HTMLElement) {
		title := e.ChildText("a")
		url := e.ChildAttr("a", "href")
		date := e.ChildText("td:last-child")

		if url != "" {
			positions = append(positions, Position{title, url, "", date})
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

	return positions
}

// Get the details of position
func getPositionDescription(url string) (description string) {

	c := colly.NewCollector()
	c.SetRequestTimeout(60 * time.Second)

	c.OnHTML("div.content-wrap", func(e *colly.HTMLElement) {
		description = e.Text

		junkText := `$(document).ready(function() {
			function initializeAddThis() {
				$.getScript('//s7.addthis.com/js/300/addthis_widget.js#pubid=mynetwork', function() {
					addthis_config = {
						username: 'mynetwork',
						ui_language  : 'en'
					}
					var addthis_share = {
						url_transforms : {
							add: {
								referring_service:
								'{{code}}'
																			}
						}
					}
				});
			}
	
							document.addEventListener('cookieConsent', function(event) {
					if (event.detail.cookieClassification.includes('Marketing')) {
						initializeAddThis();
					}
				});
					});
	`
		description = strings.Replace(description, junkText, "", -1)
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

	return description

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
	positions := getPositionsUrlsAndTitleAndDate()
	log.Println("Found ", len(positions), " open positions 🐝")

	// Extract description of the positions
	for idx, position := range positions {
		// Check if the URL has been visited before
		if visitedUrls[position.URL] {
			log.Println("URL has been visited before:", position.URL)
			continue
		}
		positions[idx].Description = getPositionDescription(position.URL)
	}
	log.Println("Extracted details of", len(positions), "open positions 🤓.")

	// Saving the positions to the database
	log.Println("Saving new positions to the database 🚀...")
	tea.SavePositionsToDB(db, positions, tableName)

	log.Println("Finished 🫡!")

}
