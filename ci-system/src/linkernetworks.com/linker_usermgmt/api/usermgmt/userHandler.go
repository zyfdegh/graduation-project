package usermgmt

import (
	"encoding/json"
	"errors"

	"github.com/Sirupsen/logrus"
	"github.com/compose/mejson"
	"github.com/emicklei/go-restful"
	"gopkg.in/mgo.v2/bson"
	"linkernetworks.com/linker_common_lib/persistence/entity"
	"linkernetworks.com/linker_common_lib/rest/response"
	"linkernetworks.com/linker_usermgmt/common"
	"linkernetworks.com/linker_usermgmt/services"
)

func (p Resource) UserService() *restful.WebService {
	ws := new(restful.WebService)
	ws.Path("/v1/user")
	ws.Consumes("*/*")
	ws.Produces(restful.MIME_JSON)

	id := ws.PathParameter(ParamID, "Storage identifier of user")
	paramID := "{" + ParamID + "}"

	// user
	ws.Route(ws.POST("/registry").To(p.UserRegistryHandler).
		Doc("Registry a new user").
		Operation("UserRegistryHandler").
		Param(ws.BodyParameter("body", "User registry request body in json format,for example {\"email\":\"...\", \"password\":\"...\", \"confirmpassword\":\"...\"}").DataType("string")))

	ws.Route(ws.POST("/login").To(p.UserLoginHandler).
		Doc("Login with an exist user").
		Operation("UserLoginHandler").
		Param(ws.BodyParameter("body", "User login request body in json format,for example {\"email\":\"...\", \"password\":\"...\"}").DataType("string")))

	// ws.Route(ws.POST("/create").To(p.UserCreateHandler).
	// 	Doc("Create a new user").
	// 	Operation("UserCreateHandler").
	// 	Param(ws.HeaderParameter("X-Auth-Token", "A valid authentication token")).
	// 	Reads(""))

	ws.Route(ws.DELETE("/" + paramID).To(p.UserDeleteHandler).
		Doc("Delete a user by its storage identifier").
		Operation("UserDeleteHandler").
		Param(ws.HeaderParameter("X-Auth-Token", "A valid authentication token")).
		Param(id))

	ws.Route(ws.GET("/active").To(p.UserActiveHandler).
		Doc("Active an user").
		Operation("UserActiveHandler").
		Param(ws.QueryParameter("uid", "identifier of an user")).
		Param(ws.QueryParameter("activeCode", "Active Code of a user")))

	ws.Route(ws.GET("/reactive/" + paramID).To(p.UserReActiveHandler).
		Doc("reactive a user by its storage identifier").
		Operation("UserReActiveHandler").
		Param(id))

	ws.Route(ws.GET("/").To(p.UserListHandler).
		Doc("Return all user items").
		Operation("UserListHandler").
		Param(ws.HeaderParameter("X-Auth-Token", "A valid authentication token")).
		// Param(ws.QueryParameter("fields", "Comma separated list of field names")).
		Param(ws.QueryParameter("count", "Count total items and return the result in X-Object-count header").DataType("boolean")).
		Param(ws.QueryParameter("skip", "Number of items to skip in the result set, default=0")).
		Param(ws.QueryParameter("limit", "Maximum number of items in the result set, default=0")).
		Param(ws.QueryParameter("sort", "Comma separated list of field names to sort")))

	ws.Route(ws.GET("/" + paramID).To(p.UserDetailHandler).
		Doc("Return a user by its storage identifier").
		Operation("UserDetailHandler").
		Param(ws.HeaderParameter("X-Auth-Token", "A valid authentication token")).
		Param(id))

	ws.Route(ws.PUT("/" + paramID).To(p.UserUpdateHandler).
		Doc("Updata a exist user by its storage identifier").
		Operation("UserUpdateHandler").
		Param(ws.HeaderParameter("X-Auth-Token", "A valid authentication token")).
		Param(id).
		Param(ws.BodyParameter("body", "User update request body in json format,for example {\"address\":\"...\", \"company\":\"...\", \"phonenum\":\"...\"}").DataType("string")))

	ws.Route(ws.PUT("/changepassword/" + paramID).To(p.UserChangePasswdHandler).
		Doc("change password of an exist user by its storage identifier").
		Operation("UserChangePasswdHandler").
		Param(ws.HeaderParameter("X-Auth-Token", "A valid authentication token")).
		Param(id).
		Param(ws.BodyParameter("body", "User login request body in json format,for example {\"password\":\"...\", \"newpassword\":\"...\",\"confirm_newpassword\":\"...\"}").DataType("string")))

	return ws
}

