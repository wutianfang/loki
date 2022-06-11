package word

import (
	"encoding/json"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/wutianfang/loki/model"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
)

type QueryResponse struct {
	Error string `json:"error"`
	Errno int    `json:"errno"`
	Data  struct {
		Word model.Word `json:"word"`
	} `json:"data"`
}

func Query(c echo.Context) error {
	response := QueryResponse{}
	params := struct {
		Word string `form:"word" query:"word"`
	}{}
	_ = c.Bind(&params)

	wordModel := model.NewWordModel()

	word, _ := wordModel.Query(params.Word)

	if word != nil {
		response.Data.Word = *word
		return c.JSON(200, response)
	}

	wordInfo, err := requestIcibaV2(params.Word)
	if err != nil {
		response.Errno = 1
		response.Error = err.Error()
		_ = c.JSON(200, response)
		return nil
	}

	newWord := model.Word{
		Word: strings.ToLower(params.Word),
		Info: *wordInfo,
	}
	err = wordModel.Insert(newWord)
	if err != nil {
		response.Errno = 1
		response.Error = err.Error()
		_ = c.JSON(200, response)
		return nil
	}
	response.Data.Word = newWord

	_ = c.JSON(200, response)

	return nil
}

func requestIciba(word string) (*model.WordInfo, error) {
	ret := &model.WordInfo{}

	client := &http.Client{}

	icibaUrl := "http://www.iciba.com/index.php?a=getWordMean&c=search&lis&word=" + word

	request, _ := http.NewRequest("GET", icibaUrl, nil)

	response, _ := client.Do(request)

	if response.StatusCode != http.StatusOK {
		return ret, nil
	}

	str, _ := ioutil.ReadAll(response.Body)

	rawResponse := struct {
		Errno    int    `json:"errno"`
		Errmsg   string `json:"errmsg"`
		BaseInfo struct {
			Exchange map[string][]string `json:"exchange"`
			Symbols  []model.WordInfo    `json:"symbols"`
		} `json:"baesInfo"`
		Sentence []model.Sentence `json:"sentence"`
	}{}

	_ = json.Unmarshal(str, &rawResponse)

	if rawResponse.Errno != 0 {
		return nil, fmt.Errorf("query iciba 错误：%s", rawResponse.Errmsg)
	}
	if rawResponse.BaseInfo.Symbols == nil {
		return nil, fmt.Errorf("query iciba 错误：单词单词不存在")
	}
	if *rawResponse.BaseInfo.Symbols[0].PhAmMp3 == "" && *rawResponse.BaseInfo.Symbols[0].PhTtsMp3 != "" {
		*rawResponse.BaseInfo.Symbols[0].PhAmMp3 = *rawResponse.BaseInfo.Symbols[0].PhTtsMp3
	}
	if *rawResponse.BaseInfo.Symbols[0].PhEnMp3 == "" && *rawResponse.BaseInfo.Symbols[0].PhTtsMp3 != "" {
		*rawResponse.BaseInfo.Symbols[0].PhEnMp3 = *rawResponse.BaseInfo.Symbols[0].PhTtsMp3
	}
	if rawResponse.BaseInfo.Symbols[0].PhEn == "" && *rawResponse.BaseInfo.Symbols[0].PhOther != "" {
		rawResponse.BaseInfo.Symbols[0].PhEn = *rawResponse.BaseInfo.Symbols[0].PhOther
	}
	if rawResponse.BaseInfo.Symbols[0].PhAm == "" && *rawResponse.BaseInfo.Symbols[0].PhOther != "" {
		rawResponse.BaseInfo.Symbols[0].PhAm = *rawResponse.BaseInfo.Symbols[0].PhOther
	}
	rawResponse.BaseInfo.Symbols[0].PhTtsMp3 = nil
	rawResponse.BaseInfo.Symbols[0].PhOther = nil

	ret = &rawResponse.BaseInfo.Symbols[0]
	ret.Exchange = rawResponse.BaseInfo.Exchange
	ret.Sentences = rawResponse.Sentence

	return ret, nil
}

