// Source: https://stackoverflow.com/questions/19965795/how-to-write-log-to-file

package logger

import (
	"flag"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

var (
	Error *log.Logger
)
var err error

var (
	// Get current file full path from runtime
	_, b, _, _ = runtime.Caller(0)

	// Root folder of this project
	ProjectRootPath = filepath.Join(filepath.Dir(b), "../../../")
)

func init() {
	// set location of log file
	var logPath = filepath.Join(ProjectRootPath, "log/scrapers_fatal_errors.log")
	err = os.MkdirAll(filepath.Dir(logPath), os.ModePerm)
	if err != nil {
		panic(err)
	}

	flag.Parse()
	var file, err = os.Create(logPath)
	if err != nil {
		panic(err)
	}

	// create multi-writer to log to both file and terminal
	multiWriter := io.MultiWriter(file, os.Stdout)

	// Get the name of package

	var programPath string
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	exePath := filepath.Dir(ex)
	if filepath.Base(exePath) == "exe" {
		programPath = strings.Split(exePath, "/")[len(strings.Split(exePath, "/"))-4]
	} else {
		programPath = filepath.Base(exePath)
	}
	Error = log.New(multiWriter, "Package "+programPath+"🦂 ", log.Ldate|log.Ltime)

}
