package services

import (
	"errors"
	"sync"

	"github.com/Sirupsen/logrus"
	"gopkg.in/mgo.v2/bson"
	"linkernetworks.com/linker_cluster/common"
	"linkernetworks.com/linker_common_lib/persistence/dao"
	"linkernetworks.com/linker_common_lib/persistence/entity"
)

var (
	clusterorderService *ClusterOrderService = nil
	onceClusterOrder    sync.Once

	SGO_ERROR_CREATE string = "E11061"
	SGO_ERROR_DELETE string = "E11062"
	SGO_ERROR_GET    string = "E11065"
	SGO_ERROR_SCALE  string = "E11063"

	SGI_ERROR_QUERY string = "E11054"
)

type ClusterOrderService struct {
	collectionName string
}

func GetClusterOrderService() *ClusterOrderService {
	onceClusterOrder.Do(func() {
		logrus.Debugf("Once called from clusterorderService ......................................")
		clusterorderService = &ClusterOrderService{"cluster_order"}
	})
	return clusterorderService

}

func (p *ClusterOrderService) CreateOrder(token string, sgo entity.ServiceGroupOrder) (newsgo entity.ServiceGroupOrder, errorCode string, err error) {
	logrus.Infoln("create order")

	tokenObj, err := GetTokenById(token)
	if err != nil {
		logrus.Errorf("get token by id error %v", err)
		return newsgo, COMMON_ERROR_INTERNAL, err
	}

	controllerURL := ""
	clusterId := sgo.ClusterId
	if len(clusterId) <= 0 {
		logrus.Infoln("cluster id is null, create order will be sent to primary controller")
		controllerURL, err = common.UTIL.ZkClient.GetControllerEndpoint()
		if err != nil {
			logrus.Errorf("get primary controller url error %v", err)
			errorCode = COMMON_ERROR_INTERNAL
			return
		}
	} else {
		_, hosts, errorCode, err := GetHostService().QueryAllByName(clusterId, 0, 0, token)
		if err != nil {
			logrus.Errorf("get hosts by cluster id error %v", err)
			return newsgo, errorCode, err
		}

		for i := 0; i < len(hosts); i++ {
			host := hosts[i]
			if host.IsMasterNode {
				controllerURL = host.IP + ":8081"

				sgo.ClusterName = host.ClusterName
				break
			}
		}

		logrus.Infoln("cluster controller url is :", controllerURL)

		err = copyBundle(token, sgo.ServiceGroupId, controllerURL)
		if err != nil {
			logrus.Warnf("copy sg to cluster error %v", err)
			return newsgo, SGO_ERROR_CREATE, err
		}

	}

	if len(controllerURL) <= 0 {
		logrus.Warnf("controller url is null, will not continue ordering")
		return newsgo, SGO_ERROR_CREATE, errors.New("can not get controller url!")
	}

	newsgo, err = OrderSGO(token, sgo, controllerURL)
	if err != nil {
		logrus.Errorf("create sg order failed %v", err)
		return newsgo, SGO_ERROR_CREATE, err
	}

	currentTime := dao.GetCurrentTime()
	objectId := bson.NewObjectId()
	clusterOrder := entity.ClusterOrder{
		ObjectId:    objectId,
		ClusterId:   sgo.ClusterId,
		ClusterName: sgo.ClusterName,
		SgoId:       newsgo.ObjectId.Hex(),
		SgId:        sgo.ServiceGroupId,
		SgiId:       newsgo.ServiceGroupInstanceId,
		Tenant_id:   tokenObj.Tenant.Id,
		User_id:     tokenObj.User.Id,
		TimeCreate:  currentTime,
		TimeUpdate:  currentTime}

	err = dao.HandleInsert(p.collectionName, clusterOrder)
	if err != nil {
		logrus.Errorf("order in cluster error %v", err)
		return newsgo, SGO_ERROR_CREATE, err
	}

	return

}

