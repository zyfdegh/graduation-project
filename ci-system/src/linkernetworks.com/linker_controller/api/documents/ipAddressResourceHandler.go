package documents

import (
	"encoding/json"
	"github.com/Sirupsen/logrus"
	"github.com/emicklei/go-restful"

	"linkernetworks.com/linker_common_lib/persistence/entity"
	"linkernetworks.com/linker_common_lib/rest/response"
	"linkernetworks.com/linker_controller/services"

	"gopkg.in/mgo.v2/bson"
	"strconv"
)

var ipCollection = "ipAddressResource"

func (p Resource) IPPoolWebService() *restful.WebService {
	ws := new(restful.WebService)
	ws.Path("/v1/ippool")
	ws.Consumes("*/*")
	ws.Produces(restful.MIME_JSON)

	ws.Route(ws.POST("/").To(p.IpPoolAddHandler).
		Doc("Store an ippool").
		Operation("IpPoolAddHandler").
		Param(ws.HeaderParameter("X-Auth-Token", "A valid authentication token")).
		Param(ws.BodyParameter("body", "Ippool body in json format,for example {\"subnet\":\"...\",  \"gateway\":\"...\",  \"pool_name\":\"...\",  \"ip_resources\": [{\"ipAddress\":\"...\", \"allocated\":\"...\"}]}").DataType("string")))

	return ws
}

func (p Resource) IPWebService() *restful.WebService {
	ws := new(restful.WebService)
	ws.Path("/v1/ips")
	ws.Consumes("*/*")
	ws.Produces(restful.MIME_JSON)

	id := ws.PathParameter(ParamID, "Storage identifier of ip")
	paramID := "{" + ParamID + "}"

	ws.Route(ws.GET("/").To(p.IpListHandler).
		Doc("Return all ip items").
		Operation("IpListHandler").
		Param(ws.HeaderParameter("X-Auth-Token", "A valid authentication token")).
		Param(ws.QueryParameter("count", "Count total items and return the result in X-Object-Count header").DataType("boolean")).
		//		Param(ws.QueryParameter("query", "Query in json format")).
		//		Param(ws.QueryParameter("fields", "Comma separated list of field names")).
		Param(ws.QueryParameter("skip", "Number of items to skip in the result set, default=0")).
		Param(ws.QueryParameter("limit", "Maximum number of items in the result set, default=0")))
	//		Param(ws.QueryParameter("sort", "Comma separated list of field names to sort"))

	ws.Route(ws.GET("/" + paramID).To(p.IpListHandler).
		Doc("Return an ip by its storage identifier").
		Operation("IpListHandler").
		Param(ws.HeaderParameter("X-Auth-Token", "A valid authentication token")).
		Param(id).
		Param(ws.QueryParameter("fields", "Comma separated list of field names")))

	ws.Route(ws.POST("").To(p.IpCreateHandler).
		Doc("Store an ip").
		Operation("IpCreateHandler").
		Param(ws.HeaderParameter("X-Auth-Token", "A valid authentication token")).
		Param(ws.BodyParameter("body", "Ip body in json format,for example {\"ipAddress\":\"...\",  \"allocated\":\"...\"}").DataType("string")))

	ws.Route(ws.PUT("/" + paramID).To(p.IpUpdateHandler).
		Doc("Update an ip by its storage identifier").
		Operation("IpUpdateHandler").
		Param(ws.HeaderParameter("X-Auth-Token", "A valid authentication token")).
		Param(id).
		Param(ws.BodyParameter("body", "Ip body in json format,for example {\"ipAddress\":\"...\",  \"allocated\":\"...\"}").DataType("string")))

	ws.Route(ws.DELETE("/" + paramID).To(p.IpDeleteHandler).
		Doc("Delete an ip by its storage identifier").
		Operation("IpDeleteHandler").
		Param(ws.HeaderParameter("X-Auth-Token", "A valid authentication token")).
		Param(id))

	ws.Route(ws.DELETE("/").To(p.IpDeleteHandler).
		Doc("Delete ip if query present, otherwise remove the entire ips ").
		Operation("IpDeleteHandler").
		Param(ws.HeaderParameter("X-Auth-Token", "A valid authentication token")).
		Param(ws.QueryParameter("count", "Count total items and return the result in X-Object-Count header").DataType("boolean")).
		//		Param(ws.QueryParameter("query", "Query in json format")).
		//		Param(ws.QueryParameter("fields", "Comma separated list of field names")).
		Param(ws.QueryParameter("skip", "Number of items to skip in the result set, default=0")).
		Param(ws.QueryParameter("limit", "Maximum number of items in the result set, default=0")))
	//		Param(ws.QueryParameter("sort", "Comma separated list of field names to sort"))

	return ws
}

// IpListHandler parses the http request and return the ip items.
// Usage :
//		GET /v1/ips
//     GET /v1/ips/ParamID
// Params :
//		ParamID : storage identifier of ip
// If successful,response code will be set to 201.
func (p *Resource) IpListHandler(req *restful.Request, resp *restful.Response) {
	logrus.Infof("IpListHandler is called!")

	x_auth_token := req.HeaderParameter("X-Auth-Token")
	code, err := services.TokenValidation(x_auth_token)
	if err != nil {
		response.WriteStatusError(code, err, resp)
		return
	}

	var skip int = queryIntParam(req, "skip", 0)
	var limit int = queryIntParam(req, "limit", 0)

	total, ips, code, err := services.GetIPResourceService().QueryAll(skip, limit, x_auth_token)
	if err != nil {
		response.WriteStatusError(code, err, resp)
		return
	}
	res := response.QueryStruct{Success: true, Data: ips}
	if c, _ := strconv.ParseBool(req.QueryParameter("count")); c {
		res.Count = total
		resp.AddHeader("X-Object-Count", strconv.Itoa(total))
	}
	resp.WriteEntity(res)
	return
}

