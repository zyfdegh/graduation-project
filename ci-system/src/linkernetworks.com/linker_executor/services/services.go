// dockerservice
package services

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/fsouza/go-dockerclient"
	"io/ioutil"
	"linkernetworks.com/linker_common_lib/httpclient"
	"linkernetworks.com/linker_common_lib/persistence/entity"
	resp "linkernetworks.com/linker_common_lib/rest/response"
	"linkernetworks.com/linker_executor/command"
	"strconv"
	"strings"
)

var (
	LIFECYCLT_INIT           = "INITED"
	LIFECYCLT_NEW            = "CREATED"
	LIFECYCLT_CONFIG_SUCCESS = "CONFIGED"
	LIFECYCLT_CONFIG_FAILED  = "UNCONFIGED"
	OPERATION_CREATE         = "CREATE"
	OPERATION_TERMINATE      = "TERMINATE"
	DockerEndpoint           string
	Zk                       *ZkClient
	Token                    string
)

const (
	Docker    = "Docker"
	Openstack = "Openstack"
)

func Terminate(appInfo entity.AppContainerInstance) (err error) {

	// call appInstance terminate api
	controllerUrl, err := Zk.GetControllerEndpoint()
	if err != nil {
		logrus.Errorf("get controller endpoint err is %+v", err)
		return
	}
	url := strings.Join([]string{"http://", controllerUrl, "/v1/appInstances/", appInfo.ObjectId.Hex()}, "")
	logrus.Debugln("TerminateAppInstance url=" + url)

	resp, err := httpclient.Http_delete(url, "", httpclient.Header{"Content-Type", "application/json"}, httpclient.Header{"X-Auth-Token", Token})
	defer resp.Body.Close()
	if err != nil {
		logrus.Errorf("delete appInfo error %s", err.Error())
		return
	}
	responseBody, _ := ioutil.ReadAll(resp.Body)
	logrus.Debugf("terminate appInstance response is %v", string(responseBody))
	if resp.StatusCode >= 400 {
		// means error
		err = errors.New("delete appInfo error, error is " + string(responseBody))
		logrus.Errorf("delete appInfo error %s", err.Error())
		return
	}
	return
}

func InitAppInstance(app *entity.App, appId, containerName string) (appInfo entity.AppContainerInstance, err error) {
	appInfo.LifeCycleStatus = LIFECYCLT_INIT
	appInfo.DockerContainerPort = ""
	appInfo.MesosSlaveHostPort = ""
	appInfo.ServiceGroupId = app.Env["LINKER_SERVICE_GROUP_ID"]
	appInfo.ServiceGroupInstanceId = app.Env["LINKER_SERVICE_GROUP_INSTANCE_ID"]
	appInfo.ServiceOrderId = app.Env["LINKER_SERVICE_ORDER_ID"]
	appInfo.MarathonAppPath = appId
	appInfo.DockerContainerName = containerName

	// reset app container id
	appPath := strings.Split(appId, "/")
	appIds := append(strings.Split(appInfo.ServiceGroupId, "/"), appPath[2:]...)
	appInfo.AppContainerId = strings.Join(appIds, "/")
	// appInfo.AppContainerId = appId

	appInstanceId, aci, err := CreateAppInstance(appInfo)
	if err != nil {
		logrus.Errorf("create appInstance err: %s", err.Error())
		return
	}
	logrus.Debugln("appInstanceId=" + appInstanceId)
	return aci, err
}

