package services

import (
	"errors"
	"github.com/Sirupsen/logrus"
	"gopkg.in/mgo.v2/bson"
	"io/ioutil"
	"encoding/json"
	"linkernetworks.com/linker_common_lib/httpclient"
	"linkernetworks.com/linker_common_lib/persistence/dao"
	"linkernetworks.com/linker_common_lib/persistence/entity"
	"strconv"
	"strings"
	"sync"
	"time"
	"linkernetworks.com/linker_cluster/common"
	"linkernetworks.com/linker_common_lib/rest/response"
)

var (
	hostService            *HostService = nil
	onceHost               sync.Once
	HOST_STATUS_CREATED    = "CREATED"
	HOST_STATUS_TERMINATED = "TERMINATED"
	HOST_STATUS_DEPLOYED   = "DEPLOYED"
	HOST_STATUS_DEPLOYING  = "DEPLOYING"
	HOST_STATUS_FAILED     = "FAILED"

	HOST_ERROR_CREATE string = "E40100"
	HOST_ERROR_UPDATE string = "E40101"
	HOST_ERROR_DELETE string = "E40102"
	HOST_ERROR_QUERY  string = "E40103"
	HOST_ERROR_TERMINATED string = "E40104"
)

type HostService struct {
	collectionName string
}

func GetHostService() *HostService {
	onceHost.Do(func() {
		logrus.Debugf("Once called from hostsService ......................................")
		hostService = &HostService{"hosts"}
	})
	return hostService
}

func (p *HostService) Create(host entity.Host, x_auth_token string) (newHost entity.Host,
	errorCode string, err error) {
	logrus.Infof("start to create host [%v]", host)
	// do authorize first
	if authorized := GetAuthService().Authorize("create_host", x_auth_token, "", p.collectionName); !authorized {
		err = errors.New("required opertion is not authorized!")
		errorCode = COMMON_ERROR_UNAUTHORIZED
		logrus.Errorf("create host [%v] error is %v", host, err)
		return
	}

	// generate ObjectId
	if !bson.IsObjectIdHex(host.ObjectId.Hex()) {
		host.ObjectId = bson.NewObjectId()
	}

	token, err := GetTokenById(x_auth_token)
	if err != nil {
		errorCode = HOST_ERROR_CREATE
		logrus.Errorf("get token failed when create host [%v], error is %v", host, err)
		return
	}

	// set token_id and user_id from token
	host.Tenant_id = token.Tenant.Id
	host.User_id = token.User.Id

	// set created_time and updated_time
	host.TimeCreate = dao.GetCurrentTime()
	host.TimeUpdate = host.TimeCreate
	host.Date = int(time.Now().Unix())

	// insert bson to mongodb
	err = dao.HandleInsert(p.collectionName, host)
	if err != nil {
		errorCode = HOST_ERROR_CREATE
		logrus.Errorf("insert host [%v] to db error is %v", host, err)
		return
	}

	newHost = host

	return
}

func (p *HostService) QueryById(objectId string, x_auth_token string) (host entity.Host,
	errorCode string, err error) {
	if !bson.IsObjectIdHex(objectId) {
		err = errors.New("invalide ObjectId.")
		errorCode = COMMON_ERROR_INVALIDATE
		return
	}

	// do authorize first
	if authorized := GetAuthService().Authorize("get_host", x_auth_token, objectId, p.collectionName); !authorized {
		err = errors.New("required opertion is not authorized!")
		errorCode = COMMON_ERROR_UNAUTHORIZED
		logrus.Errorf("get host with objectId [%v] error is %v", objectId, err)
		return
	}

	var selector = bson.M{}
	selector["_id"] = bson.ObjectIdHex(objectId)
	host = entity.Host{}
	err = dao.HandleQueryOne(&host, dao.QueryStruct{p.collectionName, selector, 0, 0, ""})
	if err != nil {
		logrus.Errorf("query host [objectId=%v] error is %v", objectId, err)
		errorCode = HOST_ERROR_QUERY
	}
	return
}

