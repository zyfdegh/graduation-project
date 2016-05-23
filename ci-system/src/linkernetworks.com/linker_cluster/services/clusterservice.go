package services

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
	"gopkg.in/mgo.v2/bson"
	"linkernetworks.com/linker_cluster/common"
	"linkernetworks.com/linker_common_lib/httpclient"
	"linkernetworks.com/linker_common_lib/persistence/dao"
	"linkernetworks.com/linker_common_lib/persistence/entity"
	"linkernetworks.com/linker_common_lib/rest/response"
)

var (
	clusterService            *ClusterService = nil
	onceCluster               sync.Once
	CLUSTER_STATUS_CREATED    = "CREATED"
	CLUSTER_STATUS_TERMINATED = "TERMINATED"
	CLUSTER_STATUS_DEPLOYED   = "DEPLOYED"
	CLUSTER_STATUS_DEPLOYING  = "DEPLOYING"
	CLUSTER_STATUS_FAILED     = "FAILED"

	CLUSTER_ERROR_CREATE string = "E40000"
	CLUSTER_ERROR_UPDATE string = "E40001"
	CLUSTER_ERROR_DELETE string = "E40002"
	CLUSTER_ERROR_UNIQUE string = "E40003"
	CLUSTER_ERROR_QUERY  string = "E40004"
)

type ClusterService struct {
	collectionName string
}

func GetClusterService() *ClusterService {
	onceCluster.Do(func() {
		logrus.Debugf("Once called from clusterService ......................................")
		clusterService = &ClusterService{"cluster"}
	})
	return clusterService

}

func (p *ClusterService) Create(cluster entity.Cluster, x_auth_token string) (newCluster entity.Cluster,
	errorCode string, err error) {
	logrus.Infof("start to create cluster [%v]", cluster)

	// do authorize first
	if authorized := GetAuthService().Authorize("create_cluster", x_auth_token, "", p.collectionName); !authorized {
		err = errors.New("required opertion is not authorized!")
		errorCode = COMMON_ERROR_UNAUTHORIZED
		logrus.Errorf("create cluster [%v] error is %v", cluster, err)
		return
	}
	/*
		// check quota
		if  isquato := GetQuotaService().A(); !isquato {
			err = errors.New("quota is not allowed")
			errorCode =
			logrus.Errorf("check quota [%v] error is %v", cluster, err)
			return
		}

		//check cash ...billing
		if isbilling := GetBillingService().B(); !isbilling {
			err = errors.New("your account is not enough")
			errorCode =
			logrus.Errorf("check cash [%v] error is %v", cluster, err)
			return
		}	*/

	name := cluster.ClusterName
	_, cluster_exists, _, err := p.QueryAllUnterminated(name, 0, 1, x_auth_token)
	if err == nil && len(cluster_exists) > 0 {
		err = errors.New("the name of cluster must be unique!")
		errorCode = CLUSTER_ERROR_UNIQUE
		logrus.Errorf("create cluster [%v] error is %v", cluster, err)
		return
	}

	// generate ObjectId
	cluster.ObjectId = bson.NewObjectId()

	token, err := GetTokenById(x_auth_token)
	if err != nil {
		errorCode = COMMON_ERROR_INTERNAL
		logrus.Errorf("get token failed when create cluster [%v], error is %v", cluster, err)
		return
	}

	// set token_id and user_id from token
	cluster.Tenant_id = token.Tenant.Id
	cluster.User_id = token.User.Id

	// set created_time and updated_time
	cluster.TimeCreate = dao.GetCurrentTime()
	cluster.TimeUpdate = cluster.TimeCreate
	cluster.Status = CLUSTER_STATUS_DEPLOYING
	if !bson.IsObjectIdHex(cluster.Flavor.ObjectId.Hex()) {
		cluster.Flavor.ObjectId = bson.NewObjectId()
	}

	// insert bson to mongodb
	err = dao.HandleInsert(p.collectionName, cluster)
	if err != nil {
		errorCode = CLUSTER_ERROR_CREATE
		logrus.Errorf("create cluster [%v] to bson error is %v", cluster, err)
		return
	}

	//TODO: check whether user has securityGroup,
	// if no, create securityGroup for user, and save mapping to db
	// if yes, using existed securityGroup when creating the VM

	//TODO: if user choose AWS, check whether user has keypair
	// if no, create new keypair for user, and save the privatekey too db
	// if no, using existed keypair when creating the VM

	instances := cluster.Instances
	hosts := make([]entity.Host, instances)

	for i, host := range hosts {
		go p.createHost(i, host, cluster, x_auth_token)
	}

	newCluster = cluster

	return

}

