package main

import (
	"bufio"
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"golang.org/x/net/http2"
	// "golang.org/x/net/http2"
)

type workerConfig struct {
	outF       *os.File
	httpConfig http.Client
	userAgent  string
	host       string
	timeout    int
	nFollow    int
}

type config struct {
	base    string
	nWorker int
	verbose bool
}

func Setup() (*workerConfig, *config, []string) {
	base := flag.String("base", "", "scheme://host to call")
	host := flag.String("host", "", "host header to pass if different from the base")
	file := flag.String("file", "uris.list", "file to examine, it's a url per line")
	nWorker := flag.Int("num-worker", 4, "number of workers")
	timeout := flag.Int("timeout", 30, "timeout in seconds")
	userAgent := flag.String("user-agent", "curl/go.jtheo", "user agent")
	nFollow := flag.Int("num-follow", 10, "number of redirects to follow")
	verbose := flag.Bool("verbose", false, "verbose")
	http2Disable := flag.Bool("http2disable", true, "http2 enabled")
	http2Insecure := flag.Bool("k", false, "http2 insecure TLS")
	logFile := flag.String("log", "", "log file, if no file is given, the log will be on the stdout")
	// errFile := flag.String("err", "", "log file, if no file is given, the log will be on the stderr")

	flag.Parse()

	if *file == "" || *base == "" {
		flag.Usage()
		os.Exit(0)
	}

	if *host == "" {
		splitURL, err := url.Parse(*base)
		if err != nil {
			log.Fatal(err)
		}

		*host = splitURL.Hostname()
	}

	var outF *os.File
	// var errF *os.File
	var err error

	if *logFile != "" {
		outF, err = os.OpenFile(*logFile, os.O_CREATE|os.O_WRONLY, 0o644)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		outF = os.Stdout
	}

	// if *errFile != "" {
	// 	errF, err := os.OpenFile(*logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}
	// }
	// var transport struct

	wc := &workerConfig{
		httpConfig: http.Client{
			Timeout: time.Duration(*timeout) * time.Second,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
		timeout:   *timeout,
		userAgent: *userAgent,
		host:      *host,
		nFollow:   *nFollow,
		outF:      outF,
	}

	if *http2Disable {
		wc.httpConfig.Transport = &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   10 * time.Second,
				KeepAlive: 10 * time.Second,
			}).DialContext,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 4 * time.Second,
			ResponseHeaderTimeout: 3 * time.Second,
		}
	} else {
		wc.httpConfig.Transport = &http2.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: *http2Insecure,
			},
		}
		wc.userAgent += " http/2"
	}

	f, err := os.Open(*file)
	if err != nil {
		fmt.Printf("Error opening the file with the list of URIs: %s: %v\n", *file, err)
		flag.Usage()
		os.Exit(1)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)

	scanner.Split(bufio.ScanLines)

	list := []string{}

	for scanner.Scan() {
		uri := scanner.Text()
		if uri == "" {
			continue
		}

		if !strings.HasPrefix(uri, "/") {
			uri = "/" + uri
		}

		list = append(list, *base+uri)
	}

	c := &config{
		verbose: *verbose,
		nWorker: *nWorker,
		base:    *base,
	}

	return wc, c, list
}
