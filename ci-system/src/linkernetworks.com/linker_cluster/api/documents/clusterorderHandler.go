package documents

import (
	"encoding/json"
	"errors"
	"strconv"

	"github.com/Sirupsen/logrus"
	"github.com/emicklei/go-restful"
	"linkernetworks.com/linker_cluster/services"
	"linkernetworks.com/linker_common_lib/persistence/entity"
	"linkernetworks.com/linker_common_lib/rest/response"
)

func (p Resource) ClusterOrderWebService() *restful.WebService {
	ws := new(restful.WebService)
	ws.Path("/v1/clusterorder")
	ws.Consumes("*/*")
	ws.Produces(restful.MIME_JSON)

	id := ws.PathParameter(ParamID, "Storage identifier of cluster")
	paramID := "{" + ParamID + "}"

	ws.Route(ws.POST("/").To(p.ClusterOrderCreateHandler).
		Doc("Creat a new order of service group in cluster").
		Operation("ClusterOrderCreateHandler").
		Param(ws.HeaderParameter("X-Auth-Token", "A valid authentication token")).
		Param(ws.BodyParameter("body", "Service order in json format,for example {\"service_group_id\":\"...\", \"cluster_id\":\"...\",  \"parameters\": [{\"appId\":\"...\",  \"paramName\":\"...\",  \"paramValue\":\"...\"}]} ").DataType("string")))

	ws.Route(ws.DELETE("/" + paramID).To(p.ClusterOrderTerminateHandler).
		Doc("Terminate a service order by its storage identifier").
		Operation("ClusterOrderTerminateHandler").
		Param(ws.HeaderParameter("X-Auth-Token", "A valid authentication token")).
		Param(ws.QueryParameter("cluster_id", "Storage identifier of a cluster")).
		Param(id))

	ws.Route(ws.GET("/").To(p.ClusterOrderListHandler).
		Doc("Returns all service group orders").
		Operation("ClusterOrderListHandler").
		Param(ws.HeaderParameter("X-Auth-Token", "A valid authentication token")).
		Param(ws.QueryParameter("count", "Count total items and return the result in X-Object-Count header").DataType("boolean")).
		Param(ws.QueryParameter("skip", "Number of items to skip in the result set, default=0")).
		Param(ws.QueryParameter("limit", "Maximum number of items in the result set, default=0")).
		Param(ws.QueryParameter("sort", "Comma separated list of field names to sort")))

	ws.Route(ws.GET("/" + paramID).To(p.ClusterOrderGetHandler).
		Doc("Return a service group order ").
		Operation("ClusterOrderGetHandler").
		Param(ws.HeaderParameter("X-Auth-Token", "A valid authentication token")).
		Param(ws.QueryParameter("cluster_id", "Storage identifier of a cluster")).
		Param(id))

	ws.Route(ws.GET("/operations/" + paramID).To(p.ClusterOrderAuthOperationHandler).
		Doc("Return authorized operations of a service group order by its storage identifier").
		Operation("ClusterOrderAuthOperationHandler").
		Param(ws.HeaderParameter("X-Auth-Token", "A valid authentication token")).
		Param(ws.QueryParameter("cluster_id", "Storage identifier of a cluster")).
		Param(id))

	ws.Route(ws.GET("/" + paramID + "/scaleInfo").To(p.ClusterOrderAppHandler).
		Doc("Return authorized operations of a service group order by its storage identifier").
		Operation("ClusterOrderAppHandler").
		Param(ws.HeaderParameter("X-Auth-Token", "A valid authentication token")).
		Param(ws.QueryParameter("appId", "App path in the group")).
		Param(ws.QueryParameter("cluster_id", "Storage identifier of a cluster")).
		Param(id))

	ws.Route(ws.PUT("/" + paramID + "/scaleApp").To(p.ClusterOrderAppScaleHandler).
		Doc("Update the number of app in service order by its storage identifier").
		Operation("ClusterOrderAppScaleHandler").
		Param(ws.HeaderParameter("X-Auth-Token", "A valid authentication token")).
		Param(id).
		Param(ws.QueryParameter("appId", "App path in the group")).
		Param(ws.QueryParameter("num", "New number of app")).
		Param(ws.QueryParameter("cluster_id", "Storage identifier of a cluster")))

	ws.Route(ws.GET("/instance/" + paramID).To(p.ClusterOrderInstanceGetHandler).
		Doc("Return a service group order instance ").
		Operation("ClusterOrderInstanceGetHandler").
		Param(ws.HeaderParameter("X-Auth-Token", "A valid authentication token")).
		Param(ws.QueryParameter("cluster_id", "Storage identifier of a cluster")).
		Param(id))

	return ws

}

func (p *Resource) ClusterOrderCreateHandler(req *restful.Request, resp *restful.Response) {
	logrus.Infoln("ClusterOrderCreateHandler is called!")

	x_auth_token := req.HeaderParameter("X-Auth-Token")
	// Stub an order to be populated from the body
	sgo := entity.ServiceGroupOrder{}

	// Populate the user data
	err := json.NewDecoder(req.Request.Body).Decode(&sgo)
	if err != nil {
		logrus.Errorf("convert body to sgo failed, error is %v", err)
		response.WriteStatusError(services.COMMON_ERROR_INVALIDATE, err, resp)
		return
	}

	newsgo, errorCode, err := services.GetClusterOrderService().CreateOrder(x_auth_token, sgo)
	if err != nil {
		response.WriteStatusError(errorCode, err, resp)
		return
	}

	res := response.QueryStruct{Success: true, Data: newsgo}
	resp.WriteEntity(res)
	return
}

