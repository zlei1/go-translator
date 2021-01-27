package engines

type Translator interface {
}

type Result struct {
	engine     string
	sl         string   //来源语言
	tl         string   //目标语言
	text       string   //需要翻译的文本
	phonetic   string   //音标
	paraphrase string   //简单释义
	explains   []string //分行解释
}