func (p *HostService) QueryContainersById(objectId string,
	skip, limit int, x_auth_token string) (total int,
	appInstances []entity.AppContainerInstance, errorCode string, err error) {
	if !bson.IsObjectIdHex(objectId) {
		err = errors.New("invalide ObjectId.")
		errorCode = COMMON_ERROR_INVALIDATE
		return
	}

	// do authorize first
	if authorized := GetAuthService().Authorize("get_host", x_auth_token, objectId, p.collectionName); !authorized {
		err = errors.New("required opertion is not authorized!")
		errorCode = COMMON_ERROR_UNAUTHORIZED
		logrus.Errorf("get host with objectId [%v] error is %v", objectId, err)
		return
	}

	// get host by id
	var selector = bson.M{}
	selector["_id"] = bson.ObjectIdHex(objectId)
	host := entity.Host{}
	err = dao.HandleQueryOne(&host, dao.QueryStruct{p.collectionName, selector, 0, 0, ""})
	if err != nil {
		logrus.Errorf("query host [objectId=%v] failed, error is %v", objectId, err)
		errorCode = HOST_ERROR_QUERY
	}

	// get cluster's master ip address to call controller's api get acis
	masters, err := GetClusterService().getMasterNodes(host.ClusterId)
	if err != nil {
		logrus.Errorf("get cluster [%v] masters failed, error is %v", host.ClusterId, err)
		return
	}

	if len(masters) <= 0 {
		err = errors.New("Can not find master node.")
		logrus.Errorf("%v", err)
		errorCode = CLUSTER_ERROR_QUERY
		return
	}

	if masters[0].Status != HOST_STATUS_DEPLOYED {
		err = errors.New("Master node is not actived")
		logrus.Errorf("%v", err)
		errorCode = CLUSTER_ERROR_QUERY
		return
	}

	masterControllerEndpoint := strings.Join([]string{"http://", masters[0].IP,
		":8081", "/v1/appInstances", "?count=true", "&skip=",
		strconv.Itoa(skip), "&limit=", strconv.Itoa(limit), "&slave=", host.IP}, "")
	logrus.Debugf("get appInstance from api [%v]", masterControllerEndpoint)

	resp, err := httpclient.Http_get(masterControllerEndpoint, "",
		httpclient.Header{"X-Auth-Token", x_auth_token})
	if err != nil {
		logrus.Errorf("get appInstance request failed, error is %v", err)
		appInstances = make([]entity.AppContainerInstance, 0)
		return
	}
	data, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		err = errors.New("get hosts by cluster id failed")
		logrus.Errorf("get appInstance by slave [%v] failed, error is %v", host.IP, string(data))
		return
	}
	// logrus.Debugf("get appInstance by slave [%v] returns %v", host.IP, string(data))
	appInstances = []entity.AppContainerInstance{}
	err = getRetFromResponse(data, &appInstances)
	if err != nil {
		logrus.Errorf("parse appInstance response failed, error is %v", err)
	}
	total, err = getCountFromResponse(data)
	if err != nil {
		logrus.Errorf("get count from appIstance response failed, error is %v", err)
	}
	return
}

func (p *HostService) QueryAllByName(cluster_id string, skip int,
	limit int, x_auth_token string) (total int, hosts []entity.Host,
	errorCode string, err error) {
	if strings.TrimSpace(cluster_id) == "" {
		return p.QueryAll(skip, limit, x_auth_token)
	}
	query := bson.M{}
	query["cluster_id"] = cluster_id
	
	return p.queryByQuery(query, skip, limit, x_auth_token, false)
}

func (p *HostService) QueryAll(skip int, limit int, x_auth_token string) (total int,
	hosts []entity.Host, errorCode string, err error) {
	return p.queryByQuery(bson.M{}, skip, limit, x_auth_token, false)
}

func (p *HostService) queryByQuery(query bson.M, skip int, limit int,
	x_auth_token string, skipAuth bool) (total int, hosts []entity.Host,
	errorCode string, err error) {
	authQuery := bson.M{}
	if !skipAuth {
		// get auth query from auth first
		authQuery, err = GetAuthService().BuildQueryByAuth("list_host", x_auth_token)
		if err != nil {
			logrus.Errorf("get auth query by token [%v] error is %v", x_auth_token, err)
			errorCode = COMMON_ERROR_INTERNAL
			return
		}
	}

	selector := generateQueryWithAuth(query, authQuery)
	hosts = []entity.Host{}
	// fix : "...." sort by time_create
	queryStruct := dao.QueryStruct{p.collectionName, selector, skip, limit, "lable"}
	total, err = dao.HandleQueryAll(&hosts, queryStruct)
	if err != nil {
		logrus.Errorf("query hosts by query [%v] error is %v", query, err)
		errorCode = HOST_ERROR_QUERY

	}
	return
}

