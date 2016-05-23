package services

import (
	"bytes"
	"code.google.com/p/go-uuid/uuid"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/fsouza/go-dockerclient"
	"gopkg.in/gomail.v2"
	"gopkg.in/mgo.v2/bson"
	"io/ioutil"
	"linkernetworks.com/linker_common_lib/httpclient"
	"linkernetworks.com/linker_common_lib/persistence/dao"
	"linkernetworks.com/linker_common_lib/persistence/entity"
	"linkernetworks.com/linker_common_lib/rest/response"
	"linkernetworks.com/linker_controller/common"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	COMMON_ERROR_INVALIDATE   = "E12002"
	COMMON_ERROR_UNAUTHORIZED = "E12004"
	COMMON_ERROR_UNKNOWN      = "E12001"
	COMMON_ERROR_INTERNAL     = "E12003"

	userMgmtEndpoint string
)

func GetUserMgmtEndpoint() (endpoint string, err error) {
	if userMgmtEndpoint != "" {
		endpoint = userMgmtEndpoint
		return endpoint, nil
	} else {
		endpoint, err := common.UTIL.ZkClient.GetUserMgmtEndpoint()
		if err != nil {
			logrus.Errorf("get userMgmt endpoint err is %v", err)
		}
		return endpoint, err
	}

}

func SetUserMgmtEndpoint(endpoint string) {
	userMgmtEndpoint = endpoint
}

func GetClusterMgmtEndpoint() (endpoint string, err error) {
	endpoint, err = common.UTIL.ZkClient.GetClusterMgmtEndpoint()
	if err != nil {
		logrus.Errorf("get clusterMgmt endpoint err is %v", err)
		return
	}

	return endpoint, err

}

func TokenValidation(tokenId string) (errorCode string, err error) {
	userUrl, err := GetUserMgmtEndpoint()
	if err != nil {
		logrus.Errorf("get userMgmt endpoint err is %v", err)
		return "E12003", err
	}
	url := strings.Join([]string{"http://", userUrl, "/v1/token/?", "token=", tokenId}, "")
	logrus.Debugln("token validation url=" + url)

	resp, err := httpclient.Http_get(url, "",
		httpclient.Header{"Content-Type", "application/json"})
	defer resp.Body.Close()
	if err != nil {
		logrus.Errorf("http get token validate error %v", err)
		return "E12003", err
	}

	data, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		logrus.Errorf("token validation failed %v", string(data))
		errorCode, err = getErrorFromResponse(data)
		return
	}

	return "", nil
}

func GetTokenById(token string) (currentToken *entity.Token, err error) {
	userUrl, err := GetUserMgmtEndpoint()
	if err != nil {
		logrus.Errorf("get userMgmt endpoint err is %v", err)
		return nil, err
	}
	url := strings.Join([]string{"http://", userUrl, "/v1/token/", token}, "")
	logrus.Debugln("get token url=" + url)

	resp, err := httpclient.Http_get(url, "",
		httpclient.Header{"Content-Type", "application/json"}, httpclient.Header{"X-Auth-Token", token})
	defer resp.Body.Close()
	if err != nil {
		logrus.Errorf("http get token error %v", err)
		return nil, err
	}

	data, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		logrus.Errorf("get token by id failed %v", string(data))
		return nil, errors.New("get token by id failed")
	}

	currentToken = new(entity.Token)
	err = getRetFromResponse(data, currentToken)
	return
}

func GenerateToken() (token string, err error) {
	userUrl, err := GetUserMgmtEndpoint()
	if err != nil {
		logrus.Errorf("get userMgmt endpoint err is %v", err)
		return "", err
	}

	user := entity.User{}
	user.Email = common.UTIL.Props.MustGet("usermgmt.username")
	user.Username = common.UTIL.Props.MustGet("usermgmt.username")
	user.Password = common.UTIL.Props.MustGet("usermgmt.password")
	user.Tenantname = common.UTIL.Props.MustGet("usermgmt.tenantname")

	body, _ := json.Marshal(user)
	logrus.Debugln("body=" + string(body))
	url := strings.Join([]string{"http://", userUrl, "/v1/token"}, "")
	logrus.Debugln("generate token url: " + url)
	response, err := httpclient.Http_post(url, string(body), httpclient.Header{"Content-Type", "application/json"})
	defer response.Body.Close()
	if err != nil {
		logrus.Errorf("post to generate token error %s", err.Error())
		return
	}

	data, _ := ioutil.ReadAll(response.Body)
	if response.StatusCode >= 400 {
		logrus.Errorf("generate token failed %v", string(data))
		return "", errors.New("get generate token failed")
	}

	return getIdFromResponse(data)

}

func GetUserById(token string, userId string) (currentUser *entity.User, err error) {
	userUrl, err := GetUserMgmtEndpoint()
	if err != nil {
		logrus.Errorf("get userMgmt endpoint err is %v", err)
		return nil, err
	}
	url := strings.Join([]string{"http://", userUrl, "/v1/user/", userId}, "")
	logrus.Debugln("token validation url=" + url)

	resp, err := httpclient.Http_get(url, "",
		httpclient.Header{"Content-Type", "application/json"}, httpclient.Header{"X-Auth-Token", token})
	defer resp.Body.Close()
	if err != nil {
		logrus.Errorf("http get token error %v", err)
		return nil, err
	}

	data, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		logrus.Errorf("get user by id failed %v", string(data))
		return nil, errors.New("get user by id failed")
	}

	currentUser = new(entity.User)
	err = getRetFromResponse(data, currentUser)
	return
}

func getRetFromResponse(data []byte, obj interface{}) (err error) {
	var resp *response.Response
	resp = new(response.Response)
	err = json.Unmarshal(data, &resp)
	if err != nil {
		return err
	}

	jsonout, err := json.Marshal(resp.Data)
	if err != nil {
		return err
	}

	json.Unmarshal(jsonout, obj)

	return
}