func AllocateIP(appInfo entity.AppContainerInstance) (allocatedip string, err error) {
	controllerUrl, err := Zk.GetControllerEndpoint()
	if err != nil {
		logrus.Errorf("get controller endpoint err is %+v", err)
		return
	}
	url := strings.Join([]string{"http://", controllerUrl, "/v1/appInstances/", appInfo.ObjectId.Hex(), "/allocate?resource=ipAddressResource:[ipAddress]"}, "")
	logrus.Debugln("AllocateIP url=" + url)
	response, err := httpclient.Http_get(url, "", httpclient.Header{"Content-Type", "application/json"}, httpclient.Header{"X-Auth-Token", Token})
	defer response.Body.Close()
	if err != nil {
		logrus.Errorf("allocte ip error %s", err.Error())
		return
	}

	// Read a document from response
	document := map[string]interface{}{}
	// Handle JSON parsing manually here, instead of relying on go-restful's
	// req.ReadEntity. This is because ReadEntity currently parses JSON with
	// UseNumber() which turns all numbers into strings. See:
	// https://github.com/emicklei/mora/pull/31
	decoder := json.NewDecoder(response.Body)
	err = decoder.Decode(&document)
	if err != nil {
		logrus.Errorf("decode payload err is %v", err)
		return
	}
	var alocateipresp resp.Response
	jsonout, err := json.Marshal(document)
	if err != nil {
		logrus.Errorf("marshal token err is %v", err)
		return
	}

	err = json.Unmarshal(jsonout, &alocateipresp)
	if !alocateipresp.Success {
		err = errors.New(alocateipresp.Error.ErrorMsg)
		return
	}

	allocatedip = alocateipresp.Data.(string)

	return allocatedip, err
}

func Update(appInfo *entity.AppContainerInstance, containerlongid string, mesosSlaveIp string, cpu float32, mem int16) (err error) {
	logrus.Debugf("update is called to update appInfo [%v]", appInfo)
	// get docker container info
	container, err := GetDockerEnv(appInfo.DockerContainerName)
	if err != nil {
		logrus.Errorf("inspect docker container %s err: %s", appInfo.DockerContainerName, err.Error())
		return
	}
	// generate appInfo json, set status to CREATED
	SetAppInfo(appInfo, container)
	containerIp, err := GetDockerIP(appInfo.DockerContainerName)
	if err != nil {
		logrus.Errorf("get docker container %s ip err: %s", appInfo.DockerContainerName, err.Error())
		return
	}
	appInfo.DockerContainerIp = containerIp
	appInfo.DockerContainerLongID = containerlongid
	appInfo.MesosSlaveIp = mesosSlaveIp
	appInfo.Cpus = cpu
	appInfo.Mem = mem

	// create appInstance to controller:/v1/appInstances
	err = UpdateAppInstance(appInfo, appInfo.ObjectId.Hex())
	if err != nil {
		logrus.Errorf("update appInstance err: %s", err.Error())
	}
	return
}

func GetConfigSteps(appId string) (configSteps []entity.ConfigStep, err error) {
	controllerUrl, err := Zk.GetControllerEndpoint()
	if err != nil {
		logrus.Errorf("get controller endpoint err is %+v", err)
		return
	}
	url := strings.Join([]string{"http://", controllerUrl, "/v1/appInstances/", appId, "/steps"}, "")
	logrus.Debugln("GetConfigSteps url=" + url)
	response, err := httpclient.Http_get(url, "", httpclient.Header{"Content-Type", "application/json"}, httpclient.Header{"X-Auth-Token", Token})
	defer response.Body.Close()
	if err != nil {
		logrus.Errorf("get config step error %s", err.Error())
		return
	}

	// Read a document from response
	document := map[string]interface{}{}
	// Handle JSON parsing manually here, instead of relying on go-restful's
	// req.ReadEntity. This is because ReadEntity currently parses JSON with
	// UseNumber() which turns all numbers into strings. See:
	// https://github.com/emicklei/mora/pull/31
	decoder := json.NewDecoder(response.Body)
	err = decoder.Decode(&document)
	if err != nil {
		logrus.Errorf("decode payload err is %v", err)
		return
	}
	var aciresp resp.Response
	jsonout, err := json.Marshal(document)
	if err != nil {
		logrus.Errorf("marshal token err is %v", err)
		return
	}

	err = json.Unmarshal(jsonout, &aciresp)
	if !aciresp.Success {
		err = errors.New(aciresp.Error.ErrorMsg)
		return
	}

	stepObjects := aciresp.Data.([]interface{})
	var steps = make([]entity.ConfigStep, len(stepObjects))
	for i, step := range stepObjects {
		stepmap := step.(map[string]interface{})
		steps[i].ConfigType = stepmap["config_type"].(string)
		steps[i].Execute = stepmap["execute"].(string)
		steps[i].Scope = stepmap["scope"].(string)
	}
	return steps, err
}

