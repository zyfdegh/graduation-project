package documents

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/emicklei/go-restful"
	"github.com/jmoiron/jsonq"
	"gopkg.in/mgo.v2/bson"
	"linkernetworks.com/linker_cicd_platform/api/response"
	"linkernetworks.com/linker_cicd_platform/persistence/entity"
	"linkernetworks.com/linker_cicd_platform/util"
	conentity "linkernetworks.com/linker_common_lib/persistence/entity"
)

var (
	opsenvCollection = "linker_ops_env"
	cicdGroupName    = "linkerops"
	Zk               *linker_util.ZkClient
)

func (u Resource) OpsEnvWebService() *restful.WebService {
	ws := new(restful.WebService)
	ws.Path("/v1/opsenvs")
	ws.Consumes("*/*")
	ws.Produces(restful.MIME_JSON)

	id := ws.PathParameter(ParamID, "Storage identifier of the app configuration")
	paramID := "{" + ParamID + "}"

	ws.Route(ws.POST("").To(u.EnvsCreateHandler).
		Doc("Create linker ops envs").
		Operation("EnvsCreateHandler").
		Param(ws.QueryParameter("envname", "Env name")).
		Param(ws.HeaderParameter("X-Auth-Token", "A valid authentication token")).
		Reads(""))

	ws.Route(ws.DELETE("/" + paramID).To(u.EnvsDeleteHandler).
		Doc("Delete linker ops envs").
		Operation("EnvsDeleteHandler").
		Param(ws.HeaderParameter("X-Auth-Token", "A valid authentication token")).
		Reads(id))

	ws.Route(ws.PUT("/" + paramID).To(u.EnvsUpdateHandler).
		Doc("Updates linker ops envs").
		Operation("EnvsUpdateHandler").
		Param(ws.HeaderParameter("X-Auth-Token", "A valid authentication token")).
		Reads(id))

	ws.Route(ws.GET("/").To(u.EnvsListHandler).
		Doc("Returns all Env items").
		Operation("EnvsListHandler").
		Param(ws.QueryParameter("count", "Count total items and return the result in X-Object-Count header").DataType("boolean")).
		Param(ws.QueryParameter("query", "Query in json format")).
		Param(ws.QueryParameter("fields", "Comma separated list of field names")).
		Param(ws.QueryParameter("skip", "Number of items to skip in the result set, default=0")).
		Param(ws.QueryParameter("limit", "Maximum number of items in the result set, default=0")).
		Param(ws.QueryParameter("sort", "Comma separated list of field names to sort")))

	ws.Route(ws.GET("/" + paramID).To(u.EnvsListHandler).
		Doc("Return an Env by its storage identifier ").
		Operation("EnvsListHandler").
		Param(id).
		Param(ws.QueryParameter("fields", "Comma separated list of field names")))

	ws.Route(ws.POST("/notify").To(u.EnvsNotifyHandler).
		Doc("Notify linker ops envs").
		Operation("EnvsNotifyHandler").
		Param(ws.QueryParameter("orderid", "Order Id")).
		Param(ws.QueryParameter("status", "Order Status")).
		Param(ws.HeaderParameter("X-Auth-Token", "A valid authentication token")).
		Reads(""))

	return ws
}

// Create the CICD Linker Ops Env through Controller
func (u *Resource) EnvsCreateHandler(req *restful.Request, resp *restful.Response) {
	//analyze the request
	logrus.Debugf("EnvsCreateHandler is called")
	var envname string = req.QueryParameter("envname")
	var token string = req.HeaderParameter("X-Auth-Token")
	var orderString = "{\"service_group_id\":\"/linkerops\",\"parameters\": []}"

	//call controller for the order
	err, serviceGroupInstanceId, serviceOrderId, serviceOfferingInstanceId, userId, tenantId := u.createEnv(orderString, token)
	if err != nil {
		logrus.Errorf("call controller to create ops env %s err is %v", err)
		response.WriteError(response.ErrCreateOpsenv, resp)
		return
	}

	//save the cicd env into database
	newId := bson.NewObjectId()
	envObj := entity.OpsEnv{Id: newId.Hex(), Name: envname, UserId: userId, TenantId: tenantId,
		ServiceGroupInstanceId: serviceGroupInstanceId, ServiceOrderId: serviceOrderId,
		Status: "Ordered", ServiceOfferingInstanceId: serviceOfferingInstanceId}
	selector := make(bson.M)
	selector[ParamID] = newId

	_, _, err = u.Dao.HandleInsert(opsenvCollection, selector, ConvertOpsEnvToBson(envObj))

	if err != nil {
		logrus.Errorf("insert env err is %v", err)
		response.WriteError(response.ErrDBInsert, resp)
		return
	}

	response.WriteResponse(envObj, resp)
	return
}

