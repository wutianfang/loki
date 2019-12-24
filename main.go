package main

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/wutianfang/loki/controller/unit"
)

func main() {

	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Static("/static", "/Users/wutianfang/go/src/github.com/wutianfang/loki/static/")

	e.GET("/unit/list", unit.List)
	e.GET("/unit/detail", unit.Detail)


	e.Logger.Fatal(e.Start(":1323"))


}