func GetTenantIdByUserId(token string, userId string) (tenantId string, err error) {
	userUrl, err := GetUserMgmtEndpoint()
	if err != nil {
		logrus.Errorf("get userMgmt endpoint err is %v", err)
		return "", err
	}
	url := strings.Join([]string{"http://", userUrl, "/v1/tenant/tenantId?", "userId=" + userId}, "")
	logrus.Debugln("get tenantid by userid url=" + url)

	resp, err := httpclient.Http_get(url, "",
		httpclient.Header{"Content-Type", "application/json"}, httpclient.Header{"X-Auth-Token", token})
	defer resp.Body.Close()
	if err != nil {
		logrus.Errorf("http get tenantid by userid error %v", err)
		return "", err
	}

	data, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		logrus.Errorf("get tenantid by userid failed %v", string(data))
		return "", errors.New("get tenant id by userid failed")
	}

	return getIdFromResponse(data)

}

func GetHostsByClusterId(clusterId string) (hosts []entity.Host, err error) {
	token, err := GenerateToken()
	if err != nil {
		return nil, err
	}

	clusterUrl, err := common.UTIL.ZkClient.GetClusterMgmtEndpoint()
	if err != nil {
		logrus.Errorf("get cluster endpoint err is %v", err)
		return nil, err
	}
	url := strings.Join([]string{"http://", clusterUrl, "/v1/host/?", "cluster_id=", clusterId}, "")
	logrus.Debugln("get hosts by cluster id url=" + url)

	resp, err := httpclient.Http_get(url, "",
		httpclient.Header{"Content-Type", "application/json"}, httpclient.Header{"X-Auth-Token", token})
	defer resp.Body.Close()
	if err != nil {
		logrus.Errorf("http get hosts by cluster id error %v", err)
		return nil, err
	}

	data, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		logrus.Errorf("get hosts by cluster id failed %v", string(data))
		return nil, errors.New("get hosts by cluster id failed")
	}

	hosts = []entity.Host{}
	err = getRetFromResponse(data, &hosts)
	return
}

func GetAllCluster() (clusters []entity.Cluster, err error) {
	token, err := GenerateToken()
	if err != nil {
		return nil, err
	}

	clusterUrl, err := common.UTIL.ZkClient.GetClusterMgmtEndpoint()
	if err != nil {
		logrus.Errorf("get cluster endpoint err is %v", err)
		return nil, err
	}
	url := strings.Join([]string{"http://", clusterUrl, "/v1/cluster/"}, "")
	logrus.Debugln("get all cluster url=" + url)

	resp, err := httpclient.Http_get(url, "",
		httpclient.Header{"Content-Type", "application/json"}, httpclient.Header{"X-Auth-Token", token})
	defer resp.Body.Close()
	if err != nil {
		logrus.Errorf("http get all clusters error %v", err)
		return nil, err
	}

	data, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		logrus.Errorf("get hosts by cluster id failed %v", string(data))
		return nil, errors.New("get all clusters failed")
	}

	clusters = []entity.Cluster{}
	err = getRetFromResponse(data, &clusters)
	return
}

func DeleteClusterByUserId(userId string) (err error) {
	token, err := GenerateToken()
	if err != nil {
		return err
	}

	clusterUrl, err := common.UTIL.ZkClient.GetClusterMgmtEndpoint()
	if err != nil {
		logrus.Errorf("get cluster endpoint err is %v", err)
		return err
	}
	url := strings.Join([]string{"http://", clusterUrl, "/v1/cluster/", "userid=", userId}, "")
	logrus.Debugln("delete cluster by userid url=" + url)

	resp, err := httpclient.Http_delete(url, "",
		httpclient.Header{"Content-Type", "application/json"}, httpclient.Header{"X-Auth-Token", token})
	defer resp.Body.Close()
	if err != nil {
		logrus.Errorf("delete cluster by user id error %v", err)
		return err
	}

	return
}

func getIdFromResponse(data []byte) (tokenId string, err error) {
	var resp *response.Response
	resp = new(response.Response)
	err = json.Unmarshal(data, &resp)
	if err != nil {
		return "", err
	}

	json := resp.Data.(map[string]interface{})
	idobj := json["id"]
	if idobj == nil {
		logrus.Errorln("no id field")
		return "", errors.New("no id field in response!")
	}
	return idobj.(string), nil

}

func getErrorFromResponse(data []byte) (errorCode string, err error) {
	var resp *response.Response
	resp = new(response.Response)
	err = json.Unmarshal(data, &resp)
	if err != nil {
		return "E12003", err
	}

	errorCode = resp.Error.Code
	err = errors.New(resp.Error.ErrorMsg)
	return
}

func SendMail(host string, username string, passwd string, to string, subject string, body string) {
	m := gomail.NewMessage()
	m.SetHeader("From", username)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/plain", body)

	d := gomail.NewPlainDialer(host, 25, username, passwd)

	err := d.DialAndSend(m)
	if err != nil {
		logrus.Warnln("send active user email error %v", err)
	}

	return
}

func getCurrentTime() (t string) {
	t = time.Now().Format(time.RFC3339)
	return
}

func getWaitTimeForNextHour() int64 {
	currentTime := time.Now()

	//add 1 hour
	dur, _ := time.ParseDuration("1h")
	newTime := currentTime.Add(dur)

	//remove minute
	minute := strconv.Itoa(currentTime.Minute())
	dur, _ = time.ParseDuration("-" + minute + "m")
	newTime = newTime.Add(dur)

	//remove second
	second := strconv.Itoa(currentTime.Second())
	dur, _ = time.ParseDuration("-" + second + "s")
	newTime = newTime.Add(dur)

	nextHour := newTime.Unix()
	current := currentTime.Unix()

	return nextHour - current

}

func GetPastDay(current time.Time) time.Time {
	dur, _ := time.ParseDuration("-24h")
	return current.Add(dur)
}