func (p *Resource) ClusterOrderTerminateHandler(req *restful.Request, resp *restful.Response) {
	logrus.Infoln("ClusterOrderTerminateHandler is called!")

	x_auth_token := req.HeaderParameter("X-Auth-Token")
	clusterId := req.QueryParameter("cluster_id")
	sgObjId := req.PathParameter(ParamID)
	if len(sgObjId) <= 0 {
		logrus.Errorf("sg obj id can not be null!")
		response.WriteStatusError(services.COMMON_ERROR_INVALIDATE, errors.New("service group obj id can not be null"), resp)
		return
	}

	errorCode, err := services.GetClusterOrderService().TerminateOrder(x_auth_token, clusterId, sgObjId)
	if err != nil {
		response.WriteStatusError(errorCode, err, resp)
		return
	}

	response.WriteSuccess(resp)
	return
}

func (p *Resource) ClusterOrderListHandler(req *restful.Request, resp *restful.Response) {
	logrus.Infoln("ClusterOrderListHandler")

	x_auth_token := req.HeaderParameter("X-Auth-Token")
	var skip int = queryIntParam(req, "skip", 0)
	var limit int = queryIntParam(req, "limit", 0)

	total, orders, errorCode, err := services.GetClusterOrderService().QueryAll(x_auth_token, skip, limit)
	if err != nil {
		response.WriteStatusError(errorCode, err, resp)
		return
	}

	res := response.QueryStruct{Success: true, Data: orders}
	if c, _ := strconv.ParseBool(req.QueryParameter("count")); c {
		res.Count = total
		resp.AddHeader("X-Object-Count", strconv.Itoa(total))
	}
	resp.WriteEntity(res)
	return

}

func (p *Resource) ClusterOrderGetHandler(req *restful.Request, resp *restful.Response) {
	logrus.Infoln("ClusterOrderGetHandler is called")

	x_auth_token := req.HeaderParameter("X-Auth-Token")
	sgoId := req.PathParameter(ParamID)
	clusterId := req.QueryParameter("cluster_id")

	sgo, code, err := services.GetClusterOrderService().QuerySGOById(x_auth_token, clusterId, sgoId)
	if err != nil {
		response.WriteStatusError(code, err, resp)
		return
	}

	res := response.QueryStruct{Success: true, Data: sgo}
	resp.WriteEntity(res)
	return
}

func (p *Resource) ClusterOrderAuthOperationHandler(req *restful.Request, resp *restful.Response) {
	x_auth_token := req.HeaderParameter("X-Auth-Token")
	sgoId := req.PathParameter(ParamID)
	clusterId := req.QueryParameter("cluster_id")

	operations, code, err := services.GetClusterOrderService().GetAuthOperations(x_auth_token, clusterId, sgoId)
	if err != nil {
		response.WriteStatusError(code, err, resp)
		return
	}

	res := response.QueryStruct{Success: true, Data: operations}
	resp.WriteEntity(res)
	return
}

func (p *Resource) ClusterOrderAppHandler(req *restful.Request, resp *restful.Response) {
	x_auth_token := req.HeaderParameter("X-Auth-Token")

	appId := req.QueryParameter("appId")
	sgoId := req.PathParameter(ParamID)
	clusterId := req.QueryParameter("cluster_id")
	logrus.Debugf("get app %v info from oreder %v, clusterId %v ", appId, sgoId, clusterId)

	app, code, err := services.GetClusterOrderService().GetAppInOrder(x_auth_token, clusterId, sgoId, appId)
	if err != nil {
		response.WriteStatusError(code, err, resp)
		return
	}

	response.WriteResponse(app, resp)
	return
}

func (p *Resource) ClusterOrderAppScaleHandler(req *restful.Request, resp *restful.Response) {
	x_auth_token := req.HeaderParameter("X-Auth-Token")

	sgoId := req.PathParameter(ParamID)
	appId := req.QueryParameter("appId")
	numStr := req.QueryParameter("num")
	clusterId := req.QueryParameter("cluster_id")
	code, err := services.GetClusterOrderService().ScaleAppByOrderId(x_auth_token, clusterId, sgoId, appId, numStr)
	if err != nil {
		response.WriteStatusError(code, err, resp)
		return
	}

	response.WriteSuccess(resp)
	return
}

func (p *Resource) ClusterOrderInstanceGetHandler(req *restful.Request, resp *restful.Response) {
	x_auth_token := req.HeaderParameter("X-Auth-Token")

	sgiId := req.PathParameter(ParamID)
	clusterId := req.QueryParameter("cluster_id")
	sgi, code, err := services.GetClusterOrderService().GetOrderInstance(x_auth_token, clusterId, sgiId)
	if err != nil {
		response.WriteStatusError(code, err, resp)
		return
	}

	res := response.QueryStruct{Success: true, Data: sgi}
	resp.WriteEntity(res)
	return
}
