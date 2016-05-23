package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"github.com/Sirupsen/logrus"
	"github.com/jmoiron/jsonq"
	"github.com/magiconair/properties"
	exec "github.com/mesos/mesos-go/executor"
	mesos "github.com/mesos/mesos-go/mesosproto"
	"github.com/rackspace/gophercloud"
	"github.com/rackspace/gophercloud/openstack"
	"github.com/rackspace/gophercloud/openstack/compute/v2/servers"
	"gopkg.in/mgo.v2/bson"
	"linkernetworks.com/linker_common_lib/persistence/entity"
	linkercommand "linkernetworks.com/linker_executor/command"
	linkerclient "linkernetworks.com/linker_executor/services"
	"math"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	propertiesFile            = flag.String("config", "executor.properties", "the configuration file")
	zkFlag                    = flag.String("zk", "", "zookeeper url")
	zkUrl                     string
	hostname                  string
	controllerUrl             string
	dockerEndpoint            string
	openstackIdentityEndpoint string
	openstackUsername         string
	openstackPassword         string
	openstackTenandId         string
	openstackRegion           string
	dockerStartCheckTime      int
	MESOS_SANDBOX             string
	props                     *properties.Properties
	isWhat                    string
	appInfo                   entity.AppContainerInstance
	taskKilled                bool
	taskKilledMutex           sync.Mutex
)

const (
	HOST_IP = "HOST_IP"
)

type LinkerDockerExecutor struct {
	tasksLaunched int
}

func newExampleExecutor() *LinkerDockerExecutor {
	return &LinkerDockerExecutor{tasksLaunched: 0}
}

func setTaskKilled() {
	taskKilledMutex.Lock()
	taskKilled = true
	taskKilledMutex.Unlock()
}

func (exec *LinkerDockerExecutor) Registered(driver exec.ExecutorDriver, execInfo *mesos.ExecutorInfo, fwinfo *mesos.FrameworkInfo, slaveInfo *mesos.SlaveInfo) {
	hostname = slaveInfo.GetHostname()
	logrus.Infoln("Registered Executor on slave ", slaveInfo.GetHostname())
}

func (exec *LinkerDockerExecutor) Reregistered(driver exec.ExecutorDriver, slaveInfo *mesos.SlaveInfo) {
	logrus.Infoln("Re-registered Executor on slave ", slaveInfo.GetHostname())
}

func (exec *LinkerDockerExecutor) Disconnected(exec.ExecutorDriver) {
	logrus.Infoln("Executor disconnected.")
}

func getValueFromEnv(envData map[string]interface{}, key string) (value string, err error) {
	for k, v := range envData {
		// logrus.Infoln("find env: " + key)
		if k == key {
			value = v.(string)
			logrus.Infoln("find env " + key + "=" + value)
			return
		}
	}
	err = errors.New("Can not find env " + key)
	return
}

func (exec *LinkerDockerExecutor) LaunchTask(driver exec.ExecutorDriver, taskInfo *mesos.TaskInfo) {
	go func() {
		// notifyStatus(mesos.TaskState_TASK_RUNNING, driver, taskInfo)
		exec.launchNewTask(driver, taskInfo)
	}()
}