func (p *ClusterService) createHost(index int, host entity.Host, cluster entity.Cluster, x_auth_token string) (errorCode string, err error) {
	logrus.Infof("start to create host [%v] of cluster [%v]", index, cluster)
	// the first node of cluster is the mater
	if index == 0 {
		host.IsMasterNode = true
	}
	host.ObjectId = bson.NewObjectId()
	host.ClusterId = cluster.ObjectId.Hex()
	host.ClusterName = cluster.ClusterName
	host.Status = HOST_STATUS_DEPLOYING
	host.Flavor = cluster.Flavor
	host.Lable = index
	host.HostName = "node" + strconv.Itoa(index+1)
	host.Lable = index
	_, _, err = GetHostService().Create(host, x_auth_token)
	if err != nil {
		logrus.Errorf("save hosts failed, err is %v", err)
		errorCode = HOST_ERROR_CREATE
		_, _, err = p.UpdateStateById(cluster.ObjectId.Hex(), CLUSTER_STATUS_FAILED, x_auth_token)
		if err != nil {
			logrus.Errorf("update cluster status err is %v", err)
		}
	}

	if cluster.Flavor.ProviderType == "AWS" {
		logrus.Infof("start to create aws host")
		cloudProvider := cluster.Flavor.ProviderType
		securityGroupId := "sg-12389e77"
		securityKey := "linkerAWS"
		hostName := host.HostName
		objId := host.ObjectId.Hex()
		errorCode, err = p.CreateAWSOrAliyunHost(cluster, objId, cloudProvider, securityGroupId, securityKey, hostName, x_auth_token)
		if err != nil {
			logrus.Errorf("create aws aliyun is err")
			return
		}

	} else if cluster.Flavor.ProviderType == "Aliyun" {
		logrus.Infof("start to create aliyun host")
		cloudProvider := cluster.Flavor.ProviderType
		securityGroupId := "sg-25z0k9nc9"
		securityKey := "Baoyun5820"
		hostName := host.HostName
		objId := host.ObjectId.Hex()
		errorCode, err = p.CreateAWSOrAliyunHost(cluster, objId, cloudProvider, securityGroupId, securityKey, hostName, x_auth_token)
		if err != nil {
			logrus.Errorf("create aliyun host is err")
			return
		}
	}
	
	
	
	//TODO: call cloud_proxy api to create VM on AWS or AliCloud

	//TODO: update host status

	return
}

