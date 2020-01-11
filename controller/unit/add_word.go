package unit

import (
	"github.com/labstack/echo/v4"
	"github.com/wutianfang/loki/common"
	"github.com/wutianfang/loki/model"
	"strings"
)


type AddWordParams struct{
	UnitId int `form:"unit_id" query:"unit_id"`
	Word string `form:"word" query:"word"`

}

type AddWordResponse struct {
	common.CommonResponse
}

func AddWord(c echo.Context) error {

	params := AddWordParams{}
	response := AddWordResponse{}

	_ = c.Bind(&params)
	params.Word = strings.ToLower(params.Word)


	if params.UnitId==0 || params.Word == "" {
		response.Errno = 2
		response.Error = "参数错误！"
		return c.JSON(200, response)
	}

	wordModel := model.NewWordModel()

	wordInfo,err := wordModel.Query(params.Word)
	if err!= nil {
		response.Errno = 3
		response.Error = "单词查询错误！"
		return c.JSON(200, response)
	}
	if wordInfo == nil {
		response.Errno = 3
		response.Error = "单词未缓存"
		return c.JSON(200, response)
	}

	unitModel := model.NewUnitModel()

	err = unitModel.AddWord(params.UnitId, params.Word)
	if err!= nil {
		response.Errno = 4
		response.Error = err.Error()
		return c.JSON(200, response)
	}

	return c.JSON(200, response)
}
