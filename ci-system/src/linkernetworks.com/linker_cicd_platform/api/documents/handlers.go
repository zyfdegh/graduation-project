package documents

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/compose/mejson"
	"github.com/emicklei/go-restful"
	"github.com/jmoiron/jsonq"
	"github.com/magiconair/properties"
	"gopkg.in/mgo.v2/bson"
	"linkernetworks.com/linker_cicd_platform/api/response"
	"linkernetworks.com/linker_cicd_platform/persistence/dao"
	"linkernetworks.com/linker_cicd_platform/util"
)

type Resource struct {
	Props *properties.Properties
	Dao   *dao.Dao
	Util  *linker_util.Util
}

type QueryStruct struct {
	Success bool        `json:"success"`
	Count   interface{} `json:"count,omitempty"`
	Prev    string      `json:"prev_url,omitempty"`
	Next    string      `json:"next_url,omitempty"`
	Data    interface{} `json:"data"`
}

type UpdateStruct struct {
	Created bool   `json:"created"`
	Url     string `json:"url"`
}

func (d *Resource) WelcomeHandler(req *restful.Request, resp *restful.Response) {
	// Write response back to client
	response.WriteResponse("Welcome", resp)
}

func (d *Resource) handleList(collectionName string, operation string, req *restful.Request, resp *restful.Response) {
	// Get selector from `_id` path parameter and `query` query parameter
	selector, one, err := getSelector(req)
	if err != nil {
		return
	}
	// parse request
	var fields bson.M = getFields(req)
	var skip int = queryIntParam(req, "skip", 0)
	var limit int = queryIntParam(req, "limit", 0)
	var sort string = req.QueryParameter("sort")
	var extended_json = req.QueryParameter("extended_json")

	if one {
		_, _, jsonDocument, err := d.Dao.HandleQuery(collectionName, selector, one, fields, skip, limit, sort, extended_json)

		if err != nil {
			logrus.Errorf("handle query err is %v", err)
			response.WriteError(response.ErrDBQuery, resp)
			return
		}

		response.WriteResponse(jsonDocument, resp)
		return
	}

	_, lenth, jsonDocuments, err := d.Dao.HandleQuery(collectionName, selector, one, fields, skip, limit, sort, extended_json)
	if err != nil {
		logrus.Errorf("handle query err is %v", err)
		response.WriteError(response.ErrDBQuery, resp)
		return
	}

	//currently add auth check here! TODO: need refactor
	var res QueryStruct
	var filterCount int
	res = QueryStruct{Success: true, Data: jsonDocuments}

	// Get limit amount
	limitnum := 10
	if limit > 0 {
		limitnum = limit
	}

	// If got full limit set next link
	if lenth == limitnum {
		res.Prev, res.Next = d.prevnexturl(req)
	}

	// Count documents if count parameter is included in query
	if c, _ := strconv.ParseBool(req.QueryParameter("count")); c {
		res.Count = filterCount
		resp.AddHeader("X-Object-Count", strconv.Itoa(filterCount))
	}

	// Write result back to client
	resp.WriteEntity(res)
}

func (d *Resource) handleListByUser(collectionName string, req *restful.Request, resp *restful.Response) {
	// Get selector from `_id` path parameter and `query` query parameter
	selector, one, err := getSelector(req)
	if err != nil {
		return
	}
	// parse request
	var fields bson.M = getFields(req)
	var skip int = queryIntParam(req, "skip", 0)
	var limit int = queryIntParam(req, "limit", 0)
	var sort string = req.QueryParameter("sort")
	var extended_json = req.QueryParameter("extended_json")

	var token string = req.HeaderParameter("X-Auth-Token")
	userId, tenantId, _, _ := d.getUserAndTenantId(token)
	selector["user_id"] = userId
	selector["tenant_id"] = tenantId

	if one {
		_, _, jsonDocument, err := d.Dao.HandleQuery(collectionName, selector, one, fields, skip, limit, sort, extended_json)

		if err != nil {
			logrus.Errorf("handle query err is %v", err)
			response.WriteError(response.ErrDBQuery, resp)
			return
		}

		response.WriteResponse(jsonDocument, resp)
		return
	}

	_, lenth, jsonDocuments, err := d.Dao.HandleQuery(collectionName, selector, one, fields, skip, limit, sort, extended_json)
	if err != nil {
		logrus.Errorf("handle query err is %v", err)
		response.WriteError(response.ErrDBQuery, resp)
		return
	}

	//currently add auth check here! TODO: need refactor
	var res QueryStruct
	var filterCount int
	res = QueryStruct{Success: true, Data: jsonDocuments}

	// Get limit amount
	limitnum := 10
	if limit > 0 {
		limitnum = limit
	}

	// If got full limit set next link
	if lenth == limitnum {
		res.Prev, res.Next = d.prevnexturl(req)
	}

	// Count documents if count parameter is included in query
	if c, _ := strconv.ParseBool(req.QueryParameter("count")); c {
		res.Count = filterCount
		resp.AddHeader("X-Object-Count", strconv.Itoa(filterCount))
	}

	// Write result back to client
	resp.WriteEntity(res)
}

