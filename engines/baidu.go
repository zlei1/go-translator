package engines

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"

	uuid "github.com/satori/go.uuid"
)

const (
	baiduApiURL = "https://fanyi-api.baidu.com/api/trans/vip/translate"
)

type BaiduTranslator struct {
	Q      string `json:"q"`
	From   string `json:"from"`
	To     string `json:"to"`
	AppKey string `json:"appKey"`
	Salt   string `json:"salt"`
	Sign   string `json:"sign"`
}

type BaiduResponseResult struct {
	Src string `json:"src"`
	Dst string `json:"dst"`
}

type BaiduResponse struct {
	From      string `json:"from,omitempty"`
	To        string `json:"to",omitempty`
	TransResult []BaiduResponseResult `json:"trans_result",omitempty`
	ErrorCode string  `json:"error_code,omitempty"`
	ErrorMsg string  `json:"error_msg,omitempty"`
}

func generateBaiduSign(app_key, app_secret, word, u1 string) string {
	str := fmt.Sprintf("%s%s%s%s", app_key, word, u1, app_secret)

	has := md5.Sum([]byte(str))
	sign := fmt.Sprintf("%x", has)

	return sign
}

func NewBaiduTranslator(sl, tl, word string) *BaiduTranslator {
	config := readConfigFile()

	u1 := uuid.NewV4().String()
	sign := generateBaiduSign(config.BaiduAppKey, config.BaiduAppSecret, word, u1)

	return &BaiduTranslator{
		Q:      word,
		From:   sl,
		To:     tl,
		AppKey: config.BaiduAppKey,
		Salt:   u1,
		Sign:   sign,
	}
}

func (bd *BaiduTranslator) Perform() error {
	data := make(url.Values, 0)
	data["q"] = []string{bd.Q}
	data["from"] = []string{bd.From}
	data["to"] = []string{bd.To}
	data["appid"] = []string{bd.AppKey}
	data["salt"] = []string{bd.Salt}
	data["sign"] = []string{bd.Sign}

	var resp *http.Response
	resp, err := http.PostForm(baiduApiURL, data)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var result BaiduResponse
	json.Unmarshal(body, &result)

	baiduConsole(bd.Q, &result, os.Stdout)

	return nil
}

func baiduConsole(q string, resp *BaiduResponse, w io.Writer) {
	if len(resp.ErrorCode) != 0 {
		fmt.Println(resp.ErrorCode)
		os.Exit(0)
	}

	fmt.Fprintln(w, "@", q)

	fmt.Fprintln(w, "[翻译]")
	for key, item := range resp.TransResult {
		fmt.Fprintln(w, "\t", key+1, ".", item.Dst)
	}
}
