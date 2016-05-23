package documents

import (
	"encoding/json"
	"errors"
	"strconv"

	"github.com/Sirupsen/logrus"
	"github.com/emicklei/go-restful"
	"linkernetworks.com/linker_common_lib/persistence/entity"
	"linkernetworks.com/linker_common_lib/rest/response"
	"linkernetworks.com/linker_controller/services"
)

func (p Resource) BillingWebService() *restful.WebService {
	ws := new(restful.WebService)
	ws.Path("/v1/billing")
	ws.Consumes("*/*")
	ws.Produces(restful.MIME_JSON)

	id := ws.PathParameter(ParamID, "Storage identifier of billing model")
	paramID := "{" + ParamID + "}"

	ws.Route(ws.GET("/").To(p.BillingModelListHandler).
		Doc("Return all billing model items belongs to the user.").
		Operation("BillingModelListHandler").
		Param(ws.HeaderParameter("X-Auth-Token", "A valid authentication token")).
		Param(ws.QueryParameter("count", "Count total items and return the result in X-Object-count header").DataType("boolean")).
		// Param(ws.QueryParameter("query", "Query in json format")).
		// Param(ws.QueryParameter("fields", "Comma separated list of field names")).
		Param(ws.QueryParameter("skip", "Number of items to skip in the result set, default=0")).
		Param(ws.QueryParameter("limit", "Maximum number of items in the result set, default=0")).
		Param(ws.QueryParameter("sort", "Comma separated list of field names to sort")))

	ws.Route(ws.GET("/" + paramID).To(p.BillingModelDetailHandler).
		Doc("Return a billing model by its storage identifier").
		Operation("BillingModelDetailHandler").
		Param(ws.HeaderParameter("X-Auth-Token", "A valid authentication token")).
		Param(id))

	ws.Route(ws.GET("/all").To(p.BillingModelListAllHandler).
		Doc("Return all billing models.").
		Operation("BillingModelListAllHandler").
		Param(ws.QueryParameter("count", "Count total items and return the result in X-Object-count header").DataType("boolean")).
		// Param(ws.QueryParameter("fields", "Comma separated list of field names")).
		Param(ws.QueryParameter("skip", "Number of items to skip in the result set, default=0")).
		Param(ws.QueryParameter("limit", "Maximum number of items in the result set, default=0")).
		Param(ws.QueryParameter("sort", "Comma separated list of field names to sort")))

	ws.Route(ws.POST("/").To(p.BillingModelCreateHandler).
		Doc("Store a billing model").
		Operation("BillingModelCreateHandler").
		Param(ws.HeaderParameter("X-Auth-Token", "A valid authentication token")).
		Param(ws.BodyParameter("body", "Billing model body in json format,for example {\"totalprice\": , \"refs\":[], \"price\": , \"modelid\":\"...\"}").DataType("string")))

	ws.Route(ws.PUT("/" + paramID).To(p.BillingModelUpdateHandler).
		Doc("Update a billing model by its storage identifier").
		Operation("BillingModelUpdateHandler").
		Param(ws.HeaderParameter("X-Auth-Token", "A valid authentication token")).
		Param(id).
		Param(ws.BodyParameter("body", "Billing model body in json format,for example {\"totalprice\": , \"refs\":[], \"price\": , \"modelid\":\"...\"}").DataType("string")))

	return ws
}

func (p *Resource) BillingModelListHandler(req *restful.Request, resp *restful.Response) {
	logrus.Infof("BillingModelListHandler is called!")

	token := req.HeaderParameter("X-Auth-Token")
	limit := queryIntParam(req, "limit", 0)
	skip := queryIntParam(req, "skip", 0)
	sort := req.QueryParameter("sort")

	bms, total, errorCode, err := services.GetBillingService().QueryBillingModels(token, skip, limit, sort)
	if err != nil {
		response.WriteStatusError(errorCode, err, resp)
		return
	}
	res := response.QueryStruct{Success: true, Data: bms}
	if c, _ := strconv.ParseBool(req.QueryParameter("count")); c {
		res.Count = total
		resp.AddHeader("X-Object-Count", strconv.Itoa(total))
	}
	resp.WriteEntity(res)
	return
}

