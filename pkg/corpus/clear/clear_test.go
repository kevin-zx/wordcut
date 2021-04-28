// 这个包能够对文字进行处理，保留汉字和删除重复部分
package clear

import "testing"

func TestPureCorpusWithLetter(t *testing.T) {
	// type args struct {
	// 	content string
	// 	sep     string
	// }
	// tests := []struct {
	// 	name string
	// 	args args
	// 	want string
	// }{
	// 	{
	// 		name: "1",
	// 		args: args{
	// 			content: "外面的露天小花园晚上看👀 也泰好看了bulingbuling的 一眼看过去超级显眼 进去感觉到了另外一个世界～",
	// 			sep:     "",
	// 		},
	// 		want: "外面的露天小花园晚上看也泰好看了\nbulingbuling\n的一眼看过去超级显眼进去感觉到了另外一个世界",
	// 	},
	// }
	// for _, tt := range tests {
	// 	t.Run(tt.name, func(t *testing.T) {
	// 		if got := PureCorpusWithLetter(tt.args.content, tt.args.sep); got != tt.want {
	// 			t.Errorf("PureCorpusWithLetter() = %v, want %v", got, tt.want)
	// 		}
	// 	})
	// }

}
