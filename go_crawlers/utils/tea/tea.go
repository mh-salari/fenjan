// https://go.dev/doc/tutorial/call-module-code
package tea

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"

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

var (
	// Get current file full path from runtime
	_, b, _, _ = runtime.Caller(0)

	// Root folder of this project
	ProjectRootPath = filepath.Join(filepath.Dir(b), "../../")
)

// getDbConnectionString function reads the database connection details from environment variables
func GetDbConnectionString() string {

	// Load environment variables from .env file
	err := godotenv.Load(ProjectRootPath + "/.env")
	if err != nil {
		log.Fatal("Error loading .env file", err)
	}

	// Make the string needed to connect to the database base on the environment variables
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	username := os.Getenv("DB_USERNAME")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")
	return username + ":" + password + "@tcp(" + host + ":" + port + ")/" + dbname
}

// CreateTableIfNotExists function creates the __tableName__ table in the database if it doesn't already exist
func CreateTableIfNotExists(db *sql.DB, tableName string) {
	// SQL statement to create the table
	query := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
		id INT AUTO_INCREMENT PRIMARY KEY,
		title VARCHAR(500) NOT NULL,
		url VARCHAR(255) NOT NULL,
		description TEXT NOT NULL,
        date VARCHAR(255),
		scraped_on TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`, tableName)

	// Execute the SQL statement
	_, err := db.Exec(query)
	if err != nil {
		log.Fatal(err)
	}
}

// SavePositionsToDB function saves the scraped positions to a MySQL database
func SavePositionsToDB(db *sql.DB, positions []Position, tableName string) {
	// Prepare the SQL statement
	query := fmt.Sprintf("INSERT INTO %s (title, url, description, date, scraped_on) VALUES (?, ?, ?, ?, NOW())", tableName)

	stmt, err := db.Prepare(query)
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

// GetUrlsFromDB function return a map of all urls in the "positions" table
func GetUrlsFromDB(db *sql.DB, tableName string) map[string]bool {
	// SELECT statement to retrieve URLs from the table
	query := fmt.Sprintf("SELECT url FROM %s", tableName)

	rows, err := db.Query(query)
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

// https://play.golang.org/p/Qg_uv_inCek
// contains checks if a string is present in a slice
func Contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}

// https://github.com/gocolly/colly/issues/657
func RetryRequest(r *colly.Response, maxRetries int) int {
	retriesLeft := maxRetries
	if x, ok := r.Ctx.GetAny("retriesLeft").(int); ok {
		retriesLeft = x
	}
	if retriesLeft > 0 {
		r.Ctx.Put("retriesLeft", retriesLeft-1)
		log.Println("Retrying 🧌!")
		r.Request.Retry()
	}
	return retriesLeft
}
