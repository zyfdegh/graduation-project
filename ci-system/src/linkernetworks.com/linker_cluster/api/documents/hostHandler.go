package documents

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/emicklei/go-restful"
	"linkernetworks.com/linker_cluster/services"
	"linkernetworks.com/linker_common_lib/persistence/entity"
	"linkernetworks.com/linker_common_lib/rest/response"
)

func (p Resource) HostWebService() *restful.WebService {
	ws := new(restful.WebService)
	ws.Path("/v1/host")
	ws.Consumes("*/*")
	ws.Produces(restful.MIME_JSON)
	id := ws.PathParameter(ParamID, "Storage identifier of host")
	paramID := "{" + ParamID + "}"

	ws.Route(ws.POST("/").To(p.HostCreateHandler).
		Doc("Store a host").
		Operation("HostCreateHandler").
		Param(ws.HeaderParameter("X-Auth-Token", "A valid authentication token")).
		Param(ws.BodyParameter("body", "").DataType("string")))

	ws.Route(ws.GET("/").To(p.HostListHandler).
		Doc("Return hosts").
		Operation("HostListHandler").
		Param(ws.HeaderParameter("X-Auth-Token", "A valid authentication token")).
		Param(ws.QueryParameter("count", "Count total items and return the result in X-Object-Count header").DataType("boolean")).
		Param(ws.QueryParameter("skip", "Number of items to skip in the result set, default=0")).
		Param(ws.QueryParameter("limit", "Maximum number of items in the result set, default=0")).
		Param(ws.QueryParameter("sort", "")).
		Param(ws.QueryParameter("cluster_id", "The storage identifier of cluster")))

	ws.Route(ws.GET("/" + paramID).To(p.HostGetHandler).
		Doc("Return a host").
		Operation("HostGetHandler").
		Param(ws.HeaderParameter("X-Auth-Token", "A valid authentication token")).
		Param(id))

	ws.Route(ws.GET("/" + "unterminated").To(p.HostListUnterminateHandler).
		Doc("Return all unterminate Host").
		Operation("HostListUnterminateHandler").
		Param(ws.HeaderParameter("X-Auth-Token", "A valid authentication token")).
		Param(ws.QueryParameter("count", "Count total items and return the result in X-Object-count header").DataType("boolean")).
		Param(ws.QueryParameter("skip", "Number of items to skip in the result set, default=0")).
		Param(ws.QueryParameter("cluster_id", "The storage identifier of cluster")).
		Param(ws.QueryParameter("limit", "Maximum number of items in the result set, default=0")))

	ws.Route(ws.GET("/" + paramID + "/containers").To(p.HostGetContainerHandler).
		Doc("Return the containers running on the host").
		Operation("HostGetContainerHandler").
		Param(ws.HeaderParameter("X-Auth-Token", "A valid authentication token")).
		Param(ws.QueryParameter("count", "Count total items and return the result in X-Object-Count header").DataType("boolean")).
		Param(ws.QueryParameter("skip", "Number of items to skip in the result set, default=0")).
		Param(ws.QueryParameter("limit", "Maximum number of items in the result set, default=0")).
		Param(id))

	ws.Route(ws.DELETE("/" + paramID).To(p.HostDeleteHandler).
		Doc("Detele a Host by its storage identifier").
		Operation("HostDeleteHandler").
		Param(ws.HeaderParameter("X-Auth-Token", "A valid authentication token")).
		Param(id))

	ws.Route(ws.DELETE("/").To(p.HostDeleteHostsHandler).
		Doc("Detele a Host by its storage identifier").
		Operation("HostDeleteHandler").
		Param(ws.HeaderParameter("X-Auth-Token", "A valid authentication token")).
		Param(ws.QueryParameter("ids", "The storage identifier of hosts")))

	ws.Route(ws.PUT("/" + paramID).To(p.HostUpdateHandler).
		Doc("Update host by its storage identifier").
		Operation("HostUpdateHandler").
		Param(ws.HeaderParameter("X-Auth-Token", "A valid authentication token")).
		Param(id).
		Param(ws.BodyParameter("body", "").DataType("string").DataType("string")))

	return ws
}