// IpDeleteHandler parses the http request and delete the ip items.
// Usage :
//		DELETE /v1/ips
//     DELETE /v1/ips/{ParamID}
// Params :
//		ParamID : storage identifier of ip
// If successful,response code will be set to 201.
func (p *Resource) IpDeleteHandler(req *restful.Request, resp *restful.Response) {
	logrus.Infof("IpDeleteHandler is called!")
	x_auth_token := req.HeaderParameter("X-Auth-Token")
	code, err := services.TokenValidation(x_auth_token)
	if err != nil {
		response.WriteStatusError(code, err, resp)
		return
	}

	objectId := req.PathParameter(ParamID)

	code, err = services.GetIPResourceService().DeleteById(objectId, x_auth_token)

	if err != nil {
		response.WriteStatusError(code, err, resp)
		return
	}
	// Write success response
	response.WriteSuccess(resp)
	return
}

// IpCreateHandler parses the http request and store an ip.
// Usage :
//		POST /v1/ips
// If successful,response code will be set to 201.
func (p *Resource) IpCreateHandler(req *restful.Request, resp *restful.Response) {
	x_auth_token := req.HeaderParameter("X-Auth-Token")
	code, err := services.TokenValidation(x_auth_token)
	if err != nil {
		response.WriteStatusError(code, err, resp)
		return
	}

	// Stub an app to be populated from the body
	ipAddressResource := entity.IpAddressResource{}

	// Populate the user data
	err = json.NewDecoder(req.Request.Body).Decode(&ipAddressResource)
	if err != nil {
		logrus.Errorf("convert body to app failed, error is %v", err)
		response.WriteStatusError(services.COMMON_ERROR_INVALIDATE, err, resp)
		return
	}

	newIPAddressResource, code, err := services.GetIPResourceService().Create(ipAddressResource, x_auth_token)
	if err != nil {
		response.WriteStatusError(code, err, resp)
		return
	}

	res := response.Response{Success: true, Data: newIPAddressResource}
	resp.WriteEntity(res)
	return

}

// IpUpdateHandler parses the http request and update an ip.
// Usage :
//		PUT /v1/ips
// If successful,response code will be set to 201.
func (p *Resource) IpUpdateHandler(req *restful.Request, resp *restful.Response) {
	x_auth_token := req.HeaderParameter("X-Auth-Token")
	code, err := services.TokenValidation(x_auth_token)
	if err != nil {
		response.WriteStatusError(code, err, resp)
		return
	}

	objectId := req.PathParameter(ParamID)

	// Stub an app to be populated from the body
	ipAddressResource := entity.IpAddressResource{}

	// Populate the user data
	err = json.NewDecoder(req.Request.Body).Decode(&ipAddressResource)
	if err != nil {
		logrus.Errorf("convert body to ipAddressResource failed, error is %v", err)
		response.WriteStatusError(services.COMMON_ERROR_INVALIDATE, err, resp)
		return
	}

	created, code, err := services.GetIPResourceService().UpdateById(objectId, ipAddressResource, x_auth_token)
	if err != nil {
		response.WriteStatusError(code, err, resp)
		return
	}
	p.successUpdate(objectId, created, req, resp)
}

// IpPoolAddHandler parses the http request and store an ippool.
// Usage :
//		POST /v1/ippool
// If successful,response code will be set to 201.
func (p *Resource) IpPoolAddHandler(req *restful.Request, resp *restful.Response) {
	x_auth_token := req.HeaderParameter("X-Auth-Token")
	code, err := services.TokenValidation(x_auth_token)
	if err != nil {
		response.WriteStatusError(code, err, resp)
		return
	}

	// Read a document from request
	document := bson.M{}
	// Handle JSON parsing manually here, instead of relying on go-restful's
	// req.ReadEntity. This is because ReadEntity currently parses JSON with
	// UseNumber() which turns all numbers into strings. See:
	// https://github.com/emicklei/mora/pull/31
	decoder := json.NewDecoder(req.Request.Body)
	err = decoder.Decode(&document)
	if err != nil {
		logrus.Errorf("decode ip pool err is %v", err)
		response.WriteError(err, resp)
		return
	}

	var pool *entity.IpAddressPool
	pool = new(entity.IpAddressPool)
	poolout, err := json.Marshal(document)
	if err != nil {
		logrus.Errorf("marshal pool err is %v", err)
		response.WriteError(err, resp)
		return
	}
	err = json.Unmarshal(poolout, &pool)
	if err != nil {
		logrus.Errorf("unmarshal pool err is %v", err)
		response.WriteError(err, resp)
		return
	}
	newIPAddressResource, code, err := services.GetIPResourceService().CreatePool(pool, x_auth_token)
	if err != nil {
		response.WriteStatusError(code, err, resp)
		return
	}

	res := response.Response{Success: true, Data: newIPAddressResource}
	resp.WriteEntity(res)
	return
}