type iCibaWordInfo struct {
	Props struct {
		PageProps struct {
			InitialReduxState struct {
				Word struct {
					WordInfo struct {
						BaesInfo struct {
							WordName string              `json:"word_name"`
							Exchange map[string][]string `json:"exchange"`
							Symbols  []struct {
								PhEn     string `json:"ph_en"`
								PhAm     string `json:"ph_am"`
								PhOther  string `json:"ph_other"`
								PhEnMp3  string `json:"ph_en_mp3"`
								PhAmMp3  string `json:"ph_am_mp3"`
								PhTtsMp3 string `json:"ph_tts_mp3"`
								Parts    []struct {
									Part  string   `json:"part"`
									Means []string `json:"means"`
								} `json:"parts"`
							} `json:"symbols"`
						} `json:"baesInfo"`
						NewSentence []struct {
							Sentences []struct {
								Id     int    `json:"id"`
								Type   int    `json:"type"`
								Cn     string `json:"cn"`
								EN     string `json:"en"`
								TtsUrl string `json:"ttsUrl"`
							} `json:"sentences"`
						} `json:"new_sentence"`
					} `json:"wordInfo"`
				} `json:"word"`
			} `json:"initialReduxState"`
		} `json:"pageProps"`
	} `json:"props"`
}

func requestIcibaV2(word string) (*model.WordInfo, error) {
	ret := &model.WordInfo{}

	client := &http.Client{}
	icibaUrl := "http://www.iciba.com/word?w=" + word
	request, _ := http.NewRequest("GET", icibaUrl, nil)
	response, _ := client.Do(request)

	if response.StatusCode != http.StatusOK {
		return ret, nil
	}
	str, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	_ = response.Body.Close()

	r, _ := regexp.Compile("{\"props\":.*}")
	jsonStr := r.FindString(string(str))

	rawResponse := iCibaWordInfo{}
	err = json.Unmarshal([]byte(jsonStr), &rawResponse)

	wordInfo := rawResponse.Props.PageProps.InitialReduxState.Word.WordInfo
	if len(wordInfo.BaesInfo.Symbols) > 0 {
		ret.PhAm = wordInfo.BaesInfo.Symbols[0].PhAm
		ret.PhEn = wordInfo.BaesInfo.Symbols[0].PhEn
		ret.PhOther = &wordInfo.BaesInfo.Symbols[0].PhOther
		if wordInfo.BaesInfo.Symbols[0].PhAmMp3 == "" && wordInfo.BaesInfo.Symbols[0].PhTtsMp3 != "" {
			wordInfo.BaesInfo.Symbols[0].PhAmMp3 = wordInfo.BaesInfo.Symbols[0].PhTtsMp3
		}
		if wordInfo.BaesInfo.Symbols[0].PhEnMp3 == "" && wordInfo.BaesInfo.Symbols[0].PhTtsMp3 != "" {
			wordInfo.BaesInfo.Symbols[0].PhEnMp3 = wordInfo.BaesInfo.Symbols[0].PhTtsMp3
		}
		ret.PhAmMp3 = &wordInfo.BaesInfo.Symbols[0].PhAmMp3
		ret.PhEnMp3 = &wordInfo.BaesInfo.Symbols[0].PhEnMp3
		ret.PhTtsMp3 = &wordInfo.BaesInfo.Symbols[0].PhTtsMp3
	}
	for _, part := range wordInfo.BaesInfo.Symbols[0].Parts {
		wordInfoPart := model.WordInfoPart{
			Part:  part.Part,
			Means: part.Means,
		}
		ret.Parts = append(ret.Parts, wordInfoPart)
	}

	for _, rawSentence := range wordInfo.NewSentence[0].Sentences {
		sentence := model.Sentence{
			NetworkId: rawSentence.Id,
			NetworkCn: rawSentence.Cn,
			NetworkEn: rawSentence.EN,
			TtsMp3:    rawSentence.TtsUrl,
		}
		ret.Sentences = append(ret.Sentences, sentence)
	}
	ret.Exchange = wordInfo.BaesInfo.Exchange

	return ret, nil
}
