package extractor

import (
	"log"
	"math"
	"sort"
	"sync"

	"github.com/kevin-zx/wordcut/pkg/corpus/clear"
)

// Builder 构建分词
type Builder struct {
	rightRank      []int
	letters        letters
	maxLen         int
	wm             []*Word
	singleWmn      map[rune]*singWm
	corpusFloatLen float64
}

type singWm struct {
	start int
	end   int
	count int
	freq  float64
}
type letters []rune

// Word 关键词信息
type Word struct {
	Word           string
	Count          int
	Freq           float64
	Poly           float64  // poly 凝固度 越大说明这个词越不可能是两个词
	Flex           float64  // 自由度 越高说明这个词是一个词概率越大
	rightNeighbour []string // 左相邻字符串
	leftNeighbour  []string // 右相邻字符串
	Score          float64
}

// 对自由度进行计算
func (w *Word) generateFlex() {
	if w.rightNeighbour != nil {
		flex := genflex(w.rightNeighbour, float64(w.Count))
		if flex == 0 || w.Flex == 0 || w.Flex > flex {
			w.Flex = flex
		}
		w.rightNeighbour = nil
	}
	if w.leftNeighbour != nil {
		flex := genflex(w.leftNeighbour, float64(w.Count))
		if flex == 0 || w.Flex == 0 || w.Flex > flex {
			w.Flex = flex
		}
		w.leftNeighbour = nil
	}
}

func genflex(rawNeighbour []string, wordCount float64) float64 {
	tr := ""
	tl := 0.0
	flex := 0.0
	for i, r := range rawNeighbour {
		tl++
		if tr == "" {
			tr = r
			continue
		}

		if tr != r || i == len(rawNeighbour)-1 {
			freq := tl / wordCount
			flex += -(freq * math.Log(freq))
			tl = 1
			tr = r
		}
	}
	return flex
}

var lock *sync.Mutex

// NewBuilder 新建Builder
func NewBuilder(corpus []rune, maxLen int) *Builder {
	b := Builder{
		letters:        corpus,
		maxLen:         maxLen,
		corpusFloatLen: float64(len(corpus)),
	}
	lock = new(sync.Mutex)
	b.wm = []*Word{}
	return &b
}

// 最小词频次
const minCount = 10
const minPoly = 5.0
const minFlex = 1.0

// Extract 进行分词计算
func (b *Builder) Extract() []*Word {
	// b.genRank()

	// log.Printf("start calculate right direct\n")
	log.Printf("start gen right rank\n")

	b.rightRank = rankWords(b.letters, b.maxLen)
	b.singleWordN()
	log.Printf("end gen right rank\n")

	b.calculateSideNew()
	return b.score()
}

func (b *Builder) score() []*Word {
	for _, wi := range b.wm {
		wi.Poly = math.Log(wi.Poly + 1)
		wi.Score = wi.Poly * wi.Flex * wi.Freq

	}
	sort.Slice(b.wm, func(i, j int) bool {
		return b.wm[i].Score > b.wm[j].Score
	})
	return b.wm
}

func (b *Builder) singleWordN() {
	b.singleWmn = make(map[rune]*singWm)
	letter := rune(0)
	start := 0
	for i, index := range b.rightRank {
		wr := b.letters[index]
		if letter == 0 {
			letter = wr
		}
		if letter != wr {
			b.singleWmn[letter] = &singWm{start: start, end: i - 1, count: i - start}
			start = i
			letter = wr
		}
	}
	for _, wi := range b.singleWmn {
		wi.freq = float64(wi.count) / b.corpusFloatLen
	}
}

func (b *Builder) calculateSideNew() {
	// 单词的第一个字符
	currLetter := rune(0)
	// // 每个不同字符积累的
	// var words map[string]*Word
	// isLetter := false

	blockLetter := rune(0)
	rankStart := 0
	rankEnd := 0
	// c := runtime.NumCPU()
	c := 1
	tasks := make(chan []int, c)
	w := sync.WaitGroup{}
	w.Add(c)
	for i := 0; i < c; i++ {
		go func() {
			for t := range tasks {
				b.calculateBlock(t[0], t[1])
			}
			w.Done()
		}()
	}
	// 循环排序结果
	for i, index := range b.rightRank {
		if i%10000 == 0 {
			log.Printf("cut words %d/%d = %.3f\n", i, len(b.rightRank), float64(i)/b.corpusFloatLen)
		}

		currLetter = b.letters[index]
		if blockLetter == 0 {
			blockLetter = b.letters[index]
			continue
		}

		if blockLetter != currLetter {
			rankEnd = i - 1
			if blockLetter != ' ' {
				tasks <- []int{rankStart, rankEnd}
			}
			rankStart = i
			blockLetter = currLetter
		}
	}
	close(tasks)
	w.Wait()
}

