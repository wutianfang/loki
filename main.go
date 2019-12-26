package main

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/wutianfang/loki/common/conf"
	"github.com/wutianfang/loki/controller/unit"
	"github.com/wutianfang/loki/controller/word"
)

func main() {

	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Static("/static", "/Users/wutianfang/go/src/github.com/wutianfang/loki/static/")
	e.Static("/word_mp3", conf.MP3_FILE_PATH)


	e.GET("/unit/list", unit.List)
	e.GET("/unit/detail", unit.Detail)
	e.Any("/unit/add_word", unit.AddWord)
	e.GET("/unit/word_list", unit.WordList)

	e.GET("/word/query", word.Query)



	e.Logger.Fatal(e.Start(":1323"))


}

