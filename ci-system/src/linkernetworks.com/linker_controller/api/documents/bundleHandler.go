package documents

import (
	"encoding/json"
	"errors"

	"github.com/Sirupsen/logrus"
	"github.com/emicklei/go-restful"
	"linkernetworks.com/linker_common_lib/persistence/entity"
	"linkernetworks.com/linker_common_lib/rest/response"
	"linkernetworks.com/linker_controller/services"
)

func (p Resource) BundleWebService() *restful.WebService {
	ws := new(restful.WebService)
	ws.Path("/v1/bundle")
	ws.Consumes("*/*")
	ws.Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/").To(p.BundleExportlHandler).
		Doc("Export a bundle by service group storage identifier").
		Operation("BundleExportlHandler").
		Param(ws.HeaderParameter("X-Auth-Token", "A valid authentication token")).
		Param(ws.QueryParameter("sgid", "An valid service group id")))

	ws.Route(ws.POST("/").To(p.BundleImportHandler).
		Doc("Import a bundle").
		Operation("BundleImportHandler").
		Param(ws.HeaderParameter("X-Auth-Token", "A valid authentication token")).
		// Param(ws.QueryParameter("controllerURL", "The destination controller url. for example: 192.168.120.51:8081")).
		Param(ws.BodyParameter("body", "Bundle body in json format,for example {\"servicegroup\": .., \"configrationpackages\": []}").DataType("string")))

	return ws
}

func (p *Resource) BundleExportlHandler(req *restful.Request, resp *restful.Response) {
	logrus.Infoln("BundleExportlHandler is called!")
	token := req.HeaderParameter("X-Auth-Token")
	sgid := req.QueryParameter("sgid")

	if len(sgid) <= 0 {
		logrus.Errorln("invalid request! id parameter should not be null!")
		response.WriteStatusError("E12002", errors.New("invalid request! id parameter should not be null!"), resp)
	}

	bundle, errorCode, err := services.GetBundleService().ExportBundle(token, sgid)
	if err != nil {
		response.WriteStatusError(errorCode, err, resp)
		return
	}

	res := response.QueryStruct{Success: true, Data: bundle}
	resp.WriteEntity(res)
	return
}

func (p *Resource) BundleImportHandler(req *restful.Request, resp *restful.Response) {
	logrus.Infoln("BundleImportHandler is called!")
	token := req.HeaderParameter("X-Auth-Token")

	bundle := entity.Bundle{}

	// Populate the user data
	err := json.NewDecoder(req.Request.Body).Decode(&bundle)
	if err != nil {
		logrus.Errorf("convert body to bundle failed, error is %v", err)
		response.WriteStatusError(services.COMMON_ERROR_INVALIDATE, err, resp)
		return
	}

	errorCode, err := services.GetBundleService().ImportBundle(token, bundle)
	if err != nil {
		response.WriteStatusError(errorCode, err, resp)
		return
	}

	response.WriteSuccess(resp)
	return
}
