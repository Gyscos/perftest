package main

import (
	"bufio"
	"math/rand"
	"os"
	"strings"
)

func ReadUrls(filename string) ([]Query, error) {
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

func Randomize(queries []Query) []Query {
	dest := make([]Query, len(queries))

	ids := rand.Perm(len(queries))
	for i, v := range ids {
		dest[v] = queries[i]
	}

	return dest
}
