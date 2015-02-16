package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"
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

func (t *Tester) testUrlChannel(queries <-chan Query, tc chan<- QueryTime, wg *sync.WaitGroup) {
	defer wg.Done()
	for q := range queries {
		times, err := t.testUrl(q.api, q.url)
		if err != nil {
			log.Println("Error testing URL:", err)
			continue
		}
		tc <- QueryTime{q.api, times}
		time.Sleep(1 * time.Second)
	}
}

func (t *Tester) Run(queries []Query, n int, threads int, forceAnalyze bool, depth int) error {
	var apiTimes TimeSet = make(map[string]TimeSerie)
	for i := 0; i < n; i++ {
		log.Printf("-----   CYCLE %3v   -----\n", i)

		var wg sync.WaitGroup
		tc := make(chan QueryTime, 10)
		qc := make(chan Query, 10)
		for j := 0; j < threads; j++ {
			wg.Add(1)
			go t.testUrlChannel(qc, tc, &wg)
		}

		var wg2 sync.WaitGroup
		wg2.Add(1)
		go func() {
			for t := range tc {
				apiTimes.Add(t.api, t.times)
			}
			wg2.Done()
		}()

		for _, query := range queries {
			qc <- query
		}
		close(qc)
		wg.Wait()
		close(tc)
		wg2.Wait()
	}

	for api, times := range apiTimes {

		// meanTime := times.Mean()
		fmt.Fprintf(t.w, "API %v: %v\n", api, times.PrettyPrint(depth))
	}

	// Now, sum the component times for each call and url
	return nil
}

func (t *Tester) testUrl(api string, targetURL string) (Times, error) {
	url := t.host + "/api/" + api + "?version=3&token=" + t.token + "&mentos&stats&admin&url=" + targetURL

	log.Println(url)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var tr TimedResponse
	// dec := json.NewDecoder(resp.Body)
	// err = dec.Decode(&tr)
	err = json.Unmarshal(b, &tr)
	if err != nil {
		return nil, err
	}
	result := tr.GetTimes()
	if result == nil {
		t.el.Println("Error reading time from:", string(b))
		return nil, errors.New("could not read time. Logged.")
	}

	return result, nil
}