func (exec *LinkerDockerExecutor) launchNewTask(driver exec.ExecutorDriver, taskInfo *mesos.TaskInfo) {
	logrus.Infoln("Launching task", taskInfo.GetName(), "with command", taskInfo.Command.GetValue())

	data := string(taskInfo.GetData()[:len(taskInfo.GetData())])
	//logrus.Infoln("Data: " + data)

	// Mesos task id
	taskId := taskInfo.GetTaskId()
	if taskId != nil {
		logrus.Infoln("TaskId: " + taskId.GetValue())
	}

	// Parse data: marathon json data
	jsondata := map[string]interface{}{}

	result := json.NewDecoder(strings.NewReader(data))
	result.Decode(&jsondata)

	jq := jsonq.NewQuery(jsondata)

	// id
	appId, _ := jq.String("id")
	logrus.Infoln(appId)

	// container name
	newId := bson.NewObjectId().Hex()
	appTmpId := strings.Replace(appId, "/", "_", -1)
	containerName := "mesos" + appTmpId + "-" + newId

	// be careful here, in appOjb, only Container or Openstack has value here.
	// get group id from env
	envMap, _ := jq.Object("env")
	offeringId, err := getValueFromEnv(envMap, "LINKER_SERVICE_OFFERING_ID")
	if err != nil {
		logrus.Errorf("get LINKER_SERVICE_OFFERING_ID from env error is %v", err)
		notifyStatus(mesos.TaskState_TASK_FAILED, driver, taskInfo)
		exec.Shutdown(driver)
		return
	}

	instanceId, err := getValueFromEnv(envMap, "LINKER_SERVICE_GROUP_INSTANCE_ID")
	if err != nil {
		logrus.Errorf("get LINKER_SERVICE_GROUP_INSTANCE_ID from env error is %v", err)
		notifyStatus(mesos.TaskState_TASK_FAILED, driver, taskInfo)
		exec.Shutdown(driver)
		return
	}

	// init token
	err = linkerclient.InitToken()
	if err != nil {
		logrus.Errorf("init token error is %v", err)
		notifyStatus(mesos.TaskState_TASK_FAILED, driver, taskInfo)
		exec.Shutdown(driver)
		return
	}

	containerType, appObj, err := linkerclient.CheckContainerType(offeringId, instanceId, appId)
	if err != nil {
		logrus.Infoln(err)
		notifyStatus(mesos.TaskState_TASK_FAILED, driver, taskInfo)
		exec.Shutdown(driver)
		return
	}
	isWhat = containerType

	appInfo, err = linkerclient.InitAppInstance(appObj, appId, containerName)
	if err != nil {
		logrus.Errorf("init app info error is %v", err)
		notifyStatus(mesos.TaskState_TASK_FAILED, driver, taskInfo)
		exec.Shutdown(driver)
		return
	}
	logrus.Debugf("INITED: appInfo=", appInfo)

	if linkerclient.Openstack == containerType {
		// Create Openstance instance ...
		logrus.Infoln("image: " + appObj.Openstack.Image + " flavor: " + appObj.Openstack.Flavor)
		opts := gophercloud.AuthOptions{
			IdentityEndpoint: openstackIdentityEndpoint,
			Username:         openstackUsername,
			Password:         openstackPassword,
			TenantID:         openstackTenandId,
		}

		provider, openstackError := openstack.AuthenticatedClient(opts)

		if openstackError != nil {
			logrus.Infoln("login into openstack error!")
			notifyStatus(mesos.TaskState_TASK_FAILED, driver, taskInfo)
			exec.Shutdown(driver)
			return
		}

		client, openstackError := openstack.NewComputeV2(provider, gophercloud.EndpointOpts{
			Region: openstackRegion,
		})

		if openstackError != nil {
			logrus.Infoln("New openstack compute client error!")
			notifyStatus(mesos.TaskState_TASK_FAILED, driver, taskInfo)
			exec.Shutdown(driver)
			return
		}

		server, openstackError := servers.Create(client, servers.CreateOpts{
			Name:       containerName,
			FlavorName: appObj.Openstack.Flavor,
			ImageName:  appObj.Openstack.Image,
		}).Extract()

		if openstackError != nil {
			logrus.Infoln("Create openstack instance error!")
			notifyStatus(mesos.TaskState_TASK_FAILED, driver, taskInfo)
			exec.Shutdown(driver)
			return
		}

		logrus.Infoln(server.ID + " " + server.Name + " " + server.Status)

		linkerclient.Update(&appInfo, "", "", 0, 0)
	} else {
		allocatedip, err := linkerclient.AllocateIP(appInfo)
		logrus.Debugf("allocated ip address is %v", allocatedip)
		if err != nil {
			logrus.Errorf("allocate ip error is %v", err)
			notifyStatus(mesos.TaskState_TASK_FAILED, driver, taskInfo)
			exec.Shutdown(driver)
			return
		}
		// Create Docker instance ...
		var commandTextBuffer bytes.Buffer
		commandTextBuffer.WriteString("docker run -d ")

		// Set memory
		memV, _ := jq.Int("mem")
		memValue := int16(memV)
		logrus.Infoln(memValue)
		if memValue > 0 {
			commandTextBuffer.WriteString("-m " + strconv.Itoa(memV) + "m ")
		}

		// Set cpu
		cpuV, _ := jq.Float("cpus")
		cpuValue := float32(cpuV)
		logrus.Infoln(cpuValue)
		if cpuValue > 0 {
			commandTextBuffer.WriteString("-c " + strconv.Itoa(calCpus(cpuV)) + " ")
		}

		// Set env
		envData, _ := jq.Object("env")

		serviceGroupId := ""
		// Map to storage env data.
		var envs = make(map[string]string)
		for key, _ := range envData {
			logrus.Infoln(key + ":")
			envvalue, _ := jq.String("env", key)
			logrus.Infoln(envvalue)

			if key == "LINKER_SERVICE_GROUP_ID" {
				serviceGroupId = envvalue
			}

			envs[key] = strings.TrimSpace(envvalue)
			commandTextBuffer.WriteString("-e '" + strings.TrimSpace(key) + "=" + strings.TrimSpace(envvalue) + "' ")
		}

		// set some parameters in container's env
		// MARATHON_APP_ID, MARATHON_APP_VERSION, HOST, MESOS_TASK_ID, MESOS_SANDBOX
		version, _ := jq.String("version")
		logrus.Infoln("version: " + version)
		commandTextBuffer.WriteString("-e 'MARATHON_APP_ID=" + appId + "' ")
		commandTextBuffer.WriteString("-e 'MARATHON_APP_VERSION=" + version + "' ")

		if hostname == "" {
			hostname, _, _ = linkercommand.ExecCommand("hostname")
		}
		commandTextBuffer.WriteString("-e 'HOST=" + hostname + "' ")
		commandTextBuffer.WriteString("-e 'MESOS_TASK_ID=" + taskId.GetValue() + "' ")
		commandTextBuffer.WriteString("-e 'MESOS_SANDBOX=" + MESOS_SANDBOX + "' ")
		//-e WEAVE_CIDR=192.168.0.10/8
		commandTextBuffer.WriteString("-e 'WEAVE_CIDR=" + allocatedip + "' ")
		
		// compose service group id
		appContainerId := composeAppContainerId(serviceGroupId, appId)
		commandTextBuffer.WriteString("-e 'LINKER_APP_CONTAINER_ID=" + appContainerId + "' ")
		
		// Get docker container infomation
		docker_image := appObj.Container.Docker.Image
		docker_network := appObj.Container.Docker.Network
		docker_privileged := strconv.FormatBool(appObj.Container.Docker.Privileged)
		docker_pull_image := appObj.Container.Docker.ForcePullImage

		if docker_pull_image {
			dockerPullCommand := "docker pull " + docker_image
			logrus.Infoln("pull command: " + dockerPullCommand)
			linkercommand.ExecCommand(dockerPullCommand)
			logrus.Infoln(dockerPullCommand)
		} else {
			dockerListCommand := "docker images | awk '{print $1}' | grep \"^.*" + docker_image + "$\""
			listresult, _, _ := linkercommand.ExecCommand(dockerListCommand)
			logrus.Infoln("list command: " + dockerListCommand)
			if listresult == "" {
				dockerPullCommand := "docker pull " + docker_image
				linkercommand.ExecCommand(dockerPullCommand)
			}
		}

		commandTextBuffer.WriteString("--privileged=" + docker_privileged + " ")
		commandTextBuffer.WriteString("--net " + strings.ToLower(docker_network) + " ")

		if appObj.Container.Docker.Parameters != nil {
			for _, element := range appObj.Container.Docker.Parameters {
				logrus.Infoln("key: " + element.Key)
				logrus.Infoln("value: " + element.Value)
				commandTextBuffer.WriteString(" --" + element.Key + "=" + element.Value + " ")
			}
		}

		//support portmapping
		needExposePort, err := getValueFromEnv(envMap, "LINKER_EXPOSE_PORTS")
		if err != nil {
			needExposePort = "false"
		}
		logrus.Infof("LINKER_EXPOSE_PORTS=%v", needExposePort)

		if strings.ToLower(needExposePort) == "true" {
			commandTextBuffer.WriteString("-P ")
		}

		//support volume
		for _, volume := range appObj.Container.Volumes {
			baseDir := strings.Join([]string{MESOS_SANDBOX, appId}, "")
			volumeConfig := volume.ContainerPath
			if volume.HostPath != "" {
				volumeConfig = strings.Join([]string{baseDir, volume.HostPath}, "") + ":" + volumeConfig
				if volume.Mode != "" {
					switch volume.Mode {
					case "RW":
						volumeConfig += ":rw"
					case "RO":
						volumeConfig += ":ro"
					default:
						logrus.Errorf("Unsupported volume mode: " + volume.Mode)
						notifyStatus(mesos.TaskState_TASK_FAILED, driver, taskInfo)
						exec.Shutdown(driver)
						return
					}
				}
			} else {
				logrus.Errorf("Host path is required with mode")
				notifyStatus(mesos.TaskState_TASK_FAILED, driver, taskInfo)
				exec.Shutdown(driver)
				return
			}
			logrus.Infoln("add volume mapping,", volumeConfig)
			commandTextBuffer.WriteString("-v " + volumeConfig + " ")
		}

		command := strings.Join([]string{commandTextBuffer.String(), "--name ", containerName, " ", docker_image}, "")
		logrus.Infoln(command)
		logrus.Infoln(" Executor_Docker_Mapping " + strconv.Itoa(os.Getpid()) + " " + containerName)

		// run docker command
		taskKilledMutex.Lock()
		if taskKilled == true {
			logrus.Infoln("task killed before run container")
			notifyStatus(mesos.TaskState_TASK_FAILED, driver, taskInfo)
			taskKilledMutex.Unlock()
			return
		}
		containerIdResult, errString, err := linkercommand.ExecCommand(command)
		if err != nil {
			// lanuch docker run command error ...
			logrus.Errorln("Lanuch docker run error: " + errString)
			logrus.Errorln("Got error", err)
			// error happened, just quit.
			notifyStatus(mesos.TaskState_TASK_FAILED, driver, taskInfo)
			exec.Shutdown(driver)
			taskKilledMutex.Unlock()
			return
		}
		taskKilledMutex.Unlock()

		logrus.Infoln(containerIdResult)
		containerLongId := ""
		if len(containerIdResult) != 0 {
			array := strings.Split(containerIdResult, "\n")
			containerLongId = array[0]
		}

		clientErr := linkerclient.Update(&appInfo, containerLongId, os.Getenv(HOST_IP), cpuValue, memValue)
		logrus.Debugf("CREATED: appInfo=", appInfo)
		if clientErr != nil {
			logrus.Errorf("Update app instance: "+containerName+" failed, error is %v", clientErr)
			_, err = stopDockerInstance(containerName)
			if err != nil {
				logrus.Errorf("Delete container: "+containerName+" failed, error is %v", err)
			}
			// delete appInstance
			err = linkerclient.Terminate(appInfo)
			if err != nil {
				logrus.Errorf("Terminate appInstance error is %v", err)
			} else {
				logrus.Infoln("App Instance: " + appInfo.ObjectId.Hex() + " is terminated!")
			}
			// error happened, just quit.
			notifyStatus(mesos.TaskState_TASK_FAILED, driver, taskInfo)
			exec.Shutdown(driver)
			return
		} else {
			configStatus, configErr := linkerclient.ConfigDockerInstance(&appInfo)
			// allow config status is false,
			logrus.Infof("app config status is %v", configStatus)
			if configErr != nil {
				// stop docker container
				logrus.Errorf("Config Docker Instance failed, error is %v", configErr)
				logrus.Infof("Will stop the docker instance")
				_, err = stopDockerInstance(containerName)
				if err != nil {
					logrus.Errorf("Delete container: "+containerName+" failed, error is %v", err)
				}
				// delete appInstance
				err = linkerclient.Terminate(appInfo)
				if err != nil {
					logrus.Errorf("Terminate appInstance error is %v", err)
				} else {
					logrus.Infoln("App Instance: " + appInfo.ObjectId.Hex() + " is terminated!")
				}
				// error happened, just quit
				notifyStatus(mesos.TaskState_TASK_FAILED, driver, taskInfo)
				exec.Shutdown(driver)
				return
			}
		}
		logrus.Debugf("CONFIGED: appInfo=", appInfo)
		logrus.Infoln("Executor ending, start to check  status...")
		go checkContainerStatus(containerName, taskInfo.GetTaskId(), driver)
	}
	taskKilledMutex.Lock()
	if taskKilled == true {
		logrus.Infoln("task killed so report failed")
		notifyStatus(mesos.TaskState_TASK_FAILED, driver, taskInfo)
	} else {
		notifyStatus(mesos.TaskState_TASK_RUNNING, driver, taskInfo)
	}
	taskKilledMutex.Unlock()
}