// CheckAndGenerateToken parses the http request and registry a new user.
// Usage :
//		POST /v1/user/registry
// If successful,response code will be set to 201.
func (p *Resource) UserRegistryHandler(req *restful.Request, resp *restful.Response) {
	logrus.Infof("UserRegistryHandler is called!")

	doc := bson.M{}
	decoder := json.NewDecoder(req.Request.Body)
	err := decoder.Decode(&doc)
	if err != nil {
		logrus.Errorf("decode user err is %v", err)
		response.WriteStatusError(common.COMMON_ERROR_INVALIDATE, err, resp)
		return
	}

	email, password, confirmPassword, alias, address, company, phonenum, infoSource, paraErr := userRegistryParamCheck(doc)
	if paraErr != nil {
		response.WriteStatusError(common.COMMON_ERROR_INVALIDATE, paraErr, resp)
		return
	}

	if len(email) == 0 || len(password) == 0 || len(confirmPassword) == 0 {
		logrus.Errorln("parameter can not be null!")
		response.WriteStatusError(common.COMMON_ERROR_INVALIDATE, errors.New("Invalid parameter"), resp)
		return
	}

	userParam := common.UserParam{
		Email:           email,
		Password:        password,
		ConfirmPassword: confirmPassword,
		Alias:           alias,
		Address:         address,
		Company:         company,
		Phonenum:        phonenum,
		InfoSource:      infoSource}
	errorCode, userId, err := services.GetUserService().UserRegistry(userParam)
	if err != nil {
		response.WriteStatusError(errorCode, err, resp)
		return
	}

	p.successUpdate(userId, true, req, resp)

}

func userRegistryParamCheck(doc interface{}) (email string, password string, confirmPassword string, alias string, address string, company string, phonenum string, infoSource string, paraErr error) {
	var document interface{}
	document, paraErr = mejson.Marshal(doc)
	if paraErr != nil {
		logrus.Errorf("marshal user err is %v", paraErr)
		return
	}

	docJson := document.(map[string]interface{})
	emailDoc := docJson["email"]
	if emailDoc == nil {
		logrus.Errorln("invalid parameter ! email can not be null")
		paraErr = errors.New("Invalid parameter!")
		return
	} else {
		email = emailDoc.(string)
	}

	passwordDoc := docJson["password"]
	if passwordDoc == nil {
		logrus.Errorln("invalid parameter ! password can not be null")
		paraErr = errors.New("Invalid parameter!")
		return
	} else {
		password = passwordDoc.(string)
	}

	confirmpasswordDoc := docJson["confirmpassword"]
	if confirmpasswordDoc == nil {
		logrus.Errorln("invalid parameter ! confirmpassword can not be null")
		paraErr = errors.New("Invalid parameter!")
		return
	} else {
		confirmPassword = confirmpasswordDoc.(string)
	}

	aliasDoc := docJson["alias"]
	if aliasDoc == nil {
		logrus.Errorln("invalid parameter! alias can not be null")
		paraErr = errors.New("Invalid parameter!")
		return
	} else {
		alias = aliasDoc.(string)
	}

	addressDoc := docJson["address"]
	if addressDoc != nil {
		address = addressDoc.(string)
	}

	companyDoc := docJson["company"]
	if companyDoc != nil {
		company = companyDoc.(string)
	}

	phonenumDoc := docJson["phonenumber"]
	if phonenumDoc != nil {
		phonenum = phonenumDoc.(string)
	}

	infosourceDoc := docJson["infosource"]
	if infosourceDoc != nil {
		infoSource = infosourceDoc.(string)
	}

	return
}

// UserReActiveHandler parses the http request and reactive a user.
// Usage :
//		GET v1/user/reactive/{ParamID}
// Params :
//		ParamID : storage identifier of user
// If successful,response code will be set to 201.
func (p *Resource) UserReActiveHandler(req *restful.Request, resp *restful.Response) {
	logrus.Infoln("UserReActiveHandler is called!")

	id := req.PathParameter("_id")

	errorCode, err := services.GetUserService().UserReactive(id)
	if err != nil {
		response.WriteStatusError(errorCode, err, resp)
		return
	}

	response.WriteSuccess(resp)
	return
}