func HashString(password string) string {
	encry := sha256.Sum256([]byte(password))
	return hex.EncodeToString(encry[:])
}

//keep 3 values after dot
func formatValue(value float64) float64 {
	formatedValue := fmt.Sprintf("%.3f", value)
	ret, err := strconv.ParseFloat(formatedValue, 64)
	if err != nil {
		logrus.Warnln("convert value error %v", err)
		return value
	}
	return ret
}

func ParseStringToSecond(value string, location *time.Location, layout string) int {
	timeValue, _ := time.ParseInLocation(layout, value, location)
	ret := int(timeValue.Unix())

	return ret
}

func generateQueryWithAuth(oriQuery bson.M, authQuery bson.M) (query bson.M) {
	if len(authQuery) == 0 {
		query = oriQuery
	} else {
		query = bson.M{}
		query["$and"] = []bson.M{oriQuery, authQuery}
	}
	logrus.Debugf("generated query [%v] with auth [%v], result is [%v]", oriQuery, authQuery, query)
	return
}

func filterInstanceByScope(scope string, instances []string,
	me string) (refinedInstances []string) {

	logrus.Debugf("filterInstanceByScope: scope is [%v]", scope)
	refinedInstances = instances
	logrus.Debugf("instances tobe refined is [%v], me is [%v]", instances, me)
	switch scope {
	case "ALL":
		refinedInstances = instances
	case "ONLYME":
		refinedInstances = make([]string, 0, len(instances))
		for _, instance := range instances {
			if instance == me {
				refinedInstances = append([]string{me})
			}
		}
	case "WITHOUTME":
		for k, instance := range instances {
			if instance == me {
				refinedInstances = append(instances[:k], instances[k+1:]...)
			}
		}
	default:
		logrus.Errorf("unexpected scope: %v", scope)
	}
	logrus.Debugf("refinedInstances are [%v]", strings.Join(refinedInstances, ","))
	return
}

func GetConfigStepsByAciId(aciId, callFrom, x_auth_token string) (steps []entity.ConfigStep,
	err error) {
	var aci entity.AppContainerInstance
	logrus.Debugf("start to get config steps by aciId [%v]", aciId)
	// get aci struct by aciId
	if len(aciId) > 0 {
		aci, _, err = GetAciService().queryById(aciId)
		if err != nil {
			logrus.Errorf("can not get aci by objectId [%v], error is %v", aciId, err)
			return
		}
	} else {
		logrus.Errorf("aciId is not set")
		err = errors.New("Invilid parameters, need appIstanceId")
		return
	}

	//get appPackage by aci.AppContainerId
	_, acps, _, err := GetAcpService().QueryByAppPath(aci.AppContainerId, x_auth_token)
	if err != nil {
		logrus.Errorf("find appPackage by appPath [%v] failed, error: %v", aci.AppContainerId, err)
		steps = make([]entity.ConfigStep, 0)
		err = nil
		return
	}

	if len(acps) <= 0 {
		logrus.Warnf("can not find appPackage by appPath [%v]", aci.AppContainerId)
		steps = make([]entity.ConfigStep, 0)
		err = nil
		return
	}

	sgi, _, err := GetSgiService().QueryById(aci.ServiceGroupInstanceId, x_auth_token)
	if err != nil {
		logrus.Errorf("find sgi by objectId [%v] failed, error: %v", aci.ServiceGroupInstanceId, err)
		return
	}

	// check acp conditions，return config commands
	// get service group instance by aci.ServiceGroupInstanceId
	acp := acps[0]
	for _, config := range acp.Configurations {
		matched := true
		for _, precondition := range config.Preconditions {
			condition := precondition.Condition
			if subMatch, _ := checkGroupInstance(condition, &sgi); !subMatch {
				matched = false
				break
			}
		}
		if matched {
			logrus.Debugln("matched config=" + config.Name + ", steps.length=" + strconv.Itoa(len(config.Steps)))
			steps = make([]entity.ConfigStep, len(config.Steps))
			for numS, step := range config.Steps {
				steps[numS] = step
				execCmd := step.Execute
				// parse execCmd to findout all needed paramaeters, the parameter should be wraped by %.
				// for example: /linker/webconfig.sh %/linkerapp/dbproxy/haproxy.[docker_container_ip]:[docker_container_port]%
				logrus.Debugln("execCmd: " + execCmd)

				//FIXME: if this call is from notify, noneed to allocate again, now is hardcode.
				if strings.Contains(strings.ToLower(callFrom), "notify") && strings.Contains(execCmd, "ALLOCATE-") {
					exec := "echo notify"
					steps[numS].Execute = exec
					logrus.Debug("add exec=" + exec)
					continue
				} else {
					neededParams := strings.Split(execCmd, " ")
					exec := neededParams[0]
					for _, neededParam := range neededParams[1:] {
						logrus.Debug("neededParam=" + neededParam)
						if strings.HasPrefix(neededParam, "%") && strings.HasSuffix(neededParam, "%") {
							neededParam = strings.Trim(neededParam, "%")
							if strings.Contains(neededParam, "ALLOCATE-") {
								replacedParam, err := getAllocatedInfo(neededParam, aci.ObjectId.Hex())
								if err != nil {
									logrus.Errorf("generate config step failed, aci id is %v, neededParam is %v, error is %v", aci.ObjectId.Hex(), neededParam, err)
									logrus.Errorf("will continue next command.")
									continue
								}
								exec = exec + " " + replacedParam
							} else {
								replacedParam := getGroupInstanceInfo(&aci,
									neededParam, &sgi, step.Scope)
								exec = exec + " " + replacedParam
							}
						} else {
							logrus.Debug("skip param replacement: " + neededParam)
							exec = exec + " " + neededParam
						}
					}
					steps[numS].Execute = exec
					logrus.Debug("add exec=" + exec)
				}
			}
			break
		} else {
			steps = make([]entity.ConfigStep, 0)
		}
	}
	return
}

