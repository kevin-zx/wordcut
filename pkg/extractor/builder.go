package extractor

import (
	"log"
	"math"
	"sort"
	"strings"

	"github.com/kevin-zx/wordcut/pkg/corpus/clear"
)

// Builder 构建分词
type Builder struct {
	left           []int
	right          []int
	letters        letters
	reverseLetters letters
	maxLen         int
	wm             map[string]*Word
	singleWm       map[string]*Word
	corpusFloatLen float64
}

type letters []rune

// Word 关键词信息
type Word struct {
	Word           []rune
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

// NewBuilder 新建Builder
func NewBuilder(corpus []rune, maxLen int) *Builder {
	b := Builder{
		letters:        corpus,
		reverseLetters: nil,
		maxLen:         maxLen,
		corpusFloatLen: float64(len(corpus)),
	}
	b.wm = make(map[string]*Word)
	return &b
}

// 最小词频次
const minCount = 10
const minPoly = 5.0
const minFlex = 1.0

// Extract 进行分词计算
func (b *Builder) Extract() []*Word {
	// b.genRank()

	log.Printf("start calculate right direct\n")
	log.Printf("start gen right rank\n")

	b.right = rankWords(b.letters, b.maxLen)
	b.singleWord()
	log.Printf("end gen right rank\n")

	b.calculateSide(true)
	b.right = nil

	log.Printf("start calculate left direct\n")

	b.reverseLetters = reverseRunes(b.letters)
	b.letters = nil

	log.Printf("start gen left rank\n")

	b.left = rankWords(b.reverseLetters, b.maxLen)

	log.Printf("end gen left rank\n")

	b.calculateSide(false)

	log.Printf("start calculate score\n")
	return b.score()
}

func (b *Builder) score() []*Word {
	war := []*Word{}
	for w, wi := range b.wm {
		delete(b.wm, w)
		wi.Poly = math.Log(wi.Poly + 1)
		wi.Score = wi.Poly * wi.Flex * wi.Freq
		// if wi.Score < 15 {
		//   continue
		// }
		war = append(war, wi)
	}
	sort.Slice(war, func(i, j int) bool {
		return war[i].Score > war[j].Score
	})
	return war
}

func (b *Builder) singleWord() {
	b.singleWm = make(map[string]*Word)
	tw := rune(0)
	for _, index := range b.right {
		wr := b.letters[index]
		w := string(wr)
		if tw == rune(0) || tw != wr {
			tw = wr
			b.singleWm[w] = &Word{
				Word:  []rune{wr},
				Count: 1,
			}
			continue
		}

		b.singleWm[w].Count++
	}
	for _, wi := range b.singleWm {
		wi.Freq = float64(wi.Count) / b.corpusFloatLen
	}
}

func (b *Builder) calculateSide(right bool) {
	// rank 是给字符串排序
	// right 是右方向排序
	rank := b.right
	if !right {
		// left 是左方向排序
		rank = b.left
	}
	// 调整字符的排序方向
	rankedLetters := b.letters
	if !right {
		rankedLetters = b.reverseLetters
	}

	// 单词的第一个字符
	prefixLetter := rune(0)
	// 每个不同字符积累的
	var words map[string]*Word
	isLetter := false

	// 循环排序结果
	for i, index := range rank {
		if i%100000 == 0 {
			log.Printf("caculateside all:%d, current:%d, process: %.5f\n", len(rank), i, float64(i)/float64(len(rank)))
		}
		// 如果 prefixLetter 是个空就给 prefixLetter 一个开头的字符
		if prefixLetter == rune(0) {
			prefixLetter = rankedLetters[index]
			words = make(map[string]*Word)
		}

		if prefixLetter != rankedLetters[index] {
			// 如果 首字符 和 新的index 的 首字符 不同
			if len(words) > 0 { // 如果积累的关键词 不等于0 则进行处理
				//
				for w, wi := range words {
					if wi.Count < minCount {
						delete(words, w)
						continue
					}
					wi.Freq = float64(wi.Count) / b.corpusFloatLen
					wi.generateFlex()
				}
				if len(words) > 0 {
					b.ployWord(words, right)
				}
				if right {
					// 右方向添加词
					for w, wi := range words {
						if wi.Poly < minPoly || wi.Flex <= minFlex {
							continue
						}
						b.wm[w] = wi
					}
				} else {
					// 左方向筛选词
					for w, wi := range words {
						if wi.Poly < minPoly || wi.Flex <= minFlex {
							delete(b.wm, w)
						}
					}
				}

				prefixLetter = rankedLetters[index]
				words = make(map[string]*Word)
			} else {
				// 如果 words 没有东西则 words 置空
				// log.Printf("tmpWm 为空, tempW为：%s， letters[index]为：%s \n", string(tmpW), string(letters[index]))
				prefixLetter = rune(0)
				continue
			}
		}

		// 看当前首个字符是不是英文字母
		isLetter = clear.IsLetter(prefixLetter)
		for l := 2; l <= b.maxLen; l++ {
			// 这块逻辑就是为了判断当前长度是不是一个单词的，如果不是一个单词的结束则不计入关键词
			if isLetter {
				if index > 0 && clear.IsLetter(rankedLetters[index-1]) {
					continue
				}
				if index+l < int(b.corpusFloatLen) && clear.IsLetter(rankedLetters[index+l]) {
					continue
				}
			}
			if index+l > int(b.corpusFloatLen) {
				continue
			}
			currentLenWordrunes := rankedLetters[index : index+l]
			if !right {
				currentLenWordrunes = reverseRunes(currentLenWordrunes)
			}

			currentLenWord := string(currentLenWordrunes)
			if !isLetter && strings.Contains(currentLenWord, " ") {
				continue
			}

			if _, ok := words[currentLenWord]; !ok {
				if !right {
					// 左侧不需要建立新的word
					if wi, ok := b.wm[currentLenWord]; ok {
						// left 如果主词没有的话就不需要判断了 但是letter因为给单独的poly 所以单拎出来
						if len(currentLenWordrunes) > 2 {
							if _, ok := words[string(currentLenWordrunes[1:])]; !ok && isLetter {
								continue
							}
						}
						words[currentLenWord] = wi
					} else {
						continue
					}
				} else {
					// 右方向直接建立word
					words[currentLenWord] = &Word{
						Word: currentLenWordrunes,
						// flex:  -1,
						Count: 0,
					}
				}

			}
			// 加一下相邻字符
			if index+l < len(rank) {
				if right {
					words[currentLenWord].rightNeighbour = append(words[currentLenWord].rightNeighbour, string(rankedLetters[index+l]))
				} else {
					words[currentLenWord].leftNeighbour = append(words[currentLenWord].leftNeighbour, string(rankedLetters[index+l]))
				}
			}
			if right {
				words[currentLenWord].Count++
			}
		}
	}
}

func reverseRunes(rs []rune) []rune {
	nr := []rune{}
	for i := len(rs) - 1; i >= 0; i-- {
		nr = append(nr, rs[i])
	}
	return nr
}

func (b *Builder) ployWord(words map[string]*Word, right bool) {
	for _, w := range words {
		// poly 主词
		main := ""
		// poly 副词
		neighbour := ""
		if right {
			main = string(w.Word[:len(w.Word)-1])
			neighbour = string(w.Word[len(w.Word)-1:])
		} else {
			main = string(w.Word[1:])
			neighbour = string(w.Word[:1])
		}

		mainInfo := words[main]
		if len(w.Word) == 2 {
			mainInfo = b.singleWm[main]
		}

		neighbourInfo := b.singleWm[neighbour]
		p := 0.0
		if mainInfo == nil || neighbourInfo == nil {
			p = 300
		} else {
			p = w.Freq / (mainInfo.Freq * neighbourInfo.Freq)
		}
		if w.Poly == 0 || w.Poly > p {
			w.Poly = p
		}

	}
}

// 生成关键词排序
func (b *Builder) genRank() {
	b.right = rankWords(b.letters, b.maxLen)
	b.left = rankWords(b.reverseLetters, b.maxLen)
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