// UserCreateHandler parses the http request and create a new user.
// Usage :
//		POST /v1/user/create
// If successful,response code will be set to 201.
func (p *Resource) UserCreateHandler(req *restful.Request, resp *restful.Response) {
	logrus.Infof("UserCreateHandler is called!")
	token := req.HeaderParameter("X-Auth-Token")

	document := bson.M{}
	decoder := json.NewDecoder(req.Request.Body)
	err := decoder.Decode(&document)
	if err != nil {
		logrus.Errorf("decode user err is %v", err)
		response.WriteStatusError(common.COMMON_ERROR_INTERNAL, err, resp)
		return
	}

	username := document["email"]
	alias := document["alias"]
	password := document["password"]
	confirmpassword := document["confirmpassword"]
	tenantname := document["tenantname"]
	// if len(username) == 0 || len(password) == 0 || len(confirmpassword) == 0 || len(tenantname) == 0 || len(alias) == 0 {
	if username == nil || alias == nil || password == nil || confirmpassword == nil || tenantname == nil {
		logrus.Errorf("invalid parameter, parameter should not be null!")
		response.WriteStatusError(common.COMMON_ERROR_INVALIDATE, errors.New("invalid parameter, parameter should not be null!"), resp)
		return
	}

	userId, errorCode, err := services.GetUserService().UserCreate(token, username.(string), password.(string), confirmpassword.(string), alias.(string), tenantname.(string))
	if err != nil {
		response.WriteStatusError(errorCode, err, resp)
		return
	}

	p.successUpdate(userId, true, req, resp)
}

// UserLoginHandler parses the http request and login with an exist user.
// Usage :
//		POST v1/user/login
// If successful,response code will be set to 201.
func (p *Resource) UserLoginHandler(req *restful.Request, resp *restful.Response) {
	logrus.Infof("UserLoginHandler is called!")

	document := bson.M{}
	decoder := json.NewDecoder(req.Request.Body)
	err := decoder.Decode(&document)
	if err != nil {
		logrus.Errorf("decode user err is %v", err)
		response.WriteStatusError(services.USER_ERROR_LOGIN, err, resp)
		return
	}

	document, err = mejson.Unmarshal(document)
	if err != nil {
		logrus.Errorf("unmarshal user err is %v", err)
		response.WriteStatusError(services.USER_ERROR_LOGIN, err, resp)
		return
	}

	username := document["email"].(string)
	password := document["password"].(string)
	if len(username) == 0 || len(password) == 0 {
		logrus.Errorf("username and password can not be null!")
		response.WriteStatusError(common.COMMON_ERROR_INVALIDATE, errors.New("Username or password can not be null"), resp)
		return
	}

	errorCode, loginRes, err := services.GetUserService().UserLogin(username, password)
	if err != nil {
		response.WriteStatusError(errorCode, err, resp)
		return
	}

	response.WriteResponse(loginRes, resp)

	return

}

// UserUpdateHandler parses the http request and updata a exist user.
// Usage :
//		PUT /v1/user/{ParamID}
// Params :
//		ParamID : storage identifier of user
// If successful,response code will be set to 201.
func (p *Resource) UserUpdateHandler(req *restful.Request, resp *restful.Response) {
	logrus.Infof("UserUpdateHanlder is called!")
	token := req.HeaderParameter("X-Auth-Token")
	id := req.PathParameter(ParamID)
	if len(id) <= 0 {
		logrus.Warnln("user id should not be null for update operation")
		response.WriteStatusError(common.COMMON_ERROR_INVALIDATE, errors.New("user id should not be null for update operation"), resp)
		return
	}

	newuser := entity.User{}

	// Populate the user data
	err := json.NewDecoder(req.Request.Body).Decode(&newuser)
	if err != nil {
		logrus.Errorf("convert body to user failed, error is %v", err)
		response.WriteStatusError(common.COMMON_ERROR_INVALIDATE, err, resp)
		return
	}

	created, id, errorCode, err := services.GetUserService().UserUpdate(token, newuser, id)
	if err != nil {
		response.WriteStatusError(errorCode, err, resp)
		return
	}

	p.successUpdate(id, created, req, resp)
}