func checkGroupInstance(condition string, sgi *entity.ServiceGroupInstance) (result bool, err error) {
	logrus.Debug("condition: " + condition)
	exps := strings.Split(condition, " ")
	appId := exps[0]
	expr := exps[1]
	count, err := strconv.Atoi(exps[2])
	if err != nil {
		logrus.Error("parse condition error, appCount is: " + exps[2])
		return
	}
	instanceIds := entity.GetAppInstanceIdsFromGroupInstance(sgi, appId)
	logrus.Debug("instanceIds:" + strings.Join(instanceIds, ","))
	// the value of expr maybe
	// -eq	检测两个数是否相等，相等返回 true。	[ $a -eq $b ] 返回 true。
	// -ne	检测两个数是否相等，不相等返回 true。	[ $a -ne $b ] 返回 true。
	// -gt	检测左边的数是否大于右边的，如果是，则返回 true。	[ $a -gt $b ] 返回 false。
	// -lt	检测左边的数是否小于右边的，如果是，则返回 true。	[ $a -lt $b ] 返回 true。
	// -ge	检测左边的数是否大等于右边的，如果是，则返回 true。	[ $a -ge $b ] 返回 false。
	// -le	检测左边的数是否小于等于右边的，如果是，则返回 true。	[ $a -le $b ] 返回 true。
	logrus.Debugf("check the number of %s in groupInstance whether satisfy %s %d", appId, expr, count)
	switch expr {
	case "-eq":
		result = len(instanceIds) == count
	case "-ne":
		result = len(instanceIds) != count
	case "-gt":
		result = len(instanceIds) > count
	case "-lt":
		result = len(instanceIds) < count
	case "-ge":
		result = len(instanceIds) >= count
	case "-le":
		result = len(instanceIds) <= count
	default:
		logrus.Debugf("wrong.")
	}
	return
}

func getAllocatedInfo(neededParam string, owner string) (returnParam string, err error) {
	//Remove ALLOCATE
	src := strings.Trim(neededParam, "ALLOCATE-")
	return _getAllocatedInfo(src, owner)
}

func _getAllocatedInfo(resourceRegx string, owner string) (returnParam string, err error) {
	//Remove ALLOCATE
	names := strings.Split(resourceRegx, ":")
	collectionName := names[0]
	fullName := names[1]
	// find value object
	selector := bson.M{}
	//get collection name and resources
	allocateValue, err := findAllocatedValue(collectionName, selector, owner)
	if err != nil {
		logrus.Errorf("getAllocatedInfo failed, resourceRegx is %v, owner is %v, error is %v",
			resourceRegx, owner, err)
		return
	}
	//make reg
	var reg = regexp.MustCompile(`\[[a-zA-Z0-9_\.]+\]`)
	dest := reg.ReplaceAllFunc([]byte(fullName), func(in []byte) (out []byte) {
		instr := string(in)
		attrName := strings.TrimPrefix(instr, "[")
		attrName = strings.TrimSuffix(attrName, "]")
		//get the value to replace
		for i := 0; i < allocateValue.NumField(); i++ {
			valueField := allocateValue.Field(i)
			typeField := allocateValue.Type().Field(i)
			if typeField.Tag.Get("json") == attrName {
				f := valueField.Interface()
				v := reflect.ValueOf(f)
				switch v.Kind() {
				case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
					out = []byte(strconv.FormatInt(v.Int(), 10))
				case reflect.String:
					out = []byte(v.String())
				}
				logrus.Debug("value=" + string(out))
				break
			}
		}
		return
	})
	returnParam = string(dest)
	return
}

func findAllocatedValue(collectionName string, selector bson.M,
	owner string) (allocateValue reflect.Value, err error) {
	// find resource by owner, if not exist, allocate one
	owner = strings.TrimSpace(owner)
	selector["allocated"] = owner
	if collectionName == "ipAddressResource" {
		ipAddressResource := entity.IpAddressResource{}
		err = dao.HandleQueryOne(&ipAddressResource, dao.QueryStruct{collectionName, selector, 0, 0, ""})
		if err != nil {
			if notFound := strings.Contains(err.Error(), "not found"); notFound {
				// the resource for owner is not exist, should allocate one for it.
				logrus.Infof("the resource for owner(%v) is not exist, should allocate one for it.", owner)
				allocateValue, err = allocateNew(collectionName, selector, owner)
				return

			} else {
				logrus.Errorf("allocate resource error, selector is %v, error is %v",
					selector, err)
				return
			}
		}
		logrus.Debug("allocated ipAddressResource [%v] ", ipAddressResource)
		allocateValue = reflect.ValueOf(ipAddressResource)
		return
	} else {
		//TODO: modify here to support other resources.
		err = errors.New("unsupported resource " + collectionName)
		return
	}
}

/*
 * allocate new resource for owner.
 */
func allocateNew(collectionName string, selector bson.M,
	owner string) (newValue reflect.Value, err error) {
	// find out one resource that not be allocated.
	selector["allocated"] = "false"

	if collectionName == "ipAddressResource" {
		ipAddressResource := entity.IpAddressResource{}
		err = dao.HandleQueryOne(&ipAddressResource, dao.QueryStruct{collectionName, selector, 0, 0, ""})
		if err != nil {
			logrus.Errorf("allocate new resource error, selector is %v, error is %v",
				selector, err)
			return
		}
		newValue = reflect.ValueOf(ipAddressResource)
		ipAddressResource.Allocated = owner
		// update ip resource to db
		_, _, err = GetIPResourceService().updateById(ipAddressResource.ObjectId.Hex(), ipAddressResource)
		if err != nil {
			logrus.Errorf("update ip resource failed, error is %v", err)
		}
		return
	} else {
		//TODO: modify here to support other resources.
		err = errors.New("unsupported resource " + collectionName)
		return
	}
}

