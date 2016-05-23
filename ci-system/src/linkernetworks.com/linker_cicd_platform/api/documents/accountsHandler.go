package documents

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"

	"linkernetworks.com/linker_cicd_platform/api/response"
	"linkernetworks.com/linker_cicd_platform/persistence/entity"
	util "linkernetworks.com/linker_cicd_platform/util"

	"github.com/Sirupsen/logrus"
	"github.com/emicklei/go-restful"
	"gopkg.in/mgo.v2/bson"
)

var (
	accountCollection = "accounts"
)

//Define RESTful API
func (p Resource) AccountsWebService() *restful.WebService {
	ws := new(restful.WebService)
	ws.Path("/v1/accounts")
	ws.Consumes("*/*")
	ws.Produces(restful.MIME_JSON)

	//usage:GET  /v1/accounts/gerrit?opsenv_id=<ID>
	ws.Route(ws.GET("/gerrit").
		To(p.GerritAccountListHandler).
		Param(ws.QueryParameter("opsenv_id", "OpsEnv ID")).
		Doc("Access gerrit account").
		Operation("GerritAccountListHandler").
		Produces(restful.MIME_JSON).
		Reads(""))

	//usage:PUT  /v1/accounts/gerrit
	ws.Route(ws.PUT("/gerrit").
		To(p.GerritAccountUpdateHandler).
		Doc("Update gerrit account").
		Operation("GerritAccountUpdateHandler").
		Produces(restful.MIME_JSON).
		Reads(""))

	//usage:GET  /v1/accounts/openldap?opsenv_id=<ID>
	ws.Route(ws.GET("/openldap").
		To(p.LdapAccountListHandler).
		Param(ws.QueryParameter("opsenv_id", "OpsEnv ID")).
		Doc("Access openldap account").
		Operation("LdapAccountListHandler").
		Produces(restful.MIME_JSON).
		Reads(""))

	//usage:PUT  /v1/accounts/openldap
	ws.Route(ws.PUT("/openldap").
		To(p.LdapAccountUpdateHandler).
		Doc("Update openldap account").
		Operation("LdapAccountUpdateHandler").
		Produces(restful.MIME_JSON).
		Reads(""))

	return ws
}

// Handler to list gerrit account.
// Usage:
//	  GET /v1/accounts/gerrit?opsenv_id=<ID>
// Authorization:
//	 None
// URL params:
//	 opsenv_id:	OpsEnv ID
// Returns:
//	200,OK if succeed,json like this in body.
//	{
//	  "success": true,
//	  "data": {
//	    "username": "linker",
//	    "password": "password",
//	    "new_password": "password2"
//	  }
//	}
func (p *Resource) GerritAccountListHandler(req *restful.Request,
	resp *restful.Response) {
	logrus.Debugf("GerritAccountListHandler is called")

	//get params
	opsenv_id := req.QueryParameter("opsenv_id")

	//check params
	if !bson.IsObjectIdHex(opsenv_id) {
		response.WriteStatusError(http.StatusBadRequest,
			response.ErrBadRequestURL, resp)
		return
	}

	//query db
	gerrit_account, err := p.findGerritAccountByOpsenvId(opsenv_id)
	if err != nil {
		response.WriteError(response.ErrDBQuery, resp)
		return
	}

	response.WriteResponse(gerrit_account, resp)
	return
}

