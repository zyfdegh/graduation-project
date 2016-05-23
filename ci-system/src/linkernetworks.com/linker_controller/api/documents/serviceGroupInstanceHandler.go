package documents

import (
	"github.com/Sirupsen/logrus"
	"github.com/emicklei/go-restful"
	"linkernetworks.com/linker_common_lib/rest/response"
	"linkernetworks.com/linker_controller/services"
	"strconv"
)

func (p Resource) ServiceGroupInstanceWebService() *restful.WebService {
	ws := new(restful.WebService)
	ws.Path("/v1/groupInstances")
	ws.Consumes("*/*")
	ws.Produces(restful.MIME_JSON)

	id := ws.PathParameter(ParamID, "Storage identifier of the group instance")
	paramID := "{" + ParamID + "}"

	ws.Route(ws.GET("/").To(p.SGIsListHandler).
		Doc("Return all service group instance items").
		Operation("SGIsListHandler").
		Param(ws.HeaderParameter("X-Auth-Token", "A valid authentication token")).
		Param(ws.QueryParameter("count", "Count total items and return the result in X-Object-Count header").DataType("boolean")).
		Param(ws.QueryParameter("skip", "Number of items to skip in the result set, default=0")).
		Param(ws.QueryParameter("limit", "Maximum number of items in the result set, default=0")))

	ws.Route(ws.GET("/" + paramID).To(p.SGIsGetHandler).
		Doc("Return a service group instance by its storage identifier").
		Operation("SGIsGetHandler").
		Param(ws.HeaderParameter("X-Auth-Token", "A valid authentication token")).
		Param(id).
		Param(ws.QueryParameter("fields", "Comma separated list of field names")))

	return ws
}

// SGIsListHandler parses the http request and return service group instance items.
// Usage :
//		GET /v1/groupInstances
// Params :
//If successful,response code will be set to 201.
func (p *Resource) SGIsListHandler(req *restful.Request, resp *restful.Response) {
	x_auth_token := req.HeaderParameter("X-Auth-Token")
	code, err := services.TokenValidation(x_auth_token)
	if err != nil {
		response.WriteStatusError(code, err, resp)
		return
	}
	var skip int = queryIntParam(req, "skip", 0)
	var limit int = queryIntParam(req, "limit", 0)

	total, sgis, code, err := services.GetSgiService().QueryAll(skip, limit, x_auth_token)
	if err != nil {
		response.WriteStatusError(code, err, resp)
		return
	}
	res := response.QueryStruct{Success: true, Data: sgis}
	if c, _ := strconv.ParseBool(req.QueryParameter("count")); c {
		res.Count = total
		resp.AddHeader("X-Object-Count", strconv.Itoa(total))
	}
	resp.WriteEntity(res)
	return

}

// SGIsGetHandler parses the http request and return service group instance by objectId.
// Usage :
//		GET /v1/groupInstances/{ParamID}
// Params :
//		ParamID : Storage identifier of the group instance
//If successful,response code will be set to 201.
func (p *Resource) SGIsGetHandler(req *restful.Request, resp *restful.Response) {
	x_auth_token := req.HeaderParameter("X-Auth-Token")
	code, err := services.TokenValidation(x_auth_token)
	if err != nil {
		response.WriteStatusError(code, err, resp)
		return
	}

	objectId := req.PathParameter(ParamID)

	sgi, code, err := services.GetSgiService().QueryById(objectId, x_auth_token)
	if err != nil {
		response.WriteStatusError(code, err, resp)
		return
	}
	logrus.Debugf("sgi is %v", sgi)

	res := response.QueryStruct{Success: true, Data: sgi}
	resp.WriteEntity(res)
	return
}
