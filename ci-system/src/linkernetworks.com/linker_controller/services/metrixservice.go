package services

import (
	"errors"
	"github.com/Sirupsen/logrus"
	"gopkg.in/mgo.v2/bson"
	"linkernetworks.com/linker_common_lib/persistence/entity"
	"strings"
	"sync"
)

var (
	metrixService     *MetrixService = nil
	onceMetrix        sync.Once
	PROVIDER_ALICLOUD string = "AliCloud"
)

type MetrixService struct {
}

func GetMetrixService() *MetrixService {
	onceMetrix.Do(func() {
		logrus.Debugf("Once called from metrixService ......................................")
		metrixService = &MetrixService{}
	})
	return metrixService
}

func (p *MetrixService) GetMetrix(category, x_auth_token string) (isSg bool, sgmetrix entity.SGMetrix,
	prvmetrix entity.ProviderMetrix, errorCode string, err error) {
	// get auth query from auth service first
	authQuery, err := GetAuthService().BuildQueryByAuth("list_metrix", x_auth_token)
	if err != nil {
		logrus.Errorf("get auth query by token [%v] error is %v", x_auth_token, err)
		errorCode = COMMON_ERROR_INTERNAL
		return
	}

	token, err := GetTokenById(x_auth_token)
	if err != nil {
		errorCode = COMMON_ERROR_INTERNAL
		logrus.Errorf("get token failed when get metrix by token [%v], error is %v", x_auth_token, err)
		return
	}

	logrus.Debugf("auth query is %v", authQuery)
	selector := generateQueryWithAuth(bson.M{}, authQuery)

	switch strings.ToLower(category) {
	case "serviceinstances":
		isSg = true
		sgmetrix, err = p.getServiceInstanceMetrix(selector)
		if err != nil {
			errorCode = COMMON_ERROR_INTERNAL
		}
		sgmetrix.TenantId = token.Tenant.Id
		sgmetrix.UserId = token.User.Id
	case "resources":
		isSg = false
		prvmetrix, err = p.getProviderMetrix(selector)
		if err != nil {
			errorCode = COMMON_ERROR_INTERNAL
		}
		prvmetrix.TenantId = token.Tenant.Id
		prvmetrix.UserId = token.User.Id
	default:
		err = errors.New("unsupported metrix category:" + category)
		logrus.Errorf("get metrix failed, error is %v", err)
		errorCode = COMMON_ERROR_INVALIDATE
		return
	}
	return
}

func (p *MetrixService) getProviderMetrix(selector bson.M) (metrix entity.ProviderMetrix, err error) {
	// get auth query from auth service first
	statusList := []string{SGI_STATUS_DEPLOYED, SGI_STATUS_MODIFYING,
		SGI_STATUS_DEPLOYING}
	inMap := make(map[string]interface{})
	inMap["$in"] = statusList
	oriQuery := bson.M{}
	oriQuery["life_cycle_status"] = inMap
	selector = generateQueryWithAuth(oriQuery, selector)
	_, sgis, _, err := GetSgiService().queryAllByquery(0, 0, selector)
	if err != nil {
		return
	}
	provider := entity.Provider{}
	provider.Status = make(map[string]entity.Used)
	total := entity.Used{}
	for i := 0; i < len(sgis); i++ {
		sgi := sgis[i]
		aciStatusList := []string{ACI_STATUS_CREATED, ACI_STATUS_CONFIGED}
		aciinMap := make(map[string]interface{})
		aciinMap["$in"] = aciStatusList
		aciSelector := bson.M{}
		aciSelector["service_group_instance_id"] = sgi.ObjectId.Hex()
		aciSelector["lifecycle_status"] = aciinMap
		aciCount, acis, _, err := GetAciService().queryAllByQuery(0, 0, aciSelector)
		if err != nil {
			logrus.Errorf("find acis by sgiid [%v] failed, error is %v",
				sgi.ObjectId.Hex(), err)
			continue
		}

		total.ContainerNum = total.ContainerNum + aciCount
		for j := 0; j < len(acis); j++ {
			aci := acis[j]
			status := aci.LifeCycleStatus
			used := provider.Status[status]
			used.ContainerNum = used.ContainerNum + 1
			used.Cpus = used.Cpus + aci.Cpus
			used.Mems = used.Mems + int64(aci.Mem)
			provider.Status[status] = used
			total.Cpus = total.Cpus + aci.Cpus
			total.Mems = total.Mems + int64(aci.Mem)
		}

	}
	provider.Total = total
	metrix.Providers = make(map[string]entity.Provider)
	metrix.Providers[PROVIDER_ALICLOUD] = provider
	return
}

func (p *MetrixService) getServiceInstanceMetrix(selector bson.M) (metrix entity.SGMetrix, err error) {
	total, sgis, _, err := GetSgiService().queryAllByquery(0, 0, selector)
	if err != nil {
		return
	}
	metrix.TotalNum = int64(total)
	metrix.StatusNum = make(map[string]int64)
	for i := 0; i < len(sgis); i++ {
		sgi := sgis[i]
		metrix.StatusNum[sgi.LifeCycleStatus] = metrix.StatusNum[sgi.LifeCycleStatus] + 1
	}
	return
}
