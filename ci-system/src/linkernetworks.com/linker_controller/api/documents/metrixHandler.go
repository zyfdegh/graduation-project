package documents

import (
	"github.com/Sirupsen/logrus"
	"github.com/emicklei/go-restful"
	"linkernetworks.com/linker_common_lib/rest/response"
	"linkernetworks.com/linker_controller/services"
)

func (p Resource) MetrixWebService() *restful.WebService {
	ws := new(restful.WebService)
	ws.Path("/v1/metrix")
	ws.Consumes("*/*")
	ws.Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/").To(p.MetrixListHandler).
		Doc("Return metrix data of the user.").
		Operation("MetrixServiceInstanceListHandler").
		Param(ws.QueryParameter("category", "Meterix category, supported 'serviceInstances', 'resources'")).
		Param(ws.HeaderParameter("X-Auth-Token", "A valid authentication token")))
	return ws
}

// MetrixListHandler parses the http request and return metrix data of the user.
// Usage :
//		GET /v1/metrix
// If successful,response code will be set to 201.
func (p *Resource) MetrixListHandler(req *restful.Request, resp *restful.Response) {
	logrus.Infof("MetrixListHandler is called!")
	x_auth_token := req.HeaderParameter("X-Auth-Token")
	code, err := services.TokenValidation(x_auth_token)
	if err != nil {
		response.WriteStatusError(code, err, resp)
		return
	}

	category := req.QueryParameter("category")
	isSg, sgMetrix, proMetrix, code, err := services.GetMetrixService().GetMetrix(category, x_auth_token)
	if err != nil {
		response.WriteStatusError(code, err, resp)
	}

	if isSg {
		response.WriteResponse(sgMetrix, resp)
	} else {
		response.WriteResponse(proMetrix, resp)
	}

}