func getGroupInstanceInfo(aci *entity.AppContainerInstance,
	paramStr string, sgi *entity.ServiceGroupInstance, scope string) (param string) {

	// scope means what should be returned, the value of scope maybe ALL, ONLYME, WITHOUTME
	// the returned value shoule be joined by ","
	// paramStr example:/linkerapp/dbproxy/haproxy.[docker_container_ip]:[docker_container_port]
	paraArr := strings.Split(paramStr, ".")
	appId := paraArr[0]
	instanceIds := entity.GetAppInstanceIdsFromGroupInstance(sgi, appId)
	instanceIds = filterInstanceByScope(scope, instanceIds, aci.ObjectId.Hex())
	params := make([]string, len(instanceIds))
	for i, instanceId := range instanceIds {
		// find instance by instanceId
		aci, _, err := GetAciService().queryById(instanceId)
		if err != nil {
			logrus.Errorf("find app instance by objectId [%v] failed, error is %v",
				instanceId, err)
			logrus.Errorf("will continue next instance.")
			continue
		}
		logrus.Debug("aci.AppContainerId: " + aci.AppContainerId)
		logrus.Debug("aci.DockerContainerIp: " + aci.DockerContainerIp)
		logrus.Debug("aci.DockerContainerPort: " + aci.DockerContainerPort)

		src := paraArr[1]
		var reg = regexp.MustCompile(`\[[a-z0-9_]+\]`)
		dest := reg.ReplaceAllFunc([]byte(src), func(in []byte) (out []byte) {
			instr := string(in)
			attrName := strings.TrimPrefix(instr, "[")
			attrName = strings.TrimSuffix(attrName, "]")
			aciValue := reflect.ValueOf(aci)
			for i := 0; i < aciValue.NumField(); i++ {
				valueField := aciValue.Field(i)
				typeField := aciValue.Type().Field(i)
				if typeField.Tag.Get("json") == attrName {
					f := valueField.Interface()
					v := reflect.ValueOf(f)
					switch v.Kind() {
					case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
						out = []byte(strconv.FormatInt(v.Int(), 10))
					case reflect.String:
						out = []byte(v.String())
					}
					logrus.Debug("value=" + string(out))
					break
				}
			}
			return
		})
		params[i] = string(dest)
	}
	// join params by comma, result like "10.2.1.1,10.2.3.1"
	param = strings.Join(params, ",")
	return
}

func isFirstNodeInZK() bool {
	hostname, err := os.Hostname()
	if err != nil {
		logrus.Warnln("get host name error!", err)
		return false
	}

	path, err := common.UTIL.ZkClient.GetFirstControllerPath()
	if err != nil {
		logrus.Warnln("get controller node from zookeeper error!", err)
		return false
	}

	return strings.HasPrefix(path, hostname)
}

func configAppInstance(aciId string,
	steps []entity.ConfigStep) (configStatus bool, err error) {

	// get app instance details, findout mesos-slave, container-name
	aci, _, err := GetAciService().queryById(aciId)
	if err != nil {
		logrus.Errorf("config aci [%v] failed, can not get aci, error is %v", aciId, err)
		configStatus = false
		return
	}

	// do docker exec
	dockerEndpoint := strings.Join([]string{"http://", aci.MesosSlave,
		":", common.UTIL.Props.GetString("docker.api.port", "4243")}, "")

	dockername := aci.DockerContainerName
	configStatus = true
	for _, step := range steps {
		c_type := strings.ToLower(step.ConfigType)
		command := step.Execute
		logrus.Infof("execute config command: %v, in app instance %v", command, aciId)
		if c_type == "docker" {
			execInsp, err := doDockerExec(dockerEndpoint, dockername, command)
			if err != nil {
				logrus.Errorf("execute command step err is %v", err)
				configStatus = false
				break
			}
			if execInsp.ExitCode != 0 {
				logrus.Errorf("execute command [%v] returned code is not 0", command)
				configStatus = false
			}
		} else if c_type == "command" {
			//TODO command Call
			logrus.Warnf("TODO: command Call: %v ", command)
			// configStatus = true
		}
	}

	return
}

func doDockerExec(dockerEndpoint, dockername,
	command string) (execInsp *docker.ExecInspect, err error) {

	// endpoint := "unix:///var/run/docker.sock"
	client, _ := docker.NewClient(dockerEndpoint)
	logrus.Info("dockername=" + dockername + ", command=" + command)
	createOptions := docker.CreateExecOptions{
		Container:    dockername,
		Cmd:          strings.Split(command, " "),
		AttachStdin:  false,
		AttachStdout: true,
		AttachStderr: true,
		Tty:          false,
	}

	exec, err := client.CreateExec(createOptions)
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
	err = client.StartExec(exec.ID, startOptions)
	if err != nil {
		logrus.Errorf("start exec error %v", err)
		return
	}

	logrus.Debugf("exec stdout: %v", stdout.String())
	logrus.Debugf("exec stderr: %v", stderr.String())

	execInsp, err = client.InspectExec(exec.ID)
	logrus.Debugf("exec is %v", execInsp)
	return
}

/*
  check service group instance status, check the instances array size >= instances
  check each instance's status is cofiged
*/
func isSgiFinished(sgi *entity.ServiceGroupInstance) (isFinished bool) {
	isFinished = true
	for i := range sgi.Groups {
		group := &sgi.Groups[i]
		isFinished = _isSgiFinished(group)
		if !isFinished {
			return
		}
	}
	return
}

