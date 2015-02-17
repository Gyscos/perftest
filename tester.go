package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	"github.com/Gyscos/urlspammer"
)

type Tester struct {
	host  string
	token string
	w     io.Writer
	el    *log.Logger
}

func NewTester(host string, token string, w io.Writer, el *log.Logger) *Tester {
	return &Tester{
		host:  host,
		token: token,
		w:     w,
		el:    el,
	}
}
func (t *Tester) makeUrl(api string, target string) string {

	return t.host + "/api/" + api + "?version=3&token=" + t.token + "&mentos&stats&admin&timeout=600000&url=" + target
}

func (t *Tester) Run(queries []Query, n int, threads int, forceAnalyze bool, depth int) error {
	var apiTimes TimeSet = make(map[string]TimeSerie)

	for i := 0; i < n; i++ {
		log.Printf("-----   CYCLE %3v   -----\n", i)

		// Fill the query channel
		qc := make(chan urlspammer.Query, 20)
		go func() {
			for _, q := range queries {
				api := q.api
				if forceAnalyze {
					api = "analyze"
				}
				qc <- urlspammer.Query{
					Url:  t.makeUrl(api, q.url),
					Data: q.api,
				}
			}
			close(qc)
		}()

		// Read from the time channel and add to the results
		tc := make(chan QueryTime, 10)
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			for t := range tc {
				apiTimes.Add(t.api, t.times)
			}
		}()

		urlspammer.SpamByThread(threads, qc, func(q urlspammer.Query, body []byte, d time.Duration) {
			log.Printf("[%v] (%v) %v\n", d, q.Data, q.Url)

			// Read the time object
			var tr TimedResponse
			err := json.Unmarshal(body, &tr)
			if err != nil {
				log.Println(err)
				return
			}
			// Retrieve what interests us
			result := tr.GetTimes()
			if result == nil {
				t.el.Println("Error reading time from:", string(body))
				log.Println("Error: could not read time. Logged.")
				return
			}

			// PROCESS IT!
			tc <- QueryTime{q.Data, result}
		})
		close(tc)
		// All URL have been called now.

		wg.Wait()
		// All times have been compiled now

	}

	for api, times := range apiTimes {

		// meanTime := times.Mean()
		fmt.Fprintf(t.w, "API %v: %v\n", api, times.PrettyPrint(depth))
	}

	// Now, sum the component times for each call and url
	return nil
}