func (exec *LinkerDockerExecutor) KillTask(driver exec.ExecutorDriver, taskInfo *mesos.TaskID) {
	go exec.killTask(driver, taskInfo)
}

func (exec *LinkerDockerExecutor) killTask(driver exec.ExecutorDriver, taskInfo *mesos.TaskID) {
	logrus.Infof("kill task is called.")
	setTaskKilled()

	// init token
	err := linkerclient.InitToken()
	if err != nil {
		logrus.Errorln(err)
	}

	logrus.Infoln("Task id is : " + taskInfo.GetValue())
	logrus.Debugf("TERMINATE. appInfo=", appInfo)
	finStatus := &mesos.TaskStatus{
		TaskId: taskInfo,
		State:  mesos.TaskState_TASK_KILLED.Enum(),
	}

	// containerName, _ := linkerclient.GetContainerIdByTaskId(taskInfo.GetValue())
	containerName := appInfo.DockerContainerName

	if linkerclient.Openstack == isWhat {
		opts := gophercloud.AuthOptions{
			IdentityEndpoint: openstackIdentityEndpoint,
			Username:         openstackUsername,
			Password:         openstackPassword,
			TenantID:         openstackTenandId,
		}

		provider, openstackError := openstack.AuthenticatedClient(opts)

		if openstackError != nil {
			logrus.Infoln("login into openstack error!")
			errStatus := &mesos.TaskStatus{
				TaskId: taskInfo,
				State:  mesos.TaskState_TASK_FAILED.Enum(),
			}
			_, err := driver.SendStatusUpdate(errStatus)
			if err != nil {
				logrus.Infoln("Got error", err)
			}
			return
		}

		client, openstackError := openstack.NewComputeV2(provider, gophercloud.EndpointOpts{
			Region: openstackRegion,
		})

		if openstackError != nil {
			logrus.Infoln("New openstack compute client error!")
			errStatus := &mesos.TaskStatus{
				TaskId: taskInfo,
				State:  mesos.TaskState_TASK_FAILED.Enum(),
			}
			_, err := driver.SendStatusUpdate(errStatus)
			if err != nil {
				logrus.Infoln("Got error", err)
			}
			return
		}

		// Here containerName is instance id.
		deleteResult := servers.Delete(client, containerName)
		if deleteResult.Err != nil {

		} else {
			logrus.Infoln("Instance: " + containerName + " is deleted!")
		}

	} else {

		if containerName != "" {
			logrus.Infoln("Got container name is: " + containerName)

			containerIdResult, err := stopDockerInstance(containerName)
			if err != nil {
				logrus.Errorln("stop docker container failed, error is %v", err)
			} else {
				logrus.Infoln("Container: " + containerIdResult + " is stoped!")
			}
			err = linkerclient.Terminate(appInfo)
			if err != nil {
				logrus.Errorf("Terminate appInfo error is %v", err)
			} else {
				logrus.Infoln("App Instance: " + appInfo.ObjectId.Hex() + " is terminated!")
			}
		} else {
			logrus.Infoln("no container name specified.")
		}
	}

	driver.SendStatusUpdate(finStatus)
	logrus.Infoln("Kill task")
	exec.Shutdown(driver)
}