// Create the CICD Linker Ops Env through Controller
func (u *Resource) EnvsDeleteHandler(req *restful.Request, resp *restful.Response) {
	//analyze the request
	logrus.Debugf("EnvsDeleteHandler is called")
	var token string = req.HeaderParameter("X-Auth-Token")

	selector, one, err := getSelector(req)
	_, _, opsEnvJson, err := u.Dao.HandleQuery(opsenvCollection, selector, one, bson.M{}, 0, 0, "", "true")

	if err != nil {
		logrus.Errorf("get env to delete groupInstance %s err is %v", err)
		response.WriteError(response.ErrDBQuery, resp)
		return
	}

	var env *entity.OpsEnv
	env = new(entity.OpsEnv)
	envout, _ := json.Marshal(opsEnvJson)
	json.Unmarshal(envout, &env)

	//call controller for the delete
	err = u.deleteEnv(env.ServiceOrderId, token)
	if err != nil {
		logrus.Errorf("delete groupInstance %s err is %v", err)
		response.WriteError(response.ErrDeleteOpsenv, resp)
		return
	}

	//terminate projects and this opsenv
	err = u.terminateOpsEnv(env.Id)
	if err != nil {
		logrus.Errorf("Error terminate opsenv,reason:%v,opsenv id:%s", err, env.Id)
	}

	response.WriteSuccess(resp)
	return
}

// Update the CICD Linker Ops Env Information from Controller
func (u *Resource) EnvsUpdateHandler(req *restful.Request, resp *restful.Response) {
	//analyze the request
	logrus.Debugf("EnvsUpdateHandler is called")
	var token string = req.HeaderParameter("X-Auth-Token")

	opsEnvId := req.PathParameter(ParamID)
	env, err := u.findOpsEnvById(opsEnvId)

	if err != nil {
		logrus.Errorf("get env to delete groupInstance %s err is %v", err)
		response.WriteError(response.ErrDBQuery, resp)
		return
	}

	//call controller for the latest information
	err = u.updateEnv(env, token)

	if err != nil {
		logrus.Errorf("get information from controller fail,  to update groupInstance %s err is %v", err)
		response.WriteError(response.ErrUpdateOpsenv, resp)
		return
	}

	//update the cicd env into database
	selector := bson.M{}
	selector[ParamID] = bson.ObjectIdHex(opsEnvId)
	_, _, _, err = u.Dao.HandleUpdateById(opsenvCollection, selector, ConvertOpsEnvToBson(*env))
	if err != nil {
		logrus.Errorf("update env err is %v", err)
		response.WriteError(response.ErrDBUpdate, resp)
		return
	}
	response.WriteResponse(env, resp)
	return
}