func ConfigDockerInstance(appInfo *entity.AppContainerInstance) (configStatus bool, err error) {
	// call controller's config API:/v1/appContainers/steps?appInstanceId=5591038575461d3d44000002
	configSteps, err := GetConfigSteps(appInfo.ObjectId.Hex())
	if err != nil {
		logrus.Errorf("get config step err: %s", err.Error())
		return
	}
	// logrus.Debugln("configSteps=" + strings.Join(configSteps, ","))
	// execute commands returned by controller
	configStatus = true
	for _, step := range configSteps {
		if step.ConfigType == "command" {
			logrus.Debugln("execute step: " + step.Execute)
			retstring, errstring, err := command.ExecCommand(step.Execute)
			if err != nil {
				logrus.Errorf("execute command %s error, stdout : %s,  stderr: %s", step.Execute, retstring, errstring, err.Error())
				configStatus = false
			}
		} else if step.ConfigType == "docker" {
			logrus.Debugln("execute step: " + step.Execute)
			execInsp, err := DoDockerExec(appInfo.DockerContainerName, step.Execute)
			if err != nil {
				logrus.Errorf("execute command %s err: %s", step.Execute, err.Error())
				configStatus = false
			}
			if execInsp.ExitCode != 0 {
				logrus.Errorf("execute command %s err: %s", step.Execute, "")
				configStatus = false
			}
		} else {
			logrus.Errorf("unsupported config type : " + step.ConfigType)
			configStatus = false
		}
	}

	// regenerate appInfo, may add ip addr
	containerIp, err := GetDockerIP(appInfo.DockerContainerName)
	if err != nil {
		logrus.Errorf("get ip err: %s", err.Error())
		return
	}
	appInfo.DockerContainerIp = containerIp
	// send config status to controller
	logrus.Debugln("config status=" + strconv.FormatBool(configStatus))
	if configStatus {
		appInfo.LifeCycleStatus = LIFECYCLT_CONFIG_SUCCESS
	} else {
		appInfo.LifeCycleStatus = LIFECYCLT_CONFIG_FAILED
	}
	err = UpdateAppInstance(appInfo, appInfo.ObjectId.Hex())
	if err != nil {
		logrus.Errorf("change appInstance status err: %s", err.Error())
	}
	return
}

func CreateAppInstance(appInfo entity.AppContainerInstance) (appInstanceId string, aci entity.AppContainerInstance, err error) {
	body, _ := json.Marshal(appInfo)
	logrus.Debugln("body=" + string(body))

	controllerUrl, err := Zk.GetControllerEndpoint()
	if err != nil {
		logrus.Errorf("get controller endpoint err is %+v", err)
		return
	}
	url := strings.Join([]string{"http://", controllerUrl, "/v1/appInstances"}, "")

	logrus.Debugln("CreateAppInstance url=" + url)
	response, err := httpclient.Http_post(url, string(body), httpclient.Header{"Content-Type", "application/json"}, httpclient.Header{"X-Auth-Token", Token})
	defer response.Body.Close()
	if err != nil {
		logrus.Errorf("post appInfo error %s", err.Error())
		return
	}

	// Stub an app to be populated from the body
	aciresp := resp.Response{}

	// Populate the user data
	err = json.NewDecoder(response.Body).Decode(&aciresp)

	if !aciresp.Success {
		err = errors.New(aciresp.Error.ErrorMsg)
		return
	}

	aci = entity.AppContainerInstance{}

	aciMap := aciresp.Data.(map[string]interface{})

	acijson, err := json.Marshal(aciMap)
	if err != nil {
		logrus.Errorf("marshal aci err is %v", err)
		return
	}

	err = json.Unmarshal(acijson, &aci)

	// appInstanceId = aciMap["id"].(string)
	appInstanceId = aci.ObjectId.Hex()
	return
}