func (exec *LinkerDockerExecutor) FrameworkMessage(driver exec.ExecutorDriver, msg string) {
	logrus.Infoln("Got framework message: ", msg)
}

func (exec *LinkerDockerExecutor) Shutdown(driver exec.ExecutorDriver) {
	logrus.Infoln("Shutting down the executor")
	driver.Stop()
}

func (exec *LinkerDockerExecutor) Error(driver exec.ExecutorDriver, err string) {
	logrus.Infoln("Got error message:", err)
}

func notifyStatus(status mesos.TaskState, driver exec.ExecutorDriver, taskInfo *mesos.TaskInfo) {
	dockerStatus := &mesos.TaskStatus{
		TaskId: taskInfo.GetTaskId(),
		State:  status.Enum(),
	}

	_, err := driver.SendStatusUpdate(dockerStatus)
	if err != nil {
		logrus.Infoln("Got error", err)
	}
}

func stopDockerInstance(containerName string) (containerId string, err error) {
	command := strings.Join([]string{"docker stop ", containerName}, "")
	containerId, _, err = linkercommand.ExecCommand(command)
	return
}

func pipework(envMap map[string]string, containerName string) (isOk bool) {
	addrValue, addrOk := envMap["LINKER_ADDR"]
	brValue, brOk := envMap["LINKER_BR"]
	if addrOk && brOk {
		command := strings.Join([]string{"/usr/local/bin/pipework ", brValue, " ", containerName, " ", addrValue}, "")
		result, erroutput, err := linkercommand.ExecCommand(command)
		logrus.Infoln("pipework: " + command)
		logrus.Infoln(result)
		logrus.Infoln(erroutput)
		if err != nil {
			return false
		}
	}
	return true
}

