package services

import (
	"errors"
	"sync"
	"time"
	"strings"
	"gopkg.in/mgo.v2/bson"
	"github.com/Sirupsen/logrus"
	"linkernetworks.com/linker_common_lib/persistence/entity"
	"linkernetworks.com/linker_common_lib/persistence/dao"
)

var (
	hostlogService            *HostLogService = nil
	onceHostLog              sync.Once
	HOSTLOG_STATUS_CREATED    = "CREATED"
	HOSTLOG_STATUS_FAILED     = "FAILED"

	HOSTLOG_ERROR_CREATE string = ""
	HOSTLOG_ERROR_QUERY  string = ""
)

type HostLogService struct {
	collectionName string
}

func GetHostLogService() *HostLogService {
	onceHostLog.Do(func() {
		logrus.Debugf("Once called from clusterlogService ......................................")
		hostlogService = &HostLogService{"hostlog"}
	})
	return hostlogService

}

func (p *HostLogService) Create(hostlog entity.HostLog, x_auth_token string) (newHostLog entity.HostLog,
	errorCode string, err error) {
	logrus.Infof("start to create hostlog [%v]", hostlog)	
	
	// do authorize first
	if authorized := GetAuthService().Authorize("create_hostlog", x_auth_token, "", p.collectionName); !authorized {
		err = errors.New("required opertion is not authorized!")
		errorCode = COMMON_ERROR_UNAUTHORIZED
		logrus.Errorf("create hostlog [%v] error is %v", hostlog, err)
		return
	}
	
	token, err := GetTokenById(x_auth_token)
	if err != nil {
		errorCode = HOSTLOG_ERROR_CREATE
		logrus.Errorf("get token failed when create hostlog [%v], error is %v", hostlog, err)
		return
	}
	
	// set token_id and user_id from token
	hostlog.Tenant_id = token.Tenant.Id
	hostlog.User_id = token.User.Id

	// set created_time and updated_time
	hostlog.TimeCreate = dao.GetCurrentTime()
	hostlog.TimeUpdate = hostlog.TimeCreate
	hostlog.Date = int(time.Now().Unix())
	
	// insert bson to mongodb
	err = dao.HandleInsert(p.collectionName, hostlog)
	if err != nil {
		errorCode = HOSTLOG_ERROR_CREATE
		logrus.Errorf("insert hostlog [%v] to db error is %v", hostlog, err)
		return
	}

	newHostLog = hostlog

	return
	
}

func (p *HostLogService) QueryAllById(cluster_id string, skip int,
	limit int, x_auth_token string) (total int, clusterlogs []entity.HostLog,
	errorCode string, err error) {
	// if cluster_name is empty, query all
	query := bson.M{}
	if strings.TrimSpace(cluster_id) == "" {
		query["host_status"] = bson.M{"$ne": "TERMINATED"}
		return p.queryByQuery(query, skip, limit, x_auth_token, false)
	}else {
		query["cluster_id"] = cluster_id
		query["host_status"] = bson.M{"$ne": "TERMINATED"}
		return p.queryByQuery(query, skip, limit, x_auth_token, false)
	}
	return	
		
}

func (p *HostLogService) queryByQuery(query bson.M, skip int, limit int,
	x_auth_token string, skipAuth bool) (total int, hostlogs []entity.HostLog,
	errorCode string, err error) {
	authQuery := bson.M{}
	if !skipAuth {
		// get auth query from auth service first
		authQuery, err = GetAuthService().BuildQueryByAuth("list_hostlog", x_auth_token)
		if err != nil {
			logrus.Errorf("get auth query by token [%v] error is %v", x_auth_token, err)
			errorCode = COMMON_ERROR_INTERNAL
			return
		}
	}

	selector := generateQueryWithAuth(query, authQuery)
	hostlogs = []entity.HostLog{}
	queryStruct := dao.QueryStruct{p.collectionName, selector, skip, limit, ""}
	total, err = dao.HandleQueryAll(&hostlogs, queryStruct)
	if err != nil {
		logrus.Errorf("query hostlogs by query [%v] error is %v", query, err)
		errorCode = HOSTLOG_ERROR_QUERY
	}
	return
}