func UpdateAppInstance(appInfo *entity.AppContainerInstance, appInstanceId string) (err error) {
	body, _ := json.Marshal(appInfo)

	logrus.Infoln("update appInstance: " + string(body))
	controllerUrl, err := Zk.GetControllerEndpoint()
	if err != nil {
		logrus.Errorf("get controller endpoint err is %+v", err)
		return
	}
	url := strings.Join([]string{"http://", controllerUrl, "/v1/appInstances", "/", appInstanceId}, "")

	logrus.Debugln("UpdateAppInstance url=" + url)
	_, err = httpclient.Http_put(url, string(body), httpclient.Header{"Content-Type", "application/json"}, httpclient.Header{"X-Auth-Token", Token})
	return
}

func GetDockerEnv(dockername string) (container *docker.Container, err error) {
	// endpoint := "unix:///var/run/docker.sock"
	client, _ := docker.NewClient(DockerEndpoint)

	logrus.Debugln("dockername=" + dockername)
	container, err = client.InspectContainer(dockername)
	return
}

func GetDockerIP(dockername string) (ipAddr string, err error) {
	// create exec
	command := "ip addr show " + Zk.Props.GetString("docker.interface.name", "eth1") +
		" | grep inet | grep -v 127.0.0.1 | grep -v inet6 | awk '{print $2}' | awk -F '/' '{print $1}'"

	client, _ := docker.NewClient(DockerEndpoint)

	logrus.Debugln("dockername=" + dockername + ", command=" + command)
	createOptions := docker.CreateExecOptions{}
	createOptions.Container = dockername
	createOptions.Cmd = []string{"bash", "-c", command}
	createOptions.AttachStdin = false
	createOptions.AttachStdout = true
	createOptions.AttachStderr = true
	createOptions.Tty = false
	exec, err := client.CreateExec(createOptions)
	if err != nil {
		logrus.Errorf("create exec error %v", err)
		return
	}
	// start exec
	var stdout, stderr bytes.Buffer
	startOptions := docker.StartExecOptions{
		Detach:       false,
		Tty:          false,
		OutputStream: &stdout,
		ErrorStream:  &stderr,
	}
	err = client.StartExec(exec.ID, startOptions)
	if err != nil {
		logrus.Error(err)
		return
	}

	execInsp, err := client.InspectExec(exec.ID)
	if err != nil {
		logrus.Error(err)
		return
	}

	if execInsp.ExitCode != 0 {
		logrus.Error("stderr=" + stderr.String())
	}

	// ipAddr = stdout.String()
	ipAddr = strings.TrimSuffix(stdout.String(), "\n")
	logrus.Infof("get docker ip: dockername=%s, ip=%s", dockername, ipAddr)

	return
}

func DoDockerExec(dockername string, command string) (execInsp *docker.ExecInspect, err error) {

	//dockerEndpoint := "unix:///var/run/docker.sock"
	client, _ := docker.NewClient(DockerEndpoint)

	logrus.Infof("exec command=" + command)
	createOptions := docker.CreateExecOptions{
		Container:    dockername,
		Cmd:          strings.Split(strings.TrimSpace(command), " "),
		AttachStdin:  false,
		AttachStdout: true,
		AttachStderr: true,
		Tty:          false,
	}
	exec, err := client.CreateExec(createOptions)
	logrus.Infof("create exec=" + exec.ID)
	if err != nil {
		logrus.Errorf("create exec error %v", err)
		return
	}
	var stdout, stderr bytes.Buffer
	startOptions := docker.StartExecOptions{
		Detach:       false,
		Tty:          false,
		OutputStream: &stdout,
		ErrorStream:  &stderr,
	}
	logrus.Infof("start exec=" + exec.ID)
	err = client.StartExec(exec.ID, startOptions)
	if err != nil {
		logrus.Errorf("start exec error %v", err)
		return
	}
	logrus.Infof("end exec=" + exec.ID)
	logrus.Infof("exec command out: %v", stdout.String())
	logrus.Infof("exec command err: %v", stderr.String())

	execInsp, err = client.InspectExec(exec.ID)
	if err != nil {
		logrus.Errorf("inspect exec error %v", err)
		return
	}
	logrus.Debugln("ID=" + execInsp.Container.ID)
	logrus.Debugln("ExitCode=" + strconv.Itoa(execInsp.ExitCode))
	logrus.Debugln("OpenStderr=" + strconv.FormatBool(execInsp.OpenStderr))
	logrus.Debugln("OpenStdin=" + strconv.FormatBool(execInsp.OpenStdin))
	logrus.Debugln("OpenStdout=" + strconv.FormatBool(execInsp.OpenStdout))
	logrus.Debugln("Running=" + strconv.FormatBool(execInsp.Running))
	logrus.Debugln("EntryPoint=" + execInsp.ProcessConfig.EntryPoint)
	logrus.Debugln("Arguments=" + strings.Join(execInsp.ProcessConfig.Arguments, " "))
	logrus.Debugln("User=" + execInsp.ProcessConfig.User)
	return
}

