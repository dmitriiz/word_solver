package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"sort"
	"strings"
	"unicode"
)

type WordStat struct {
	stat map[rune]int
}

type CrossWord struct {
	width  int
	height int
	data   []string
}

type Word struct {
	X          int
	Y          int
	Size       int
	Horizontal bool
}

type Point struct {
	w1 *Word
	w2 *Word
	n1 int
	n2 int
}

type WordEntry struct {
	Word          *Word
	Intersections []Point
	MatchedWords  []string
	Guess         *string
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

func loadCrossWord(name string) *CrossWord {
	file, err := os.ReadFile(name)
	if err != nil {
		log.Fatal(err)
	}

	first := -1
	last := -1
	size := 0

	data := strings.Split(string(file), "\n")
	for i, line := range data {
		s := strings.TrimRightFunc(line, unicode.IsSpace)
		n := len(s)
		if n > 0 {
			if first < 0 {
				first = i
			}
			last = i
			if size < n {
				size = n
			}
		}
		data[i] = s
	}

	if first < 0 || last < 0 || size == 0 {
		log.Fatal("Invalid crossword")
	}

	data = data[first : last+1]
	for i, line := range data {
		n := len(line)
		if n < size {
			data[i] = line + strings.Repeat(" ", size-n)
		}
	}

	return &CrossWord{
		width:  size,
		height: last - first + 1,
		data:   data,
	}
}

func (cw *CrossWord) Print() {
	for _, line := range cw.data {
		fmt.Println(line)
	}
}

func (cw *CrossWord) Get(x, y int) rune {
	return rune(cw.data[y][x])
}

func (cw *CrossWord) Put(x, y int, c rune) {
	b := []uint8(cw.data[y])
	b[x] = uint8(c)
	cw.data[y] = string(b)
}

func (cw *CrossWord) findWords(horizontal bool) []Word {
	var x int
	var y int
	var f func(int, int) (int, int)
	if horizontal {
		x = cw.width
		y = cw.height
		f = func(x, y int) (int, int) {
			return x, y
		}
	} else {
		x = cw.height
		y = cw.width
		f = func(x, y int) (int, int) {
			return y, x
		}
	}

	result := make([]Word, 0)
	for j := 0; j < y; j++ {
		n1 := -1
		n2 := -1
		for i := 0; i < x; i++ {
			c := cw.Get(f(i, j))
			if unicode.IsSpace(c) {
				if n1 >= 0 {
					n := n2 - n1 + 1
					if n > 2 {
						x, y := f(n1, j)
						result = append(result, Word{
							X:          x,
							Y:          y,
							Size:       n,
							Horizontal: horizontal,
						})
					}
					n1 = -1
					n2 = -1
				}
			} else {
				if n1 < 0 {
					n1 = i
					n2 = i
				} else {
					n2 = i
				}
			}
		}
		if n1 >= 0 {
			n := n2 - n1 + 1
			if n > 2 {
				x, y := f(n1, j)
				result = append(result, Word{
					X:          x,
					Y:          y,
					Size:       n,
					Horizontal: horizontal,
				})
			}
		}
	}

	return result
}

func (w *Word) Intersect(other *Word) (Point, bool) {
	if w.Horizontal != other.Horizontal {
		return Point{}, false
	}

	if w.Horizontal {
		if w.X <= other.X && w.X+w.Size > other.X && other.Y <= w.Y && other.Y+other.Size > w.Y {
			return Point{
				w1: w,
				w2: other,
				n1: other.X - w.X,
				n2: w.Y - other.Y,
			}, true
		}
	} else {
		if w.Y <= other.Y && w.Y+w.Size > other.Y && other.X <= w.X && other.X+other.Size > w.X {
			return Point{
				w1: w,
				w2: other,
				n1: other.Y - w.Y,
				n2: w.X - other.X,
			}, true
		}
	}

	return Point{}, false
}

func (w *Word) filterWords(crossford *CrossWord, words []string) []string {
	mask := "^"
	exp := false
	for i := 0; i < w.Size; i++ {
		var c rune
		if w.Horizontal {
			c = crossford.Get(w.X+i, w.Y)
		} else {
			c = crossford.Get(w.X, w.Y+i)
		}
		if c == '?' {
			mask += "."
		} else {
			mask += string(c)
			exp = true
		}
	}
	mask += "$"

	var re *regexp.Regexp
	if exp {
		re, _ = regexp.Compile(mask)
	}

	result := make([]string, 0)
	for _, word := range words {
		if len(word) != w.Size {
			continue
		}
		if exp && !re.MatchString(word) {
			continue
		}
		result = append(result, word)
	}
	return result
}

func (e *WordEntry) matchWords(allWords map[*Word]*WordEntry) []string {
	result := make([]string, 0)
	for _, word := range e.MatchedWords {
		if len(e.Intersections) == 0 {
			result = append(result, word)
		} else {
			matched := true
			for _, p := range e.Intersections {
				var o *Word
				var n int
				if p.w1 == e.Word {
					o = p.w2
					n = p.n2
				} else {
					o = p.w1
					n = p.n1
				}
				other := allWords[o]

				if other.Guess == nil {
					continue
				}

				c := (*other.Guess)[n]
				if c != word[n] {
					matched = false
					break
				}
			}

			if matched {
				result = append(result, word)
			}
		}
	}
	return result
}

func (e *WordEntry) Push(crossword *CrossWord) {
	if e.Guess == nil {
		log.Fatal("No guess")
	}
	for i := 0; i < e.Word.Size; i++ {
		c := rune((*e.Guess)[i])
		if e.Word.Horizontal {
			crossword.Put(e.Word.X+i, e.Word.Y, c)
		} else {
			crossword.Put(e.Word.X, e.Word.Y+i, c)
		}
	}
}

func mapWordsToEntries(words []Word) []WordEntry {
	result := make([]WordEntry, 0)
	for _, word := range words {
		result = append(result, WordEntry{
			Word:          &word,
			Intersections: make([]Point, 0),
		})
	}
	return result
}

func solve(crossword *CrossWord, allWords map[*Word]*WordEntry, restWords []*WordEntry) bool {
	first := restWords[0]
	matchedWords := first.matchWords(allWords)
	if len(matchedWords) == 0 {
		return false
	}
	rest := restWords[1:]
	if len(rest) == 0 {
		first.Guess = &matchedWords[0]
		return true
	}
	for _, word := range matchedWords {
		first.Guess = &word
		if solve(crossword, allWords, rest) {
			return true
		}
	}
	return false
}

func main() {
	re := flag.String("regex", "", "filter words by regexp")
	ss := flag.Bool("sort", false, "sort output by size")
	cw := flag.String("crossword", "", "crossword file")

	flag.Parse()
	args := flag.Args()

	if len(args) < 1 {
		log.Fatal("Please provide source letters")
	}

	letters := args[0]

	words := loadWords()
	filtered := filterWords(words, letters, *re, *ss)

	if *cw == "" {
		fmt.Println("Filtered words:")
		for _, word := range filtered {
			fmt.Println(word)
		}
		return
	}

	crossword := loadCrossWord(*cw)
	//crossword.Print()

	hWords := mapWordsToEntries(crossword.findWords(true))
	vWords := mapWordsToEntries(crossword.findWords(false))
	for _, hw := range hWords {
		for _, vw := range vWords {
			p, ok := hw.Word.Intersect(vw.Word)
			if ok {
				hw.Intersections = append(hw.Intersections, p)
				vw.Intersections = append(vw.Intersections, p)
			}
		}
	}

	rest := make([]*WordEntry, 0)
	allWords := make(map[*Word]*WordEntry)
	for _, e := range append(hWords, vWords...) {
		rest = append(rest, &e)
		allWords[e.Word] = &e
		e.MatchedWords = e.Word.filterWords(crossword, filtered)
		if len(e.MatchedWords) == 0 {
			log.Fatal("No matches")
		}
	}

	n := len(allWords)
	if n == 0 {
		log.Fatal("No words found in crossword")
	}
	fmt.Println("Found words:", n)

	solved := solve(crossword, allWords, rest)
	if solved {
		for _, e := range rest {
			e.Push(crossword)
		}
		crossword.Print()
	} else {
		fmt.Println("No solution")
	}
}