// Handler to change gerrit http password.
// Note that the http password differs from the login password
// in the Gerrit web UI.This password is used in http request to
// call gerrit RESTful API.
// Usage:
//	 PUT /v1/accounts/gerrit
// Authorization:
//	 None
// URL params:
//	 None
// Body:
//	 Accepts json like this:
//	 {
//	     "opsenv_id":"562d88707227b029162d56be",
//	     "gerrit_account":{
//	         "username":"linker",
//	         "password":"password",
//	         "new_password":"password2"
//	     }
//	 }
// Returns:
//	200,OK if update successfully,new json will be returned in response body.
//	400,BadRequst
//	401,Auth failure
//	500,InternalServerError
func (p *Resource) GerritAccountUpdateHandler(req *restful.Request,
	resp *restful.Response) {
	logrus.Debugf("GerritAccountUpdateHandler is called")

	//parse request
	var accounts *entity.Accounts
	decoder := json.NewDecoder(req.Request.Body)
	err := decoder.Decode(&accounts)
	if err != nil {
		logrus.Errorln("Error decoding request body to entity.")
		response.WriteResponseStatus(http.StatusBadRequest,
			response.ErrBadRequestBody, resp)
		return
	}

	//get opsenv id
	opsenv_id := accounts.OpsEnvId
	if len(opsenv_id) == 0 {
		logrus.Errorln("Provide opsenv id in request body.")
		response.WriteStatusError(http.StatusBadRequest,
			response.ErrBadRequestBody, resp)
		return
	}

	//query gerrit ip and port
	host, port, err := p.queryGerritInfo(opsenv_id)
	if err != nil {
		logrus.Debugln("cannot query gerrit ip and port")
		response.WriteError(response.ErrDBQuery, resp)
		return
	}

	//get params
	var gerrit_account = accounts.GerritAccount
	username := gerrit_account.Username
	password := gerrit_account.Password
	new_password := gerrit_account.NewPassword

	//check params
	if len(username) == 0 || len(password) == 0 || len(new_password) == 0 {
		logrus.Debugln("Empty username or password")
		response.WriteStatusError(http.StatusBadRequest,
			response.ErrBadRequestBody, resp)
		return
	}

	//	logrus.Debugf("Update gerrit account. Host %s, Port %s, Username %s, Password %s, NewPassword %s",
	//		host, port, username, password, new_password)

	//call gerrit
	gerrit_resp, err := changeGerritHTTPPasswd(host, port, username,
		password, new_password)
	if err != nil {
		logrus.Debugf("Fail to update account")
		response.WriteError(response.ErrUpdateGerritPasswd, resp)
		return
	}

	switch {
	case gerrit_resp.StatusCode == 200:
		//succeed
		//save update to db
		err = p.updateGerritAccountToDb(opsenv_id, &gerrit_account)
		if err != nil {
			logrus.Errorln("Fail to insert or update gerrit account to db")
			response.WriteError(response.ErrDBUpdate, resp)
			return
		}
		//response
		body, _ := ConvertToBson(gerrit_account)
		response.WriteResponseStatus(http.StatusOK, body, resp)
		return
	case gerrit_resp.StatusCode == 401:
		logrus.Errorf("%v", response.NewError(response.ErrGerritAuthFailure))
		response.WriteStatusError(gerrit_resp.StatusCode,
			response.ErrGerritAuthFailure, resp)
		return
	default:
		logrus.Errorln("Cannot change gerrit HTTP password,unknown reason.")
		response.WriteStatusError(gerrit_resp.StatusCode,
			response.ErrUpdateGerritPasswd, resp)
		return
	}
}

// Handler to list openldap account.
// Usage:
//	  GET /v1/accounts/openldap?opsenv_id=<ID>
// Authorization:
//	 None
// URL params:
//	 opsenv_id:	OpsEnv ID
// Returns:
//	200,OK if succeed,json like this in body.
//	{
//	  "success": true,
//	  "data": {
//	    "dn": "uid=linker@linkernetworks.com,dc=linkernetworks,dc=com",
//	    "password": "password",
//	    "new_password": "password2"
//	  }
//	}
func (p *Resource) LdapAccountListHandler(req *restful.Request,
	resp *restful.Response) {
	logrus.Debugf("LdapAccountListHandler is called")

	//get params
	opsenv_id := req.QueryParameter("opsenv_id")

	//check params
	if !bson.IsObjectIdHex(opsenv_id) {
		response.WriteStatusError(http.StatusBadRequest,
			response.ErrBadRequestURL, resp)
		return
	}

	//query db
	openldap_account, err := p.findOpenldapAccoutByOpsenvId(opsenv_id)
	if err != nil {
		response.WriteError(response.ErrDBQuery, resp)
		return
	}

	//return
	response.WriteResponse(openldap_account, resp)
	return
}