func SetAppInfo(appInfo *entity.AppContainerInstance, container *docker.Container) (err error) {
	appInfo.DockerContainerIp = container.NetworkSettings.IPAddress
	appInfo.LifeCycleStatus = LIFECYCLT_NEW
	// logrus.Debugln("ports=" + container.NetworkSettings.Ports)
	appInfo.DockerContainerPort = ""
	appInfo.MesosSlaveHostPort = ""
	for containerPort, portBinding := range container.NetworkSettings.Ports {
		logrus.Debugln("containerPort=" + containerPort)
		appInfo.DockerContainerPort = appInfo.DockerContainerPort + containerPort.Port() + ","
		if len(portBinding) > 0 {
			appInfo.MesosSlaveHostPort = appInfo.MesosSlaveHostPort + portBinding[0].HostPort + ","
		}
	}
	// remove the last comma if exist
	logrus.Debugln("appInfo.ContainerPort=" + appInfo.DockerContainerPort + "; appInfo.HostPort=" + appInfo.MesosSlaveHostPort)
	appInfo.DockerContainerPort = strings.TrimSuffix(appInfo.DockerContainerPort, ",")
	appInfo.MesosSlaveHostPort = strings.TrimSuffix(appInfo.MesosSlaveHostPort, ",")
	logrus.Debugln("appInfo.ContainerPort=" + appInfo.DockerContainerPort + "; appInfo.HostPort=" + appInfo.MesosSlaveHostPort)

	//volume mapping
	appInfo.Volumes = container.Volumes

	envs := container.Config.Env
	for _, env := range envs {
		envArr := strings.Split(env, "=")
		switch envArr[0] {
		case "MARATHON_APP_ID":
			// appInfo.AppContainerId = envArr[1]
			appInfo.MarathonAppPath = envArr[1]
		case "MARATHON_APP_VERSION":
			appInfo.MarathonAppVersion = envArr[1]
		case "LINKER_SERVICE_GROUP_ID":
			appInfo.ServiceGroupId = envArr[1]
		case "LINKER_SERVICE_GROUP_INSTANCE_ID":
			appInfo.ServiceGroupInstanceId = envArr[1]
		case "LINKER_SERVICE_ORDER_ID":
			appInfo.ServiceOrderId = envArr[1]
		case "HOST":
			appInfo.MesosSlave = envArr[1]
		case "MESOS_TASK_ID":
			appInfo.MesosTaskId = envArr[1]
		case "MESOS_SANDBOX":
			appInfo.MesosSandbox = envArr[1]
		default:
			// logrus.Debugf("no need.")
		}
	}
	// reset appContainerId
	appPath := strings.Split(appInfo.MarathonAppPath, "/")
	appId := append(strings.Split(appInfo.ServiceGroupId, "/"), appPath[2:]...)
	appInfo.AppContainerId = strings.Join(appId, "/")
	return
}

func GetSgFromResponse(data []byte) (jsonout []byte, err error) {
	// create serviceGroupOrder
	var response *resp.Response
	response = new(resp.Response)
	err = json.Unmarshal(data, &response)
	if err != nil {
		return
	}

	var jmap map[string]interface{}

	switch response.Data.(type) {
	case []interface{}:
		jmaps := response.Data.([]interface{})
		jmap = jmaps[0].(map[string]interface{})
	case interface{}:
		jmap = response.Data.(map[string]interface{})
	}
	jsonout, err = json.Marshal(jmap)
	if err != nil {
		fmt.Println(err)
		return
	}
	return
}