func (p *ClusterOrderService) TerminateOrder(token string, clusterId string, sgoId string) (errorCode string, err error) {
	logrus.Infoln("terminate order")

	controllerURL, errorCode, err := getControllerURL(token, clusterId)
	if err != nil {
		logrus.Errorf("get controllerURL error %v", err)
		return errorCode, err
	}
	logrus.Infof("controllerURL is: %v", controllerURL)

	if len(controllerURL) <= 0 {
		logrus.Warnf("controller url is null, will not continue terminating")
		return SGO_ERROR_DELETE, errors.New("can not get controller!")
	}

	err = TerminateSGO(token, sgoId, controllerURL)
	if err != nil {
		logrus.Errorf("terminate sg order failed %v", err)
		errorCode = SGO_ERROR_DELETE
		return
	}

	err = p.DeleteBySgoId(sgoId)
	if err != nil {
		logrus.Errorf("remove cluster order failed %v", err)
		return SGO_ERROR_DELETE, err
	}

	return
}

func (p *ClusterOrderService) DeleteBySgoId(sgoId string) (err error) {
	logrus.Infof("delete cluster order by sgo id %v", sgoId)

	selector := bson.M{}
	selector["sgo_id"] = sgoId
	err = dao.HandleDelete(p.collectionName, true, selector)
	if err != nil {
		logrus.Errorf("delete cluster order error %v", err)
		return
	}

	return
}

func (p *ClusterOrderService) QuerySGOById(token string, clusterId string, sgoId string) (sgo entity.ServiceGroupOrder, errorCode string, err error) {
	controllerURL, errorCode, err := getControllerURL(token, clusterId)
	if err != nil {
		logrus.Errorf("get controllerURL error %v", err)
		return sgo, errorCode, err
	}
	logrus.Infof("controllerURL is: %v", controllerURL)

	if len(controllerURL) <= 0 {
		logrus.Warnf("controller url is null, can not get sgo by id")
		return sgo, SGO_ERROR_GET, errors.New("can not get controller!")
	}

	sgo, err = QuerySGO(token, sgoId, controllerURL)
	if err != nil {
		logrus.Errorf("Query sg order failed %v", err)
		errorCode = SGO_ERROR_GET
		return
	}

	return
}

func (p *ClusterOrderService) GetAuthOperations(token string, clusterId string, sgoId string) (operations map[string]int, errorCode string, err error) {
	controllerURL, errorCode, err := getControllerURL(token, clusterId)
	if err != nil {
		logrus.Errorf("get controllerURL error %v", err)
		return operations, errorCode, err
	}
	logrus.Infof("controllerURL is: %v", controllerURL)

	if len(controllerURL) <= 0 {
		logrus.Warnf("controller url is null")
		return operations, SGO_ERROR_GET, errors.New("can not get controller!")
	}

	operations, err = QueryAuthOrder(token, sgoId, controllerURL)
	if err != nil {
		logrus.Errorf("Query sg order failed %v", err)
		errorCode = SGO_ERROR_GET
		return
	}

	return
}

func (p *ClusterOrderService) GetAppInOrder(token string, clusterId string, sgoId string, appId string) (app *entity.App, errorCode string, err error) {
	controllerURL, errorCode, err := getControllerURL(token, clusterId)
	if err != nil {
		logrus.Errorf("get controllerURL error %v", err)
		return app, errorCode, err
	}
	logrus.Infof("controllerURL is: %v", controllerURL)

	if len(controllerURL) <= 0 {
		logrus.Warnf("controller url is null")
		return app, SGO_ERROR_GET, errors.New("can not get controllerURL!")
	}

	app, err = QueryAppInOrder(token, sgoId, appId, controllerURL)
	if err != nil {
		logrus.Errorf("Query sg order failed %v", err)
		errorCode = SGO_ERROR_GET
		return
	}

	return
}