// Get Notify from Controller to Update Env Information
func (u *Resource) EnvsNotifyHandler(req *restful.Request, resp *restful.Response) {
	//analyze the request
	logrus.Debugf("EnvsNotifyHandler is called")
	var serviceOrderId string = req.QueryParameter("orderid")
	var status string = req.QueryParameter("status")

	//get cicd env information from database
	selector := make(bson.M)
	selector["service_order_id"] = serviceOrderId
	_, _, opsEnvJson, err := u.Dao.HandleQuery(opsenvCollection, selector, true, bson.M{}, 0, 0, "", "true")

	if err != nil {
		logrus.Errorf("get env to update %s err is %v", serviceOrderId, err)
		response.WriteError(response.ErrDBQuery, resp)
		return
	}

	var env *entity.OpsEnv
	env = new(entity.OpsEnv)
	envout, _ := json.Marshal(opsEnvJson)
	json.Unmarshal(envout, &env)

	//get user token from controller
	token, err := u.getUserToken(env.UserId, env.TenantId)

	if len(token) <= 0 {
		logrus.Errorf("no token for update linker ops env information")
		response.WriteStatusError(http.StatusUnauthorized, response.ErrUnauthorized, resp)
		return
	}

	//call controller for the latest information
	u.updateEnv(env, token)
	if len(status) > 0 {
		env.Status = status
	}

	if err != nil {
		logrus.Errorf("get information from controller fail,  to update groupInstance %s err is %v", env.Id, err)
		response.WriteError(response.ErrUpdateOpsenv, resp)
		return
	}

	envselector := bson.M{}
	envselector["_id"] = bson.ObjectIdHex(env.Id)

	//save the cicd env into database
	_, _, _, err = u.Dao.HandleUpdateById(opsenvCollection, envselector, ConvertOpsEnvToBson(*env))
	if err != nil {
		logrus.Errorf("update env err is %v", err)
		response.WriteError(response.ErrDBUpdate, resp)
		return
	}
	response.WriteResponse(env, resp)
	return
}

// List cicd envs from database
func (u *Resource) EnvsListHandler(req *restful.Request, resp *restful.Response) {
	//analyze the request
	logrus.Debugf("EnvsListHandler is called")

	u.handleListByUser(opsenvCollection, req, resp)
	return
}

