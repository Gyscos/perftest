package main

import (
	"fmt"
	"strings"
)

// Lots of times I guess ?
type Times map[string]interface{}

// TimeSeries map component names to a serie of time results
type TimeSerie map[string][]float64

func (ts *TimeSerie) Add(t Times) {
	computeTotal(t)
	ts.AddPrefix(t, "")
}

func (ts *TimeSerie) AddPrefix(t Times, prefix string) {
	for key, value := range t {
		switch value := value.(type) {
		case map[string]interface{}:
			ts.AddPrefix(Times(value), prefix+key+".")

		case float64:
			(*ts)[prefix+key] = append((*ts)[prefix+key], value)
		}
	}
}

func computeTotal(t Times) float64 {
	if v, ok := t["total"]; ok {
		return v.(float64)
	} else {
		total := 0.0

		for _, value := range t {
			switch value := value.(type) {
			case map[string]interface{}:
				total += computeTotal(Times(value))
			case float64:
				total += value
			}
		}

		t["total"] = total
		return total
	}
}

func (ts *TimeSerie) Mean() map[string]float64 {
	result := make(map[string]float64)

	for k, v := range *ts {
		sum := 0.0
		invN := 1.0 / float64(len(v))

		for _, t := range v {
			sum += t * invN
		}

		result[k] = sum
	}

	return result
}

type TimeSet map[string]TimeSerie

func (ts *TimeSet) Add(api string, time Times) {
	serie, ok := (*ts)[api]
	if !ok {
		serie = TimeSerie(make(map[string][]float64))
		(*ts)[api] = serie
	}
	serie.Add(time)
}

type TimedResponse struct {
	Stats struct {
		Times struct {
			Processing Times
		}
	}
}

func (tr *TimedResponse) GetTimes() Times {
	return tr.Stats.Times.Processing
}

func sumAll(times []*Times) Times {
	result := Times(make(map[string]interface{}))
	n := len(times)

	for _, t := range times {
		for key, value := range *t {
			//result[key] += value
			var currentFloat float64
			current := result[key]
			if current != nil {
				currentFloat = current.(float64)
			}
			result[key] = currentFloat + value.(float64)/float64(n)
		}
	}

	return result
}

func (t *TimeSerie) PrettyPrint(depth int) string {
	meanTimes := t.Mean()
	if depth > 0 {
		for k, v := range meanTimes {
			l := strings.Split(k, ".")
			if len(l) > depth {
				if len(l) == depth+1 && l[len(l)-1] == "total" {
					// Keep it but rename it
					meanTimes[strings.Join(l[:len(l)-1], ".")] = v
				}
				delete(meanTimes, k)
			}
		}
	}
	return fmt.Sprint(meanTimes)
}