func (b *Builder) calculateBlock(start, end int) {
	if end-start < 10 {
		return
	}
	first := b.letters[b.rightRank[start]]
	isfirstLetter := clear.IsDigital(first) || clear.IsLetter(first)

	for l := 2; l < b.maxLen; l++ {
		pres := []rune{}
		suffixes := []rune{}
		var word []rune
		count := 0
		isletter := false

		for i := start; i <= end; i++ {
			index := b.rightRank[i]
			if index+l > len(b.letters) {
				continue
			}
			cs := b.letters[index : index+l]
			if !isfirstLetter {
				for i, c := range cs {
					if c == ' ' && i != len(cs)-1 {
						isletter = true
					}
				}
			} else {
				isletter = true
			}

			if isletter {
				if cs[len(cs)-1] == ' ' {
					continue
				}
				// fmt.Println(string(cs))

			}

			if word == nil {
				word = cs
			}
			if !runeEqual(word, cs) {
				if count > 5 && len(pres) != 0 && len(suffixes) != 0 {
					b.calculateLenBlock(word, pres, suffixes, count)
				}

				word = cs
				pres = []rune{}
				suffixes = []rune{}
				count = 1
			}
			if index > 0 {
				pre := b.letters[index-1]
				if pre == ' ' {
					if index > 1 {
						pre = b.letters[index-2]
					}
				}
				if isletter {
					if clear.IsLetter(pre) || clear.IsDigital(pre) {
						continue
					}
				}
				if pre != ' ' {
					pres = append(pres, pre)
				}
			}
			if index+l < len(b.letters) {
				suf := b.letters[index+l]
				if suf == ' ' {
					if index+l+1 < len(b.letters) {
						suf = b.letters[index+l+1]
					}
				}
				if isletter {
					if clear.IsLetter(suf) || clear.IsDigital(suf) {
						continue
					}
				}
				if suf != ' ' {
					suffixes = append(suffixes, suf)
				}
			}

			count++
			// words = append(words, w)
		}
	}
}

func runeEqual(r1, r2 []rune) bool {
	if len(r1) != len(r2) {
		return false
	}
	for i := 0; i < len(r1); i++ {
		if r1[i] != r2[i] {
			return false
		}
	}
	return true
}

func (b *Builder) calculateLenBlock(word, prefix, suffix []rune, count int) {
	flex := b.blockGenFlex(word, prefix, suffix, count)
	if flex < minFlex {
		return
	}
	ploy := b.blockPloyWord(word, count)
	if ploy < minPoly {
		return
	}
	lock.Lock()
	b.wm = append(b.wm, &Word{
		Word:  string(word),
		Count: count,
		Freq:  float64(count) / b.corpusFloatLen,
		Poly:  ploy,
		Flex:  flex,
		Score: 0.0,
	})
	lock.Unlock()

}
func (b *Builder) blockPloyWord(word []rune, wordCount int) float64 {
	pr := b.ployOne(word[:len(word)-1], word[len(word)-1], wordCount)
	if pr < minPoly {
		return pr
	}
	return math.Min(pr, b.ployOne(word[1:], word[0], wordCount))
}

func (b *Builder) ployOne(mainWord []rune, fix rune, wordCount int) float64 {
	fsw := b.singleWmn[fix]
	msw := b.singleWmn[mainWord[0]]
	mc := 0
	if len(mainWord) == 1 {
		mc = msw.count
	} else {
		mc = b.getOneWordCount(mainWord, msw.start, msw.end)
	}
	return (b.corpusFloatLen * float64(wordCount)) / (float64(fsw.count * mc))
}
func (b *Builder) getOneWordCount(word []rune, rankStart int, rankEnd int) int {
	count := 0
	wordL := len(word)

	for {
		if rankStart >= len(b.rightRank) {
			break
		}
		letterIndex := b.rightRank[rankStart]
		rankStart++
		if letterIndex+wordL > len(b.letters) {
			continue
		}
		cw := b.letters[letterIndex : letterIndex+wordL]
		if b.letters[letterIndex] != word[0] {
			break
		}
		if !runeEqual(cw, word) {
			if count == 0 {
				continue
			} else {
				break
			}
		}

		count++

	}
	return count
}

func (b *Builder) findFirst(word []rune, rankStart int, rankEnd int) int {

	wordL := len(word)
	alreadyEq := false
	for {
		if rankEnd == rankStart {
			return rankEnd
		}
		mid := (rankEnd-rankStart)/2 + rankStart
		letterIndex := b.rightRank[mid]
		cw := b.letters[letterIndex : letterIndex+wordL]
		if !runeEqual(cw, word) {
			if !alreadyEq {
				rankEnd = mid
			} else {
				rankStart = mid
				break
			}
		} else {
			rankEnd = mid
			alreadyEq = true
		}

	}
	return -1
}

func (b *Builder) blockGenFlex(word, prefix, suffix []rune, count int) float64 {
	// 因为只从右方向计算，右方向的字符 （suffix）就是有序的，左方向（prefix）就需要排序下
	sort.Slice(prefix, func(i, j int) bool {
		return prefix[i] < prefix[j]
	})
	leftFlex := oneDirectFlex(prefix, float64(count))
	if leftFlex < minFlex {
		return leftFlex
	}
	return math.Min(leftFlex, oneDirectFlex(suffix, float64(count)))

}

func oneDirectFlex(neighbourLetters []rune, count float64) float64 {
	tr := rune(0)
	tl := 0.0
	flex := 0.0
	for i, r := range neighbourLetters {
		tl++
		if tr == 0 {
			tr = r
			continue
		}

		if tr != r || i == len(neighbourLetters)-1 {
			freq := tl / count
			flex += -(freq * math.Log(freq))
			tl = 1
			tr = r
		}
	}
	return flex
}

// rankWords 给切分的关键词排序
func rankWords(words []rune, maxLen int) []int {
	wus := make([]int, len(words))
	for i := range words {
		wus[i] = i
	}
	sort.Slice(wus, func(i, j int) bool {
		for l := 0; l < maxLen; l++ {
			if wus[i]+l >= len(words) || wus[j]+l >= len(words) {
				break
			}
			if words[wus[i]+l] == words[wus[j]+l] {
				continue
			}
			return words[wus[i]+l] < words[wus[j]+l]
		}
		return false
	})

	return wus
}
