package documents

import (
	"encoding/json"
	"linkernetworks.com/linker_cluster/services"
	"linkernetworks.com/linker_common_lib/rest/response"
	"linkernetworks.com/linker_common_lib/persistence/entity"
	"github.com/emicklei/go-restful"
	"github.com/Sirupsen/logrus"
)

func (p Resource) FlavorWebService() *restful.WebService {
	ws := new(restful.WebService)
	ws.Path("/v1/flavor")
	ws.Consumes("*/*")
	ws.Produces(restful.MIME_JSON)
	
	id := ws.PathParameter(ParamID, "Storage identifier of flavor")
	paramID := "{" + ParamID + "}"

	ws.Route(ws.POST("/").To(p.FlavorCreateHandler).
		Doc("Store a flavor").
		Operation("FlavorCreateHandler").
		Param(ws.HeaderParameter("X-Auth-Token", "A valid authentication token")).
		Param(ws.BodyParameter("body", "").DataType("string")))
	
	ws.Route(ws.GET("/").To(p.FlavorListHandler).
		Doc("Store a flavor").
		Operation("FlavorListHandler").
		Param(ws.HeaderParameter("X-Auth-Token", "A valid authentication token")).
		Param(ws.QueryParameter("supplier", "The name of supplier wanted to query")))
		
	ws.Route(ws.PUT("/" + paramID).To(p.FlavorUpdateHandler).
		Doc("Update a cluster by its storage identifier").
		Operation("FlavorUpdateHandler").
		Param(ws.HeaderParameter("X-Auth-Token", "A valid authentication token")).
		Param(id).
		Param(ws.BodyParameter("body", "").DataType("string")))


	return ws
}

func (p *Resource) FlavorUpdateHandler(req *restful.Request, resp *restful.Response) {
	x_auth_token := req.HeaderParameter("X-Auth-Token")
	code, err := services.TokenValidation(x_auth_token)
	if err != nil {
		response.WriteStatusError(code, err, resp)
		return
	}
	
	objectId := req.PathParameter(ParamID)
	
	// Stub an flavor to be populated from the body
	flavor := entity.Flavor{}

	// Populate the user data
	err = json.NewDecoder(req.Request.Body).Decode(&flavor)
	if err != nil {
		logrus.Errorf("convert body to flavor failed, error is %v", err)
		response.WriteStatusError(services.COMMON_ERROR_INVALIDATE, err, resp)
		return
	}
	
	created, code, err := services.GetFlavorService().UpdateById(objectId, flavor, x_auth_token)
	if err != nil {
		response.WriteStatusError(code, err, resp)
		return
	}
	p.successUpdate(objectId, created, req, resp)
	
}

func (p *Resource) FlavorCreateHandler(req *restful.Request, resp *restful.Response) {
	x_auth_token := req.HeaderParameter("X-Auth-Token")
	code, err := services.TokenValidation(x_auth_token)
	if err != nil {
		response.WriteStatusError(code, err, resp)
		return
	}
	
	flavor := entity.Flavor{}
	err = json.NewDecoder(req.Request.Body).Decode(&flavor)
	if err != nil {
		logrus.Errorf("convert body to flavor failed, error is %v", err)
		response.WriteStatusError(services.COMMON_ERROR_INVALIDATE, err, resp)
		return
	}
	
	newFlavor, code, err := services.GetFlavorService().Create(flavor, x_auth_token)
	if err != nil {
		response.WriteStatusError(code, err, resp)
		return
	}

	res := response.QueryStruct{Success: true, Data: newFlavor}
	resp.WriteEntity(res)
	return
}

func (p *Resource) FlavorListHandler(req *restful.Request, resp *restful.Response) {
	
	x_auth_token := req.HeaderParameter("X-Auth-Token")
	code, err := services.TokenValidation(x_auth_token)
	if err != nil {
		response.WriteStatusError(code, err, resp)
		return
	}

	var name string = req.QueryParameter("supplier")
	_, flavors, code, err := services.GetFlavorService().QueryAllByName(name, 0, 0, x_auth_token)
	if err != nil {
		response.WriteStatusError(code, err, resp)
		return
	}

	res := response.QueryStruct{Success: true, Data: flavors}
	resp.WriteEntity(res)
	return

}
