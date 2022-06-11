package model

import (
	"github.com/go-xorm/xorm"
	"github.com/wutianfang/loki/common/conf"
	"io"
	"net/http"
	"os"
	"time"
)

type WordModel struct {
	engine *xorm.Engine
}

type Word struct {
	Word       string
	CreateTime time.Time
	Info       WordInfo `xorm:"json 'info'"`
}

type Sentence struct {
	NetworkId  int    `json:"network_id"`
	NetworkEn  string `json:"network_en"`
	NetworkCn  string `json:"network_cn"`
	TtsMp3     string `json:"tts_mp3"`
	SourceType int    `json:"source_type"`
	SourceId   int    `json:"source_id"`
}

type WordInfo struct {
	PhEn      string              `json:"ph_en"`
	PhAm      string              `json:"ph_am"`
	PhOther   *string             `json:"ph_other"`
	PhEnMp3   *string             `json:"ph_en_mp3,omitempty"`
	PhAmMp3   *string             `json:"ph_am_mp3,omitempty"`
	PhTtsMp3  *string             `json:"ph_tts_mp3,omitempty"`
	Parts     []WordInfoPart      `json:"parts"`
	Exchange  map[string][]string `json:"exchange"`
	Sentences []Sentence          `json:"sentences,omitempty"`
}
type WordInfoPart struct {
	Part  string   `json:"part"`
	Means []string `json:"means"`
}

func (row Word) TableName() string {
	return "words"
}

func NewWordModel() *WordModel {
	engine, _ := xorm.NewEngine("sqlite3", conf.DB_FILE_PATH)
	return &WordModel{
		engine: engine,
	}
}

func (model *WordModel) Query(word string) (*Word, error) {
	ret := &Word{}

	session := model.engine.NewSession()
	session.Where("word=?", word)

	has, _ := session.Get(ret)
	if has == false {
		return nil, nil
	}
	phAmMp3 := "/word_mp3/am/" + ret.Word[0:2] + "/" + ret.Word + ".mp3"
	phEnMp3 := "/word_mp3/en/" + ret.Word[0:2] + "/" + ret.Word + ".mp3"

	ret.Info.PhAmMp3 = &phAmMp3
	ret.Info.PhEnMp3 = &phEnMp3
	return ret, nil
}

func (model *WordModel) Insert(word Word) (err error) {
	err = downloadFile(*word.Info.PhEnMp3, conf.MP3_FILE_PATH+"/en/"+word.Word[0:2]+"/"+word.Word+".mp3")
	if err != nil {
		return err
	}
	err = downloadFile(*word.Info.PhAmMp3, conf.MP3_FILE_PATH+"/am/"+word.Word[0:2]+"/"+word.Word+".mp3")
	if err != nil {
		return err
	}
	word.Info.PhAmMp3 = nil
	word.Info.PhEnMp3 = nil
	word.CreateTime = time.Now()

	_, err = model.engine.NewSession().Insert(word)
	return err
}

func downloadFile(sourceFile string, targetFile string) error {
	resp, err := http.Get(sourceFile)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(targetFile)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return nil
	}

	return nil
}