// UserChangePasswdHandler parses the http request and change
// password of an exist user.
// Usage :
//		PUT v1/user/changepassword/{ParamID}
// Params :
//		ParamID : storage identifier of user
// If successful,response code will be set to 201.
func (p *Resource) UserChangePasswdHandler(req *restful.Request, resp *restful.Response) {
	logrus.Infof("UserChangePasswdHandler is called!")
	token := req.HeaderParameter("X-Auth-Token")
	id := req.PathParameter(ParamID)
	if len(id) <= 0 {
		logrus.Warnln("user id should not be null for change password operation")
		response.WriteStatusError(common.COMMON_ERROR_INVALIDATE, errors.New("user id should not be null for update operation"), resp)
		return
	}

	document := bson.M{}
	decoder := json.NewDecoder(req.Request.Body)
	err := decoder.Decode(&document)
	if err != nil {
		logrus.Errorf("decode change password object err is %v", err)
		response.WriteStatusError(common.COMMON_ERROR_INTERNAL, err, resp)
		return
	}

	document, err = mejson.Unmarshal(document)
	if err != nil {
		logrus.Errorf("unmarshal change password obejct err is %v", err)
		response.WriteStatusError(common.COMMON_ERROR_INTERNAL, err, resp)
		return
	}

	password := document["password"]
	newpwd1 := document["newpassword"]
	newpwd2 := document["confirm_newpassword"]
	if password == nil || newpwd1 == nil || newpwd2 == nil {
		logrus.Errorln("invalid parameter! password and newpassword field should not be null")
		response.WriteStatusError(common.COMMON_ERROR_INVALIDATE, errors.New("invalid parameter!password, newpassword and confirm_newpassword should not be null!"), resp)
		return
	}

	created, errorCode, err := services.GetUserService().UserChangePassword(token, id, password.(string), newpwd1.(string), newpwd2.(string))
	if err != nil {
		response.WriteStatusError(errorCode, err, resp)
		return
	}

	p.successUpdate(id, created, req, resp)

}

// UserDeleteHandler parses the http request and delete a user.
// Usage :
//		DELETE /v1/user/{ParamID}
// Params :
//		ParamID : storage identifier of user
// If successful,response code will be set to 201.
func (p *Resource) UserDeleteHandler(req *restful.Request, resp *restful.Response) {
	logrus.Infof("UserDeleteHandler is called!")
	token := req.HeaderParameter("X-Auth-Token")
	id := req.PathParameter(ParamID)
	if len(id) <= 0 {
		logrus.Warnln("user id should not be null for delete operation")
		response.WriteStatusError(common.COMMON_ERROR_INVALIDATE, errors.New("user id should not be null for delete operation"), resp)
		return
	}

	errorCode, err := services.GetUserService().UserDelete(token, id)
	if err != nil {
		response.WriteStatusError(errorCode, err, resp)
		return
	}

	response.WriteSuccess(resp)
}

// UserListHandler parses the http request and return the user items.
// Usage :
//		GET /v1/user
//		GET /v1/user/{ParamID}
// Params :
//		ParamID : storage identifier of user
// If successful,response code will be set to 201.
func (p *Resource) UserListHandler(req *restful.Request, resp *restful.Response) {
	logrus.Infof("UserListHandler is called!")

	token := req.HeaderParameter("X-Auth-Token")
	limitnum := queryIntParam(req, "limit", 10)
	skipnum := queryIntParam(req, "skip", 0)
	sort := req.QueryParameter("sort")

	ret, count, errorCode, err := services.GetUserService().UserList(token, limitnum, skipnum, sort)
	if err != nil {
		response.WriteStatusError(errorCode, err, resp)
		return
	}

	p.successList(ret, limitnum, count, req, resp)
}

func (p *Resource) UserDetailHandler(req *restful.Request, resp *restful.Response) {
	logrus.Infof("UserDetailHandler is called!")

	token := req.HeaderParameter("X-Auth-Token")
	id := req.PathParameter(ParamID)
	if len(id) <= 0 {
		logrus.Warnln("user id should not be null for user detail operation")
		response.WriteStatusError(common.COMMON_ERROR_INVALIDATE, errors.New("user id should not be null for get user operation"), resp)
		return
	}

	ret, errorCode, err := services.GetUserService().UserDetail(token, id)
	if err != nil {
		response.WriteStatusError(errorCode, err, resp)
		return
	}

	response.WriteResponse(ret, resp)
}

// UserActiveHandler parses the http request and active an user.
// Usage :
//		GET /v1/user/active
// If successful,response code will be set to 201.
func (p *Resource) UserActiveHandler(req *restful.Request, resp *restful.Response) {
	logrus.Infof("UserActiveHandler is called!")

	id := req.QueryParameter("uid")
	code := req.QueryParameter("activeCode")
	if len(id) <= 0 || len(code) <= 0 {
		response.WriteStatusError(common.COMMON_ERROR_INVALIDATE, errors.New("uid and activeCode can not be null"), resp)
		return
	}

	errorCode, err := services.GetUserService().UserActive(code, id)
	if err != nil {
		response.WriteStatusError(errorCode, err, resp)
		return
	}

	response.WriteSuccess(resp)
}