//call controller to order linker ops env
func (u *Resource) createEnv(orderString, token string) (err error, serviceGroupInstanceId, serviceOrderId, serviceOfferingInstanceId, userId, tenantId string) {
	logrus.Debugln("body=" + orderString)

	controllerUrl, err := u.Util.ZkClient.GetControllerEndpoint()
	if err != nil {
		logrus.Errorf("get controller endpoint err is %+v", err)
		return
	}
	url := strings.Join([]string{"http://", controllerUrl, "/v1/serviceGroupOrders"}, "")

	logrus.Debugln("Create Service Group Order for Linker Ops url=" + url)
	resp, err := linker_util.Http_post(url, "application/json", token, orderString)
	if err != nil {
		logrus.Errorf("post service group order for linker ops error %s", err.Error())
		return
	}

	logrus.Infof("create Service Group Order for Linker Ops response: %v", resp)
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
func (u *Resource) deleteEnv(orderid, token string) (err error) {
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

func (u *Resource) updateEnv(env *entity.OpsEnv, token string) (err error) {
	instanceId := env.ServiceGroupInstanceId
	controllerUrl, err := u.Util.ZkClient.GetControllerEndpoint()
	if err != nil {
		logrus.Errorf("get controller endpoint err is %+v", err)
		return
	}
	//get service group instance from controller
	url := strings.Join([]string{"http://", controllerUrl, "/v1/groupInstances/", instanceId}, "")
	logrus.Debugln("Get linker ops instance from controller url=" + url)

	resp, err := linker_util.Http_get_resp(url, "application/json", token, "")
	if err != nil {
		logrus.Errorf("Get linker ops instance from controller error %s", err.Error())
		return
	}

	logrus.Infof("Get linker ops instance from controller response: %v", resp)
	defer resp.Body.Close()
	data, _ := ioutil.ReadAll(resp.Body)

	jsonout, err := GetDataFromResponse(data)
	if err != nil {
		logrus.Errorf("get json obj err is %v", err)
		return
	}

	var sgi *conentity.ServiceGroupInstance
	sgi = new(conentity.ServiceGroupInstance)
	json.Unmarshal(jsonout, &sgi)

	env.Status = sgi.LifeCycleStatus

	//get gerrit information from controller
	gerritInstanceIds := conentity.GetAppInstanceIdsFromGroupInstance(sgi, "/linkerops/code/linker-gerrit")
	for _, gerritInstanceId := range gerritInstanceIds {
		dockerContainerIp, mesosSlaveIp, dockerContainerPort, mesosSlaveHostPort, _ := u.getContainerIp(gerritInstanceId, controllerUrl, token)

		env.GerritDockerIP = dockerContainerIp
		env.GerritHttpPort = "8080"

		port := u.findExternalPort(mesosSlaveHostPort, dockerContainerPort, "8080")
		env.GerritInfo = "http://" + mesosSlaveIp + ":" + port
		env.GerritInternalInfo = "http://" + dockerContainerIp + ":8080"
	}

	//get jenkins information from controller
	jenkinsInstanceIds := conentity.GetAppInstanceIdsFromGroupInstance(sgi, "/linkerops/ci/linker-jenkins")
	for _, jenkinsInstanceId := range jenkinsInstanceIds {
		dockerContainerIp, mesosSlaveIp, dockerContainerPort, mesosSlaveHostPort, _ := u.getContainerIp(jenkinsInstanceId, controllerUrl, token)

		env.JenkinsDockerIP = dockerContainerIp
		env.JenkinsHttpPort = "8080"

		port := u.findExternalPort(mesosSlaveHostPort, dockerContainerPort, "8080")
		env.JenkinsInfo = "http://" + mesosSlaveIp + ":" + port
		env.JenkinsInternalInfo = "http://" + dockerContainerIp + ":8080"
	}

	//get mysql information from controller
	mysqlInstanceIds := conentity.GetAppInstanceIdsFromGroupInstance(sgi, "/linkerops/db/linker-mysql")
	for _, mysqlInstanceId := range mysqlInstanceIds {
		dockerContainerIp, mesosSlaveIp, dockerContainerPort, mesosSlaveHostPort, _ := u.getContainerIp(mysqlInstanceId, controllerUrl, token)

		env.MysqlDockerIP = dockerContainerIp

		port := u.findExternalPort(mesosSlaveHostPort, dockerContainerPort, "3306")
		env.MysqlInfo = "mysql://" + mesosSlaveIp + ":" + port
		env.MysqlInternalInfo = "mysql://" + dockerContainerIp + ":3306"
	}

	//get nexus information from controller
	nexusInstanceIds := conentity.GetAppInstanceIdsFromGroupInstance(sgi, "/linkerops/repo/linker-nexus")
	for _, nexusInstanceId := range nexusInstanceIds {
		dockerContainerIp, mesosSlaveIp, dockerContainerPort, mesosSlaveHostPort, _ := u.getContainerIp(nexusInstanceId, controllerUrl, token)

		env.NexusDockerIP = dockerContainerIp
		env.NexusHttpPort = "8081"

		port := u.findExternalPort(mesosSlaveHostPort, dockerContainerPort, "8081")
		env.NexusInfo = "http://" + mesosSlaveIp + ":" + port + "/nexus"
		env.NexusInternalInfo = "http://" + dockerContainerIp + ":8081/nexus"
	}

	//get openldap information from controller
	ldapInstanceIds := conentity.GetAppInstanceIdsFromGroupInstance(sgi, "/linkerops/ldap/linker-openldap")
	for _, ldapInstanceId := range ldapInstanceIds {
		dockerContainerIp, mesosSlaveIp, dockerContainerPort, mesosSlaveHostPort, _ := u.getContainerIp(ldapInstanceId, controllerUrl, token)

		env.LdapDockerIP = dockerContainerIp

		port := u.findExternalPort(mesosSlaveHostPort, dockerContainerPort, "389")
		env.LdapInfo = "ldap://" + mesosSlaveIp + ":" + port
		env.LdapInternalInfo = "ldap://" + dockerContainerIp + ":389"
	}
	return
}

//get container ip from controller
func (u *Resource) getContainerIp(appInstanceId, controllerUrl, token string) (containerIp, mesosSlaveIp, dockerContainerPort, mesosSlaveHostPort string, err error) {
	url := strings.Join([]string{"http://", controllerUrl, "/v1/appInstances/", appInstanceId}, "")
	logrus.Debugln("url: " + url)
	resp, err := linker_util.Http_get(url, "application/json", token, "")

	if err != nil {
		logrus.Errorf("Get container instance from controller error %s", err.Error())
		return
	}

	respData := map[string]interface{}{}
	result := json.NewDecoder(strings.NewReader(string(resp)))
	result.Decode(&respData)

	jq := jsonq.NewQuery(respData)
	containerIp, _ = jq.String("data", "docker_container_ip")
	mesosSlaveIp, _ = jq.String("data", "mesos_slave")
	dockerContainerPort, _ = jq.String("data", "docker_container_port")
	mesosSlaveHostPort, _ = jq.String("data", "mesos_slave_host_port")
	return
}

//get user token from controller
func (u *Resource) getUserToken(userId, tenantId string) (token string, err error) {
	logrus.Debugln("userId=" + userId)
	logrus.Debugln("tenantId=" + tenantId)

	adminInfo := "{\"email\":\"" + u.Util.Props.MustGet("controller.admin.name") + "\",\"password\":\"" + u.Util.Props.MustGet("controller.admin.password") + "\"}"

	controllerUrl, err := u.Util.ZkClient.GetControllerEndpoint()
	if err != nil {
		logrus.Errorf("get controller endpoint err is %+v", err)
		return
	}
	url := strings.Join([]string{"http://", controllerUrl, "/v1/user/login"}, "")

	logrus.Debugln("Admin loging url=" + url)
	resp, err := linker_util.Http_post(url, "application/json", token, adminInfo)
	if err != nil {
		logrus.Errorf("admin login error %s", err.Error())
		return
	}

	logrus.Infof("Get Admin Token response: %v", resp)
	jsondata := map[string]interface{}{}
	dec := json.NewDecoder(strings.NewReader(resp))
	dec.Decode(&jsondata)
	jq := jsonq.NewQuery(jsondata)
	admintoken, err := jq.String("data", "id")

	//TODO get user token
	userInfo := "{\"tenant_id\":\"" + tenantId + "\",\"user_id\":\"" + userId + "\"}"

	url = strings.Join([]string{"http://", controllerUrl, "/v1/token/regenerate"}, "")

	logrus.Debugln("User token url=" + url)
	resp, err = linker_util.Http_post(url, "application/json", admintoken, userInfo)
	if err != nil {
		logrus.Errorf("admin login error %s", err.Error())
		return
	}

	logrus.Infof("Get User Token: %v", resp)
	jsondata = map[string]interface{}{}
	dec = json.NewDecoder(strings.NewReader(resp))
	dec.Decode(&jsondata)
	jq = jsonq.NewQuery(jsondata)
	token, err = jq.String("data", "id")
	return
}

func (u *Resource) findOpsEnvById(opsEnvId string) (opsEnv *entity.OpsEnv, err error) {
	selector := bson.M{}
	selector[ParamID] = bson.ObjectIdHex(opsEnvId)
	_, _, opsEnvJson, err := u.Dao.HandleQuery(opsenvCollection, selector, true, bson.M{}, 0, 1, "", "true")

	opsEnv = new(entity.OpsEnv)
	envout, _ := json.Marshal(opsEnvJson)
	json.Unmarshal(envout, &opsEnv)
	return
}

//terminate projects ,if ok ,terminate opsenv
func (u *Resource) terminateOpsEnv(opsEnvId string) (err error) {
	//## This function will do the following
	//terminate project envs
	//terminate jobs
	//terminate projects
	err = u.terminateProjects(opsEnvId)
	if err != nil {
		logrus.Errorf("Error terminate projects,reason:%v,opsenv id:%s", err, opsEnvId)
		return
	} else {
		//terminate opsenv
		change := bson.M{"status": entity.OPSENV_STATUS_TERMINATED}
		selector := bson.M{}
		selector[ParamID] = bson.ObjectIdHex(opsEnvId)
		err = u.Dao.HandleUpdateByQueryPartial(opsenvCollection, selector, change)
		return
	}
}

func (u *Resource) findExternalPort(mesosPorts, dockerPorts, internalPort string) (port string) {
	mesosPortsArray := strings.Split(mesosPorts, ",")
	dockerPortsArray := strings.Split(dockerPorts, ",")
	port = ""
	if len(mesosPortsArray) == len(dockerPortsArray) {
		for index, value := range dockerPortsArray {
			if internalPort == value {
				port = mesosPortsArray[index]
				break
			}
		}
	} else {
		logrus.Errorf("External and internal ports aren't same length:%s vs %s", mesosPortsArray, dockerPortsArray)
	}
	return port
}