func (p *ClusterOrderService) ScaleAppByOrderId(token string, clusterId string, sgoId string, appId string, numStr string) (errorCode string, err error) {
	controllerURL, errorCode, err := getControllerURL(token, clusterId)
	if err != nil {
		logrus.Errorf("get controllerURL error %v", err)
		return errorCode, err
	}
	logrus.Infof("controllerURL is: %v", controllerURL)

	if len(controllerURL) <= 0 {
		logrus.Warnf("controller url is null")
		return SGO_ERROR_SCALE, errors.New("can not get controller!")
	}

	err = ScaleAppByOrderId(token, sgoId, appId, numStr, controllerURL)
	if err != nil {
		logrus.Errorf("Query sg order failed %v", err)
		errorCode = SGO_ERROR_SCALE
		return
	}

	return
}

func (p *ClusterOrderService) GetOrderInstance(token string, clusterId string, sgiId string) (sgi entity.ServiceGroupInstance, errorCode string, err error) {
	controllerURL, errorCode, err := getControllerURL(token, clusterId)
	if err != nil {
		logrus.Errorf("get controllerURL error %v", err)
		return sgi, errorCode, err
	}
	logrus.Infof("controllerURL is: %v", controllerURL)

	if len(controllerURL) <= 0 {
		logrus.Warnf("controller url is null")
		return sgi, SGI_ERROR_QUERY, errors.New("can not get controller!")
	}

	sgi, err = GetSGI(token, sgiId, controllerURL)
	if err != nil {
		logrus.Errorf("Query sg instance failed %v", err)
		errorCode = SGI_ERROR_QUERY
		return
	}

	return
}

func (p *ClusterOrderService) QueryAll(x_auth_token string, skip int, limit int) (total int,
	clusters []entity.ClusterOrder, errorCode string, err error) {
	return p.queryByQuery(bson.M{}, skip, limit, x_auth_token, false)
}

func getControllerURL(tokenId string, clusterId string) (controllerURL string, errorCode string, err error) {
	if len(clusterId) <= 0 {
		logrus.Infof("get sgo from primary controller!")
		controllerURL, err = common.UTIL.ZkClient.GetControllerEndpoint()
		if err != nil {
			logrus.Errorf("get primary controller url error %v", err)
			return "", COMMON_ERROR_INTERNAL, err
		}
	} else {
		logrus.Infof("get sgo from cluster, clusterId:%v", clusterId)
		_, hosts, errorCode, err := GetHostService().QueryAllByName(clusterId, 0, 0, tokenId)
		if err != nil {
			logrus.Errorf("get hosts by cluster id error %v", err)
			return "", errorCode, err
		}

		for i := 0; i < len(hosts); i++ {
			host := hosts[i]
			if host.IsMasterNode {
				controllerURL = host.IP + ":8081"
				return controllerURL, "", nil
			}
		}
	}

	return
}

func (p *ClusterOrderService) queryByQuery(query bson.M, skip int, limit int,
	x_auth_token string, skipAuth bool) (total int, clusters []entity.ClusterOrder,
	errorCode string, err error) {
	authQuery := bson.M{}
	if !skipAuth {
		// get auth query from auth service first
		authQuery, err = GetAuthService().BuildQueryByAuth("list_clusterorder", x_auth_token)
		if err != nil {
			logrus.Errorf("get auth query by token [%v] error is %v", x_auth_token, err)
			errorCode = COMMON_ERROR_INTERNAL
			return
		}
	}

	selector := generateQueryWithAuth(query, authQuery)
	clusters = []entity.ClusterOrder{}
	queryStruct := dao.QueryStruct{p.collectionName, selector, skip, limit, ""}
	total, err = dao.HandleQueryAll(&clusters, queryStruct)
	if err != nil {
		logrus.Errorf("query clusterorders by query [%v] error is %v", query, err)
		errorCode = SGO_ERROR_GET
	}
	return
}

// func generateQueryWithAuth(oriQuery bson.M, authQuery bson.M) (query bson.M) {
// 	if len(authQuery) == 0 {
// 		query = oriQuery
// 	} else {
// 		query = bson.M{}
// 		query["$and"] = []bson.M{oriQuery, authQuery}
// 	}
// 	logrus.Debugf("generated query [%v] with auth [%v], result is [%v]", oriQuery, authQuery, query)
// 	return
// }
