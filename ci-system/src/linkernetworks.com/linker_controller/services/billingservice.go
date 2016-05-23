package services

import (
	"container/list"
	"errors"
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/compose/mejson"
	"gopkg.in/mgo.v2/bson"
	"linkernetworks.com/linker_common_lib/persistence/dao"
	"linkernetworks.com/linker_common_lib/persistence/entity"
	"linkernetworks.com/linker_controller/common"
)

var billingService *BillingService = nil
var oncebilling sync.Once

var f_rfc3339 = "2006-01-02T15:04:05Z07:00"
var secInDay = 86400

var CLUSTER_STATUS_FAILED = "FAILED"
var CLUSTER_STATUS_TERMINATED = "TERMINATED"

var HOST_STATUS_CREATED = "CREATED"
var HOST_STATUS_TERMINATED = "TERMINATED"
var HOST_STATUS_DEPLOYED = "DEPLOYED"
var HOST_STATUS_DEPLOYING = "DEPLOYING"
var HOST_STATUS_FAILED = "FAILED"

var BILLING_ERROR_GET = "E11034"
var BILLING_ERROR_CREATE = "E11031"
var BILLING_ERROR_UPDATE = "E11032"

var userMap = map[string]string{}

type BillingService struct {
	collectionName string
}

func GetBillingService() *BillingService {
	oncebilling.Do(func() {
		logrus.Debugf("Once called from billingService ......................................")
		billingService = &BillingService{"billing"}

		billingService.Initialize()
	})
	return billingService
}

func (p *BillingService) Initialize() {
	logrus.Infoln("initialize billing service...")

	enable_billing := common.UTIL.Props.GetBool("enable_billing", true)
	if !enable_billing {
		logrus.Infoln("skip billing process!")
		return
	}

	// waitTime := getWaitTimeForNextHour()
	waitTime := int64(30)

	//currently, the billing process will be invoked every day
	intervalTime := int64(86400)

	go p.startBillingTimer(waitTime, intervalTime)

}

func (p *BillingService) startBillingTimer(waitTime int64, intervalTime int64) {
	logrus.Infoln("waiting for billing process to start")
	t := time.NewTimer(time.Second * time.Duration(waitTime))
	<-t.C

	logrus.Infoln("begin to do billing process")
	p.startBilling()

	logrus.Infoln("set ticker for billing interval check")
	ticker := time.NewTicker(time.Second * time.Duration(intervalTime))
	go p.run(ticker)
}

func (p *BillingService) run(ticker *time.Ticker) {
	for t := range ticker.C {
		logrus.Debugln("ticker ticked: ", t)
		p.startBilling()
	}
}

func (p *BillingService) startBilling() {
	if !isFirstNodeInZK() {
		logrus.Infoln("current node is not first node in zk, will skip billing process")
		return
	}

	//for shared host
	// p.handleSharedHostCost()

	//for unshared host(cluster)
	p.handleUnsharedHostCost()

}

