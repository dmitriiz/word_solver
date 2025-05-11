package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
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

func readLetters(reader *bufio.Reader) string {
	fmt.Println("Please type letters to use:")
	data, err := reader.ReadString('\n')
	if err != nil {
		log.Fatal(err)
	}
	return strings.ToLower(strings.TrimSpace(data))
}

func loadWords() []string {
	// downloaded from https://raw.githubusercontent.com/dwyl/english-words/refs/heads/master/words_alpha.txt
	data, err := os.ReadFile("words_alpha.txt")
	if err != nil {
		log.Fatal(err)
	}
	return strings.Split(string(data), "\r\n")
}

func filterWords(words []string, letters string) []string {
	result := make([]string, 0)
	n := len(letters)
	s := buildWordStat(letters)
	for _, word := range words {
		if len(word) > n || len(word) < 3 {
			continue
		}
		stat := buildWordStat(word)
		if stat.SubsetOf(s) {
			result = append(result, word)
		}
	}
	return result
}

func main() {
	reader := bufio.NewReader(os.Stdin)
	letters := readLetters(reader)

	words := loadWords()
	filtered := filterWords(words, letters)

	fmt.Println("Filtered words:")
	for _, word := range filtered {
		fmt.Println(word)
	}
}