func (p *Resource) HostListUnterminateHandler(req *restful.Request, resp *restful.Response) {
	logrus.Infof("HostListUnterminateHandler is called!")
	x_auth_token := req.HeaderParameter("X-Auth-Token")
	code, err := services.TokenValidation(x_auth_token)
	if err != nil {
		response.WriteStatusError(code, err, resp)
		return
	}

	var skip int = queryIntParam(req, "skip", 0)
	var limit int = queryIntParam(req, "limit", 0)
	var name string = req.QueryParameter("cluster_id")
	
	total, hosts, code, err := services.GetHostService().QueryAllUnterminated(name, skip, limit, x_auth_token)
	if err != nil {
		response.WriteStatusError(code, err, resp)
		return
	}
	res := response.QueryStruct{Success: true, Data: hosts}
	if c, _ := strconv.ParseBool(req.QueryParameter("count")); c {
		res.Count = total
		resp.AddHeader("X-Object-Count", strconv.Itoa(total))
	}
	resp.WriteEntity(res)
	return
}

func (p *Resource) HostDeleteHostsHandler(req *restful.Request, resp *restful.Response) {
	logrus.Infof("HostDeleteHostsHandler is called!")
	x_auth_token := req.HeaderParameter("X-Auth-Token")
	code, err := services.TokenValidation(x_auth_token)
	if err != nil {
		response.WriteStatusError(code, err, resp)
		return
	}
	ids := req.QueryParameter("ids")
	idAll := strings.Split(ids, ",")
	for _, val := range idAll {
		host, _, err := services.GetHostService().QueryById(val, x_auth_token)
		istrue := host.IsMasterNode
		clusterId := host.ClusterId
		if !istrue {
			code, err = services.GetHostService().DeleteById(val, x_auth_token)
			logrus.Infof("id is", val)
			if err != nil {
				response.WriteStatusError(code, err, resp)
				return
			}
			code, err = services.GetClusterService().ChangeClusterInstancesAccHost(host, x_auth_token)
			if err != nil {
				response.WriteStatusError(code, err, resp)
				return
			}
		}else{
			code, err = services.GetClusterService().DeleteById(clusterId, x_auth_token)
			if err != nil {
				response.WriteStatusError(code, err, resp)
				return
			}
		}
		
	}

	response.WriteSuccess(resp)
	return
}

func (p *Resource) HostUpdateHandler(req *restful.Request, resp *restful.Response) {
	logrus.Infof("HostUpdateHandler is called!")
	x_auth_token := req.HeaderParameter("X-Auth-Token")
	code, err := services.TokenValidation(x_auth_token)
	if err != nil {
		response.WriteStatusError(code, err, resp)
		return
	}
	objectId := req.PathParameter(ParamID)

	host := entity.Host{}
	err = json.NewDecoder(req.Request.Body).Decode(&host)
	if err != nil {
		logrus.Errorf("convert body to host failed, error is %v", err)
		response.WriteStatusError(services.COMMON_ERROR_INVALIDATE, err, resp)
		return
	}
	created, code, err := services.GetHostService().UpdateById(objectId, host, x_auth_token)
	if err != nil {
		response.WriteStatusError(code, err, resp)
		return
	}
	p.successUpdate(objectId, created, req, resp)

}

func (p *Resource) HostDeleteHandler(req *restful.Request, resp *restful.Response) {
	logrus.Infof("HostDeleteHandler is called!")
	x_auth_token := req.HeaderParameter("X-Auth-Token")
	code, err := services.TokenValidation(x_auth_token)
	if err != nil {
		response.WriteStatusError(code, err, resp)
		return
	}

	objectId := req.PathParameter(ParamID)
	host, _, err := services.GetHostService().QueryById(objectId, x_auth_token)
	istrue := host.IsMasterNode
	clusterId := host.ClusterId
	if !istrue {
		code, err = services.GetHostService().DeleteById(objectId, x_auth_token)
		if err != nil {
			response.WriteStatusError(code, err, resp)
			return
		}
		code, err = services.GetClusterService().ChangeClusterInstancesAccHost(host, x_auth_token)
		if err != nil {
			response.WriteStatusError(code, err, resp)
			return
		}
	}else{
		code, err = services.GetClusterService().DeleteById(clusterId, x_auth_token)
		if err != nil {
			response.WriteStatusError(code, err, resp)
			return
		}
	}
	
	
	
	// Write success response
	response.WriteSuccess(resp)
	return
}