func (p *ClusterService) CreateAWSOrAliyunHost(cluster entity.Cluster, objId string,
	cloudProvider string, securityGroupId string, securityKey string,
	hostName string, x_auth_token string) (errorCode string, err error) {
	
	instanceType := cluster.Flavor.Type
	time_now := time.Now().Format("20150102150405")
	timeString := time_now + "000"
	timeStamp := timeString

	createString := strings.Join([]string{"linker_cloud_pay2015", "cloudProvider=", cloudProvider, "&hostName=", hostName, "&instanceType=", instanceType, "&securityGroupId=", securityGroupId, "&securityKey=", securityKey, "&timeStamp=", timeStamp}, "")
	logrus.Infoln("createString is %v", createString)

	signature := HashString(createString)

	createString2 := strings.Join([]string{"cloudProvider=", cloudProvider, "&hostName=", hostName, "&instanceType=", instanceType, "&securityGroupId=", securityGroupId, "&securityKey=", securityKey, "&timeStamp=", timeStamp}, "")
	logrus.Infoln("createString2 is %v", createString2)

	ip := common.UTIL.Props.GetString("http.proxy.url", "")
	port := common.UTIL.Props.GetString("http.proxy.port", "")

	url := strings.Join([]string{"http://", ip, ":", port, "/createInstance?", createString2, "&signature=", signature}, "")
	logrus.Infoln("url is %v", url)

	resp, errquery := httpclient.Http_get(url, "", httpclient.Header{"Content-Type", "application/json"})
	if errquery != nil {
		logrus.Errorf("get cloud proxy create host is %v", errquery)
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
		logrus.Errorf("err is %v", err)
	}

	cloudProxy := entity.CloudProxy{}
	err = getRetFromResponse(data, &cloudProxy)
	if err != nil {
		logrus.Errorf("parse cloudProxy failed, error is %v", err)
		return
	}

	istrue := respo.Success
	cloudProxyId := cloudProxy.InstanceId
	logrus.Infoln("cloudProxyId is %v", cloudProxyId)
	requestId := cloudProxy.RequestId
	logrus.Infoln("requestId is %v", requestId)

	if istrue && len(cloudProxyId) > 0 {
		cloudProvider := cluster.Flavor.ProviderType
		instanceId := cloudProxyId
		time_now := time.Now().Format("20150102150405")
		timeString := time_now + "000"
		timeStamp := timeString
		describeString := strings.Join([]string{"linker_cloud_pay2015", "cloudProvider=", cloudProvider, "&instanceId=", instanceId, "&timeStamp=", timeStamp}, "")
		logrus.Infoln("describeString is %v", describeString)
		signature := HashString(describeString)
		describeString2 := strings.Join([]string{"cloudProvider=", cloudProvider, "&instanceId=", instanceId, "&timeStamp=", timeStamp}, "")
		logrus.Infoln("describeString2 is %v", describeString2)

		url := strings.Join([]string{"http://", ip, ":", port, "/describeInstances?", describeString2, "&signature=", signature}, "")
		logrus.Infoln("url is %v", url)
		resp, errquery := httpclient.Http_get(url, "", httpclient.Header{"Content-Type", "application/json"})

		if errquery != nil {
			logrus.Errorf("errquery is %v", errquery)
			err = errors.New("get cloud proxy is error!")
			return
		}
		data, _ := ioutil.ReadAll(resp.Body)
		if resp.StatusCode >= 400 {
			err = errors.New("get cloud proxy failed")
			logrus.Errorf("get cloud proxy failed, error is %v", string(data))
			return
		}
		hostDeacribe := entity.HostDeacribe{}
		err = getRetFromResponse(data, &hostDeacribe)
		if err != nil {
			logrus.Errorf("parse hostDeacribe failed, error is %v", err)
			return
		}
		//FIXING must to PublicIpAddressList not IpAddress
		hostIp := hostDeacribe.IpAddress
		
		
		//FIXING  where to get dockerversion
		dockerversion := ""
		_, _, err = GetHostService().UpdateHostById(objId, HOST_STATUS_DEPLOYED, hostIp, cloudProxyId, dockerversion, x_auth_token)
		if err != nil {
			logrus.Errorf("delete host [objectId=%v] error is %v", objId, err)
			errorCode = HOST_ERROR_UPDATE
			return
		}
	}else {
		_, _, err = GetHostService().UpdateStateById(objId,HOST_STATUS_TERMINATED,x_auth_token)
		if err != nil {
			logrus.Errorf("update host status err is %v", err)
			return
		}
	}

	return
}

func (p *ClusterService) QueryAllByName(cluster_name string, skip int,
	limit int, x_auth_token string) (total int, clusters []entity.Cluster,
	errorCode string, err error) {
	// if cluster_name is empty, query all
	if strings.TrimSpace(cluster_name) == "" {
		return p.QueryAll(skip, limit, x_auth_token)
		return
	}
	query := bson.M{}
	query["cluster_name"] = cluster_name
	return p.queryByQuery(query, skip, limit, x_auth_token, false)
}

func (p *ClusterService) QueryAllById(cluster_id string, skip int,
	limit int, x_auth_token string) (total int, clusters []entity.Cluster,
	errorCode string, err error) {
	// if cluster_name is empty, query all
	if strings.TrimSpace(cluster_id) == "" {
		return p.QueryAll(skip, limit, x_auth_token)
		return
	}
	query := bson.M{}
	query["_id"] = bson.ObjectIdHex(cluster_id)
	return p.queryByQuery(query, skip, limit, x_auth_token, false)
}

