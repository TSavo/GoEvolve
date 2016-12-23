package goevolve

import (
	"bufio"
	"os"
)

type Dictionary []string

func readLines(path string) []string {
	file, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines
}

func NewDictionary(name string) *Dictionary {
	x := Dictionary(readLines(name))
	return &x
}

func (dict *Dictionary) RandomWord() string {
	return (*dict)[rng.Int()%len(*dict)]
}

var USDict = NewDictionary("US.dic")