func _isSgiFinished(rg *entity.RefinedGroup) (isFinished bool) {
	isFinished = true
	if rg.RefinedGroups != nil {
		for i := range rg.RefinedGroups {
			group := &rg.RefinedGroups[i]
			isFinished = _isSgiFinished(group)
			if !isFinished {
				return
			}
		}
	}
	if rg.RefinedApps != nil {
		for j := range rg.RefinedApps {
			app := &rg.RefinedApps[j]
			if len(app.InstanceIds) != app.Instances {
				isFinished = false
				return
			}
			for _, aciId := range app.InstanceIds {
				appInstance, _, err := GetAciService().queryById(aciId)
				if err != nil {
					logrus.Errorf("get aci by aci id error is %v", err)
					isFinished = false
					return
				}
				if appInstance.LifeCycleStatus != ACI_STATUS_CONFIGED {
					isFinished = false
					return
				}
			}
		}
	}
	return
}

func doNotifyByAci(aci *entity.AppContainerInstance, sgi *entity.ServiceGroupInstance, callFrom, x_auth_token string) (err error) {
	aciId := aci.ObjectId.Hex()
	logrus.Infof("start to do notify caused by aci [%v] changed", aci)
	total, acps, _, err := GetAcpService().QueryByAppPath(aci.AppContainerId, x_auth_token)
	if err != nil {
		logrus.Errorf("find acp by appPath [%v] failed, error is %v", aci.AppContainerId, err)
		return
	}
	// no acp founded, no need to notify
	if total <= 0 {
		logrus.Warnf("not found acp with appPath [%v], skip notify", aci.AppContainerId)
		return
	}

	// do notify by acp and sgi
	acp := acps[0]
	for _, dep := range acp.Notifies {
		logrus.Debugf("check to notify [%v], caused by delete aci with aciId [%v]", dep.NotifyPath, aciId)
		instanceIds := entity.GetAppInstanceIdsFromGroupInstance(sgi, dep.NotifyPath)
		instanceIds = filterInstanceByScope(dep.Scope, instanceIds, aciId)
		for _, instanceId := range instanceIds {
			// here callFrom supports deleteNotify, updateNotify
			steps, notifyError := GetConfigStepsByAciId(instanceId, callFrom, x_auth_token)
			if notifyError != nil {
				err = notifyError
				logrus.Errorf("can not get configuration steps by instance [%v], error is %v", instanceId, err)
				return
			}
			logrus.Infof("change aci [%v] status to created", instanceId)
			notifyError = GetAciService().updateAciStatus(instanceId, ACI_STATUS_CREATED)
			if notifyError != nil {
				logrus.Errorf("update appinstance %s status to %s error is %v",
					instanceId, ACI_STATUS_CREATED, notifyError)
			}
			configStatus, notifyError := configAppInstance(instanceId, steps)
			if notifyError != nil {
				err = notifyError
				logrus.Errorf("execute configs error, %v", err)
				logrus.Infof("change aci [%v] status to unconfiged", instanceId)
				err2 := GetAciService().updateAciStatus(instanceId, ACI_STATUS_UNCONFIGED)
				if err2 != nil {
					logrus.Errorf("update appinstance %s status to %s error is %v",
						instanceId, ACI_STATUS_UNCONFIGED, err2)
				}
				return
			}
			if !configStatus {
				logrus.Error("execute configs error, ExitCode is not 0")
				logrus.Infof("change aci [%v] status to unconfiged", instanceId)
				err2 := GetAciService().updateAciStatus(instanceId, ACI_STATUS_UNCONFIGED)
				if err2 != nil {
					logrus.Errorf("update appinstance %s status to %s error is %v",
						instanceId, ACI_STATUS_UNCONFIGED, err2)
				}
				// here no return error, contact jzhang@linkernetworks.com
				// err = errors.New("execute configs error, ExitCode is not 0")
				// return
			} else {
				logrus.Infof("change aci [%v] status to configed", instanceId)
				err2 := GetAciService().updateAciStatus(instanceId, ACI_STATUS_CONFIGED)
				if err2 != nil {
					logrus.Errorf("update appinstance %s status to %s error is %v",
						instanceId, ACI_STATUS_CONFIGED, err2)
				}
			}
		}
	}
	return nil
}

func updateSgoStatus(sgi *entity.ServiceGroupInstance, x_auth_token string) {
	if isFinished := isSgiFinished(sgi); isFinished {
		logrus.Infof("sgi is finished, start to change sgi and sgo status to DEPLOYED")

		logrus.Infof("updateSgoStatus sgi life cycle status is %s \n", sgi.LifeCycleStatus)

		// update sgo status
		_, _, err := GetSgoService().UpdateStateBySgiId(sgi.ObjectId.Hex(),
			SGO_STATUS_DEPLOYED, x_auth_token)
		if err != nil {
			logrus.Errorf("update sgo status error is %v", err)
		}

		if sgi.LifeCycleStatus == SGI_STATUS_REPARING {
			logrus.Infof("sgi is finished and in repair status, call repair service")
			//update the instance lifecycle from repairing to Deployed
			_, _, err = GetSgiService().UpdateRepairIdAndStatusById(sgi.ObjectId.Hex(), "", SGI_STATUS_DEPLOYED, x_auth_token)
			if err != nil {
				//call reparing service as repair Failed
				logrus.Errorf("update sgi status for repair error is %v", err)
				GetRepairPolicyService().AnalyzeNotify(REPAIR_ACTION_FAILURE, sgi.RepairId)
				return
			}
			GetRepairPolicyService().AnalyzeNotify(REPAIR_ACTION_SUCCESS, sgi.RepairId)
			return
		}

		// update sgi status
		_, _, err = GetSgiService().UpdateStateById(sgi.ObjectId.Hex(),
			SGI_STATUS_DEPLOYED, x_auth_token)
		if err != nil {
			logrus.Errorf("update sgi status error is %v", err)
		}

		//check notification
		logrus.Infof("start to check the notification")
		err = notificationBySgiId(sgi.ObjectId.Hex(), x_auth_token)
		if err != nil {
			logrus.Errorf("notification error is %v", err)
		}

	} else {
		logrus.Infof("sgi is not finished, start to change sgi and sgo status to DEPLOYING")
		if sgi.LifeCycleStatus == SGI_STATUS_REPARING {
			logrus.Infof("sgi is not finished and in repair status, no changes to lifecycle")
		}

		// update sgi status
		_, _, err := GetSgiService().UpdateStateById(sgi.ObjectId.Hex(),
			SGI_STATUS_DEPLOYING, x_auth_token)
		if err != nil {
			logrus.Errorf("update sgi status error is %v", err)
		}
		// update sgo status
		_, _, err = GetSgoService().UpdateStateBySgiId(sgi.ObjectId.Hex(),
			SGO_STATUS_DEPLOYING, x_auth_token)
		if err != nil {
			logrus.Errorf("update sgo status error is %v", err)
		}
	}
}