func (d *Resource) successUpdate(id string, created bool, req *restful.Request, resp *restful.Response) {
	// Updated document API location
	docpath := d.documentLocation(req, id)

	// Content-Location header
	resp.AddHeader("Content-Location", docpath)

	// Information about updated document
	info := UpdateStruct{created, docpath}

	if created {
		response.WriteResponseStatus(http.StatusCreated, info, resp)
	} else {
		response.WriteResponse(info, resp)
	}
}

func (d *Resource) handleUpdate(collectionName string, one bool, selector, document bson.M, req *restful.Request, resp *restful.Response) {
	// Update document(/s)
	// var (
	// 	docid string
	// 	err   error
	// )

	// Write info about updated document
	if one {
		// docid, created, err := d.Dao.HandleUpdateById(collectionName, selector, document)
		docid, _, created, err := d.Dao.HandleUpdateById(collectionName, selector, document)
		if err != nil {
			logrus.Errorf("handle update by id err is %v", err)
			response.WriteError(response.ErrDBQuery, resp)
			return
		}
		d.successUpdate(docid, created, req, resp)
		return
	} else {
		err := d.Dao.HandleUpdateByQuery(collectionName, selector, document)
		if err != nil {
			logrus.Errorf("handle update by query err is %v", err)
			response.WriteError(response.ErrDBQuery, resp)
			return
		}
		// Write success response
		response.WriteSuccess(resp)
	}
}

func (d *Resource) handleInsert(collectionName string, selector, document bson.M, req *restful.Request, resp *restful.Response) {
	id, _, err := d.Dao.HandleInsert(collectionName, selector, document)

	if err != nil {
		logrus.Errorf("handle insert err is %v", err)
		response.WriteError(response.ErrDBInsert, resp)
		return
	}

	d.successUpdate(id, true, req, resp)
}

func (d *Resource) handleDelete(collectionName string, one bool, selector bson.M, req *restful.Request, resp *restful.Response) {
	err := d.Dao.HandleDelete(collectionName, one, selector)

	if err != nil {
		logrus.Errorf("handle delete err is %v", err)
		response.WriteError(response.ErrDBDelete, resp)
		return
	}

	// Write success response
	response.WriteSuccess(resp)
}

//
// Return document location URL
//
func (d *Resource) documentLocation(req *restful.Request, id string) (location string) {
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

func (d *Resource) prevnexturl(req *restful.Request) (prev string, next string) {
	return d.collectionurl(false, req), d.collectionurl(true, req)
}

func (d *Resource) collectionurl(next bool, req *restful.Request) string {
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

func (u *Resource) getUserAndTenantId(token string) (userId string, tenantId string, role string, email string) {
	// get controller url to send request.
	controllerUrl, err := u.Util.ZkClient.GetControllerEndpoint()
	if err != nil {
		logrus.Errorf("get controller endpoint err is %+v", err)
		return
	}
	url := strings.Join([]string{"http://", controllerUrl, "/v1/token/ids/"}, "")
	logrus.Debugln("get userid token url=" + url)

	// get userid & tetantid
	response, err := linker_util.Http_get(url, "application/json", token, "")
	if err != nil {
		logrus.Errorf("Get userid by token error %s", err.Error())
		return
	}

	// parse json data to get userid & tenantid
	jsondata := map[string]interface{}{}
	result := json.NewDecoder(strings.NewReader(response))
	result.Decode(&jsondata)

	jq := jsonq.NewQuery(jsondata)
	userId, _ = jq.String("data", "userid")
	tenantId, _ = jq.String("data", "tenantid")
	role, _ = jq.String("data", "role")
	email, _ = jq.String("data", "email")
	return
}

func (d *Resource) handleListByProjectId(collectionName string, req *restful.Request, resp *restful.Response) {
	// Get selector from `_id` path parameter and `query` query parameter
	selector, one, err := getSelector(req)
	if err != nil {
		return
	}

	// parse request
	var fields bson.M = getFields(req)
	var skip int = queryIntParam(req, "skip", 0)
	var limit int = queryIntParam(req, "limit", 0)
	var sort string = req.QueryParameter("sort")
	var extended_json = req.QueryParameter("extended_json")

	projectid := req.QueryParameter("projectid")
	selector["project"] = projectid

	if one {
		_, _, jsonDocument, err := d.Dao.HandleQuery(collectionName, selector, one, fields, skip, limit, sort, extended_json)

		if err != nil {
			logrus.Errorf("handle query err is %v", err)
			response.WriteError(response.ErrDBQuery, resp)
			return
		}

		response.WriteResponse(jsonDocument, resp)
		return
	}

	_, lenth, jsonDocuments, err := d.Dao.HandleQuery(collectionName, selector, one, fields, skip, limit, sort, extended_json)
	if err != nil {
		logrus.Errorf("handle query err is %v", err)
		response.WriteError(response.ErrDBQuery, resp)
		return
	}

	//currently add auth check here! TODO: need refactor
	var res QueryStruct
	var filterCount int
	res = QueryStruct{Success: true, Data: jsonDocuments}

	// Get limit amount
	limitnum := 10
	if limit > 0 {
		limitnum = limit
	}

	// If got full limit set next link
	if lenth == limitnum {
		res.Prev, res.Next = d.prevnexturl(req)
	}

	// Count documents if count parameter is included in query
	if c, _ := strconv.ParseBool(req.QueryParameter("count")); c {
		res.Count = filterCount
		resp.AddHeader("X-Object-Count", strconv.Itoa(filterCount))
	}

	// Write result back to client
	resp.WriteEntity(res)
}