func (p *BillingService) handleSharedHostCost() {
	logrus.Infoln("handle shared host cost")
	logrus.Infoln("getting service group instances by status")

	//query sgi by status
	statusList := []string{SGO_STATUS_DEPLOYED, SGO_STATUS_MODIFYING,
		SGO_STATUS_TERMINATING, SGO_STATUS_DEPLOYING}

	inMap := make(map[string]interface{})
	inMap["$in"] = statusList

	queryMap := make(map[string]interface{})
	queryMap["life_cycle_status"] = inMap

	querybson, err := mejson.Unmarshal(queryMap)
	if err != nil {
		logrus.Errorln("format querymap for sgo error %v", err)
		return
	}

	queryStruct := dao.QueryStruct{
		CollectionName: GetSgoService().collectionName,
		Selector:       querybson,
		Skip:           0,
		Limit:          0,
		Sort:           ""}

	// jsonDocs := []interface{}{}
	jsonDocs := []entity.ServiceGroupOrder{}
	_, err = dao.HandleQueryAll(&jsonDocs, queryStruct)
	// _, _, jsonDocs, err := p.Dao.HandleQuery(sgoCollection, queryMap,
	// 	false, bson.M{}, 0, 0, "", "true")
	if err != nil {
		logrus.Errorln("query sgo by status error %v", err)
		return
	}

	allDocs := jsonDocs

	//query sgi by status and updatetime
	queryMap["life_cycle_status"] = SGO_STATUS_TERMINATED
	querybson, err = mejson.Unmarshal(queryMap)
	if err != nil {
		logrus.Errorln("format querymap for sgo error %v", err)
		return
	}

	queryStruct = dao.QueryStruct{
		CollectionName: GetSgoService().collectionName,
		Selector:       querybson,
		Skip:           0,
		Limit:          0,
		Sort:           ""}

	jsonDocs = []entity.ServiceGroupOrder{}
	_, err = dao.HandleQueryAll(&jsonDocs, queryStruct)
	// _, _, jsonDocs, err = p.Dao.HandleQuery(sgoCollection, queryMap,
	// 	false, bson.M{}, 0, 0, "", "true")
	if err != nil {
		logrus.Errorln("query sgo by status error %v", err)
		return
	}

	terminateDocs := jsonDocs
	currentTime := time.Now()
	location := currentTime.Location()
	pastDayTime := GetPastDay(currentTime)
	for i := 0; i < len(terminateDocs); i++ {
		// record := terminateDocs[i].(map[string]interface{})
		record := terminateDocs[i]

		// updatetime := record["time_update"].(string)
		updatetime := record.TimeUpdate
		updateTime, _ := time.ParseInLocation(f_rfc3339, updatetime, location)

		if updateTime.Before(currentTime) && updateTime.After(pastDayTime) {
			allDocs = append(allDocs, terminateDocs[i])
		}
	}

	runningCost := p.calculateRunningCost(allDocs)

	p.calculateDesignCost(allDocs, runningCost)
}

func (p *BillingService) handleUnsharedHostCost() {
	logrus.Infoln("handle unshared Host cost")

	clusters, err := GetAllCluster()
	if err != nil {
		logrus.Errorf("get all cluster error %v", err)
		return
	}

	location := time.Now().Location()
	currenttime := int(time.Now().Unix())
	pastDay := GetPastDay(time.Now())
	pastDaytime := int(pastDay.Unix())

	for i := 0; i < len(clusters); i++ {
		cluster := clusters[i]
		if cluster.Status == CLUSTER_STATUS_FAILED {
			logrus.Debugln("cluster [%v] status is failed or terminated, skipped!", cluster.ClusterName)
			continue
		}
		if cluster.Status == CLUSTER_STATUS_TERMINATED {
			updatetime := cluster.TimeUpdate
			updateTime, _ := time.ParseInLocation(f_rfc3339, updatetime, location)
			if updateTime.Before(pastDay) {
				continue
			}
		}

		userid := cluster.User_id
		tenantid := cluster.Tenant_id
		// clusterPrice := float64(0.0)
		logrus.Debugln("handle hosts")
		price := cluster.Flavor.Price
		price = price * 24
		hosts, err := GetHostsByClusterId(cluster.ObjectId.Hex())
		if err != nil {
			logrus.Errorln("get hosts by cluster id error %v", err)
			return
		}
		if hosts != nil && len(hosts) > 0 {
			for i := 0; i < len(hosts); i++ {
				hostPrice := float64(0.0)
				host := hosts[i]
				status := host.Status
				if status == HOST_STATUS_DEPLOYING ||
					status == HOST_STATUS_FAILED {
					logrus.Debugln("host status is deploying or failed, skipped!")
					continue
				} else if status == HOST_STATUS_CREATED ||
					status == HOST_STATUS_DEPLOYED {
					createtime := ParseStringToSecond(host.TimeCreate, location, f_rfc3339)
					if createtime <= pastDaytime {
						hostPrice = hostPrice + price
					} else {
						runningTime := currenttime - createtime
						rate := float64(runningTime) / float64(secInDay)

						cost := price * rate
						hostPrice += cost
					}
				} else if status == HOST_STATUS_TERMINATED {
					updatetime := ParseStringToSecond(host.TimeUpdate, location, f_rfc3339)
					if updatetime <= pastDaytime {
						continue
					} else {
						createtime := ParseStringToSecond(host.TimeCreate, location, f_rfc3339)
						var runningTime int
						if createtime <= pastDaytime {
							runningTime = updatetime - pastDaytime
						} else {
							runningTime = updatetime - createtime
						}

						rate := float64(runningTime) / float64(secInDay)
						cost := price * rate
						hostPrice += cost
					}
				} else {
					logrus.Warnf("unsupported host status for billing %v", status)
					continue
				}

				p.createAccountRecord(hostPrice, "", "", userid, tenantid, account_ConsumeType, account_desc_vm_consume, host.HostName, account_Status_success)

			}
		}

	}
}

