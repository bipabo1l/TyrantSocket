package controllers

import (
	"github.com/astaxie/beego"
)

type MainController struct {
	beego.Controller
}

// @router / [get]
func (c *MainController) Get() {
	c.TplName = "index.html"
}


// @router /main/ [get]
func (this *MainController) GetPageCVE() {
	this.TplName = "getStatus.html"
}