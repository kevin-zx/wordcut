package main

import (
	"encoding/csv"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/kevin-zx/wordcut/pkg/corpus/clear"
	"github.com/kevin-zx/wordcut/pkg/extractor"
)

func main() {

	// for _, w := range words {
	//   fmt.Printf("%v\n", w)
	// }

	// fmt.Println(' ')
	run()
}

func run() {
	// data, err := ioutil.ReadFile("../../data/西游记.txt")
	// data, err := ioutil.ReadFile("./data/西游记.txt")
	// data, err := ioutil.ReadFile("./data/content.data")
	data, err := ioutil.ReadFile("../../data/content.data")
	if err != nil {
		panic(err)
	}
	rs := clear.PureCorpusWithLettersAndDigitals(string(data))
	b := extractor.NewBuilder(rs, 10)
	// ioutil.WriteFile("f.data", []byte(string(rs[:500000])), os.ModePerm)
	log.Printf("start %d \n", time.Now().UnixNano())
	words := b.Extract()
	log.Printf("end   %d\n", time.Now().UnixNano())
	log.Printf("end   %d\n", time.Now().Unix())
	f, err := os.Create("result.csv")
	if err != nil {
		panic(err)
	}
	csvw := csv.NewWriter(f)
	for _, w := range words {
		csvw.Write([]string{w.Word, strconv.Itoa(w.Count),
			strconv.FormatFloat(w.Flex, 'f', 2, 64),
			strconv.FormatFloat(w.Poly, 'f', 2, 64),
			strconv.FormatFloat(w.Score, 'f', 8, 64),
			strconv.FormatFloat(w.Freq, 'f', 8, 64),
		})
	}
}
