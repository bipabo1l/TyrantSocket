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
func (this *AgentController) GetAgent() {
	this.TplName = "status/index.html"
}

// @router /agentstatus [get]
func (this *AgentController) Get() {
	var agentSercver models.AgentServer
	result, num, err := agentSercver.GetStatus()
	if err != nil {
		log.Println("GetStatus 接口调用错误")
		this.Data["json"] = utils.AjaxReturn(result, "result invalid", 0)
	} else {
		this.Data["json"] = utils.AjaxReturn(result, "success", num)
	}
	this.ServeJSON()
}

// @router /stop [post]
func (this *AgentController) Post() {
	ip := this.Input().Get("ip")
	var agentSercver models.AgentServer
	err := agentSercver.StopStatus(ip)
	if err != nil {
		panic(err)
	}
	this.Data["json"] = utils.AjaxReturn(1, "success", 1)
	this.ServeJSON()
}

// @router /agentiprange [get]
func (this *AgentController) Range() {
	var agentSercver models.AgentServer
	result,  err := agentSercver.GetIPRange()
	if err != nil {
		log.Println("GetIPRange 接口调用错误")
		this.Data["json"] = utils.AjaxReturn(result, "result invalid", 0)
	} else {
		this.Data["json"] = utils.AjaxReturn(result, "success", 1)
	}
	this.ServeJSON()
}