package main

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type response struct {
	err        error
	initialURI string
	finalURI   string
	status     int
	nRedir     int
	latency    time.Duration
}

func (r response) String() string {
	if r.err == nil {
		r.err = fmt.Errorf("no")
	}

	ret := fmt.Sprintf("%q,%q,%03d,%d,%v,%v",
		r.initialURI, r.finalURI, r.status, r.nRedir, r.err, r.latency)

	return ret
}

func worker(nextURL string, wc workerConfig) {
	start := time.Now()

	var i int

	var r response

	if nextURL == "" {
		return
	}

	if !strings.HasPrefix(nextURL, "http") {
		nextURL = "http://" + nextURL
	}

	fullParse, err := url.Parse(nextURL)
	if err != nil {
		fmt.Fprintln(wc.outF, "parse failed", response{
			initialURI: "",
			status:     0,
			finalURI:   nextURL,
			err:        err,
			nRedir:     0,
			latency:    time.Since(start),
		})

		return
	}

	if fullParse.Path == "" {
		return
	}

	if !strings.HasPrefix(fullParse.Path, "/") {
		r = response{
			initialURI: fullParse.Path,
			finalURI:   nextURL,
			status:     0,
			err:        fmt.Errorf("the path is missing the initial slash: %s", fullParse.Path),
			nRedir:     0,
			latency:    0,
		}
		fmt.Fprintln(wc.outF, r)

		return
	}

	for i < wc.nFollow {

		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, nextURL, nil)
		if err != nil {
			r = response{
				initialURI: fullParse.Path,
				finalURI:   nextURL,
				status:     0,
				err:        fmt.Errorf("new Request failed: %w", err),
				nRedir:     i,
				latency:    time.Since(start),
			}
			fmt.Fprintln(wc.outF, r)

			return
		}

		req.Header.Add("User-Agent", wc.userAgent)
		req.Host = wc.host

		resp, err := wc.httpConfig.Do(req)
		if err != nil {
			r = response{
				initialURI: fullParse.Path,
				finalURI:   nextURL,
				status:     0,
				err:        fmt.Errorf("do get %v: %w", resp, err),
				nRedir:     i,
				latency:    time.Since(start),
			}

			break
		}

		defer resp.Body.Close()

		if resp.StatusCode < 300 || resp.StatusCode >= 400 {
			r = response{
				initialURI: fullParse.Path,
				finalURI:   resp.Request.URL.Path,
				status:     resp.StatusCode,
				err:        nil,
				nRedir:     i,
				latency:    time.Since(start),
			}

			break
		}

		if resp.StatusCode >= 300 && resp.StatusCode < 400 {
			// location := resp.Header.Get("Location")
			//
			// if strings.HasPrefix(location, "/") {
			location := resp.Header.Get("Location")
			if err != nil {
				panic(err)
			}
			if strings.HasPrefix(location, "/") {
				location = resp.Request.URL.Scheme + "://" + resp.Request.URL.Host + location
			}
			// }
			// nextURL = resp.Header.Get("Location")
			nextURL = location
			i++
		}
	}

	if r == (response{}) {
		r = response{
			initialURI: fullParse.Path,
			finalURI:   nextURL,
			status:     0,
			err:        fmt.Errorf("looping redirects over %d times", wc.nFollow),
			nRedir:     i,
			latency:    time.Since(start),
		}
	}
	r.latency = r.latency.Round(time.Microsecond)

	if r.finalURI == "" {
		r.finalURI = "/"
	}
	fmt.Fprintln(wc.outF, r)
}