//Handler to update openldap password.
//Usage:
//	PUT  /v1/accounts/openldap
//Authorization:
//	None
//URL params:
//	None
//Body:
//	Accepts json like this:
//	{
//		"opsenv_id":"562d88707227b029162d56be",
//		"openldap_account":{
//		    "dn":"uid=linker@linkernetworks.com,dc=linkernetworks,dc=com",
//		    "password":"password",
//		    "new_password":"password2"
//		}
//	}
//Returns:
//	200,OK if update successfully
//	500,InternalServerError if failed
func (p *Resource) LdapAccountUpdateHandler(req *restful.Request, resp *restful.Response) {
	logrus.Debugf("LdapAccountUpdateHandler is called")

	//parse body and convert to entity
	var accounts *entity.Accounts
	decoder := json.NewDecoder(req.Request.Body)
	err := decoder.Decode(&accounts)
	if err != nil {
		logrus.Errorln("Error decoding request body to entity.")
		response.WriteStatusError(http.StatusBadRequest,
			response.ErrBadRequestBody, resp)
		return
	}

	//get params
	opsenv_id := accounts.OpsEnvId
	host, port, err := p.queryLdapInfo(opsenv_id)
	openldap_account := accounts.OpenldapAccount
	account_dn := openldap_account.DN
	password := openldap_account.Password
	new_password := openldap_account.NewPassword

	//check params
	if len(account_dn) == 0 || len(password) == 0 || len(new_password) == 0 {
		response.WriteStatusError(http.StatusBadRequest,
			response.ErrBadRequestBody, resp)
		return
	}

	//	logrus.Debugf("Update openldap account. Host %s, Port %s, DN %s, Password %s, NewPassword %s",
	//		host, port, account_dn, password, new_password)

	//call openldap
	err = changeOpenldapPasswd(host, port, account_dn, password, new_password)
	if err != nil {
		logrus.Debugf("Cannot change openldap password.")
		response.WriteError(response.ErrUpdateLdapPasswd, resp)
		return
	}

	//succeed
	//save update to db
	err = p.updateOpenldapAccountToDb(opsenv_id, &openldap_account)
	if err != nil {
		logrus.Errorln("Fail to insert or update openldap account to db")
		response.WriteError(response.ErrDBUpdate, resp)
		return
	}

	//return
	body, _ := ConvertToBson(openldap_account)
	response.WriteResponseStatus(http.StatusOK, body, resp)
	return
}

//update gerrit account or insert accounts
func (p *Resource) updateGerritAccountToDb(opsenv_id string,
	gerritAccount *entity.GerritAccount) (err error) {
	if gerritAccount == nil {
		return errors.New("Nil pointer")
	}
	//query
	selector := bson.M{}
	selector["opsenv_id"] = opsenv_id
	count, _, _, err := p.Dao.HandleQuery(accountCollection, selector, false,
		bson.M{}, 0, 0, "", "true")

	//not found,insert
	if count == 0 {
		//new entity
		objId := bson.NewObjectId()
		accounts := entity.Accounts{
			Id:            objId.Hex(),
			OpsEnvId:      opsenv_id,
			GerritAccount: *gerritAccount,
		}
		//insert
		selector2 := make(bson.M)
		selector2[ParamID] = objId
		_, _, err := p.Dao.HandleInsert(accountCollection, selector2,
			ConvertAccountsToBson(accounts))
		return err
	}

	//found,update
	update := bson.M{}
	update["gerrit_account"] = *gerritAccount
	err = p.Dao.HandleUpdateByQueryPartial(accountCollection, selector, update)
	return
}

//update openldap account or insert accounts
func (p *Resource) updateOpenldapAccountToDb(opsenv_id string,
	openldapAccount *entity.OpenldapAccount) (err error) {
	if openldapAccount == nil {
		return errors.New("Nil pointer")
	}
	//query
	selector := bson.M{}
	selector["opsenv_id"] = opsenv_id
	count, _, _, err := p.Dao.HandleQuery(accountCollection, selector, false,
		bson.M{}, 0, 0, "", "true")

	//not found,insert
	if count == 0 {
		//new entity
		objId := bson.NewObjectId()
		accounts := entity.Accounts{
			Id:              objId.Hex(),
			OpsEnvId:        opsenv_id,
			OpenldapAccount: *openldapAccount,
		}
		//insert
		selector2 := make(bson.M)
		selector2[ParamID] = objId
		_, _, err := p.Dao.HandleInsert(accountCollection, selector2,
			ConvertAccountsToBson(accounts))
		return err
	}

	//found,update
	update := bson.M{}
	update["openldap_account"] = *openldapAccount
	err = p.Dao.HandleUpdateByQueryPartial(accountCollection, selector, update)
	return
}

