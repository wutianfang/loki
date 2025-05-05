package word

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/wutianfang/loki/model"
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
	params.Word = strings.ToLower(params.Word)

	wordModel := model.NewWordModel()
	word, _ := wordModel.Query(params.Word)
	if word != nil {
		response.Data.Word = *word
		go checkMp3File(word)
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
	err = wordModel.Insert(&newWord)
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

func checkMp3File(word *model.Word) {
	wordModel := model.NewWordModel()
	if !wordModel.CheckFile(word) {
		wordInfo, err := requestIcibaV2(word.Word)
		if err != nil {
			return
		}
		wordModel.ReDownload(&model.Word{
			Word: strings.ToLower(word.Word),
			Info: *wordInfo,
		})
	}
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
	//Props struct {
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
	//} `json:"props"`
}

type iCibaWordInfoV2 struct {
	PageProps struct {
		Query struct {
			W string `json:"w"`
		} `json:"query"`
		InitialReduxState struct {
			User struct {
				IsLogin  bool `json:"isLogin"`
				UserInfo struct {
				} `json:"userInfo"`
			} `json:"user"`
			Store struct {
				PrevPath struct {
				} `json:"prevPath"`
			} `json:"store"`
			Word struct {
				WordInfo struct {
					CetSix []struct {
						Word     string `json:"word"`
						Count    int    `json:"count"`
						Kd       string `json:"kd"`
						Sentence []struct {
							Sentence string `json:"sentence"`
							Come     string `json:"come"`
						} `json:"Sentence"`
					} `json:"cetSix"`
					ErrWords []struct {
						WordName string   `json:"word_name"`
						Means    []string `json:"means"`
					} `json:"err_words"`
					Slang []struct {
						Tokens string `json:"tokens"`
						Type   string `json:"type"`
						List   []struct {
							Explanation string `json:"explanation"`
							Example     []struct {
								En string `json:"en"`
								Zh string `json:"zh"`
							} `json:"example"`
						} `json:"list"`
						Class []string `json:"class"`
					} `json:"slang"`
					Exchanges    []string `json:"exchanges"`
					StemsAffixes []struct {
						Type      string `json:"type"`
						TypeValue string `json:"type_value"`
						TypeExp   string `json:"type_exp"`
						WordParts []struct {
							WordPart     string `json:"word_part"`
							StemsAffixes []struct {
								ValueEn   string `json:"value_en"`
								ValueCn   string `json:"value_cn"`
								WordBuile string `json:"word_buile"`
							} `json:"stems_affixes"`
						} `json:"word_parts"`
					} `json:"stems_affixes"`
					TradeMeans []struct {
						WordTrade string   `json:"word_trade"`
						WordMean  []string `json:"word_mean"`
					} `json:"trade_means"`
					Gaokao []struct {
						Word     string `json:"word"`
						Count    int    `json:"count"`
						Kd       string `json:"kd"`
						Sentence []struct {
							Sentence string `json:"sentence"`
							Come     string `json:"come"`
						} `json:"Sentence"`
					} `json:"gaokao"`
					Kaoyan []struct {
						Word     string `json:"word"`
						Count    int    `json:"count"`
						Kd       string `json:"kd"`
						Sentence []struct {
							Sentence string `json:"sentence"`
							Come     string `json:"come"`
						} `json:"Sentence"`
					} `json:"kaoyan"`
					Bidec struct {
						WordName string `json:"word_name"`
						Parts    []struct {
							PartId   string `json:"part_id"`
							PartName string `json:"part_name"`
							WordId   string `json:"word_id"`
							Means    []struct {
								MeanId    string `json:"mean_id"`
								PartId    string `json:"part_id"`
								WordMean  string `json:"word_mean"`
								Sentences []struct {
									En string `json:"en"`
									Cn string `json:"cn"`
								} `json:"sentences"`
							} `json:"means"`
						} `json:"parts"`
					} `json:"bidec"`
					Synonym []struct {
						PartName string `json:"part_name"`
						Means    []struct {
							WordMean string   `json:"word_mean"`
							Cis      []string `json:"cis"`
						} `json:"means"`
					} `json:"synonym"`
					Phrase []struct {
						CizuName string `json:"cizu_name"`
						Jx       []struct {
							JxEnMean string `json:"jx_en_mean"`
							JxCnMean string `json:"jx_cn_mean"`
							Lj       []struct {
								LjLy string `json:"lj_ly"`
								LjLs string `json:"lj_ls"`
							} `json:"lj"`
						} `json:"jx"`
					} `json:"phrase"`
					Collins []struct {
						Entry []struct {
							Def     string `json:"def"`
							Tran    string `json:"tran"`
							Posp    string `json:"posp"`
							Example []struct {
								Ex      string `json:"ex"`
								Tran    string `json:"tran"`
								TtsMp3  string `json:"tts_mp3"`
								TtsSize string `json:"tts_size"`
							} `json:"example"`
						} `json:"entry"`
					} `json:"collins"`
					EeMean []struct {
						PartName string `json:"part_name"`
						Means    []struct {
							WordMean  string `json:"word_mean"`
							Sentences []struct {
								Sentence string `json:"sentence"`
							} `json:"sentences"`
						} `json:"means"`
					} `json:"ee_mean"`
					Derivation []struct {
						YuyuanName string `json:"yuyuan_name"`
					} `json:"derivation"`
					BaesInfo struct {
						WordName string              `json:"word_name"`
						IsCRI    string              `json:"is_CRI"`
						Exchange map[string][]string `json:"exchange"`
						Symbols  []struct {
							PhEn       string `json:"ph_en"`
							PhAm       string `json:"ph_am"`
							PhOther    string `json:"ph_other"`
							PhEnMp3    string `json:"ph_en_mp3"`
							PhAmMp3    string `json:"ph_am_mp3"`
							PhTtsMp3   string `json:"ph_tts_mp3"`
							PhEnMp3Bk  string `json:"ph_en_mp3_bk"`
							PhAmMp3Bk  string `json:"ph_am_mp3_bk"`
							PhTtsMp3Bk string `json:"ph_tts_mp3_bk"`
							Parts      []struct {
								Part  string   `json:"part"`
								Means []string `json:"means"`
							} `json:"parts"`
						} `json:"symbols"`
						BaesElse []struct {
							WordName string `json:"word_name"`
							IsCRI    string `json:"is_CRI"`
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
						} `json:"baesElse"`
						WordTag       []int `json:"word_tag"`
						TranslateType int   `json:"translate_type"`
						Frequence     int   `json:"frequence"`
					} `json:"baesInfo"`
					NewSentence []struct {
						Tag       string `json:"tag"`
						Word      string `json:"word"`
						Meaning   string `json:"meaning"`
						Sentences []struct {
							Id      int    `json:"id"`
							Type    int    `json:"type"`
							Cn      string `json:"cn"`
							En      string `json:"en"`
							From    string `json:"from"`
							TtsUrl  string `json:"ttsUrl"`
							TtsSize int    `json:"ttsSize"`
							LikeNum int    `json:"likeNum"`
						} `json:"sentences"`
					} `json:"new_sentence"`
					CetFour []struct {
						Word     string `json:"word"`
						Count    int    `json:"count"`
						Kd       string `json:"kd"`
						Sentence []struct {
							Sentence string `json:"sentence"`
							Come     string `json:"come"`
						} `json:"Sentence"`
					} `json:"cetFour"`
				} `json:"wordInfo"`
				History []interface{} `json:"history"`
			} `json:"word"`
			Fy struct {
				FyHeight    string `json:"fyHeight"`
				TransedWord struct {
				} `json:"transedWord"`
				AllLanguateMap struct {
				} `json:"allLanguateMap"`
				SearchWord    string `json:"searchWord"`
				Loading       bool   `json:"loading"`
				Upbroadparams struct {
					Reqid   string `json:"reqid"`
					Version string `json:"version"`
					Ttype   string `json:"ttype"`
				} `json:"upbroadparams"`
			} `json:"fy"`
			FyPassage struct {
				BlockToggle string `json:"blockToggle"`
				PassageFile struct {
				} `json:"passageFile"`
				LanguageParams struct {
					From     string `json:"from"`
					FromType string `json:"fromType"`
					To       string `json:"to"`
					ToType   string `json:"toType"`
				} `json:"languageParams"`
				ProgressInfo struct {
				} `json:"progressInfo"`
				Tid               interface{} `json:"tid"`
				FileUrl           interface{} `json:"fileUrl"`
				FormalDownloadurl string      `json:"formalDownloadurl"`
				ImgUrl            interface{} `json:"imgUrl"`
				AllOptionsMap     interface{} `json:"allOptionsMap"`
			} `json:"fyPassage"`
			Rgfy struct {
				BillInformation struct {
				} `json:"billInformation"`
			} `json:"rgfy"`
			Grammar struct {
				Res  []interface{} `json:"res"`
				Data struct {
				} `json:"data"`
				ContractData []interface{} `json:"contractData"`
				DefaultText  string        `json:"defaultText"`
				Cache        struct {
					RefName    string        `json:"refName"`
					RepairList []interface{} `json:"repairList"`
					IgnoreList []interface{} `json:"ignoreList"`
				} `json:"cache"`
				ErrorData struct {
					标点符号错误 []interface{} `json:"标点符号错误"`
					语法错误   []interface{} `json:"语法错误"`
					拼写错误   []interface{} `json:"拼写错误"`
					句子推荐   []interface{} `json:"句子推荐"`
					句子改写   []interface{} `json:"句子改写"`
				} `json:"errorData"`
				ErrorIds struct {
					标点符号错误 []interface{} `json:"标点符号错误"`
					语法错误   []interface{} `json:"语法错误"`
					拼写错误   []interface{} `json:"拼写错误"`
					句子推荐   []interface{} `json:"句子推荐"`
					句子改写   []interface{} `json:"句子改写"`
				} `json:"errorIds"`
				RepairList []interface{} `json:"repairList"`
				CopyMap    struct {
				} `json:"copyMap"`
				IgnoreList      []interface{} `json:"ignoreList"`
				RefName         string        `json:"refName"`
				PrevInput       string        `json:"prevInput"`
				IsShowErrorType interface{}   `json:"isShowErrorType"`
				CanContract     bool          `json:"canContract"`
				Sentences       []interface{} `json:"sentences"`
				Polish          []interface{} `json:"polish"`
				Error           string        `json:"error"`
			} `json:"grammar"`
			Translate struct {
				PicStep     int         `json:"picStep"`
				PicFile     interface{} `json:"picFile"`
				PicUrl      string      `json:"picUrl"`
				PicResult   interface{} `json:"picResult"`
				PicLanguage struct {
					From     string `json:"from"`
					FromType string `json:"fromType"`
					To       string `json:"to"`
					ToType   string `json:"toType"`
				} `json:"picLanguage"`
				PicTranslating bool          `json:"picTranslating"`
				History        []interface{} `json:"history"`
				AllLanguageMap struct {
				} `json:"allLanguageMap"`
				Sentence string `json:"sentence"`
			} `json:"translate"`
		} `json:"initialReduxState"`
		Redirect  bool `json:"redirect"`
		IsCrawler bool `json:"isCrawler"`
	} `json:"pageProps"`
	NSSP bool `json:"__N_SSP"`
}

func requestIcibaV2(word string) (*model.WordInfo, error) {
	ret := &model.WordInfo{}

	client := &http.Client{}
	word = url.QueryEscape(word)
	//icibaUrl := "http://www.iciba.com/word?w=" + word
	icibaUrl := "https://www.iciba.com/_next/data/OPeO-bTu_2jVUKMSaH9b0/word.json?w=" + word
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

	//r, _ := regexp.Compile("{\"props\":.*}")
	//jsonStr := r.FindString(string(str))

	rawResponse := iCibaWordInfoV2{}
	err = json.Unmarshal(str, &rawResponse)

	wordInfo := rawResponse.PageProps.InitialReduxState.Word.WordInfo
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
			NetworkEn: rawSentence.En,
			TtsMp3:    rawSentence.TtsUrl,
		}
		ret.Sentences = append(ret.Sentences, sentence)
	}
	ret.Exchange = wordInfo.BaesInfo.Exchange

	return ret, nil
}
