package extractor

import (
	"fmt"
	"testing"
)

func Test_getOneWordCount(t *testing.T) {
	corpusRaw := `我们可以对文本进行标记，以方便在文档的不同位置间跳转。创建标记将光标移到某一行，使用ma命令进行标记。其中，m是标记命令，a是所做标记的名称。可以使用小写字母a-z或大写字母A-Z中的任意一个做为标记名称。小写字母的标记，仅用于当前缓冲区；而大写字母的标记，则可以跨越不同的缓冲区。例如，你正在编辑File1，但仍然可以使用'A命令，移动到File2中创建的标记A。跳转标记创建标记后，可以使用'a命令，移动到指定标记行的首个非空字符。这里'是单引号。也可以使用a命令，移到所做标记时的光标位置。这里是反引号（也就是数字键1左边的那一个）。列示标记利用:marks命令，可以列出所有标记。这其中也包括一些系统内置的特殊标记（Special marks）`
	corpus := []rune(corpusRaw)
	b := NewBuilder(corpus, 10)
	b.rightRank = rankWords(corpus, 10)
	for _, letterIndex := range b.rightRank {
		if letterIndex+10 > len(corpus) {
			fmt.Println(string(corpus[letterIndex:]))
		}
		fmt.Println(string(corpus[letterIndex : letterIndex+10]))
	}
	c := b.getOneWordCount([]rune("小写"), 0, len(corpus)-1)
	fmt.Println(c)
}
