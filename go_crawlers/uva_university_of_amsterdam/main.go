// package main

// import (
// 	"database/sql"
// 	"fmt"
// 	"log"
// 	"os"
// 	"strings"

// 	_ "github.com/go-sql-driver/mysql"
// 	"github.com/gocolly/colly"
// 	"github.com/joho/godotenv"
// )

// // extractPositionDetails function uses colly to scrape a position page and finds the title and text of the position
// func extractPositionDetails(url string) string {
// 	// Create a new collector
// 	c := colly.NewCollector()
// 	var description string
// 	c.OnHTML("div.content-wrap", func(e *colly.HTMLElement) {
// 		description = e.Text

// 		junkText := `$(document).ready(function() {
//             function initializeAddThis() {
//                 $.getScript('//s7.addthis.com/js/300/addthis_widget.js#pubid=mynetwork', function() {
//                     addthis_config = {
//                         username: 'mynetwork',
//                         ui_language  : 'en'
//                     }
//                     var addthis_share = {
//                         url_transforms : {
//                             add: {
//                                 referring_service:
//                                 '{{code}}'
//                                                                             }
//                         }
//                     }
//                 });
//             }

//                             document.addEventListener('cookieConsent', function(event) {
//                     if (event.detail.cookieClassification.includes('Marketing')) {
//                         initializeAddThis();
//                     }
//                 });
//                     });
// `
// 		description = strings.Replace(description, junkText, "", -1)
// 	})
// 	// Add the OnRequest function to log the URLs that are visited
// 	c.OnRequest(func(r *colly.Request) {
// 		log.Println("Visiting", r.URL)
// 	})
// 	c.Visit(url)
// 	return description
// }

// func getPositions(visitedUrls map[string]bool) []Position {

// 	c := colly.NewCollector()
// 	positions := []Position{}
// 	// Find and visit all links
// 	c.OnHTML("tr", func(e *colly.HTMLElement) {
// 		title := e.ChildText("a")
// 		url := e.ChildAttr("a", "href")
// 		date := e.ChildText("td:last-child")
// 		if url != "" {
// 			positions = append(positions, Position{title, url, "", date})
// 		}
// 	})

// 	c.OnRequest(func(r *colly.Request) {
// 		fmt.Println("Visiting", r.URL)
// 	})
// 	c.Visit("https://www.kth.se/en/om/work-at-kth/doktorander-1.572201")
// 	return positions
// }

// func main() {
// 	// Load environment variables from .env file
// 	err := godotenv.Load()
// 	if err != nil {
// 		log.Fatal("Error loading .env file")
// 	}

// 	log.Println("Connecting to the 'fenjan' database 🐰.")
// 	db, err := sql.Open("mysql", getDbConnectionString())
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	log.Println("Creating the 'uva_nl' table in the 'fenjan' database if not exists 👾.")
// 	createTableIfNotExists(db)

// 	// Get the URLs from the database
// 	visitedUrls := getUrlsFromDB(db)

// 	log.Println("Searching the KTH Royal Institute of Technology  for the Ph.D. vacancies 🦉.")
// 	positions := getPositions(visitedUrls)
// 	newPositions := []Position{}
// 	// Loop through each URL and get the title and text of each position
// 	for _, position := range positions[:] {
// 		// Check if the URL has been visited before
// 		if visitedUrls[position.URL] {
// 			log.Println("URL has been visited before:", position.URL)
// 			continue
// 		}
// 		description := extractPositionDetails(position.URL)

// 		newPositions = append(newPositions, Position{position.Title, position.URL, description, position.Date})
// 	}

// 	log.Println("Extracted details of", len(positions), "open positions 🤓.")

// 	log.Println("Saving new positions to the database 🚀...")
// 	savePositionsToDB(db, newPositions)

// 	log.Println("Finished 🫡!")

// }

package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gocolly/colly"
	"github.com/joho/godotenv"
)

type Position struct {
	Title       string `json:"title"`
	URL         string `json:"url"`
	Description string `json:"description"`
	Date        string `json:"date"`
}