func postScript(envMap map[string]string) (isOk bool) {
	confScript, script := envMap["LINKER_CONF_SCRIPT"]
	confParams, paramOk := envMap["LINKER_CONF_PARAMS"]
	if script && paramOk {
		command := strings.Join([]string{confScript, " ", confParams, " "}, "")
		//		result, erroutput, err := linkercommand.ExecCommand(command)
		logrus.Infoln("postScript: " + command)
		//		logrus.Infoln(result)
		//		logrus.Infoln(erroutput)
		//		if err != nil {
		//			return false
		//		}
	}
	return true
}

func calCpus(cpus float64) (maxCpus int) {
	const CPU_SHARES_PER_CPU int = 1024
	const MIN_CPU_SHARES int = 10

	one := float64(CPU_SHARES_PER_CPU) * float64(cpus)
	two := float64(MIN_CPU_SHARES)

	result := math.Max(one, two)

	return int(result)
}

func checkContainerStatus(containerName string, taskInfo *mesos.TaskID, driver exec.ExecutorDriver) {
	logrus.Infoln("Start to check container: " + containerName)
	time.Sleep(time.Duration(dockerStartCheckTime) * time.Second)
	dockerCheckInterval := props.GetInt("docker.checkalive.interval", 5)
	ticker := time.NewTicker(time.Second * time.Duration(dockerCheckInterval))

	go func() {
		for range ticker.C {
			result := isContainerAlive(containerName)

			if !result {
				logrus.Infoln("Find broken container...")
				finStatus := &mesos.TaskStatus{
					TaskId: taskInfo,
					State:  mesos.TaskState_TASK_LOST.Enum(),
				}

				driver.SendStatusUpdate(finStatus)

				defer driver.Stop()
			}

		}
	}()

}

