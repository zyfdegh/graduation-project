package documents

import (
	"encoding/json"
	"github.com/compose/mejson"
	"github.com/emicklei/go-restful"
	"gopkg.in/mgo.v2/bson"
	// "linkernetworks.com/linker_common_lib/persistence/entity"
	"linkernetworks.com/linker_common_lib/rest/response"
	"net/http"
	"net/url"
	"strconv"
	"strings"
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

func (p *Resource) prevnexturl(req *restful.Request) (prev string, next string) {
	return p.collectionurl(false, req), p.collectionurl(true, req)
}

func (p *Resource) collectionurl(next bool, req *restful.Request) string {
	// Get current location url
	uri, _ := url.Parse(req.Request.URL.RequestURI())
	q := uri.Query()

	// Skip/limit values
	limitnum := queryIntParam(req, "limit", 10)
	skipnum := queryIntParam(req, "skip", 0)

	// Number of documents to skip
	if next {
		q.Set("skip", strconv.Itoa(skipnum+limitnum))
	} else {
		// prev
		prevskip := skipnum - limitnum
		if prevskip < 0 {
			prevskip = 0
		}
		if prevskip == skipnum {
			return ""
		}
		q.Set("skip", strconv.Itoa(prevskip))
	}

	// URL query
	uri.RawQuery = q.Encode()

	return uri.String()
}

// func (d *Resource) GetMongoCollection(collName string, session *mgo.Session) *mgo.Collection {
// 	return session.DB(getParam("database", req)).C(req.PathParameter("collection"))
// 	return session.DB("").C(collName)
// }

func getFields(req *restful.Request) bson.M {
	selector := bson.M{}
	fields := req.QueryParameter("fields")
	if len(fields) > 0 {
		for _, v := range strings.Split(fields, ",") {
			selector[v] = 1
		}
	}
	return selector
}

func buildSpecFields() bson.M {
	selector := bson.M{}
	fieldList := []string{"user_id", "tenant_id", "state"}

	for i := 0; i < len(fieldList); i++ {
		selector[fieldList[i]] = 1
	}

	return selector
}

//
// Composes a mongo selector from request
// If _id in the path is present `one` is true and query parameter is not inclued.
//
// Param(ws.PathParameter(ParamID, "query in json format")).
// Param(ws.QueryParameter("query", "query in json format")).
func getSelector(req *restful.Request) (selector bson.M, one bool, err error) {
	selector = make(bson.M)
	// If id is included in path, dont include query
	// It only select's one item
	if id := req.PathParameter(ParamID); id != "" {
		selector[ParamID] = id
	} else {
		// Unmarshal json query if any
		if query := req.QueryParameter("query"); len(query) > 0 {
			query, err = url.QueryUnescape(query)
			if err != nil {
				return
			}
			err = json.Unmarshal([]byte(query), &selector)
			if err != nil {
				return
			}
			selector, err = mejson.Unmarshal(selector)
			if err != nil {
				return
			}
		}
	}

	// Transform string HexId to ObjectIdHex
	if selid, _ := selector[ParamID].(string); selid != "" {
		// Transform to ObjectIdHex if required
		if bson.IsObjectIdHex(selid) {
			selector[ParamID] = bson.ObjectIdHex(selid)
		} else {
			selector[ParamID] = selid
		}
		one = true
	}

	return
}

// Returns a string parameter from request path or req.Attributes
func getParam(name string, req *restful.Request) (param string) {
	// Get parameter from request path
	param = req.PathParameter(name)
	if param != "" {
		return param
	}

	// Get parameter from request attributes (set by intermediates)
	attr := req.Attribute(name)
	if attr != nil {
		param, _ = attr.(string)
	}
	return
}

func queryIntParam(req *restful.Request, name string, def int) int {
	num := def
	if strnum := req.QueryParameter(name); len(strnum) > 0 {
		num, _ = strconv.Atoi(strnum)
	}
	return num
}
