package documents

import (
	"encoding/json"
	"strconv"

	"github.com/Sirupsen/logrus"
	"github.com/emicklei/go-restful"
	"linkernetworks.com/linker_cluster/services"
	"linkernetworks.com/linker_common_lib/persistence/entity"
	"linkernetworks.com/linker_common_lib/rest/response"
)

func (p Resource) ClusterWebService() *restful.WebService {
	ws := new(restful.WebService)
	ws.Path("/v1/cluster")
	ws.Consumes("*/*")
	ws.Produces(restful.MIME_JSON)

	id := ws.PathParameter(ParamID, "Storage identifier of cluster")
	paramID := "{" + ParamID + "}"

	ws.Route(ws.POST("/").To(p.ClusterCreateHandler).
		Doc("Store a cluster").
		Operation("ClusterCreateHandler").
		Param(ws.HeaderParameter("X-Auth-Token", "A valid authentication token")).
		Param(ws.BodyParameter("body", "").DataType("string")))

	ws.Route(ws.GET("/").To(p.ClustersListHandler).
		Doc("Returns all cluster items").
		Operation("ClustersListHandler").
		Param(ws.HeaderParameter("X-Auth-Token", "A valid authentication token")).
		Param(ws.QueryParameter("count", "Count total items and return the result in X-Object-Count header").DataType("boolean")).
		Param(ws.QueryParameter("skip", "Number of items to skip in the result set, default=0")).
		Param(ws.QueryParameter("_id", "The name of cluster wanted to query")).
		Param(ws.QueryParameter("limit", "Maximum number of items in the result set, default=0")))

	ws.Route(ws.GET("/" + paramID).To(p.ClusterGetHandler).
		Doc("Return a cluster").
		Operation("ClusterGetHandler").
		Param(ws.HeaderParameter("X-Auth-Token", "A valid authentication token")).
		Param(id))

	ws.Route(ws.GET("/" + "unterminated").To(p.ClusterListUnterminateHandler).
		Doc("Return all unterminate Cluster").
		Operation("ClusterListUnterminateHandler").
		Param(ws.HeaderParameter("X-Auth-Token", "A valid authentication token")).
		Param(ws.QueryParameter("count", "Count total items and return the result in X-Object-count header").DataType("boolean")).
		Param(ws.QueryParameter("skip", "Number of items to skip in the result set, default=0")).
		Param(ws.QueryParameter("limit", "Maximum number of items in the result set, default=0")))

	ws.Route(ws.DELETE("/" + paramID).To(p.ClusterDeleteHandler).
		Doc("Detele a Cluster by its storage identifier").
		Operation("ClusterDeleteHandler").
		Param(ws.HeaderParameter("X-Auth-Token", "A valid authentication token")).
		Param(id))

	ws.Route(ws.DELETE("/").To(p.ClusterDeleteByQueryHandler).
		Doc("Detele a Cluster by query condition").
		Operation("ClusterDeleteByQueryHandler").
		Param(ws.HeaderParameter("X-Auth-Token", "A valid authentication token")).
		Param(ws.QueryParameter("userid", "Storage identifier of an user ")).
		Param(id))

	ws.Route(ws.PUT("/" + paramID).To(p.ClusterUpdateHandler).
		Doc("Update a Cluster by its storage identifier").
		Operation("ClusterUpdateHandler").
		Param(ws.HeaderParameter("X-Auth-Token", "A valid authentication token")).
		Param(id).
		Param(ws.QueryParameter("instances", "")))
		
	

	return ws

}

func (p *Resource) ClusterListUnterminateHandler(req *restful.Request, resp *restful.Response) {
	logrus.Infoln("ClusterListUnterminateHandler is called!")
	x_auth_token := req.HeaderParameter("X-Auth-Token")
	code, err := services.TokenValidation(x_auth_token)
	if err != nil {
		response.WriteStatusError(code, err, resp)
		return
	}

	var skip int = queryIntParam(req, "skip", 0)
	var limit int = queryIntParam(req, "limit", 0)

	total, clusters, code, err := services.GetClusterService().QueryAllUnterminated("", skip, limit, x_auth_token)
	if err != nil {
		response.WriteStatusError(code, err, resp)
		return
	}
	res := response.QueryStruct{Success: true, Data: clusters}
	if c, _ := strconv.ParseBool(req.QueryParameter("count")); c {
		res.Count = total
		resp.AddHeader("X-Object-Count", strconv.Itoa(total))
	}
	resp.WriteEntity(res)
	return
}