func (p *BillingService) calculateRunningCost(allDocs []entity.ServiceGroupOrder) map[string]float64 {
	cpu_cost := common.UTIL.Props.GetFloat64("cpu_cost", 0.2)
	unit_cpu_oneday_cost := cpu_cost * 24 / 1
	memory_cost := common.UTIL.Props.GetFloat64("memory_cost", 0.8)
	unit_memory_oneday_cost := memory_cost * 24 / 1024

	currenttime := int(time.Now().Unix())
	pastDaytime := int(GetPastDay(time.Now()).Unix())

	location := time.Now().Location()
	totalCost := make(map[string]float64)
	if len(allDocs) > 0 {
		for i := 0; i < len(allDocs); i++ {
			record := allDocs[i]
			sgiId := record.ServiceGroupInstanceId

			//get app instances by sgi id
			selector := bson.M{}
			selector["service_group_instance_id"] = sgiId

			queryStruct := dao.QueryStruct{
				CollectionName: GetAciService().collectionName,
				Selector:       selector,
				Skip:           0,
				Limit:          0,
				Sort:           ""}
			// appinsDocuments := []interface{}{}
			appinsDocuments := []entity.AppContainerInstance{}
			_, err := dao.HandleQueryAll(&appinsDocuments, queryStruct)
			if err != nil {
				logrus.Warnln("get app instance by sgi id error %v", err)
				continue
			}

			docs := appinsDocuments
			var allCost float64
			for j := 0; j < len(docs); j++ {
				appins := docs[j]

				state := appins.LifeCycleStatus

				cpus := float64(0)
				mem := float64(0)

				cpus = float64(appins.Cpus)
				mem = float64(appins.Mem)
				// cpusObj := appins["cpus"]
				// if cpusObj != nil {
				// 	cpus = cpusObj.(float64)
				// }
				// memObj := appins["mem"]
				// if memObj != nil {
				// 	mem = memObj.(float64)
				// }

				//failed app instances will not be charged in this unit time
				if state == ACI_STATUS_UNCONFIGED || state == ACI_STATUS_FAILED {
					continue
				}

				if state == ACI_STATUS_CREATED || state == ACI_STATUS_CONFIGED {

					// createtime := appins["time_create"].(map[string]interface{})["$date"].(int)
					createtime := ParseStringToSecond(appins.TimeCreate, location, f_rfc3339)
					if createtime <= pastDaytime {
						cost := cpus*unit_cpu_oneday_cost + mem*unit_memory_oneday_cost
						allCost += cost
					} else {
						runningTime := currenttime - createtime
						rate := float64(runningTime) / float64(secInDay)

						cost := (cpus*unit_cpu_oneday_cost + mem*unit_memory_oneday_cost) * rate
						allCost += cost
					}
				}

				if state == ACI_STATUS_TERMINATED {
					updatetime := ParseStringToSecond(appins.TimeUpdate, location, f_rfc3339)
					if updatetime <= pastDaytime {
						continue
					} else {
						createtime := ParseStringToSecond(appins.TimeCreate, location, f_rfc3339)
						var runningTime int
						if createtime <= pastDaytime {
							runningTime = updatetime - pastDaytime
						} else {
							runningTime = updatetime - createtime
						}

						rate := float64(runningTime) / float64(secInDay)
						cost := (cpus*unit_cpu_oneday_cost + mem*unit_memory_oneday_cost) * rate
						allCost += cost

					}
				}
			}

			userid := record.User_id
			servicegroupobjid := record.ServiceGroupObjId

			key := userid + "&" + servicegroupobjid
			_, ok := totalCost[key]
			if !ok {
				totalCost[key] = allCost
			}

		}
	}

	return totalCost

}