func isContainerAlive(containerName string) (result bool) {
	command := strings.Join([]string{"docker ps -f \"name=", containerName, "\" "}, "")
	commandResult, erroutput, _ := linkercommand.ExecCommand(command)
	if erroutput != "" {
		result = false
	} else {
		if strings.Contains(commandResult, containerName) {
			result = true
		} else {
			result = false
		}

	}
	return
}

func composeAppContainerId(serviceGroupId, marathonAppId string) (appContainerId string) {
	appPath := strings.Split(marathonAppId, "/")
	appIds := append(strings.Split(serviceGroupId, "/"), appPath[2:]...)
	appContainerId = strings.Join(appIds, "/")
	return appContainerId
}

func main() {
	logrus.Infoln("Starting Linker Executor (Go)")

	MESOS_SANDBOX = props.GetString("mesos.volume.base", "/mnt/mesos/sandbox/")

	dockerStartCheckTime = props.GetInt("docker.startcheck.time", 15)

	openstackIdentityEndpoint = props.GetString("openstack.identity.endpoint", "")
	openstackUsername = props.GetString("openstack.username", "")
	openstackPassword = props.GetString("openstack.password", "")
	openstackTenandId = props.GetString("openstack.tenantid", "")
	openstackRegion = props.GetString("openstack.region", "")

	dockerProtocol := props.MustGet("docker.host.protocol")
	dockerUrl := props.MustGet("docker.host.url")
	dockerEndpoint = strings.Join([]string{dockerProtocol, "://", dockerUrl}, "")

	zkClient := linkerclient.ZkClient{
		Url:   zkUrl,
		Props: props,
		Conn:  nil,
	}
	linkerclient.Zk = &zkClient
	linkerclient.DockerEndpoint = dockerEndpoint

	dconfig := exec.DriverConfig{
		Executor: newExampleExecutor(),
	}
	driver, err := exec.NewMesosExecutorDriver(dconfig)

	if err != nil {
		logrus.Infoln("Unable to create a ExecutorDriver ", err.Error())
	}

	_, err = driver.Start()
	if err != nil {
		logrus.Infoln("Got error:", err)
		return
	}
	logrus.Infoln("Executor process has started and running.")
	driver.Join()
}