func notificationBySgiId(sgiId string, x_auth_token string) (err error) {
	query := bson.M{}
	query["service_group_instance_id"] = sgiId
	sgo, _, err := GetSgoService().queryOneByQuery(query)
	if err != nil {
		logrus.Errorf("get sgo by sgiId [%v] failed, error is %v", sgiId, err)
		return
	}

	sgoId := sgo.ObjectId.Hex()
	logrus.Infof("sgorderid is %v", sgoId)
	if err != nil {
		logrus.Errorf(" error is %v", err)
	}
	query_two := bson.M{}
	query_two["service_group_id"] = sgo.ServiceGroupId
	notify := entity.Notification{}
	err = dao.HandleQueryOne(&notify, dao.QueryStruct{"notification", query_two, 0, 0, ""})
	if err != nil {
		logrus.Errorf("query notification [query=%v] error is %v", query, err)
	}
	notifyAddress := notify.NotifyAddress
	if len(notifyAddress) > 0 {
		logrus.Infof("notifyaddress is %v", notifyAddress)
		url := strings.Join([]string{notifyAddress, "?orderid=", sgoId,
			"&status=", SGO_STATUS_DEPLOYED}, "")

		logrus.Infof("notification notify url is %v", url)
		_, err := httpclient.Http_post(url, "", httpclient.Header{"Content-Type",
			"application/json"}, httpclient.Header{"X-Auth-Token", x_auth_token})

		if err != nil {
			logrus.Errorf("post notification error %s", err)
		}
	}
	return
}

func getAppNumInSg(sg *entity.ServiceGroup) int {
	count := 0
	for i := range sg.Groups {
		group := &sg.Groups[i]
		count += getAppNumInGroup(group)
	}

	return count
}

func getAppNumInGroup(group *entity.Group) int {
	count := 0
	for i := range group.Apps {
		app := &group.Apps[i]
		count += app.Instances
	}

	for i := range group.Groups {
		ingroup := &group.Groups[i]
		count += getAppNumInGroup(ingroup)
	}

	return count
}

func getAppNumInSgi(sgi *entity.ServiceGroupInstance) int {
	count := 0
	for i := range sgi.Groups {
		group := &sgi.Groups[i]
		count += getAppNumInRefinedGroup(group)
	}

	return count
}

func getAppNumInRefinedGroup(group *entity.RefinedGroup) int {
	count := 0
	for i := range group.RefinedApps {
		app := &group.RefinedApps[i]
		count += len(app.InstanceIds)
	}

	for i := range group.RefinedGroups {
		ingroup := &group.RefinedGroups[i]
		count += getAppNumInRefinedGroup(ingroup)
	}

	return count
}

func makeMarathonGroupId(sgId string) (marathonGroupId string) {
	marathonGroupId = sgId
	marathonGroupId += "-" + uuid.NewUUID().String()
	return
}

func setServiceGroupIntanceGroups(sgi *entity.ServiceGroupInstance,
	sg *entity.ServiceGroup) {

	sgi.Groups = []entity.RefinedGroup{}
	for _, group := range sg.Groups {
		sgiGroup := refineAppInstance(group)
		sgi.Groups = append(sgi.Groups, sgiGroup)
	}
}

func refineAppInstance(group entity.Group) (refinedGroup entity.RefinedGroup) {
	refinedGroup.Id = group.Id
	refinedGroup.Dependencies = group.Dependencies
	if group.Apps != nil {
		refinedGroup.RefinedApps = []entity.RefinedApp{}
		for _, app := range group.Apps {
			sgiApp := entity.RefinedApp{
				Id:          app.Id,
				Cpus:        app.Cpus,
				Instances:   app.Instances,
				Mem:         app.Mem,
				InstanceIds: []string{},
			}
			refinedGroup.RefinedApps = append(refinedGroup.RefinedApps, sgiApp)
		}
	}
	if group.Groups != nil {
		refinedGroup.RefinedGroups = []entity.RefinedGroup{}
		for _, g := range group.Groups {
			rg := refineAppInstance(g)
			refinedGroup.RefinedGroups = append(refinedGroup.RefinedGroups, rg)
		}
	}
	return
}

func _refineEnv(sgo entity.ServiceGroupOrder, group *entity.Group) {
	logrus.Debugln("_refineEnv in group:", group.Id)
	sgi_id := sgo.ServiceGroupInstanceId
	sgo_id := sgo.ObjectId.Hex()
	so_id := sgo.ServiceOfferingId
	sg_id := sgo.ServiceGroupId
	if group.Apps != nil {
		newApps := []entity.App{}
		for _, app := range group.Apps {
			app.Env["LINKER_SERVICE_GROUP_INSTANCE_ID"] = sgi_id
			app.Env["LINKER_SERVICE_ORDER_ID"] = sgo_id
			app.Env["LINKER_SERVICE_OFFERING_ID"] = so_id
			app.Env["LINKER_SERVICE_GROUP_ID"] = sg_id
			newApps = append(newApps, app)
		}
		group.Apps = append(newApps)
	}
	if group.Groups != nil {
		newGroups := []entity.Group{}
		for _, g := range group.Groups {
			_refineEnv(sgo, &g)
			newGroups = append(newGroups, g)
		}
		group.Groups = append(newGroups)
	}
	return
}