func (p *BillingService) calculateDesignCost(alltimeDocs []entity.ServiceGroupOrder,
	runningCost map[string]float64) {

	currenttime := int(time.Now().Unix())
	pastDaytime := int(GetPastDay(time.Now()).Unix())
	location := time.Now().Location()

	if len(alltimeDocs) > 0 {
		for i := 0; i < len(alltimeDocs); i++ {
			record := alltimeDocs[i]

			userid := record.User_id
			tenantid := record.Tenant_id
			servicegroupid := record.ServiceGroupId
			servicegroupobjid := record.ServiceGroupObjId

			// createtime := record["time_create"].(map[string]interface{})["$date"].(int)
			createtime := ParseStringToSecond(record.TimeCreate, location, f_rfc3339)
			updatetime := ParseStringToSecond(record.TimeUpdate, location, f_rfc3339)
			state := record.LifeCycleStatus
			if state == SGI_STATUS_DEPLOYED || state == SGI_STATUS_MODIFYING || state == SGI_STATUS_TERMINATING || state == SGI_STATUS_DEPLOYING {
				if createtime <= pastDaytime {
					p.buildAndCreateUserAccountRecord(pastDaytime, currenttime,
						runningCost, servicegroupid, servicegroupobjid, userid, tenantid)
				} else {
					p.buildAndCreateUserAccountRecord(createtime, currenttime,
						runningCost, servicegroupid, servicegroupobjid, userid, tenantid)
				}
			} else if state == SGI_STATUS_TERMINATED {
				if updatetime > pastDaytime {
					p.buildAndCreateUserAccountRecord(pastDaytime, updatetime,
						runningCost, servicegroupid, servicegroupobjid, userid, tenantid)
				}
			}

		}
	}
}

func (p *BillingService) QueryBillingModels(token string, skip int, limit int, sort string) (bms []entity.BillingModel, count int, errorCode string, err error) {
	code, err := TokenValidation(token)
	if err != nil {
		return nil, 0, code, err
	}

	authQuery, err := GetAuthService().BuildQueryByAuth("list_billing", token)
	if err != nil {
		logrus.Errorf("get auth query by token [%v] error is %v", token, err)
		return nil, 0, BILLING_ERROR_GET, err
	}

	queryStruct := dao.QueryStruct{
		CollectionName: p.collectionName,
		Selector:       authQuery,
		Skip:           skip,
		Limit:          limit,
		Sort:           sort}
	bms = []entity.BillingModel{}
	count, err = dao.HandleQueryAll(&bms, queryStruct)
	if err != nil {
		logrus.Errorln("query billing models error %v", err)
		return nil, 0, COMMON_ERROR_INTERNAL, err
	}

	return
}

func (p *BillingService) QueryAllBillingModels(skip int, limit int, sort string) (count int, bms []entity.BillingModel, errorCode string, err error) {
	queryStruct := dao.QueryStruct{
		CollectionName: p.collectionName,
		Selector:       bson.M{},
		Skip:           skip,
		Limit:          limit,
		Sort:           sort}
	bms = []entity.BillingModel{}
	count, err = dao.HandleQueryAll(&bms, queryStruct)
	if err != nil {
		logrus.Errorln("get all billing model error %v", err)
		return count, nil, BILLING_ERROR_GET, err
	}

	return
}

