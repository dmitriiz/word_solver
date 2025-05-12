package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"sort"
	"strings"
)

type WordStat struct {
	stat map[rune]int
}

func (ws WordStat) SubsetOf(other WordStat) bool {
	for letter, count := range ws.stat {
		if other.stat[letter] < count {
			return false
		}
	}
	return true
}

func buildWordStat(word string) WordStat {
	stat := make(map[rune]int)
	for _, letter := range word {
		stat[letter]++
	}
	return WordStat{
		stat: stat,
	}
}

func loadWords() []string {
	// downloaded from https://raw.githubusercontent.com/dwyl/english-words/refs/heads/master/words_alpha.txt
	data, err := os.ReadFile("words_alpha.txt")
	if err != nil {
		log.Fatal(err)
	}
	return strings.Split(string(data), "\r\n")
}

func filterWords(words []string, letters string, re string, ss bool) []string {
	var rx *regexp.Regexp
	if re != "" {
		var err error
		rx, err = regexp.Compile(re)
		if err != nil {
			log.Fatal("Invalid regular expression")
		}
	}

	result := make([]string, 0)
	n := len(letters)
	s := buildWordStat(letters)
	for _, word := range words {
		if len(word) > n || len(word) < 3 {
			continue
		}
		stat := buildWordStat(word)
		if stat.SubsetOf(s) {
			if rx != nil && !rx.MatchString(word) {
				continue
			}
			result = append(result, word)
		}
	}

	if ss {
		sort.Slice(result, func(i, j int) bool {
			return len(result[i]) > len(result[j])
		})
	}

	return result
}

func main() {
	re := flag.String("regex", "", "filter words by regexp")
	ss := flag.Bool("sort", false, "sort output by size")

	flag.Parse()
	args := flag.Args()

	if len(args) < 1 {
		log.Fatal("Please provide source letters")
	}

	letters := args[0]

	words := loadWords()
	filtered := filterWords(words, letters, *re, *ss)

	fmt.Println("Filtered words:")
	for _, word := range filtered {
		fmt.Println(word)
	}
}
