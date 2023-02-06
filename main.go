package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {
	// This is not really the best solution I could have come up
	wc, c, list := Setup()

	start := time.Now()

	if c.verbose {
		// fmt.Printf("\n\n%#v\n\n%#v\n\n", *wc, *c)
		fmt.Printf("Start processing %d uris against %s with host %s and %d workers\n\n", len(list), c.base, wc.host, c.nWorker)
		fmt.Fprintf(wc.outF, "%q,%q,%s,%s,%s,%s\n", "Initial URI", "Final URI", "Status Code", "Nr. Rediretc", "Error", "Latency")
	}

	limiter := make(chan struct{}, c.nWorker)

	var wg sync.WaitGroup

	for _, p := range list {
		limiter <- struct{}{}

		wg.Add(1)

		go func(p string) {
			worker(p, *wc)
			wg.Done()
			<-limiter
		}(p)
	}

	wg.Wait()

	if c.verbose {
		fmt.Println("Total time is", time.Since(start))
	}
}
