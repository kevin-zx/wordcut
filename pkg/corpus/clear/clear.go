// 这个包能够对文字进行处理，保留汉字和删除重复部分
package clear

import (
	"runtime"
	"strings"
	"sync"
	"unicode"
)

func PureCorpusWithLettersAndDigitals(content string) []rune {
	core := runtime.NumCPU()
	for strings.Contains(content, "  ") {
		content = strings.ReplaceAll(content, "  ", " ")
	}
	var rs []rune
	tasks := make(chan string, core*2)
	lock := sync.Mutex{}
	wg := sync.WaitGroup{}
	wg.Add(core)
	for i := 0; i < core; i++ {
		go func() {
			for c := range tasks {
				tmpRs := []rune{}
				s := false
				l := 0
				for _, r := range c {
					if unicode.IsPunct(r) {
						continue
					}
					if unicode.Is(unicode.Han, r) {
						if s {
							s = false
							if l > 1 {
								if len(tmpRs) > 0 && tmpRs[len(tmpRs)-1] != ' ' {
									tmpRs = append(tmpRs, ' ')
								}
							}
						}
						tmpRs = append(tmpRs, r)
					}
					if IsDigital(r) || IsLetter(r) {
						l++
						if !s {
							s = true
							if len(tmpRs) > 0 && tmpRs[len(tmpRs)-1] != ' ' {
								tmpRs = append(tmpRs, ' ')
							}
						}
						tmpRs = append(tmpRs, r)
					}
				}
				lock.Lock()
				rs = append(rs, tmpRs...)
				lock.Unlock()
			}
			wg.Done()
		}()
	}
	for _, line := range strings.Split(content, "\n") {
		tasks <- line
	}
	close(tasks)
	wg.Wait()
	return rs
}

// IsLetter 是否是字母
func IsLetter(r rune) bool {
	return (r >= 97 && r <= 122) || (r >= 65 && r <= 90)
}

func IsDigital(r rune) bool {
	return r >= '0' && r <= '9'
}