func (p *ClusterService) QueryAll(skip int, limit int, x_auth_token string) (total int,
	clusters []entity.Cluster, errorCode string, err error) {
	return p.queryByQuery(bson.M{}, skip, limit, x_auth_token, false)
}

func (p *ClusterService) queryByQuery(query bson.M, skip int, limit int,
	x_auth_token string, skipAuth bool) (total int, clusters []entity.Cluster,
	errorCode string, err error) {
	authQuery := bson.M{}
	if !skipAuth {
		// get auth query from auth service first
		authQuery, err = GetAuthService().BuildQueryByAuth("list_cluster", x_auth_token)
		if err != nil {
			logrus.Errorf("get auth query by token [%v] error is %v", x_auth_token, err)
			errorCode = COMMON_ERROR_INTERNAL
			return
		}
	}

	selector := generateQueryWithAuth(query, authQuery)
	clusters = []entity.Cluster{}
	queryStruct := dao.QueryStruct{p.collectionName, selector, skip, limit, ""}
	total, err = dao.HandleQueryAll(&clusters, queryStruct)
	if err != nil {
		logrus.Errorf("query clusters by query [%v] error is %v", query, err)
		errorCode = CLUSTER_ERROR_QUERY
	}
	return
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

func (p *ClusterService) UpdateStateById(objectId string, newState string, x_auth_token string) (created bool,
	errorCode string, err error) {
	logrus.Infof("start to update cluster by objectId [%v] status to %v", objectId, newState)
	// do authorize first
	if authorized := GetAuthService().Authorize("update_cluster", x_auth_token, objectId, p.collectionName); !authorized {
		err = errors.New("required opertion is not authorized!")
		errorCode = COMMON_ERROR_UNAUTHORIZED
		logrus.Errorf("update cluster with objectId [%v] status to [%v] failed, error is %v", objectId, newState, err)
		return
	}
	// validate objectId
	if !bson.IsObjectIdHex(objectId) {
		err = errors.New("invalide ObjectId.")
		errorCode = COMMON_ERROR_INVALIDATE
		return
	}
	cluster, _, err := p.QueryById(objectId, x_auth_token)
	if err != nil {
		logrus.Errorf("get cluster by objeceId [%v] failed, error is %v", objectId, err)
		return
	}
	if cluster.Status == newState {
		logrus.Infof("this cluster [%v] is already in state [%v]", cluster, newState)
		return false, "", nil
	}
	var selector = bson.M{}
	selector["_id"] = bson.ObjectIdHex(objectId)

	change := bson.M{"status": newState, "time_update": dao.GetCurrentTime()}
	err = dao.HandleUpdateByQueryPartial(p.collectionName, selector, change)
	if err != nil {
		logrus.Errorf("update cluster with objectId [%v] status to [%v] failed, error is %v", objectId, newState, err)
	}
	created = true
	return

}

func (p *ClusterService) QueryById(objectId string, x_auth_token string) (cluster entity.Cluster,
	errorCode string, err error) {
	if !bson.IsObjectIdHex(objectId) {
		err = errors.New("invalide ObjectId.")
		errorCode = COMMON_ERROR_INVALIDATE
		return
	}

	// do authorize first
	if authorized := GetAuthService().Authorize("get_cluster", x_auth_token, objectId, p.collectionName); !authorized {
		err = errors.New("required opertion is not authorized!")
		errorCode = COMMON_ERROR_UNAUTHORIZED
		logrus.Errorf("get cluster with objectId [%v] error is %v", objectId, err)
		return
	}

	var selector = bson.M{}
	selector["_id"] = bson.ObjectIdHex(objectId)
	cluster = entity.Cluster{}
	err = dao.HandleQueryOne(&cluster, dao.QueryStruct{p.collectionName, selector, 0, 0, ""})
	if err != nil {
		logrus.Errorf("query cluster [objectId=%v] error is %v", objectId, err)
		errorCode = CLUSTER_ERROR_QUERY
	}
	return

}

func (p *ClusterService) DeleteByQuery(userId string, token string) (errorCode string, err error) {
	logrus.Infof("start to delete Cluster with userid [%v]", userId)

	authQuery, err := GetAuthService().BuildQueryByAuth("delete_clusters", token)
	if err != nil {
		logrus.Errorf("get auth query by token [%v] error is %v", token, err)
		errorCode = COMMON_ERROR_INTERNAL
		return errorCode, err
	}

	andQuery := bson.M{}
	if len(userId) > 0 {
		andQuery["user_id"] = userId
	}

	selector := generateQueryWithAuth(andQuery, authQuery)

	clusters := []entity.Cluster{}
	_, err = dao.HandleQueryAll(&clusters, dao.QueryStruct{p.collectionName, selector, 0, 0, ""})
	if err != nil {
		logrus.Errorf("get all cluster by userId error %v", err)
		errorCode = CLUSTER_ERROR_UPDATE
		return
	}

	for i := 0; i < len(clusters); i++ {
		cluster := clusters[i]
		_, err := p.DeleteById(cluster.ObjectId.Hex(), token)
		if err != nil {
			logrus.Errorf("delete cluster by id error %v", err)
			continue
		}
	}

	return
}

func (p *ClusterService) DeleteById(objectId string, x_auth_token string) (errorCode string, err error) {
	logrus.Infof("start to delete Cluster with objectId [%v]", objectId)
	// do authorize first
	if authorized := GetAuthService().Authorize("delete_cluster", x_auth_token, objectId, p.collectionName); !authorized {
		err = errors.New("required opertion is not authorized!")
		errorCode = COMMON_ERROR_UNAUTHORIZED
		logrus.Errorf("delete cluster with objectId [%v] error is %v", objectId, err)
		return
	}
	if !bson.IsObjectIdHex(objectId) {
		err = errors.New("invalide ObjectId.")
		errorCode = COMMON_ERROR_INVALIDATE
		return
	}

	//fix : limit 1? 0?
	_, hosts, _, err := GetHostService().QueryAllByName(objectId, 0, 0, x_auth_token)
	for _, v := range hosts {
		_, err = GetHostService().DeleteById(v.ObjectId.Hex(), x_auth_token)
		if err != nil {
			logrus.Errorf("delete host [objectId=%v] error is %v", objectId, err)
			errorCode = HOST_ERROR_DELETE
		}
	}
	
	_, _, err = p.UpdateStateById(objectId, CLUSTER_STATUS_TERMINATED, x_auth_token)
	if err != nil {
		logrus.Errorf("delete cluster [objectId=%v] error is %v", objectId, err)
		errorCode = CLUSTER_ERROR_DELETE
	}
	return
}

func (p *ClusterService) UpdateById(objectId string, instances int, cluster entity.Cluster, x_auth_token string) (created bool,
	errorCode string, err error) {
	logrus.Infof("start to update cluster")
	// do authorize first
	if authorized := GetAuthService().Authorize("update_cluster", x_auth_token, objectId, p.collectionName); !authorized {
		err = errors.New("required opertion is not authorized!")
		errorCode = COMMON_ERROR_UNAUTHORIZED
		logrus.Errorf("update cluster with objectId [%v] error is %v", objectId, err)
		return
	}
	if !bson.IsObjectIdHex(objectId) {
		err = errors.New("invalide ObjectId.")
		errorCode = COMMON_ERROR_INVALIDATE
		return
	}
	// change cluster status to deploying
	
	
	cluster, _, _ = p.QueryById(objectId, x_auth_token)
	total := cluster.Instances
	logrus.Infof("total is ", total)
	_, _, err = p.UpdateStateById(cluster.ObjectId.Hex(), CLUSTER_STATUS_DEPLOYING, x_auth_token)
	if err != nil {
		logrus.Errorf("delete cluster [objectId=%v] error is %v", objectId, err)
		errorCode = CLUSTER_ERROR_DELETE
	}
	
	
	var selector = bson.M{}
	selector["_id"] = bson.ObjectIdHex(objectId)

	// reset objectId and updated_time
	cluster.ObjectId = bson.ObjectIdHex(objectId)
	cluster.Instances = instances
	
	logrus.Infof("start to change cluster")
	if instances < total {
		logrus.Infof("start to delete hosts")
		number := total - instances
		_, hosts, _, errQuery := GetHostService().QueryAllUnterminated(objectId, 0, 0, x_auth_token)
		length := len(hosts)
		if errQuery != nil {
			err = errors.New("query hosts is err")
			logrus.Errorf("query host is err")
			return
		}
		for i := 0; i < number; i++ {
			_, err = GetHostService().DeleteById(hosts[length - i - 1].ObjectId.Hex(), x_auth_token)
			if err != nil {
				logrus.Errorf("delete hosts err is %v", err)
				return
			}
		}
		//after delete host then change cluster status and instance
		cluster.TimeUpdate = dao.GetCurrentTime()
		cluster.Status = CLUSTER_STATUS_DEPLOYED
		err = dao.HandleUpdateByQueryPartial(p.collectionName, selector, &cluster)
		if err != nil {
			logrus.Errorf("update cluster [%v] error is %v", cluster, err)
			errorCode = CLUSTER_ERROR_UPDATE
			return
		}
		
		created = true
	} else if instances > total {
		logrus.Infof("start to creat new hosts")
		number := instances - total
		newHosts := make([]entity.Host, number)
		
		_, hosts, _, errQuery := GetHostService().QueryAllUnterminated(objectId, 0, 0, x_auth_token)
			if errQuery != nil {
				err = errors.New("query hosts is err")
				logrus.Errorf("query host is err")
				return
			}
		length := len(hosts)
		lable := hosts[length - 1].Lable
		
		for i, hostss := range  newHosts{
			go p.createHost(i + lable + 1, hostss, cluster, x_auth_token)
		}
		
		cluster.Status = CLUSTER_STATUS_DEPLOYED
		err = dao.HandleUpdateByQueryPartial(p.collectionName, selector, &cluster)
		if err != nil {
			logrus.Errorf("update cluster [%v] error is %v", cluster, err)
			errorCode = CLUSTER_ERROR_UPDATE
			return
		}
			
	}
	return
}



func (p *ClusterService) ChangeClusterInstancesAccHost(host entity.Host, x_auth_token string) (errorCode string, err error) {
	logrus.Infof("start to change cluster instances")
	cluster, _, err := p.QueryById(host.ClusterId, x_auth_token)
	instances := cluster.Instances
	logrus.Infof("cluster instances is %v", instances)
	var selector = bson.M{}
	selector["_id"] = cluster.ObjectId
	cluster.TimeUpdate = dao.GetCurrentTime()
	cluster.Instances = instances - 1
	if cluster.Instances == 0 {
		cluster.Status = CLUSTER_STATUS_TERMINATED
	}
	err = dao.HandleUpdateByQueryPartial(p.collectionName, selector, &cluster)
	if err != nil {
		logrus.Errorf("update cluster instances [%v] error is %v", cluster, err)
		errorCode = CLUSTER_ERROR_UPDATE
		return
	}
	return
}

func (p *ClusterService) getMasterNodes(cluster_id string) (masters []entity.Host, err error) {
	if !bson.IsObjectIdHex(cluster_id) {
		err = errors.New("invalidate cluster_id:" + cluster_id)
		logrus.Errorf("invalidate cluster_id [%v]", cluster_id)
		return
	}
	query := bson.M{}
	query["cluster_id"] = cluster_id
	query["ismasternode"] = true
	_, masters, _, err = GetHostService().queryByQuery(query, 0, 0, "", true)
	if err != nil {
		logrus.Errorf("query master nodes by query [%v] failed, error is %v", query, err)
	}
	return
}

func (p *ClusterService) QueryAllUnterminated(cluster_name string, skip int, limit int, x_auth_token string) (total int,
	clusters []entity.Cluster, errorCode string, err error) {
	query := bson.M{}
	if cluster_name == "" {
		query["status"] = bson.M{"$ne": "TERMINATED"}
		return p.queryByQuery(query, skip, limit, x_auth_token, false)
	} else {
		query["cluster_name"] = cluster_name
		query["status"] = bson.M{"$ne": "TERMINATED"}
		return p.queryByQuery(query, skip, limit, x_auth_token, false)
	}
	return

}
