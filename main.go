package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

func main() {
	// These should never really change
	var token string
	var dataFile string

	// Host will probably always be localhost...
	var host string
	// Repeat shouldn't change the results too much..
	var repeat int

	// This changes the test configuration
	var threads int
	var threadsPerCPU int

	var errlog string
	var benchbaseServer string
	var benchHost string
	var benchRev string

	var random bool
	var forceAnalyze bool

	flag.StringVar(&token, "token", "", "Token to use.")
	flag.StringVar(&dataFile, "f", "urls.txt", "URL file to test.")

	flag.StringVar(&host, "host", "http://localhost:8080", "API Server to run calls on.")
	flag.IntVar(&repeat, "repeat", 10, "Number of times to call each URL.")

	flag.IntVar(&threads, "threads", 1, "Number of parallel calls to make.")
	flag.IntVar(&threadsPerCPU, "threadsPerCPU", 0, "Number of parallel calls to make for each CPU. Overrides threads if non zero.")

	flag.StringVar(&errlog, "err", "", "Error log file. Blank for stdout.")
	flag.StringVar(&benchbaseServer, "benchbase", "", "Benchbase server to send results to.")
	flag.StringVar(&benchHost, "benchHost", "", "Identifier to label the benchmark with.")
	flag.StringVar(&benchRev, "benchRev", "", "Revision to indicate in the benchmark.")

	flag.BoolVar(&forceAnalyze, "analyze", false, "Always call analyze, no matter the API given in the input file.")
	flag.BoolVar(&random, "random", false, "Randomize the URL list (still stable between cycles).")

	flag.Parse()

	if threadsPerCPU != 0 {
		threads = threadsPerCPU * runtime.NumCPU()
	}

	// Prepare error handler?
	var errW io.Writer = os.Stdout
	if errlog != "" {
		f, err := os.Create(errlog)
		if err == nil {
			errW = f
			defer f.Close()
		}
	}
	errlogger := log.New(errW, "", log.LstdFlags)

	// Read URL file
	urls, err := ReadUrls(dataFile)
	if err != nil {
		log.Fatal(err)
	}

	// Randomize?
	if random {
		urls = Randomize(urls)
	}

	tester := NewTester(host, token, errlogger)
	times, err := tester.Run(urls, repeat, threads, forceAnalyze)
	if err != nil {
		log.Fatal("Error during RUN:", err)
	}

	if benchHost == "" {
		benchHost, _ = os.Hostname()
	}

	if benchRev == "" {
		benchRev = getRev()
	}

	// Now make the benchmark
	bench := MakeBenchmark(times)

	bench.Conf["ForceAnalyze"] = fmt.Sprint(forceAnalyze)

	bench.Subj["Threads"] = fmt.Sprint(threads)
	bench.Subj["ThreadPerCPU"] = fmt.Sprint(threadsPerCPU)
	bench.Subj["Host"] = benchHost
	bench.Subj["Rev"] = benchRev

	b, err := json.Marshal(&bench)
	if err != nil {
		log.Fatal("???", err)
	}
	if benchbaseServer != "" {
		url := benchbaseServer + "/push"
		resp, err := http.Post(url, "application/json", bytes.NewReader(b))
		if err != nil {
			log.Fatal("Could not send data:", err)
		}
		defer resp.Body.Close()
	} else {
		log.Println(string(b))
	}
}

func getRev() string {
	b, err := exec.Command("svnversion").CombinedOutput()
	if err != nil {
		log.Println("Error getting svn version:", err)
		return ""
	}
	return strings.Trim(string(b), " \t\n")
}
