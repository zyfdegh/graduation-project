package documents

import (
	"github.com/emicklei/go-restful"
	"linkernetworks.com/linker_common_lib/rest/response"
	"linkernetworks.com/linker_controller/services"
)

func (p Resource) UserAccountBalanceWebService() *restful.WebService {
	ws := new(restful.WebService)
	ws.Path("/v1/userAccountBalance")
	ws.Consumes("*/*")
	ws.Produces(restful.MIME_JSON)

	// id := ws.PathParameter(ParamID, "Storage identifier of service group")
	// paramID := "{" + ParamID + "}"

	ws.Route(ws.GET("/").To(p.UabsListHandler).
		Doc("Return all user account balance items according to query parameter.").
		Operation("UabsListHandler").
		Param(ws.HeaderParameter("X-Auth-Token", "A valid authentication token")).
		Param(ws.QueryParameter("userId", "The name of services wanted to query")))

	return ws
}

func (p *Resource) UabsListHandler(req *restful.Request, resp *restful.Response) {
	token := req.HeaderParameter("X-Auth-Token")
	userId := req.QueryParameter("userId")

	uabs, errorCode, err := services.GetUserAccountBalanceService().QueryByUserId(token, userId)
	if err != nil {
		response.WriteStatusError(errorCode, err, resp)
		return
	}
	res := response.QueryStruct{Success: true, Data: uabs}
	resp.WriteEntity(res)
	return
}