func (p *BillingService) QueryBillingModelById(token string, id string) (bm *entity.BillingModel, errorCode string, err error) {
	if len(token) <= 0 || len(id) <= 0 {
		logrus.Error("token or billing id should not be null!")
		return nil, BILLING_ERROR_GET, errors.New("invalid parameter for query billing model")
	}

	code, err := TokenValidation(token)
	if err != nil {
		return nil, code, err
	}

	if authorized := GetAuthService().Authorize("get_billing", token, id, p.collectionName); !authorized {
		logrus.Errorln("required opertion is not allowed!")
		return nil, COMMON_ERROR_UNAUTHORIZED, errors.New("Required opertion is not authorized!")
	}

	selector := bson.M{}
	selector["_id"] = bson.ObjectIdHex(id)

	queryStruct := dao.QueryStruct{
		CollectionName: p.collectionName,
		Selector:       selector,
		Skip:           0,
		Limit:          0,
		Sort:           ""}

	bm = new(entity.BillingModel)
	err = dao.HandleQueryOne(bm, queryStruct)
	if err != nil {
		logrus.Error("get billing model by id error %v", err)
		return nil, BILLING_ERROR_GET, err
	}

	return

}

func (p *BillingService) CreateBillingModel(bm *entity.BillingModel, token string) (id string, errorCode string, err error) {
	code, err := TokenValidation(token)
	if err != nil {
		return "", code, err
	}

	if authorized := GetAuthService().Authorize("create_billing", token, "", p.collectionName); !authorized {
		logrus.Errorln("required opertion is not allowed!")
		return "", COMMON_ERROR_UNAUTHORIZED, errors.New("required opertion is not authorized!")
	}

	Token, erro := GetTokenById(token)
	if erro != nil {
		logrus.Errorln("invalid token for billing model create!")
		return "", BILLING_ERROR_CREATE, errors.New("no token for operation")
	}

	bm.User_id = Token.User.Id
	bm.Tenant_id = Token.Tenant.Id
	bm.ObjectId = bson.NewObjectId()
	currentTime := getCurrentTime()
	bm.TimeCreate = currentTime
	bm.TimeUpdate = currentTime

	err = dao.HandleInsert(p.collectionName, bm)
	if err != nil {
		logrus.Errorln("create billing model error %v", err)
		return "", BILLING_ERROR_CREATE, err
	}

	return bm.ObjectId.Hex(), "", nil

}

func (p *BillingService) UpdateBillingModel(bm *entity.BillingModel, token string, billingModelId string) (id string, errorCode string, err error) {
	code, err := TokenValidation(token)
	if err != nil {
		return "", code, err
	}

	if authorized := GetAuthService().Authorize("update_billing", token, billingModelId, p.collectionName); !authorized {
		logrus.Errorln("required opertion is not allowed!")
		return "", COMMON_ERROR_UNAUTHORIZED, errors.New("required opertion is not authorized!")
	}

	Token, erro := GetTokenById(token)
	if erro != nil {
		logrus.Errorln("invalid token for billing model create!")
		return "", BILLING_ERROR_UPDATE, errors.New("no token for operation")
	}

	bm.User_id = Token.User.Id
	bm.Tenant_id = Token.Tenant.Id

	currentTime := getCurrentTime()
	bm.TimeUpdate = currentTime

	selector := bson.M{}
	selector["_id"] = bson.ObjectIdHex(billingModelId)

	queryStruct := dao.QueryStruct{
		CollectionName: p.collectionName,
		Selector:       selector,
		Skip:           0,
		Limit:          0,
		Sort:           ""}

	_, err = dao.HandleUpdateOne(bm, queryStruct)
	if err != nil {
		logrus.Errorln("update billing model error %v", err)
		return "", BILLING_ERROR_UPDATE, err
	}

	return billingModelId, "", nil

}