func refineEnv(sgo entity.ServiceGroupOrder, so *entity.ServiceGroup) {
	newGroups := []entity.Group{}
	for _, group := range so.Groups {
		_refineEnv(sgo, &group)
		newGroups = append(newGroups, group)
	}
	so.Groups = append(newGroups)
}

func refineParam(sgo entity.ServiceGroupOrder, so *entity.ServiceGroup) {
	for _, offeringParam := range sgo.OfferingParameters {
		appId := offeringParam.AppId
		app, _ := entity.GetAppFromServiceGroup(so, appId)
		_refineParam(app, &offeringParam)
	}
}

func _refineParam(app *entity.App, offeringParam *entity.OfferingParameter) {
	container := app.Container
	newParamValue := strings.Join([]string{offeringParam.ParamName, "=", offeringParam.ParamValue}, "")
	newParameters := []entity.Parameter{}
	replaceFlag := false
	for _, param := range app.Container.Docker.Parameters {
		paramArr := strings.Split(param.Value, "=")
		if paramArr[0] == offeringParam.ParamName {
			replaceFlag = true
			param.Value = newParamValue
			newParameters = append(newParameters, param)
			logrus.Debugf("refined params:%v", param)
		} else {
			newParameters = append(newParameters, param)
		}
	}

	if !replaceFlag {
		newParam := entity.Parameter{}
		newParam.Editable = true
		newParam.Description = "new from order"
		newParam.Key = "env"
		newParam.Value = newParamValue
		newParameters = append(newParameters, newParam)
		logrus.Debugf("add params:%v", newParam)
	}
	container.Docker.Parameters = newParameters
	app.Container = container
}

func generateMarathonGroup(sg entity.ServiceGroup,
	sgo entity.ServiceGroupOrder) (sgjncson string, err error) {

	sg.Id = sgo.MarathonGroupId
	sgjson, _ := json.Marshal(sg)

	var sgnc *entity.ServiceGroupNoContainer
	err = json.Unmarshal(sgjson, &sgnc)

	if err != nil {
		logrus.Errorf("Unmarshal ServiceGroupNoContainer err is %v", err)
		return
	}

	jsonresult, _ := json.Marshal(*sgnc)
	sgjncson = string(jsonresult)
	return
}

func removeDockerContainerInfoInApp(app *entity.App) (appncjson string, err error) {
	appjson, _ := json.Marshal(app)

	var appnc *entity.AppNoContainer
	err = json.Unmarshal(appjson, &appnc)

	if err != nil {
		logrus.Errorf("Unmarshal AppNoContainer err is %v", err)
		return
	}

	jsonresult, _ := json.Marshal(*appnc)
	appncjson = string(jsonresult)
	return
}

func postToMarathonGroup(reqBody string) (err error) {
	marathonEndpoint, err := common.UTIL.ZkClient.GetMarathonEndpoint()
	if err != nil {
		logrus.Errorf("get marathon endpoint failed, error is %v", err)
		return
	}
	logrus.Debugf("marathon endpoint is %v", marathonEndpoint)

	url := strings.Join([]string{"http://", marathonEndpoint, "/v2/groups"}, "")
	// body := ioutil.NopCloser(strings.NewReader(reqBody))
	logrus.Debugf("the body of group which send to marathon is %v", reqBody)

	resp, err := httpclient.Http_post(url, reqBody,
		httpclient.Header{"Content-Type", "application/json"})
	defer resp.Body.Close()
	if err != nil {
		logrus.Errorf("post group to marathon failed, error is %v", err)
		return
	}

	data, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		logrus.Errorf("marathon returned error code is %v", resp.StatusCode)
		logrus.Errorf("detail is %v", string(data))
		err = errors.New(string(data))
		return
	}
	return
}

func deleteToMarathonGroup(groupId string) (err error) {
	marathonEndpoint, err := common.UTIL.ZkClient.GetMarathonEndpoint()
	if err != nil {
		logrus.Errorf("get marathon endpoint failed, error is %v", err)
		return
	}
	logrus.Debugf("marathon endpoint is %v", marathonEndpoint)

	url := strings.Join([]string{"http://", marathonEndpoint, "/v2/groups",
		groupId, "?force=true"}, "")
	logrus.Debugf("the id of group which send to marathon is %v", groupId)

	resp, err := httpclient.Http_delete(url, "",
		httpclient.Header{"Content-Type", "application/json"})
	defer resp.Body.Close()
	if err != nil {
		logrus.Errorf("delete group [%v] in marathon failed, error is %v", groupId, err)
		return
	}

	data, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		logrus.Errorf("marathon returned error code is %v", resp.StatusCode)
		logrus.Errorf("detail is %v", string(data))
		err = errors.New(string(data))
		return
	}
	return
}

func putToMarathonApp(appId, reqBody string) (err error) {
	marathonEndpoint, err := common.UTIL.ZkClient.GetMarathonEndpoint()
	if err != nil {
		logrus.Errorf("get marathon endpoint failed, error is %v", err)
		return
	}
	logrus.Debugf("marathon endpoint is %v", marathonEndpoint)

	url := strings.Join([]string{"http://", marathonEndpoint, "/v2/apps",
		appId}, "")
	logrus.Debugf("change app url of marathon is %v", url)

	resp, err := httpclient.Http_put(url, reqBody,
		httpclient.Header{"Content-Type", "application/json"})
	defer resp.Body.Close()
	if err != nil {
		logrus.Errorf("update app [%v] in marathon failed, error is %v", reqBody, err)
		return
	}

	data, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		logrus.Errorf("marathon returned error code is %v", resp.StatusCode)
		logrus.Errorf("detail is %v", string(data))
		err = errors.New(string(data))
		return
	}
	return
}
