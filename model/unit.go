package model

import (
	"github.com/go-xorm/xorm"
	"github.com/wutianfang/loki/common/conf"
	"time"
)

type UnitModel struct {
	engine *xorm.Engine
}

type Unit struct {
	Id int
	Name string
	Count int
	CreateTime  time.Time
}

type UnitWordRelation struct {
	//Id int
	UnitId int
	Word string
	CreateTime  time.Time
}

func (row Unit) TableName() string {
	return "units"
}

func NewUnitModel() UnitModel {

	engine,_ := xorm.NewEngine("sqlite3", conf.DB_FILE_PATH)
/*
	if err!= nil {
		return nil
	}
	*/
	return UnitModel{
		engine :engine,
	}
}

func (engine *UnitModel) GetList() []Unit {

	session := engine.engine.NewSession()

	units := []Unit{}
	_ = session.Find(&units)

	return units
}

func (engine UnitModel) GetOne(id int ) Unit {

	session := engine.engine.NewSession()
	session.Where("id=?", id)

	unit := Unit{}
	_,_ = session.Get(&unit)

	return unit
}

func (engine *UnitModel) AddWord(unit_id int,word string) error {
	one := UnitWordRelation{
		UnitId:unit_id,
		Word:word,
	}
	_, err := engine.engine.NewSession().Insert(one)
	if err!=nil &&  err.Error() == "UNIQUE constraint failed: unit_word_relation.id" {
		return nil
	}

	return err
}

func (engine *UnitModel) GetWordList(unit_id int) ([]Word, error){
	retWords := []Word{}

	wordRelation := []UnitWordRelation{}

	session := engine.engine.NewSession()
	session.Where("unit_id=?", unit_id)
	err := session.Find(&wordRelation)

	words := []string{}
	for _,relation_row := range wordRelation {
		words = append(words, relation_row.Word)
	}

	wordSession := engine.engine.NewSession()
	wordSession.In("word", words)
	err = wordSession.Find(&retWords)

	for index, row := range retWords {
		phAmMp3 := "/word_mp3/am/" + row.Word[0:2] + "/" + row.Word + ".mp3"
		phEnMp3 := "/word_mp3/en/" + row.Word[0:2] + "/" + row.Word + ".mp3"
		retWords[index].Info.PhAmMp3 = &phAmMp3
		retWords[index].Info.PhEnMp3 = &phEnMp3

	}

	return retWords, err
}