func CheckContainerType(offeringId, instanceId, appPath string) (iswhat string, app *entity.App, err error) {
	logrus.Infof("appPath: %s", appPath)
	logrus.Infof("offeringId: %s", offeringId)
	logrus.Infof("instanceId: %s", instanceId)

	controllerUrl, err := Zk.GetControllerEndpoint()
	if err != nil {
		logrus.Errorf("get controller endpoint err is %+v", err)
		return
	}

	url := strings.Join([]string{"http://", controllerUrl, "/v1/serviceOfferings/", offeringId, "/containerinfo?serviceGroupInstanceId=", instanceId, "&appId=", appPath}, "")
	logrus.Infof("url: %s", url)

	//var respBody string
	response, _ := httpclient.Http_get(url, "", httpclient.Header{"Content-Type", "application/json"}, httpclient.Header{"X-Auth-Token", Token})
	defer response.Body.Close()
	data, _ := ioutil.ReadAll(response.Body)

	if response.StatusCode >= 400 {
		logrus.Errorf("get App from offering error %v", string(data))
		err = errors.New("get App from offering error.")
		return
	}
	logrus.Debugf("body: " + string(data))
	jsonout, err := GetSgFromResponse(data)
	if err != nil {
		logrus.Errorf("get json obj err is %v", err)
		return
	}
	app = new(entity.App)
	err = json.Unmarshal(jsonout, app)
	if err != nil {
		logrus.Errorf("Unmarshal json obj err is %v", err)
		return
	}

	var isDocker interface{} = app.Container.Type
	if isDocker != nil && isDocker != "" {
		iswhat = Docker
	} else {
		iswhat = Openstack
	}
	return
}

func InitToken() (err error) {
	user := entity.User{}
	user.Email = Zk.Props.MustGet("usermgmt.username")
	user.Username = Zk.Props.MustGet("usermgmt.username")
	user.Password = Zk.Props.MustGet("usermgmt.password")
	user.Tenantname = Zk.Props.MustGet("usermgmt.tenantname")

	body, _ := json.Marshal(user)
	logrus.Debugln("body=" + string(body))

	//init user management endpoints
	userMgmtHA := Zk.Props.GetBool("usermgmt.ha", true)

	var userMgmtUrl string

	if userMgmtHA == true {
		//user management is true, check the zookepper
		userMgmtUrl, err = Zk.GetUserMgmtEndpoint()
		if err != nil {
			logrus.Errorf("get usermgmt endpoint err is %+v", err)
			return
		}
	} else {
		//user management is false, just use the endpoint in propertiey file
		userMgmtUrl = Zk.Props.MustGet("usermgmt.endpoint")
	}

	url := strings.Join([]string{"http://", userMgmtUrl, "/v1/token"}, "")
	fmt.Println("url: " + url)
	response, err := httpclient.Http_post(url, string(body), httpclient.Header{"Content-Type", "application/json"})
	defer response.Body.Close()
	if err != nil {
		logrus.Errorf("post to usermgmt error %s", err.Error())
		return
	}

	// Read a document from response
	document := map[string]interface{}{}
	// Handle JSON parsing manually here, instead of relying on go-restful's
	// req.ReadEntity. This is because ReadEntity currently parses JSON with
	// UseNumber() which turns all numbers into strings. See:
	// https://github.com/emicklei/mora/pull/31
	decoder := json.NewDecoder(response.Body)
	err = decoder.Decode(&document)
	if err != nil {
		logrus.Errorf("decode payload err is %v", err)
		return
	}
	var tokenresp resp.Response
	jsonout, err := json.Marshal(document)
	if err != nil {
		logrus.Errorf("marshal token err is %v", err)
		return
	}

	err = json.Unmarshal(jsonout, &tokenresp)
	if !tokenresp.Success {
		err = errors.New(tokenresp.Error.ErrorMsg)
		return
	}

	token := tokenresp.Data.(map[string]interface{})
	Token = token["id"].(string)
	return
}
