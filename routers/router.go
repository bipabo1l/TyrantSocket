package routers

import (
	"TyrantSocket/controllers"
	"github.com/astaxie/beego"
)

func init() {
	beego.Include(&controllers.AgentController{})
	beego.Include(&controllers.MainController{})
}
