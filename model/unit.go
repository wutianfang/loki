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