// BillingModelListAllHandler parses the http request and return all billing models.
// Usage :
//		GET /v1/billing/all
// If successful,response code will be set to 201.
func (p *Resource) BillingModelListAllHandler(req *restful.Request, resp *restful.Response) {
	logrus.Infoln("BillingModelListAllHandler is called!")
	var skip int = queryIntParam(req, "skip", 0)
	var limit int = queryIntParam(req, "limit", 0)
	sort := req.QueryParameter("sort")

	total, bms, errorCode, err := services.GetBillingService().QueryAllBillingModels(skip, limit, sort)
	if err != nil {
		response.WriteStatusError(errorCode, err, resp)
		return
	}
	res := response.QueryStruct{Success: true, Data: bms}
	if c, _ := strconv.ParseBool(req.QueryParameter("count")); c {
		res.Count = total
		resp.AddHeader("X-Object-Count", strconv.Itoa(total))
	}
	resp.WriteEntity(res)
	return
}

func (p *Resource) BillingModelDetailHandler(req *restful.Request, resp *restful.Response) {
	logrus.Infof("BillingModelListHandler is called!")
	token := req.HeaderParameter("X-Auth-Token")
	id := req.PathParameter(ParamID)
	if len(id) <= 0 {
		logrus.Warnln("billing id should not be null for get detail operation")
		response.WriteStatusError("E11040", errors.New("billing id should not be null for get billing detail operation"), resp)
		return
	}

	bm, errorCode, err := services.GetBillingService().QueryBillingModelById(token, id)
	if err != nil {
		response.WriteStatusError(errorCode, err, resp)
		return
	}

	res := response.QueryStruct{Success: true, Data: bm}
	resp.WriteEntity(res)
	return
}

// BillingModelListHandler parses the http request and return a billing model.
// Usage :
//		GET /v1/billing
//		GET /v1/billing/{ParamID}
// Params :
//		ParamID : Storage identifier of billing model
// If successful,response code will be set to 201.
// func (p *Resource) BillingModelListHandler(req *restful.Request, resp *restful.Response) {
// 	logrus.Infof("BillingModelListHandler is called!")
// 	token := req.HeaderParameter("X-Auth-Token")
// 	code, err := p.TokenValidation(token)
// 	if err != nil {
// 		response.WriteStatusError(code, err, resp)
// 		return
// 	}

// 	p.handleList(BILLINGCOLLECTION, "list_billing", req, resp, false)
// }

// BillingModelCreateHandler parses the http request and store a billing model.
// Usage :
//		POST /v1/billing
// If successful,response code will be set to 201.
func (p *Resource) BillingModelCreateHandler(req *restful.Request, resp *restful.Response) {
	logrus.Infof("BillingModelCreateHandler is called!")
	token := req.HeaderParameter("X-Auth-Token")

	// Read a document from request
	document := new(entity.BillingModel)
	decoder := json.NewDecoder(req.Request.Body)
	err := decoder.Decode(document)
	if err != nil {
		logrus.Errorf("decode billing model err is %v", err)
		response.WriteStatusError("E11011", err, resp)
		return
	}

	id, errorCode, err := services.GetBillingService().CreateBillingModel(document, token)
	if err != nil {
		response.WriteStatusError(errorCode, err, resp)
		return
	}

	res := response.QueryStruct{Success: true, Data: id}
	resp.WriteEntity(res)
	return
}

// BillingModelUpdateHandler parses the http request and update a billing model.
// Usage :
//		PUT /v1/billing/{ParamID}
// Params :
//		ParamID : Storage identifier of billing model
// If successful,response code will be set to 201.
func (p *Resource) BillingModelUpdateHandler(req *restful.Request, resp *restful.Response) {
	logrus.Infof("BillingModelUpdateHandler is called!")
	token := req.HeaderParameter("X-Auth-Token")
	id := req.PathParameter(ParamID)
	if len(id) <= 0 {
		logrus.Warnln("billing model id should not be null for update operation")
		response.WriteStatusError("E11012", errors.New("billingmodel id should not be null for update operation"), resp)
		return
	}

	// Read a document from request
	document := new(entity.BillingModel)
	decoder := json.NewDecoder(req.Request.Body)
	err := decoder.Decode(document)
	if err != nil {
		logrus.Errorf("decode billing model err is %v", err)
		response.WriteStatusError("E11012", err, resp)
		return
	}

	bmid, errorCode, err := services.GetBillingService().UpdateBillingModel(document, token, id)
	if err != nil {
		response.WriteStatusError(errorCode, err, resp)
		return
	}

	res := response.QueryStruct{Success: true, Data: bmid}
	resp.WriteEntity(res)
	return

}