func (p *BillingService) buildAndCreateUserAccountRecord(start int, end int,
	runningCost map[string]float64, servicegroupid string, servicegroupobjid string,
	userid string, tenantid string) {

	rate := float64(end-start) / float64(secInDay)
	key := userid + "&" + servicegroupobjid
	cost1, ok := runningCost[key]
	if ok {

		billModel, err := p.getBillingByServiceGroupId(servicegroupobjid)
		if err != nil {
			logrus.Infof("get billing model by service group id error %v", err)
			//no billing model, only record running cost
			p.createAccountRecord(cost1, servicegroupid, servicegroupobjid,
				userid, tenantid, account_ConsumeType, account_desc_container_consume, servicegroupid, account_Status_success)
			return
		}

		designPricePerDay := billModel.TotalPrice * 24
		allcost := cost1 + designPricePerDay*rate
		p.createAccountRecord(allcost, servicegroupid, servicegroupobjid,
			userid, tenantid, account_ConsumeType, account_desc_container_consume, servicegroupid, account_Status_success)

		billingModels := list.New()
		billingModels.PushBack(billModel)

		length := billingModels.Len()
		if length <= 0 {
			return
		}

		for length > 0 {
			element := billingModels.Front()
			bm := element.Value.(*entity.BillingModel)
			price := bm.Price * 24 * rate
			sg, err := getSgById(bm.ModelId)
			if err != nil {
				logrus.Warnln("get service group by _id error %v", err)
				return
			}

			p.createAccountRecord(price, sg.Id, bm.ModelId, bm.User_id,
				bm.Tenant_id, account_InComingType, account_desc_incoming, sg.Id, account_Status_success)

			refs := bm.Refs
			if len(refs) > 0 {
				for i := 0; i < len(refs); i++ {
					billmodel, err := p.getBillingById(refs[i])
					if err != nil {
						logrus.Infoln("get billing by id error %v", err)
						continue
					}

					billingModels.PushBack(billmodel)
				}
			}
			billingModels.Remove(element)
			length = billingModels.Len()

		}

	}

}

func (p *BillingService) getBillingById(billingId string) (billingmodel *entity.BillingModel, err error) {
	selector := bson.M{}
	selector["_id"] = bson.ObjectIdHex(billingId)

	billingmodel = new(entity.BillingModel)
	err = dao.HandleQueryOne(billingmodel, dao.QueryStruct{p.collectionName, selector, 0, 0, ""})
	if err != nil {
		return
	}

	return
}

func (p *BillingService) getBillingByServiceGroupId(sgId string) (billingmodel *entity.BillingModel, err error) {

	selector := bson.M{}
	selector["modelid"] = sgId

	billingmodel = new(entity.BillingModel)
	err = dao.HandleQueryOne(billingmodel, dao.QueryStruct{p.collectionName, selector, 0, 0, ""})
	if err != nil {
		return
	}

	return
}

func (p *BillingService) createAccountRecord(price float64, sgid string, sgobjid string,
	userid string, tenantid string, tranType string, tranDesc string, tranObj string, tranStatus string) {

	username, ok := userMap[userid]
	if !ok {
		token, erro := GenerateToken()
		if erro != nil {
			logrus.Errorln("generate token error %v", erro)
			username = userid
		} else {
			user, err := GetUserById(token, userid)
			if err != nil {
				logrus.Errorln("get user by id error %v", err)
				username = userid
			} else {
				username = user.Email
			}
		}

		userMap[userid] = username
	}

	currentTime := getCurrentTime()
	formatValue := formatValue(price)
	record := entity.UserAccount{
		ObjectId:           bson.NewObjectId(),
		Username:           username,
		User_id:            userid,
		Tenant_id:          tenantid,
		Transaction_type:   tranType,
		Transaction_desc:   tranDesc,
		Transaction_object: tranObj,
		Transaction_status: tranStatus,
		Sg_id:              sgid,
		Sg_objId:           sgobjid,
		Price:              formatValue,
		Date:               int(time.Now().Unix()),
		TimeCreate:         currentTime,
		TimeUpdate:         currentTime,
	}

	logrus.Debugln("1. insert a user account")
	_, err := GetUserAccountService().createUserAccount(record)
	if err != nil {
		logrus.Warnf("save account record err is %v", err)
		return
	}

	logrus.Debugln("2. update user account balance")
	if tranType == account_ConsumeType {
		err := GetUserAccountBalanceService().updateConsumeByUserid(userid, formatValue)
		if err != nil {
			logrus.Errorf("update user account balance's consume error %v", err)
			return
		}
	} else if tranType == account_InComingType {
		err := GetUserAccountBalanceService().updateIncomeByUserid(userid, formatValue)
		if err != nil {
			logrus.Errorf("update user account balance's incoming error %v", err)
			return
		}
	} else {
		logrus.Warnln("unsupported transaction type: ", tranType)
		return
	}

	return

}
