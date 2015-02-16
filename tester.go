package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

type Tester struct {
	host  string
	token string
	w     io.Writer
}

func NewTester(host string, token string, w io.Writer) *Tester {
	return &Tester{
		host:  host,
		token: token,
		w:     w,
	}
}

func (t *Tester) Run(queries []Query, n int, forceAnalyze bool, depth int) error {
	var apiTimes TimeSet = make(map[string]TimeSerie)
	for i := 0; i < n; i++ {
		log.Printf("-----   CYCLE %3v   -----\n", i)
		for _, query := range queries {
			api := query.api
			if forceAnalyze {
				api = "analyze"
			}
			times, err := t.testUrl(api, query.url)
			if err != nil {
				log.Println("Error testing URL:", err)
				continue
			}
			apiTimes.Add(query.api, times)
			time.Sleep(1 * time.Second)
		}
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

	var tr TimedResponse
	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(&tr)
	if err != nil {
		return nil, err
	}
	result := tr.GetTimes()
	if result == nil {
		return nil, errors.New("could not read time.")
	}

	return result, nil
}
