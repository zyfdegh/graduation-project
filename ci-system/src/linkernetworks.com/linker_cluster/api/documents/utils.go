package documents

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/emicklei/go-restful"
	"linkernetworks.com/linker_common_lib/rest/response"
)

type Resource struct {
}

func (p *Resource) successUpdate(id string, created bool,
	req *restful.Request, resp *restful.Response) {
	// Updated document API location
	docpath := p.documentLocation(req, id)

	// Content-Location header
	resp.AddHeader("Content-Location", docpath)

	// Information about updated document
	info := response.UpdateStruct{created, docpath}

	if created {
		response.WriteResponseStatus(http.StatusCreated, info, resp)
	} else {
		response.WriteResponse(info, resp)
	}
}

//
// Return document location URL
//
func (p *Resource) documentLocation(req *restful.Request, id string) (location string) {
	// Get current location url
	location = strings.TrimRight(req.Request.URL.RequestURI(), "/")

	// Remove id from current location url if any
	if reqId := req.PathParameter(ParamID); reqId != "" {
		idlen := len(reqId)
		// If id is in current location remove it
		if noid := len(location) - idlen; noid > 0 {
			if id := location[noid : noid+idlen]; id == reqId {
				location = location[:noid]
			}
		}
		location = strings.TrimRight(location, "/")
	}

	// Add id of the document
	return location + "/" + id
}

func queryIntParam(req *restful.Request, name string, def int) int {
	num := def
	if strnum := req.QueryParameter(name); len(strnum) > 0 {
		num, _ = strconv.Atoi(strnum)
	}
	return num
}
