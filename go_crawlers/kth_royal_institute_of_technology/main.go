package main

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	"fenjan.ai-hue.ir/tea"
	"github.com/gocolly/colly"
)

type Position = tea.Position

// extractPositionDescription function uses colly to scrape a position page and finds the title and text of the position
func extractPositionDescription(url string) string {
	// Create a new collector
	c := colly.NewCollector()
	var description string
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
		log.Println("Visiting", r.URL)
	})
	c.Visit(url)
	return description
}

func getPositions(visitedUrls map[string]bool) []Position {

	c := colly.NewCollector()
	positions := []Position{}
	// Find and visit all links
	c.OnHTML("tr", func(e *colly.HTMLElement) {
		title := e.ChildText("a")
		url := e.ChildAttr("a", "href")
		date := e.ChildText("td:last-child")
		if url != "" {
			positions = append(positions, Position{title, url, "", date})
		}
	})

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL)
	})
	c.Visit("https://www.kth.se/en/om/work-at-kth/doktorander-1.572201")
	return positions
}

func main() {

	// Define table name for  KTH Royal Institute of Technology positions
	tableName := "kth_se"

	log.Println("Connecting to the 'fenjan' database 🐰.")
	db, err := sql.Open("mysql", tea.GetDbConnectionString())
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Creating the " + tableName + " table in the 'fenjan' database if not exists 👾.")
	tea.CreateTableIfNotExists(db, tableName)

	// Get the URLs from the database
	visitedUrls := tea.GetUrlsFromDB(db, tableName)

	log.Println("Searching the KTH Royal Institute of Technology  for the Ph.D. vacancies 🦉.")
	positions := getPositions(visitedUrls)
	newPositions := []Position{}

	// Loop through each URL and get the title and text of each position
	for _, position := range positions[:] {
		// Check if the URL has been visited before
		if visitedUrls[position.URL] {
			log.Println("URL has been visited before:", position.URL)
			continue
		}

		// Extract description of new positions and add them to a new slice
		description := extractPositionDescription(position.URL)
		newPositions = append(newPositions, Position{position.Title, position.URL, description, position.Date})
	}

	log.Println("Extracted description of", len(newPositions), "open positions 🤓.")

	log.Println("Saving new positions to the database 🚀...")
	tea.SavePositionsToDB(db, newPositions, tableName)

	log.Println("Finished 🫡!")

}
