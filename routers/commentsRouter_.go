package routers

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context/param"
)

func init() {

	beego.GlobalControllerRouter["TyrantSocket/controllers:AgentController"] = append(beego.GlobalControllerRouter["TyrantSocket/controllers:AgentController"],
		beego.ControllerComments{
			Method: "GetAgent",
			Router: `/status`,
			AllowHTTPMethods: []string{"get"},
			MethodParams: param.Make(),
			Params: nil})

	beego.GlobalControllerRouter["TyrantSocket/controllers:AgentController"] = append(beego.GlobalControllerRouter["TyrantSocket/controllers:AgentController"],
		beego.ControllerComments{
			Method: "Get",
			Router: `/agentstatus`,
			AllowHTTPMethods: []string{"get"},
			MethodParams: param.Make(),
			Params: nil})

	beego.GlobalControllerRouter["TyrantSocket/controllers:AgentController"] = append(beego.GlobalControllerRouter["TyrantSocket/controllers:AgentController"],
		beego.ControllerComments{
			Method: "Post",
			Router: `/stop`,
			AllowHTTPMethods: []string{"post"},
			MethodParams: param.Make(),
			Params: nil})

	beego.GlobalControllerRouter["TyrantSocket/controllers:AgentController"] = append(beego.GlobalControllerRouter["TyrantSocket/controllers:AgentController"],
		beego.ControllerComments{
			Method: "Range",
			Router: `/agentiprange`,
			AllowHTTPMethods: []string{"get"},
			MethodParams: param.Make(),
			Params: nil})

	beego.GlobalControllerRouter["TyrantSocket/controllers:MainController"] = append(beego.GlobalControllerRouter["TyrantSocket/controllers:MainController"],
		beego.ControllerComments{
			Method: "Get",
			Router: `/`,
			AllowHTTPMethods: []string{"get"},
			MethodParams: param.Make(),
			Params: nil})

	beego.GlobalControllerRouter["TyrantSocket/controllers:MainController"] = append(beego.GlobalControllerRouter["TyrantSocket/controllers:MainController"],
		beego.ControllerComments{
			Method: "GetPageCVE",
			Router: `/main/`,
			AllowHTTPMethods: []string{"get"},
			MethodParams: param.Make(),
			Params: nil})

	beego.GlobalControllerRouter["TyrantSocket/controllers:MainController"] = append(beego.GlobalControllerRouter["TyrantSocket/controllers:MainController"],
		beego.ControllerComments{
			Method: "Main",
			Router: `/main1/`,
			AllowHTTPMethods: []string{"get"},
			MethodParams: param.Make(),
			Params: nil})

}
