package engines

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"io"
	"os"
	"strconv"
	"time"

	"github.com/spf13/viper"
	uuid "github.com/satori/go.uuid"
)

const (
	apiURL = "https://openapi.youdao.com/api"
)

type YoudaoConfig struct {
	AppKey    string `mapstructure:"youdao_app_key"`
	AppSecret string `mapstructure:"youdao_app_secret"`
}

type ResponseWeb struct {
	Key   string   `json:"key"`
	Value []string `json:"value"`
}

type ResponseBasic struct {
	UsPhonetic string   `json:"us-phonetic"`
	Phonetic   string   `json:"phonetic"`
	UkPhonetic string   `json:"uk-phonetic"`
	UkSpeech   string   `json:"uk-speech"`
	UsSpeech   string   `json:"us-speech"`
	Explains   []string `json:"explains"`
}

type Response struct {
	ErrorCode    string                 `json:"errorCode"`
	Query        string                 `json:"query"`
	Translation  []string               `json:"translation"`
	Basic        ResponseBasic          `json:"basic"`
	Web          []ResponseWeb          `json:"web,omitempty"`
	Lang         string                 `json:"l"`
	Dict         map[string]interface{} `json:"dict,omitempty"`
	Webdict      map[string]interface{} `json:"webdict,omitempty"`
	TSpeakUrl    string                 `json:"tSpeakUrl,omitempty"`
	SpeakUrl     string                 `json:"speakUrl,omitempty"`
	ReturnPhrase []string               `json:"returnPhrase,omitempty"`
}

type YoudaoTranslator struct {
	Q        string `json:"q"`
	From     string `json:"from"`
	To       string `json:"to"`
	AppKey   string `json:"appKey"`
	Salt     string `json:"salt"`
	Sign     string `json:"sign"`
	SignType string `json:"signType"`
	Curtime  string `json:"curtime"`
}

func readConfigFile() *YoudaoConfig {
	viper.AddConfigPath("config/")
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	err := viper.ReadInConfig()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(0)
	}

	config := new(YoudaoConfig)
	err = viper.Unmarshal(config)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(0)
	}

	return config
}

func generateSign(app_key, app_secret, word, u1, stamp string) string {
	var sign string

	input := truncate(word)
	str := fmt.Sprintf("%s%s%s%s%s", app_key, input, u1, stamp, app_secret)

	buff := sha256.Sum256([]byte(str))
	for _, value := range buff {
		str := strconv.FormatUint(uint64(value), 16)
		if len([]rune(str)) == 1 {
			sign = sign + "0" + str
		} else {
			sign = sign + str
		}
	}
	return sign
}

func truncate(q string) string {
	res := make([]byte, 10)
	qlen := len([]rune(q))

	if qlen <= 20 {
		return q
	} else {
		temp := []byte(q)
		copy(res, temp[:10])
		lenstr := strconv.Itoa(qlen)
		res = append(res, lenstr...)
		res = append(res, temp[qlen-10:qlen]...)
		return string(res)
	}
}

func NewYoudaoTranslator(sl, tl, word string) *YoudaoTranslator {
	config := readConfigFile()

	u1 := uuid.NewV4().String()
	stamp := time.Now().Unix()
	stamp_str := strconv.FormatInt(stamp, 10)
	sign := generateSign(config.AppKey, config.AppSecret, word, u1, stamp_str)

	return &YoudaoTranslator{
		Q: word,
		From: sl,
		To: tl,
		AppKey: config.AppKey,
		Salt: u1,
		Sign: sign,
		SignType: "v3",
		Curtime: stamp_str,
	}
}

func (yd *YoudaoTranslator) Perform() error {
	data := make(url.Values, 0)
	data["q"] = []string{yd.Q}
	data["from"] = []string{yd.From}
	data["to"] = []string{yd.To}
	data["appKey"] = []string{yd.AppKey}
	data["salt"] = []string{yd.Salt}
	data["sign"] = []string{yd.Sign}
	data["signType"] = []string{yd.SignType}
	data["curtime"] = []string{yd.Curtime}

	var resp *http.Response
	resp, err := http.PostForm(apiURL, data)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var result Response
	json.Unmarshal(body, &result)

	console(&result, os.Stdout)

	return nil
}

func console(resp *Response, w io.Writer) {
	if resp.ErrorCode != "0" {
		fmt.Fprintln(w, "服务调用失败")
		os.Exit(0)
	}
	fmt.Fprintln(w, "@", resp.Query)

	if resp.Basic.UkPhonetic != "" {
		fmt.Fprintln(w, "英:", "[", resp.Basic.UkPhonetic, "]")
	}
	if resp.Basic.UsPhonetic != "" {
		fmt.Fprintln(w, "美:", "[", resp.Basic.UsPhonetic, "]")
	}

	fmt.Fprintln(w, "[翻译]")
	for key, item := range resp.Translation {
		fmt.Fprintln(w, "\t", key+1, ".", item)
	}
	fmt.Fprintln(w, "[延伸]")
	for key, item := range resp.Basic.Explains {
		fmt.Fprintln(w, "\t", key+1, ".", item)
	}

	fmt.Fprintln(w, "[网络]")
	for key, item := range resp.Web {
		fmt.Fprintln(w, "\t", key+1, ".", item.Key)
		fmt.Fprint(w, "\t翻译:")
		for _, val := range item.Value {
			fmt.Fprint(w, val, ",")
		}
		fmt.Fprint(w, "\n")
	}
}