func (p *HostService) DeleteById(objectId string, x_auth_token string) (errorCode string, err error) {
	logrus.Infof("start to delete Host with objectId [%v]", objectId)
	// do authorize first
	if authorized := GetAuthService().Authorize("delete_host", x_auth_token, objectId, p.collectionName); !authorized {
		err = errors.New("required opertion is not authorized!")
		errorCode = COMMON_ERROR_UNAUTHORIZED
		logrus.Errorf("delete host with objectId [%v] error is %v", objectId, err)
		return
	}
	if !bson.IsObjectIdHex(objectId) {
		err = errors.New("invalide ObjectId.")
		errorCode = COMMON_ERROR_INVALIDATE
		return
	}

	//TODO yun delete
	
	ip := common.UTIL.Props.GetString("http.proxy.url","")
	port := common.UTIL.Props.GetString("http.proxy.port","")
	
	host, _, err := p.QueryById(objectId, x_auth_token)
	if err != nil {
		logrus.Errorf("delete host [objectId=%v] error is %v", objectId, err)
		errorCode = HOST_ERROR_QUERY
		return
	}
	
	
	cloudProvider := host.Flavor.ProviderType
	instanceId := host.CloudProxyId
	time_now := time.Now().Format("20150102150405")
	timeString := time_now + "000"
	timeStamp := timeString
	
	sinatureString := strings.Join([]string{"linker_cloud_pay2015","cloudProvider=",cloudProvider,"&instanceId=",instanceId,"&timeStamp=",timeStamp},"")
	logrus.Infoln("sinatureString is %v",sinatureString)
	signature := HashString(sinatureString)
	
	deleteString := strings.Join([]string{"cloudProvider=",cloudProvider,"&instanceId=",instanceId,"&timeStamp=",timeStamp},"")
	logrus.Infoln("deleteString is %v",deleteString)
	
	url := strings.Join([]string{"http://",ip,":",port,"/deleteInstance?",deleteString,"&signature=",signature},"")
	logrus.Infoln("url is %v",url)
	
	resp,errquery := httpclient.Http_get(url,"",httpclient.Header{"Content-Type", "application/json"})
	if errquery != nil {
		logrus.Errorf("queryerr is %v", errquery)
		err = errors.New("get cloud proxy is error!")
		return
	}
	
	data, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		err = errors.New("get cloud proxy failed")
		logrus.Errorf("get cloud proxy failed, error is %v", string(data))
		return
	}
	var respo *response.Response
	respo = new(response.Response)
	err = json.Unmarshal(data, &respo)
	if err != nil {
		logrus.Errorf("err is %v",err)
	}
	istrue := respo.Success
	if istrue {
		_,_,err = p.UpdateStateById(objectId,HOST_STATUS_TERMINATED,x_auth_token)
		if err != nil {
			logrus.Errorf("delete host [objectId=%v] error is %v", objectId, err)
			errorCode = HOST_ERROR_DELETE
			return
		}
	}
	
	
	return
}

func (p *HostService) UpdateById(objectId string, host entity.Host, x_auth_token string) (created bool,
	errorCode string, err error) {
	logrus.Infof("start to update host [%v]", host)
	// do authorize first
	if authorized := GetAuthService().Authorize("update_host", x_auth_token, objectId, p.collectionName); !authorized {
		err = errors.New("required opertion is not authorized!")
		errorCode = COMMON_ERROR_UNAUTHORIZED
		logrus.Errorf("update host with objectId [%v] error is %v", objectId, err)
		return
	}

	if !bson.IsObjectIdHex(objectId) {
		err = errors.New("invalide ObjectId.")
		errorCode = COMMON_ERROR_INVALIDATE
		return
	}

	// FIXING
	//	hostquery, _, _  := p.QueryById(objectId, x_auth_token)
	var selector = bson.M{}
	selector["_id"] = bson.ObjectIdHex(objectId)

	host.ObjectId = bson.ObjectIdHex(objectId)
	host.TimeUpdate = dao.GetCurrentTime()

	logrus.Infof("start to change host")
	err = dao.HandleUpdateByQueryPartial(p.collectionName, selector, &host)
	//	created, err = dao.HandleUpdateOne(&host, dao.QueryStruct{p.collectionName, selector, 0, 0, ""})
	if err != nil {
		logrus.Errorf("update host [%v] error is %v", host, err)
		errorCode = HOST_ERROR_UPDATE
	}
	created = true
	return
}