// https://play.golang.org/p/Qg_uv_inCek
// contains checks if a string is present in a slice
func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}

// getDbConnectionString function reads the database connection details from environment variables
func getDbConnectionString() string {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	username := os.Getenv("DB_USERNAME")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")
	return username + ":" + password + "@tcp(" + host + ":" + port + ")/" + dbname
}

// createTableIfNotExists function creates the "uva_nl" table if it doesn't already exist
func createTableIfNotExists(db *sql.DB) {
	// SQL statement to create the table
	query := `CREATE TABLE IF NOT EXISTS uva_nl (
		id INT AUTO_INCREMENT PRIMARY KEY,
		title VARCHAR(255) NOT NULL,
		url VARCHAR(255) NOT NULL,
		description TEXT NOT NULL,
		date VARCHAR(255) NOT NULL
	)`

	// Execute the SQL statement
	_, err := db.Exec(query)
	if err != nil {
		log.Fatal(err)
	}
}

// savePositionsToDB function saves the scraped positions to a MySQL database
func savePositionsToDB(db *sql.DB, positions []Position) {
	// Prepare the SQL statement
	stmt, err := db.Prepare("INSERT INTO uva_nl (title, url, description, date) VALUES (?, ?, ?, ?)")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	// Loop through each position and execute the SQL statement
	for _, position := range positions {
		_, err := stmt.Exec(position.Title, position.URL, position.Description, position.Date)
		if err != nil {
			log.Fatal(err)
		}
	}
}

// getUrlsFromDB function return a map of all urls in the "positions" table
func getUrlsFromDB(db *sql.DB) map[string]bool {
	// SELECT statement to retrieve URLs from the table
	rows, err := db.Query("SELECT url FROM uva_nl")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	// Create a map to store the URLs
	urls := make(map[string]bool)
	for rows.Next() {
		var url string
		if err := rows.Scan(&url); err != nil {
			log.Fatal(err)
		}
		urls[url] = true
	}
	return urls
}

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

func getPositionsURL(numPositions int) []string {
	var positionsURL []string
	for startRow := 0; startRow < numPositions; startRow += 10 {
		url := fmt.Sprintf("https://vacatures.uva.nl/UvA/tile-search-results/?q=phd&startrow=%d", startRow)
		c := colly.NewCollector()
		c.OnHTML("a[href]", func(e *colly.HTMLElement) {
			href := "https://vacatures.uva.nl" + e.Attr("href")

			if strings.Contains(strings.ToLower(e.Text), "phd") && !contains(positionsURL, href) {
				positionsURL = append(positionsURL, href)
			}
		})
		c.OnRequest(func(r *colly.Request) {
			fmt.Println("Visiting", r.URL)
		})
		c.Visit(url)
	}
	return positionsURL
}

func extractPositionDetails(positionsURL []string, visitedUrls map[string]bool) []Position {
	positions := []Position{}
	for _, url := range positionsURL {
		var title, date, description string
		c := colly.NewCollector()
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

		c.OnRequest(func(r *colly.Request) {
			fmt.Println("Visiting", r.URL)
		})
		c.Visit(url)
	}

	return positions
}

func main() {

	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	log.Println("Connecting to the 'fenjan' database 🐰.")
	db, err := sql.Open("mysql", getDbConnectionString())
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Creating the 'kth_se' table in the 'fenjan' database if not exists 👾.")
	createTableIfNotExists(db)

	// Get the URLs from the database
	visitedUrls := getUrlsFromDB(db)

	log.Println("Searching the KTH Royal Institute of Technology  for the Ph.D. vacancies 🦉.")

	numPositions := findNumActivePositions()
	positionsURL := getPositionsURL(numPositions)
	log.Println("Found ", len(positionsURL), "open positions.")
	positions := extractPositionDetails(positionsURL, visitedUrls)
	log.Println("Extracted details of", len(positions), "open positions 🤓.")
	log.Println("Saving new positions to the database 🚀...")
	savePositionsToDB(db, positions)

	log.Println("Finished 🫡!")
}
