package services

import (
	"errors"
	"github.com/Sirupsen/logrus"
	"gopkg.in/mgo.v2/bson"
	"linkernetworks.com/linker_common_lib/persistence/dao"
	"linkernetworks.com/linker_common_lib/persistence/entity"
	"sync"
)

var (
	ipResourceService        *IPResourceService = nil
	onceIPResource           sync.Once
	IP_RESOURCE_ERROR_CREATE string = "E11071"
	IP_RESOURCE_ERROR_UPDATE string = "E11073"
	IP_RESOURCE_ERROR_DELETE string = "E11072"
	IP_RESOURCE_ERROR_QUERY  string = "E11074"
)

type IPResourceService struct {
	collectionName string
}

func GetIPResourceService() *IPResourceService {
	onceIPResource.Do(func() {
		logrus.Debugf("Once called from ip resource service ......................................")
		ipResourceService = &IPResourceService{"ipAddressResource"}
	})
	return ipResourceService
}

func (p *IPResourceService) releaseIp(owner string) (err error) {
	// findout owner's ip
	selector := bson.M{}
	selector["allocated"] = owner
	queryStruct := dao.QueryStruct{p.collectionName, selector, 0, 0, ""}
	var ipAddressResources = []entity.IpAddressResource{}
	total, err := dao.HandleQueryAll(&ipAddressResources, queryStruct)
	if err != nil {
		return
	}
	logrus.Debugf("founded [%v] ips belongs to %v", total, owner)
	if total > 0 {
		// reset ip's allocated to false
		for _, ipAddressResource := range ipAddressResources {
			logrus.Debugf("releaseIp, ip=%s", ipAddressResource.IpAddress)
			ipAddressResource.Allocated = "false"
			_, _, err := p.updateById(ipAddressResource.ObjectId.Hex(),
				ipAddressResource)
			if err != nil {
				logrus.Errorf("release ipaddr [%v] from [%v] failed, err is: %v",
					ipAddressResource.IpAddress, owner, err)
			}
		}
	}
	return
}

func (p *IPResourceService) resetIp(oldowner, newowner string) (err error) {
	// findout oldowner's ip
	selector := bson.M{}
	selector["allocated"] = oldowner
	queryStruct := dao.QueryStruct{p.collectionName, selector, 0, 0, ""}
	var ipAddressResources = []entity.IpAddressResource{}
	total, err := dao.HandleQueryAll(&ipAddressResources, queryStruct)
	if err != nil {
		return
	}
	logrus.Debugf("founded [%v] ips belongs to %v", total, oldowner)
	if total > 0 {
		// reset ip's allocated to newowner
		for _, ipAddressResource := range ipAddressResources {
			logrus.Debugf("resetIp, ip=%s", ipAddressResource.IpAddress)
			ipAddressResource.Allocated = newowner
			_, _, err := p.updateById(ipAddressResource.ObjectId.Hex(),
				ipAddressResource)
			if err != nil {
				logrus.Errorf("reset ipaddr [%v] from [%v] to [%v] failed, err is: %v",
					ipAddressResource.IpAddress, oldowner, newowner, err)
			}
		}
	}
	return
}

func (p *IPResourceService) QueryAll(skip int, limit int, x_auth_token string) (total int, ipAddressResources []entity.IpAddressResource,
	errorCode string, err error) {
	// get auth query from auth service first
	authQuery, err := GetAuthService().BuildQueryByAuth("list_ips", x_auth_token)
	if err != nil {
		logrus.Errorf("get auth query by token [%v] error is %v", x_auth_token, err)
		errorCode = COMMON_ERROR_INTERNAL
		return
	}

	logrus.Debugf("auth query is %v", authQuery)

	selector := generateQueryWithAuth(bson.M{}, authQuery)
	logrus.Debugf("selector is %v", selector)
	sort := ""
	ipAddressResources = []entity.IpAddressResource{}
	queryStruct := dao.QueryStruct{p.collectionName, selector, skip, limit, sort}
	total, err = dao.HandleQueryAll(&ipAddressResources, queryStruct)
	if err != nil {
		logrus.Errorf("list ips [token=%v] failed, error is %v", x_auth_token, err)
		errorCode = IP_RESOURCE_ERROR_QUERY
	}
	return
}

func (p *IPResourceService) DeleteById(objectId string, x_auth_token string) (errorCode string, err error) {
	logrus.Infof("start to delete ip with objectId [%v]", objectId)
	// do authorize first
	if authorized := GetAuthService().Authorize("delete_ip", x_auth_token, objectId, p.collectionName); !authorized {
		err = errors.New("required opertion is not authorized!")
		errorCode = COMMON_ERROR_UNAUTHORIZED
		logrus.Errorf("delete ip with objectId [%v] error is %v", objectId, err)
		return
	}

	if !bson.IsObjectIdHex(objectId) {
		err = errors.New("invalide ObjectId.")
		errorCode = COMMON_ERROR_INVALIDATE
		return
	}

	var selector = bson.M{}
	selector["_id"] = bson.ObjectIdHex(objectId)

	err = dao.HandleDelete(p.collectionName, true, selector)
	if err != nil {
		logrus.Errorf("delete ip [objectId=%v] error is %v", objectId, err)
		errorCode = IP_RESOURCE_ERROR_DELETE
	}
	return
}

