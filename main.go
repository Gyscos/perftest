package main

import (
	"bufio"
	"flag"
	"io"
	"log"
	"math/rand"
	"os"
	"strings"
)

func readUrls(filename string) ([]Query, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	result := make([]Query, 0)

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "#") {
			continue
		}
		words := strings.Split(line, " ")
		if len(words) != 2 {
			continue
		}
		api := words[0]
		url := words[1]
		query := Query{api, url}
		result = append(result, query)
	}

	if err = scanner.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func randomize(queries []Query) []Query {
	dest := make([]Query, len(queries))

	ids := rand.Perm(len(queries))
	for i, v := range ids {
		dest[v] = queries[i]
	}

	return dest
}

func main() {
	var random bool
	var host string
	var token string
	var dataFile string
	var output string
	var n int
	var forceAnalyze bool
	var depth int
	var threads int

	flag.IntVar(&depth, "depth", 0, "Maximum depth to return in the times object. 0 for unlimited.")
	flag.IntVar(&n, "n", 10, "Number of times to call each URL")
	flag.IntVar(&threads, "threads", 1, "Number of parallel calls to make.")
	flag.StringVar(&output, "o", "", "Output file. Blank for stdout.")
	flag.StringVar(&host, "host", "http://localhost:8080", "Hostname")
	flag.StringVar(&token, "token", "", "Token to use")
	flag.StringVar(&dataFile, "f", "urls.txt", "URL file to test")
	flag.BoolVar(&forceAnalyze, "analyze", false, "Always call analyze, no matter the API given in the input file.")
	flag.BoolVar(&random, "random", false, "Randomize the URL list (still stable between cycles).")

	flag.Parse()

	var w io.Writer = os.Stdout
	if output != "" {
		f, err := os.Create(output)
		if err == nil {
			w = f
			defer f.Close()
		}
	}

	urls, err := readUrls(dataFile)
	if err != nil {
		log.Fatal(err)
	}

	// Randomize?
	if random {
		urls = randomize(urls)
	}

	tester := NewTester(host, token, w)
	err = tester.Run(urls, n, threads, forceAnalyze, depth)
	if err != nil {
		log.Println("Error during RUN:", err)
	}
}