func init() {
	logrus.Infoln("Init Linker Executor (Go)")
	// get configuration
	flag.Parse()
	logrus.Infoln("propertiesFileValue= " + *propertiesFile)
	zkUrl = *zkFlag

	var err error
	if props, err = properties.LoadFile(*propertiesFile, properties.UTF8); err != nil {
		logrus.Errorf("[error] Unable to read properties:%v\n", err)
	}
	// set log configuration
	// Log as JSON instead of the default ASCII formatter.
	switch props.GetString("logrus.formatter", "text") {
	case "text":
		logrus.SetFormatter(&logrus.TextFormatter{})
	case "json":
		logrus.SetFormatter(&logrus.JSONFormatter{})
	default:
		logrus.SetFormatter(&logrus.TextFormatter{})
	}
	// Use the Airbrake hook to report errors that have Error severity or above to
	// an exception tracker. You can create custom hooks, see the Hooks section.
	// log.AddHook(airbrake.NewHook("https://example.com", "xyz", "development"))

	// Output to stderr instead of stdout, could also be a file.
	logrus.SetOutput(os.Stderr)

	// Only log the warning severity or above.
	level, err := logrus.ParseLevel(props.GetString("logrus.level", "info"))
	if err != nil {
		logrus.Errorf("parse log level err is %v\n", err)
		logrus.Errorf("using default level is %v \n", logrus.InfoLevel)
		level = logrus.InfoLevel
	}
	logrus.SetLevel(level)
}
