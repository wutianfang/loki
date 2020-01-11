package word

import (
	"encoding/json"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/wutianfang/loki/model"
	"io/ioutil"
	"net/http"
	"strings"
)

type QueryResponse struct{
	Error string `json:"error"`
	Errno int `json:"errno"`
	Data struct {
		Word model.Word `json:"word"`
	} `json:"data"`
}

func Query(c echo.Context) error {
	response := QueryResponse{}
	params := struct{
		Word  string `form:"word" query:"word"`
	}{}
	_ = c.Bind(&params)

	wordModel := model.NewWordModel()

	word,_ := wordModel.Query(params.Word)

	if word != nil {
		response.Data.Word = *word
		return c.JSON(200, response)
	}

	wordInfo ,err := requestIciba(params.Word)
	if err!= nil {
		response.Errno = 1
		response.Error = err.Error()
		_ = c.JSON(200, response)
		return nil
	}

	newWord := model.Word{
		Word:strings.ToLower(params.Word),
		Info:*wordInfo,
	}
	err = wordModel.Insert(newWord)
	if err!= nil {
		response.Errno = 1
		response.Error = err.Error()
		_ = c.JSON(200, response)
		return nil
	}
	response.Data.Word = newWord

	_ = c.JSON(200, response)

	return nil
}

func requestIciba(word string) (*model.WordInfo,error) {
	ret := &model.WordInfo{}

	client := &http.Client{}

	icibaUrl := "http://www.iciba.com/index.php?a=getWordMean&c=search&lis&word="+word

	request, _ := http.NewRequest("GET", icibaUrl, nil)

	response,_ := client.Do(request)

	if response.StatusCode != http.StatusOK {
		return ret,nil
	}

	str, _ := ioutil.ReadAll(response.Body)

	rawResponse := struct{
		Errno int `json:"errno"`
		Errmsg string `json:"errmsg"`
		BaseInfo struct{
			Exchange map[string][]string `json:"exchange"`
			Symbols []model.WordInfo `json:"symbols"`
		} `json:"baesInfo"`
		Sentence []model.Sentence `json:"sentence"`
	}{}

	_ = json.Unmarshal(str, &rawResponse)

	if rawResponse.Errno!=0 {
		return nil, fmt.Errorf("query iciba 错误：%s", rawResponse.Errmsg)
	}
	if rawResponse.BaseInfo.Symbols == nil {
		return nil, fmt.Errorf("query iciba 错误：单词单词不存在")
	}
	ret = &rawResponse.BaseInfo.Symbols[0]
	ret.Exchange = rawResponse.BaseInfo.Exchange
	ret.Sentences = rawResponse.Sentence

	return ret,nil
}


