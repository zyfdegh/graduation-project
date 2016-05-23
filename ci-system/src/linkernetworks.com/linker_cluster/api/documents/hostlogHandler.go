package documents

import (
	"strconv"
	"encoding/json"
	"github.com/Sirupsen/logrus"
	"github.com/emicklei/go-restful"
	"linkernetworks.com/linker_cluster/services"
	"linkernetworks.com/linker_common_lib/rest/response"
	"linkernetworks.com/linker_common_lib/persistence/entity"
)

func (p Resource) HostLogWebService() *restful.WebService{
	ws := new(restful.WebService)
	ws.Path("/v1/hostlog")
	ws.Consumes("*/*")
	ws.Produces(restful.MIME_JSON)
	
//	id := ws.PathParameter(ParamID, "Storage identifier of hostlog")
//	paramID := "{" + ParamID + "}"
	
	ws.Route(ws.POST("/").To(p.HostLogCreateHandler).
		Doc("Store a hostlog").
		Operation("HostLogCreateHandler").
		Param(ws.HeaderParameter("X-Auth-Token", "A valid authentication token")).
		Param(ws.BodyParameter("body", "").DataType("string")))
	
	ws.Route(ws.GET("/").To(p.HostLogListHandler).
		Doc("Returns all hostlog items").
		Operation("HostLogListHandler").
		Param(ws.HeaderParameter("X-Auth-Token", "A valid authentication token")).
		Param(ws.QueryParameter("count", "Count total items and return the result in X-Object-Count header").DataType("boolean")).
		Param(ws.QueryParameter("skip", "Number of items to skip in the result set, default=0")).
		Param(ws.QueryParameter("cluster_id", "The name of cluster wanted to query")).
		Param(ws.QueryParameter("limit", "Maximum number of items in the result set, default=0")))
		
	return ws
} 

func (p *Resource) HostLogListHandler(req *restful.Request, resp *restful.Response) {
	logrus.Infoln("HostLogListHandler is called!")
	x_auth_token := req.HeaderParameter("X-Auth-Token")
	code, err := services.TokenValidation(x_auth_token)
	if err != nil {
		response.WriteStatusError(code, err, resp)
		return
	}
	
	var skip int = queryIntParam(req, "skip", 0)
	var limit int = queryIntParam(req, "limit", 0)
	var name string = req.QueryParameter("cluster_id")
	
	total, hostlogs, code, err := services.GetClusterService().QueryAllById(name, skip, limit, x_auth_token)
	if err != nil {
		response.WriteStatusError(code, err, resp)
		return
	}
	res := response.QueryStruct{Success: true, Data: hostlogs}
	if c, _ := strconv.ParseBool(req.QueryParameter("count")); c {
		res.Count = total
		resp.AddHeader("X-Object-Count", strconv.Itoa(total))
	}
	resp.WriteEntity(res)
	return
	
}

func (p *Resource) HostLogCreateHandler(req *restful.Request, resp *restful.Response) {
	logrus.Infoln("HostLogCreateHandler is called!")
	x_auth_token := req.HeaderParameter("X-Auth-Token")
	code, err := services.TokenValidation(x_auth_token)
	if err != nil {
		response.WriteStatusError(code, err, resp)
		return
	}
	// Stub an acluster to be populated from the body
	hostlog := entity.HostLog{}

	err = json.NewDecoder(req.Request.Body).Decode(&hostlog)
	if err != nil {
		logrus.Errorf("convert body to hostlog failed, error is %v", err)
		response.WriteStatusError(services.COMMON_ERROR_INVALIDATE, err, resp)
		return
	}
	
	newHostLog, code, err := services.GetHostLogService().Create(hostlog, x_auth_token)
	if err != nil {
		response.WriteStatusError(code, err, resp)
		return
	}

	res := response.QueryStruct{Success: true, Data: newHostLog}
	resp.WriteEntity(res)
	return
}



