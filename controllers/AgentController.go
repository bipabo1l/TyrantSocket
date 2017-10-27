package controllers

import (
	"github.com/astaxie/beego"
	"TyrantSocket/models"
	"log"
	"TyrantSocket/utils"
)

func init() {

}

type AgentController struct {
	beego.Controller
}

// @router /status [get]
func (this *AgentController) GetAgent(){
	this.TplName = "status/index.html"
}

// @router /agentstatus [get]
func (this *AgentController) Get() {
	var agentSercver models.AgentServer
	result, num ,err := agentSercver.GetStatus()
	if err != nil {
		log.Println("GetStatus 接口调用错误")
		this.Data["json"] = utils.AjaxReturn(result, "result invalid", 0)
	} else {
		this.Data["json"] = utils.AjaxReturn(result, "success", num)
	}
	this.ServeJSON()
}