func (p *IPResourceService) QueryById(objectId string, x_auth_token string) (app entity.App,
	errorCode string, err error) {
	if !bson.IsObjectIdHex(objectId) {
		err = errors.New("invalide ObjectId.")
		errorCode = COMMON_ERROR_INVALIDATE
		return
	}
	// do authorize first
	if authorized := GetAuthService().Authorize("get_ip", x_auth_token, objectId, p.collectionName); !authorized {
		err = errors.New("required opertion is not authorized!")
		errorCode = COMMON_ERROR_UNAUTHORIZED
		logrus.Errorf("get ip with objectId [%v] error is %v", objectId, err)
		return
	}

	var selector = bson.M{}
	selector["_id"] = bson.ObjectIdHex(objectId)
	var ipAddressResource = entity.IpAddressResource{}
	err = dao.HandleQueryOne(&ipAddressResource, dao.QueryStruct{p.collectionName, selector, 0, 0, ""})
	if err != nil {
		logrus.Errorf("query ip [objectId=%v] error is %v", objectId, err)
		errorCode = IP_RESOURCE_ERROR_QUERY
	}
	return
}

func (p *IPResourceService) UpdateById(objectId string, ipAddressResource entity.IpAddressResource, x_auth_token string) (created bool,
	errorCode string, err error) {
	logrus.Infof("start to update ipAddressResource [%v]", ipAddressResource)
	// do authorize first
	if authorized := GetAuthService().Authorize("update_ip", x_auth_token, objectId, p.collectionName); !authorized {
		err = errors.New("required opertion is not authorized!")
		errorCode = COMMON_ERROR_UNAUTHORIZED
		logrus.Errorf("update ip with objectId [%v] error is %v", objectId, err)
		return
	}
	return p.updateById(objectId, ipAddressResource)
}

func (p *IPResourceService) Create(ipAddressResource entity.IpAddressResource, x_auth_token string) (newIPAddressResource entity.IpAddressResource,
	errorCode string, err error) {
	logrus.Infof("start to create ip [%v]", ipAddressResource)
	// do authorize first
	if authorized := GetAuthService().Authorize("create_ip", x_auth_token, "", p.collectionName); !authorized {
		err = errors.New("required opertion is not authorized!")
		errorCode = COMMON_ERROR_UNAUTHORIZED
		logrus.Errorf("create ip [%v] error is %v", ipAddressResource, err)
		return
	}

	token, err := GetTokenById(x_auth_token)
	if err != nil {
		errorCode = IP_RESOURCE_ERROR_CREATE
		logrus.Errorf("get token failed when create ip [%v], error is %v", ipAddressResource, err)
		return
	}

	// set token_id and user_id from token
	ipAddressResource.Tenant_id = token.Tenant.Id
	ipAddressResource.User_id = token.User.Id

	return p.createIp(ipAddressResource)
}

func (p *IPResourceService) createIp(ipAddressResource entity.IpAddressResource) (newIPAddressResource entity.IpAddressResource,
	errorCode string, err error) {
	ipAddressResource.ObjectId = bson.NewObjectId()
	// set created_time and updated_time
	ipAddressResource.TimeCreate = dao.GetCurrentTime()
	ipAddressResource.TimeUpdate = ipAddressResource.TimeCreate

	// insert bson to mongodb
	err = dao.HandleInsert(p.collectionName, ipAddressResource)
	if err != nil {
		errorCode = IP_RESOURCE_ERROR_CREATE
		logrus.Errorf("create ip [%v] to bson error is %v", ipAddressResource, err)
		return
	}

	newIPAddressResource = ipAddressResource
	return
}

func (p *IPResourceService) CreatePool(ipAddressPool *entity.IpAddressPool, x_auth_token string) (newipAddressPool *entity.IpAddressPool,
	errorCode string, err error) {
	// do authorize first
	if authorized := GetAuthService().Authorize("create_ippool", x_auth_token, "", p.collectionName); !authorized {
		err = errors.New("required opertion is not authorized!")
		errorCode = COMMON_ERROR_UNAUTHORIZED
		logrus.Errorf("create ip pool error is %v", err)
		return
	}

	token, err := GetTokenById(x_auth_token)
	if err != nil {
		errorCode = IP_RESOURCE_ERROR_CREATE
		logrus.Errorf("get token failed when create ip pool, error is %v", err)
		return
	}

	for _, ip := range ipAddressPool.IpResources {
		ip.ObjectId = bson.NewObjectId()
		ip.Gateway = ipAddressPool.Gateway
		ip.Subnet = ipAddressPool.Subnet
		ip.PoolName = ipAddressPool.PoolName
		ip.User_id = token.User.Id
		ip.Tenant_id = token.Tenant.Id

		_, code, err := p.createIp(ip)
		if err != nil {
			logrus.Errorf("create ip err is %v", err)
			errorCode = code
			return nil, errorCode, err
		}
	}

	newipAddressPool = ipAddressPool
	return
}

func (p *IPResourceService) updateById(objectId string, ipAddressResource entity.IpAddressResource) (created bool,
	errorCode string, err error) {
	// validate id
	if !bson.IsObjectIdHex(objectId) {
		err = errors.New("invalide ObjectId.")
		errorCode = COMMON_ERROR_INVALIDATE
		return
	}
	var selector = bson.M{}
	selector["_id"] = bson.ObjectIdHex(objectId)

	// reset objectId and updated_time
	ipAddressResource.ObjectId = bson.ObjectIdHex(objectId)
	ipAddressResource.TimeUpdate = dao.GetCurrentTime()

	// insert bson to mongodb
	created, err = dao.HandleUpdateOne(&ipAddressResource, dao.QueryStruct{p.collectionName, selector, 0, 0, ""})
	if err != nil {
		logrus.Errorf("update ipAddressResource [%v] error is %v", ipAddressResource, err)
		errorCode = IP_RESOURCE_ERROR_UPDATE
	}
	return
}