//call gerrit to change http password
func changeGerritHTTPPasswd(host string, port string, accountId string,
	oldPasswd string, newPasswd string) (*http.Response, error) {

	//'PUT /accounts/{account-id}/password.http'
	url := "http://" + host + ":" + port + "/a/accounts/" +
		accountId + "/password.http"
	method := "PUT"

	//Example body
	//	`{
	//		"generate":false,
	//		"http_password":"password2"
	//	}`
	body := string("{\"generate\":false,\"http_password\":\"" + newPasswd + "\"}")

	req, err := http.NewRequest(method, url, bytes.NewBuffer([]byte(body)))
	if err != nil {
		logrus.Errorf("Error making new http request,reason:%v", err)
		return nil, err
	}

	//generate digest auth string
	digest, err := util.CalcDigestHeader(accountId, oldPasswd, method, url)
	if err != nil {
		logrus.Errorf("Error generating digest,reason:%v", err)
		return nil, err
	}

	//add digest authorization in Header
	req.Header.Set("Authorization", digest)

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("charset", "UTF-8")

	//do http request
	client := http.DefaultClient
	resp, err := client.Do(req)
	defer resp.Body.Close()
	return resp, err
}

//call openldap to update password
func changeOpenldapPasswd(host string, port string, accountDN string, password string,
	newPasswd string) (err error) {
	//API disabled due to lack of openldap-devel.
	//Uncomment this line to enable API
	//Add mqu/openldap in build.sh, yum install openldap-devel and then build
	//	err = util.ChangeOpenldapPasswd(host, port, accountDN, password, newPasswd)
	return errors.New("API disabled.")
}

//query gerrit ip and port from opsenvCollection
func (p *Resource) queryGerritInfo(opsEnvId string) (host string, port string,
	err error) {
	//query
	opsenv, err := p.findOpsEnvById(opsEnvId)
	if err != nil {
		logrus.Errorln("Fail to query opsenv")
		return
	}

	host = opsenv.GerritDockerIP
	port = opsenv.GerritHttpPort
	return
}

//query openldap ip and port from opsenvCollection
func (p *Resource) queryLdapInfo(opsEnvId string) (host string, port string,
	err error) {
	//query
	opsenv, err := p.findOpsEnvById(opsEnvId)
	if err != nil {
		logrus.Errorln("Fail to query opsenv")
		return
	}

	host = opsenv.LdapDockerIP
	port = opsenv.LdapHttpPort
	return
}

//query accounts by OpsEnv ID
func (p *Resource) findAccountsByOpsenvId(opsEnvId string) (accounts *entity.Accounts,
	err error) {
	selector := bson.M{}
	selector["opsenv_id"] = opsEnvId
	_, _, document, err := p.Dao.HandleQuery(accountCollection, selector, true,
		bson.M{}, 0, 0, "", "true")

	//convert
	data, err := json.Marshal(document)
	if err != nil {
		return
	}
	err = json.Unmarshal(data, &accounts)
	if err != nil {
		return
	}
	return
}

//query gerrit account
func (p *Resource) findGerritAccountByOpsenvId(opsEnvId string) (*entity.GerritAccount,
	error) {
	accounts, err := p.findAccountsByOpsenvId(opsEnvId)
	return &accounts.GerritAccount, err
}

//query openldap account
func (p *Resource) findOpenldapAccoutByOpsenvId(opsEnvId string) (*entity.OpenldapAccount,
	error) {
	accounts, err := p.findAccountsByOpsenvId(opsEnvId)
	return &accounts.OpenldapAccount, err
}
