package model

import (
	"github.com/go-xorm/xorm"
	"time"
)

type UnitModel struct {
	engine *xorm.Engine
}

type Units struct {
	Id int
	Name string
	Count int
	CreateTime  time.Time
}

func NewUnitModel() UnitModel {

	engine,_ := xorm.NewEngine("sqlite3", "/Users/wutianfang/go/src/github.com/wutianfang/loki/db/loki.db")
/*
	if err!= nil {
		return nil
	}
	*/
	return UnitModel{
		engine :engine,
	}
}

func (engine *UnitModel) GetList() []Units {

	session := engine.engine.NewSession()

	units := []Units{}
	_ = session.Find(&units)

	return units
}

func (engine UnitModel) GetOne(id int ) Units {

	session := engine.engine.NewSession()
	session.Where("id=?", id)

	unit := Units{}
	_,_ = session.Get(&unit)

	return unit
}
