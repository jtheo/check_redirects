package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

type positiveResult struct {
	uri    string
	status int 
}

func (p positiveResult) String() string {
	return fmt.Sprintf("%s,%d", p.uri, p.status)
}

type negativeResult struct {
	uri   string
	error error
}

func (n negativeResult) String() string {
	return fmt.Sprintf("%s,%d", n.uri, n.error)
}

func main() {
	filePath := flag.String("file", "", "file with the list of uris")
	urlBase := flag.String("base", "https://www.qa.livehealthily.com", "host website")
	flag.Parse()

	if *filePath == "" {
		log.Println("I need a file with a list of redirects")
		flag.Usage()
		return
	}

	readFile, err := os.Open(*filePath)

	if err != nil {
		fmt.Println(err)
	}
	defer readFile.Close()

	fileScanner := bufio.NewScanner(readFile)
	fileScanner.Split(bufio.ScanLines)
	var uris []string

	for fileScanner.Scan() {
		uris = append(uris, fileScanner.Text())
	}

	var results []positiveResult
	var errors []negativeResult
	start := time.Now()

	fmt.Printf("Processing: ")
	for _, uri := range uris {
		fmt.Printf(".")
		url := *urlBase + uri
		resp, err := http.Get(url)
		if err != nil {
			errors = append(errors, negativeResult{uri: uri, error: err})
		}
		defer resp.Body.Close()
		results = append(results, positiveResult{uri: uri, status: resp.StatusCode})
	}
	fmt.Println()

	lUris := len(uris)
	lErrors := len(errors)
	lSuccess := len(results)

	if lErrors > 0 {
		percentError := float64(lErrors * 100 / lUris)
		fmt.Printf("%d Errors on %d uris, %3.2f%%\n", lErrors, lUris, percentError)

		for _, err := range errors {
			fmt.Println(err)
		}
		fmt.Printf("\n\n")
	}

	if len(results) > 0 {
		percentSuccess := float64(lSuccess * 100 / lUris)
		fmt.Printf("%d Success on %d uris, %3.2f%%\n", lErrors, lUris, percentSuccess)

		for _, res := range results {
			fmt.Println(res)
		}
	}

	fmt.Printf("It took %v\n", time.Since(start))
}
