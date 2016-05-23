package documents

import (
	"encoding/json"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/emicklei/go-restful"
	"github.com/jmoiron/jsonq"
	"gopkg.in/mgo.v2/bson"
	"linkernetworks.com/linker_cicd_platform/api/response"
	"linkernetworks.com/linker_cicd_platform/persistence/entity"
	"linkernetworks.com/linker_cicd_platform/util"
)

var (
	projectenvCollection = "linker_project_env"
)

func (u Resource) ProjectEnvWebService() *restful.WebService {
	ws := new(restful.WebService)
	ws.Path("/v1/projectenvs")
	ws.Consumes("*/*")
	ws.Produces(restful.MIME_JSON)

	id := ws.PathParameter(ParamID, "Storage identifier of the app configuration")
	paramID := "{" + ParamID + "}"

	ws.Route(ws.POST("").To(u.ProjectEnvsCreateHandler).
		Doc("Create linker ops project envs").
		Operation("ProjectEnvsCreateHandler").
		Param(ws.QueryParameter("projectid", "Project Id")).
		Param(ws.QueryParameter("jobid", "Job id")).
		Param(ws.HeaderParameter("X-Auth-Token", "A valid authentication token")).
		Reads(""))

	ws.Route(ws.DELETE("/" + paramID).To(u.ProjectEnvsDeleteHandler).
		Doc("Delete linker ops project envs").
		Operation("ProjectEnvsDeleteHandler").
		Param(ws.HeaderParameter("X-Auth-Token", "A valid authentication token")).
		Reads(id))

	ws.Route(ws.GET("/").To(u.ProjectEnvsListHandler).
		Doc("Returns all Env items").
		Operation("ProjectEnvsListHandler").
		Param(ws.QueryParameter("count", "Count total items and return the result in X-Object-Count header").DataType("boolean")).
		Param(ws.QueryParameter("query", "Query in json format")).
		Param(ws.QueryParameter("fields", "Comma separated list of field names")).
		Param(ws.QueryParameter("skip", "Number of items to skip in the result set, default=0")).
		Param(ws.QueryParameter("limit", "Maximum number of items in the result set, default=0")).
		Param(ws.QueryParameter("sort", "Comma separated list of field names to sort")))

	ws.Route(ws.GET("/" + paramID).To(u.ProjectEnvsListHandler).
		Doc("Return an Env by its storage identifier ").
		Operation("ProjectEnvsListHandler").
		Param(id).
		Param(ws.QueryParameter("fields", "Comma separated list of field names")))

	ws.Route(ws.POST("/notify").To(u.ProjectEnvsNotifyHandler).
		Doc("Notify linker ops Project envs").
		Operation("ProjectEnvsNotifyHandler").
		Param(ws.QueryParameter("orderid", "Order Id")).
		Param(ws.QueryParameter("status", "Order Status")).
		Param(ws.HeaderParameter("X-Auth-Token", "A valid authentication token")).
		Reads(""))

	return ws
}

// Create the CICD Linker Ops Project Env through Controller
func (u *Resource) ProjectEnvsCreateHandler(req *restful.Request, resp *restful.Response) {
	//analyze the request
	logrus.Debugf("ProjectEnvsCreateHandler is called")
	var projectId string = req.QueryParameter("projectid")
	var jobId string = req.QueryParameter("jobid")
	var token string = req.HeaderParameter("X-Auth-Token")

	//get Project Info
	selector := make(bson.M)
	selector["_id"] = bson.ObjectIdHex(projectId)
	_, _, projJson, err := u.Dao.HandleQuery(projectCollection, selector, true, bson.M{}, 0, 0, "", "true")

	if err != nil {
		logrus.Errorf("get project to create project env %s err is %v", err)
		response.WriteError(response.ErrDBQuery, resp)
		return
	}

	var project *entity.Project
	project = new(entity.Project)
	projout, _ := json.Marshal(projJson)
	json.Unmarshal(projout, &project)

	opsEnvId := project.OpsEnvId
	//get service model id
	serviceModelName := project.ServiceModelId

	orderString := "{\"service_group_id\":\"" + serviceModelName + "\",\"parameters\": []}"

	//call controller for the order
	err, serviceGroupInstanceId, serviceOrderId, serviceOfferingInstanceId, userId, tenantId := u.createEnv(orderString, token)
	if err != nil {
		logrus.Errorf("call controller to create ops env %s err is %v", err)
		response.WriteError(response.ErrCreateOpsenv, resp)
		return
	}

	//save the cicd project env into database
	newId := bson.NewObjectId()
	envObj := entity.ProjectEnv{Id: newId.Hex(), OpsEnvId: opsEnvId, ProjectId: projectId,
		JobId: jobId, UserId: userId, TenantId: tenantId, ServiceGroupInstanceId: serviceGroupInstanceId,
		ServiceOrderId: serviceOrderId, Status: "Ordered", ServiceOfferingInstanceId: serviceOfferingInstanceId}
	pselector := make(bson.M)
	pselector[ParamID] = newId

	document, err := ConvertToBson(envObj)
	if err != nil {
		logrus.Errorf("convert to bson error is %v", err)
	} else {
		_, _, err = u.Dao.HandleInsert(projectenvCollection, pselector, document)
		if err != nil {
			logrus.Errorf("insert project env err is %v", err)
			response.WriteError(response.ErrDBInsert, resp)
			return
		}
	}

	response.WriteResponse(envObj, resp)
	return
}

