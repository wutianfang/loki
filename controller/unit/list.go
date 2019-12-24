package unit

import (
	//"github.com/go-xorm/xorm"
	"github.com/labstack/echo/v4"
	_ "github.com/mattn/go-sqlite3"
	"github.com/wutianfang/loki/model"
	"time"
)

type Units struct {
	Id int
	Name string
	Count int
	CreateTime  time.Time
}



func List(c echo.Context) error {

	unitModel := model.NewUnitModel()
	units := unitModel.GetList()


	err := c.JSON(200, units)

	return err
}

func Detail(c echo.Context) error {

	unitModel := model.NewUnitModel()

	params := struct{
		Id  int `form:"id" query:"id"`
	}{}

	_= c.Bind(&params)

	unit := unitModel.GetOne(params.Id)


	err := c.JSON(200, unit)

	return err
}