func (p *HostService) QueryAllUnterminated(cluster_id string, skip int, limit int, x_auth_token string) (total int,
	hosts []entity.Host, errorCode string, err error) {
	query := bson.M{}
	if strings.TrimSpace(cluster_id) == "" {
		query["status"] = bson.M{"$ne": "TERMINATED"}
		return p.queryByQuery(query, skip, limit, x_auth_token, false)
	}else{
		query["cluster_id"] = cluster_id
		query["status"] = bson.M{"$ne": "TERMINATED"}
		return p.queryByQuery(query, skip, limit, x_auth_token, false)
	}
	return	
}

func (p *HostService) UpdateStateById(objectId string, newState string, x_auth_token string) (created bool,
	errorCode string, err error) {
	logrus.Infof("start to update host by objectId [%v] status to %v", objectId, newState)
	// do authorize first
	if authorized := GetAuthService().Authorize("update_host", x_auth_token, objectId, p.collectionName); !authorized {
		err = errors.New("required opertion is not authorized!")
		errorCode = COMMON_ERROR_UNAUTHORIZED
		logrus.Errorf("update host with objectId [%v] status to [%v] failed, error is %v", objectId, newState, err)
		return
	}
	// validate objectId
	if !bson.IsObjectIdHex(objectId) {
		err = errors.New("invalide ObjectId.")
		errorCode = COMMON_ERROR_INVALIDATE
		return
	}
	host, _, err := p.QueryById(objectId, x_auth_token)
	if err != nil {
		logrus.Errorf("get host by objeceId [%v] failed, error is %v", objectId, err)
		return
	}
	if host.Status == newState {
		logrus.Infof("this host [%v] is already in state [%v]", host, newState)
		return false, "", nil
	}
	var selector = bson.M{}
	selector["_id"] = bson.ObjectIdHex(objectId)

	change := bson.M{"status":newState,"time_update":dao.GetCurrentTime()}
	err = dao.HandleUpdateByQueryPartial(p.collectionName, selector, change)
	if err != nil {
		logrus.Errorf("update host with objectId [%v] status to [%v] failed, error is %v", objectId, newState, err)
		created = false
		return
	}
	created = true
	return

}

func (p *HostService) UpdateHostById(objectId string, newState string, ip string, CloudProxyId string, DockerVersion string, x_auth_token string) (created bool,
	errorCode string, err error) {
	logrus.Infof("start to update host by objectId [%v] status to %v", objectId, newState)
	// do authorize first
	if authorized := GetAuthService().Authorize("update_host", x_auth_token, objectId, p.collectionName); !authorized {
		err = errors.New("required opertion is not authorized!")
		errorCode = COMMON_ERROR_UNAUTHORIZED
		logrus.Errorf("update host with objectId [%v] status to [%v] failed, error is %v", objectId, newState, err)
		return
	}
	// validate objectId
	if !bson.IsObjectIdHex(objectId) {
		err = errors.New("invalide ObjectId.")
		errorCode = COMMON_ERROR_INVALIDATE
		return
	}
	host, _, err := p.QueryById(objectId, x_auth_token)
	if err != nil {
		logrus.Errorf("get host by objeceId [%v] failed, error is %v", objectId, err)
		return
	}
	if host.Status == newState {
		logrus.Infof("this host [%v] is already in state [%v]", host, newState)
		return false, "", nil
	}
	var selector = bson.M{}
	selector["_id"] = bson.ObjectIdHex(objectId)

	change := bson.M{"status":newState,"time_update":dao.GetCurrentTime(),"ip":ip,"cloudproxy_id":CloudProxyId,"docker_version":DockerVersion}
	err = dao.HandleUpdateByQueryPartial(p.collectionName, selector, change)
	if err != nil {
		logrus.Errorf("update host with objectId [%v] status to [%v] failed, error is %v", objectId, newState, err)
		created = false
		return
	}
	created = true
	return

}