// Create the CICD Linker Ops Project Env through Controller
func (u *Resource) ProjectEnvsDeleteHandler(req *restful.Request, resp *restful.Response) {
	//analyze the request
	logrus.Debugf("ProjectEnvsDeleteHandler is called")
	var token string = req.HeaderParameter("X-Auth-Token")

	selector, one, err := getSelector(req)
	_, _, projectenvJson, err := u.Dao.HandleQuery(projectenvCollection, selector, one, bson.M{}, 0, 0, "", "true")

	if err != nil {
		logrus.Errorf("get env to delete groupInstance %s err is %v", err)
		response.WriteError(response.ErrDBQuery, resp)
		return
	}

	var env *entity.ProjectEnv
	env = new(entity.ProjectEnv)
	envout, _ := json.Marshal(projectenvJson)
	json.Unmarshal(envout, &env)

	//call controller for the delete
	err = u.deleteEnv(env.ServiceOrderId, token)

	if err != nil {
		logrus.Errorf("delete groupInstance %s err is %v", err)
		response.WriteError(response.ErrDeleteOpsenv, resp)
		return
	}

	//call controller to terminate linker ops env
	err = u.deleteProjectEnv(env.ServiceOrderId, token)
	if err != nil {
		logrus.Errorf("delete env err is %v", err)
		response.WriteError(response.ErrDeleteOpsenv, resp)
		return
	}

	//terminate project env
	err = u.terminateProjectEnv(env.Id)
	if err != nil {
		logrus.Debugf("Error terminate project env,reason:%v,projectenv id:%s", err, env.Id)
	}

	response.WriteSuccess(resp)
	return
}

// Get Notify from Controller to Update Project Env Information
func (u *Resource) ProjectEnvsNotifyHandler(req *restful.Request, resp *restful.Response) {
	//analyze the request
	logrus.Debugf("ProjectEnvsNotifyHandler is called")
	var serviceOrderId string = req.QueryParameter("orderid")
	var status string = req.QueryParameter("status")

	//get cicd env information from database
	selector := make(bson.M)
	selector["service_order_id"] = serviceOrderId
	_, _, projEnvJson, err := u.Dao.HandleQuery(projectenvCollection, selector, true, bson.M{}, 0, 0, "", "true")

	if err != nil {
		logrus.Errorf("get env to update %s err is %v", serviceOrderId, err)
		response.WriteError(response.ErrDBQuery, resp)
		return
	}

	var env *entity.ProjectEnv
	env = new(entity.ProjectEnv)
	envout, _ := json.Marshal(projEnvJson)
	json.Unmarshal(envout, &env)

	env.Status = status

	envselector := bson.M{}
	envselector["_id"] = bson.ObjectIdHex(env.Id)

	//save the cicd project env into database
	document, err := ConvertToBson(*env)
	if err != nil {
		logrus.Errorf("convert to bson error is %v", err)
	} else {
		_, _, _, err = u.Dao.HandleUpdateById(projectenvCollection, envselector, document)
		if err != nil {
			logrus.Errorf("update env err is %v", err)
			response.WriteError(response.ErrDBQuery, resp)
			return
		}
	}

	response.WriteResponse(env, resp)
	return
}

// List cicd project envs from database
func (u *Resource) ProjectEnvsListHandler(req *restful.Request, resp *restful.Response) {
	//analyze the request
	logrus.Debugf("ProjectEnvsListHandler is called")

	u.handleList(projectenvCollection, "list_envs", req, resp)
	return
}