func (p *Resource) HostCreateHandler(req *restful.Request, resp *restful.Response) {
	logrus.Infof("HostCreateHandler is called!")
	x_auth_token := req.HeaderParameter("X-Auth-Token")
	code, err := services.TokenValidation(x_auth_token)
	if err != nil {
		response.WriteStatusError(code, err, resp)
		return
	}

	host := entity.Host{}
	err = json.NewDecoder(req.Request.Body).Decode(&host)
	if err != nil {
		logrus.Errorf("convert body to cluster failed, error is %v", err)
		response.WriteStatusError(services.COMMON_ERROR_INVALIDATE, err, resp)
		return
	}

	newHost, code, err := services.GetHostService().Create(host, x_auth_token)
	if err != nil {
		response.WriteStatusError(code, err, resp)
		return
	}

	res := response.QueryStruct{Success: true, Data: newHost}
	resp.WriteEntity(res)
	return

}

func (p *Resource) HostGetHandler(req *restful.Request, resp *restful.Response) {
	logrus.Infof("HostGetHandler is called!")
	x_auth_token := req.HeaderParameter("X-Auth-Token")
	code, err := services.TokenValidation(x_auth_token)
	if err != nil {
		response.WriteStatusError(code, err, resp)
		return
	}

	objectId := req.PathParameter(ParamID)
	host, code, err := services.GetHostService().QueryById(objectId, x_auth_token)
	if err != nil {
		response.WriteStatusError(code, err, resp)
		return
	}
	logrus.Debugf("host is %v", host)

	res := response.QueryStruct{Success: true, Data: host}
	resp.WriteEntity(res)
	return

}

func (p *Resource) HostGetContainerHandler(req *restful.Request, resp *restful.Response) {
	logrus.Infof("HostGetContainerHandler is called!")
	x_auth_token := req.HeaderParameter("X-Auth-Token")
	code, err := services.TokenValidation(x_auth_token)
	if err != nil {
		response.WriteStatusError(code, err, resp)
		return
	}

	objectId := req.PathParameter(ParamID)
	var skip int = queryIntParam(req, "skip", 0)
	var limit int = queryIntParam(req, "limit", 0)
	total, instances, code, err := services.GetHostService().QueryContainersById(objectId, skip, limit, x_auth_token)
	if err != nil {
		response.WriteStatusError(code, err, resp)
		return
	}
	res := response.QueryStruct{Success: true, Data: instances}
	if c, _ := strconv.ParseBool(req.QueryParameter("count")); c {
		res.Count = total
		resp.AddHeader("X-Object-Count", strconv.Itoa(total))
	}
	resp.WriteEntity(res)
	return

}

func (p *Resource) HostListHandler(req *restful.Request, resp *restful.Response) {
	logrus.Infof("HostListHandler is called!")
	x_auth_token := req.HeaderParameter("X-Auth-Token")
	code, err := services.TokenValidation(x_auth_token)
	if err != nil {
		response.WriteStatusError(code, err, resp)
		return
	}

	var skip int = queryIntParam(req, "skip", 0)
	var limit int = queryIntParam(req, "limit", 0)
	var name string = req.QueryParameter("cluster_id")

	total, hosts, code, err := services.GetHostService().QueryAllByName(name, skip, limit, x_auth_token)
	if err != nil {
		response.WriteStatusError(code, err, resp)
		return
	}
	res := response.QueryStruct{Success: true, Data: hosts}
	if c, _ := strconv.ParseBool(req.QueryParameter("count")); c {
		res.Count = total
		resp.AddHeader("X-Object-Count", strconv.Itoa(total))
	}
	resp.WriteEntity(res)
	return

}