func (p *Resource) ClusterDeleteHandler(req *restful.Request, resp *restful.Response) {
	logrus.Infoln("ClusterDeleteHandler is called!")
	x_auth_token := req.HeaderParameter("X-Auth-Token")
	code, err := services.TokenValidation(x_auth_token)
	if err != nil {
		response.WriteStatusError(code, err, resp)
		return
	}

	objectId := req.PathParameter(ParamID)

	code, err = services.GetClusterService().DeleteById(objectId, x_auth_token)
	if err != nil {
		response.WriteStatusError(code, err, resp)
		return
	}
	// Write success response
	response.WriteSuccess(resp)
	return
}

func (p *Resource) ClusterDeleteByQueryHandler(req *restful.Request, resp *restful.Response) {
	logrus.Infoln("ClusterDeleteByQueryHandler is called!")
	x_auth_token := req.HeaderParameter("X-Auth-Token")
	code, err := services.TokenValidation(x_auth_token)
	if err != nil {
		response.WriteStatusError(code, err, resp)
		return
	}

	userId := req.QueryParameter("userid")

	code, err = services.GetClusterService().DeleteByQuery(userId, x_auth_token)
	if err != nil {
		response.WriteStatusError(code, err, resp)
		return
	}
	// Write success response
	response.WriteSuccess(resp)
	return
}

func (p *Resource) ClusterCreateHandler(req *restful.Request, resp *restful.Response) {
	logrus.Infof("ClusterCreateHandler is called!")
	x_auth_token := req.HeaderParameter("X-Auth-Token")
	code, err := services.TokenValidation(x_auth_token)
	if err != nil {
		response.WriteStatusError(code, err, resp)
		return
	}
	// Stub an acluster to be populated from the body
	cluster := entity.Cluster{}

	err = json.NewDecoder(req.Request.Body).Decode(&cluster)
	if err != nil {
		logrus.Errorf("convert body to cluster failed, error is %v", err)
		response.WriteStatusError(services.COMMON_ERROR_INVALIDATE, err, resp)
		return
	}

	newCluster, code, err := services.GetClusterService().Create(cluster, x_auth_token)
	if err != nil {
		response.WriteStatusError(code, err, resp)
		return
	}

	res := response.QueryStruct{Success: true, Data: newCluster}
	resp.WriteEntity(res)
	return

}

func (p *Resource) ClustersListHandler(req *restful.Request, resp *restful.Response) {
	logrus.Infof("ClustersListHandler is called!")
	x_auth_token := req.HeaderParameter("X-Auth-Token")
	code, err := services.TokenValidation(x_auth_token)
	if err != nil {
		response.WriteStatusError(code, err, resp)
		return
	}

	var skip int = queryIntParam(req, "skip", 0)
	var limit int = queryIntParam(req, "limit", 0)
	var name string = req.QueryParameter("_id")

	total, clusters, code, err := services.GetClusterService().QueryAllById(name, skip, limit, x_auth_token)
	if err != nil {
		response.WriteStatusError(code, err, resp)
		return
	}
	res := response.QueryStruct{Success: true, Data: clusters}
	if c, _ := strconv.ParseBool(req.QueryParameter("count")); c {
		res.Count = total
		resp.AddHeader("X-Object-Count", strconv.Itoa(total))
	}
	resp.WriteEntity(res)
	return

}

func (p *Resource) ClusterGetHandler(req *restful.Request, resp *restful.Response) {
	logrus.Infof("ClusterGetHandler is called!")
	x_auth_token := req.HeaderParameter("X-Auth-Token")
	code, err := services.TokenValidation(x_auth_token)
	if err != nil {
		response.WriteStatusError(code, err, resp)
		return
	}

	objectId := req.PathParameter(ParamID)
	cluster, code, err := services.GetClusterService().QueryById(objectId, x_auth_token)
	if err != nil {
		response.WriteStatusError(code, err, resp)
		return
	}
	logrus.Debugf("cluster is %v", cluster)

	res := response.QueryStruct{Success: true, Data: cluster}
	resp.WriteEntity(res)
	return

}

func (p *Resource) ClusterUpdateHandler(req *restful.Request, resp *restful.Response) {
	logrus.Infof("ClusterUpdateHandler is called!")
	x_auth_token := req.HeaderParameter("X-Auth-Token")
	code, err := services.TokenValidation(x_auth_token)
	if err != nil {
		response.WriteStatusError(code, err, resp)
		return
	}

	objectId := req.PathParameter(ParamID)
	instances := queryIntParam(req, "instances", 0)
	cluster := entity.Cluster{}

	created, code, err := services.GetClusterService().UpdateById(objectId, instances, cluster, x_auth_token)
	if err != nil {
		response.WriteStatusError(code, err, resp)
		return
	}
	p.successUpdate(objectId, created, req, resp)
}
