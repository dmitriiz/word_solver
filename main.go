package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"slices"
	"strings"
)

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
	s := wordStat(letters)
	for _, word := range words {
		if len(word) != n {
			continue
		}
		stat := wordStat(word)
		if stat == s {
			result = append(result, word)
		}
	}
	return result
}

func wordStat(word string) string {
	stat := make(map[rune]int)
	for _, letter := range word {
		stat[letter]++
	}
	data := make([]string, 0, len(stat))
	for letter, count := range stat {
		data = append(data, fmt.Sprintf("%c:%d", letter, count))
	}
	slices.Sort(data)
	return strings.Join(data, ",")
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
