package main

import "github.com/Gyscos/benchbase"

func MakeBenchmark(times map[string]TimeSerie) *benchbase.Benchmark {

	benchmark := benchbase.NewBenchmark()

	for api, serie := range times {
		for component, time := range serie.Mean() {
			benchmark.Result[api+"."+component] = time
		}
	}

	return benchmark
}