//call controller to order linker ops env
func (u *Resource) createProjectEnv(orderString, token string) (err error, serviceGroupInstanceId, serviceOrderId, serviceOfferingInstanceId, userId, tenantId string) {
	logrus.Debugln("body=" + orderString)

	controllerUrl, err := u.Util.ZkClient.GetControllerEndpoint()
	if err != nil {
		logrus.Errorf("get controller endpoint err is %+v", err)
		return
	}
	url := strings.Join([]string{"http://", controllerUrl, "/v1/serviceGroupOrders"}, "")

	logrus.Debugln("Create Service Group Order for Linker Project Ops url=" + url)
	resp, err := linker_util.Http_post(url, "application/json", token, orderString)
	if err != nil {
		logrus.Errorf("post service group order for linker Projectops error %s", err.Error())
		return
	}

	logrus.Infof("create Service Group Order for Linker Project Ops response: %v", resp)
	jsondata := map[string]interface{}{}
	dec := json.NewDecoder(strings.NewReader(resp))
	dec.Decode(&jsondata)
	jq := jsonq.NewQuery(jsondata)
	serviceGroupInstanceId, err = jq.String("data", "service_group_instance_id")
	serviceOfferingInstanceId, err = jq.String("data", "service_offering_id")
	serviceOrderId, err = jq.String("data", "order_id")
	userId, err = jq.String("data", "user_id")
	tenantId, err = jq.String("data", "tenant_id")
	return
}

//call controller to terminate linker ops env
func (u *Resource) deleteProjectEnv(orderid, token string) (err error) {
	// call appInstance terminate api
	controllerUrl, err := u.Util.ZkClient.GetControllerEndpoint()
	if err != nil {
		logrus.Errorf("get controller endpoint err is %+v", err)
		return
	}
	url := strings.Join([]string{"http://", controllerUrl, "/v1/serviceGroupOrders/", orderid}, "")
	logrus.Debugln("Terminate linker ops url=" + url)

	resp, err := linker_util.Http_delete(url, "application/json", token, "")
	if err != nil {
		logrus.Errorf("delete linker ops error %s", err.Error())
		return
	}
	logrus.Info(resp)
	return
}

func (u *Resource) findProjectEnvById(id string) (projectEnv *entity.ProjectEnv, err error) {
	selector := bson.M{}
	selector[ParamID] = bson.ObjectIdHex(id)
	_, _, document, err := u.Dao.HandleQuery(projectenvCollection, selector, true, bson.M{}, 0, 0, "", "true")
	projectEnv = new(entity.ProjectEnv)
	data, _ := json.Marshal(document)
	json.Unmarshal(data, &projectEnv)
	return
}

func (u *Resource) terminateProjectEnvs(jobId string) (err error) {
	//query project envs
	selector := bson.M{}
	selector["job_id"] = jobId
	_, _, document, err := u.Dao.HandleQuery(projectenvCollection, selector, false, bson.M{}, 0, 0, "", "true")
	if err != nil {
		return
	}
	var projEnvs []entity.ProjectEnv = make([]entity.ProjectEnv, 0)
	data, _ := json.Marshal(document)
	err = json.Unmarshal(data, &projEnvs)
	if err != nil {
		return
	}

	//loop
	for _, projEnv := range projEnvs {
		err = u.terminateProjectEnv(projEnv.Id)
		if err != nil {
			logrus.Debugf("Error terminate project env,reason:%v,projectenv id:%s", err, projEnv.Id)
		}
	}
	return
}

//terminate project env
func (u *Resource) terminateProjectEnv(projEnvId string) (err error) {
	//delete project env
	projectEnv, err := u.findProjectEnvById(projEnvId)
	if err != nil {
		logrus.Debugf("Error find project env.reason:%v.projEnvId", err, projEnvId)
		return
	}
	token, err := u.getUserToken(projectEnv.UserId, projectEnv.TenantId)
	err = u.deleteProjectEnv(projectEnv.ServiceOrderId, token)
	if err != nil {
		logrus.Debugf("Error delete project env.reason:%v.projEnvId:", err, projEnvId)
		return
	}
	//delete env
	err = u.deleteEnv(projectEnv.ServiceOrderId, token)
	if err != nil {
		logrus.Debugf("Error delete env.reason:%v.projEnvId:", err, projEnvId)
		return
	}

	//terminate project env
	change := bson.M{"status": entity.PROJECTENV_STATUS_TERMINATED}
	selector := bson.M{}
	selector[ParamID] = bson.ObjectIdHex(projEnvId)
	err = u.Dao.HandleUpdateByQueryPartial(projectenvCollection, selector, change)
	return
}
