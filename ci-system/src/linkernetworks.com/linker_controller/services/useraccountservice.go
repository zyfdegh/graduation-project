package services

import (
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/compose/mejson"
	"gopkg.in/mgo.v2/bson"
	"linkernetworks.com/linker_common_lib/persistence/dao"
	"linkernetworks.com/linker_common_lib/persistence/entity"
)

var (
	userAccountService *UserAccountService = nil
	onceUserAccount    sync.Once

	account_Status_success    = "success"
	account_Status_failed     = "failed"
	account_Status_processing = "processing"

	account_InComingType = "incoming"
	account_ConsumeType  = "consume"
	account_RechargeType = "recharge"

	account_desc_sys_present       = "sys_present"
	account_desc_transfer_pay      = "transfer_pay"
	account_desc_vm_consume        = "vm_consume"
	account_desc_container_consume = "container_consume"
	account_desc_incoming          = "sg_referenced"

	f_datetime = "2006-01-02 15:04:05"

	USERACCOUNT_ERROR_RECHARGE = "E11101"
)

type UserAccountService struct {
	collectionName string
}

func GetUserAccountService() *UserAccountService {
	onceUserAccount.Do(func() {
		logrus.Debugf("Once called from userAccountService ......................................")
		userAccountService = &UserAccountService{"user_account"}
	})
	return userAccountService
}

func (p *UserAccountService) CreateRechargeUserAccount(tokenId string, userAccount entity.UserAccount) (ua entity.UserAccount, errorCode string, err error) {
	logrus.Infof("create recharge useraccount %v", userAccount)

	errorCode, err = TokenValidation(tokenId)
	if err != nil {
		return
	}

	if authorized := GetAuthService().Authorize("create_account", tokenId, "", p.collectionName); !authorized {
		err = errors.New("required opertion is not authorized!")
		errorCode = COMMON_ERROR_UNAUTHORIZED
		logrus.Errorf("create useraccount error is %v", err)
		return
	}

	Token, erro := GetTokenById(tokenId)
	if erro != nil {
		logrus.Errorln("invalid token for recharge user account create!")
		return ua, USERACCOUNT_ERROR_RECHARGE, errors.New("no token for operation")
	}

	userAccount.User_id = Token.User.Id
	userAccount.Tenant_id = Token.Tenant.Id
	userAccount.Username = Token.User.Username
	userAccount.Price = formatValue(userAccount.Price / 100.0)
	userAccount.TimeCreate = dao.GetCurrentTime()
	userAccount.TimeUpdate = userAccount.TimeCreate
	userAccount.Transaction_status = account_Status_processing
	userAccount.Date = int(time.Now().Unix())
	userAccount.ObjectId = bson.NewObjectId()

	err = dao.HandleInsert(p.collectionName, userAccount)
	if err != nil {
		logrus.Errorf("insert user account to db error is %v", err)
		errorCode = USERACCOUNT_ERROR_RECHARGE
		return
	}

	ua = userAccount

	return
}

func (p *UserAccountService) NotifyRechargeStatusById(status string, id string, checkKey string, signature string) (err error) {
	logrus.Infof("update useraccount recharge status to %v by id %v ", status, id)
	if status != account_Status_failed && status != account_Status_processing && status != account_Status_success {
		logrus.Errorf("not supported user account status %v", status)
		return errors.New("not supported user account status")
	}

	if !bson.IsObjectIdHex(id) {
		err = errors.New("invalidate ObjectId for update user account")
		return
	}

	currentHashValue := HashString(checkKey)
	if signature != currentHashValue {
		logrus.Errorf("wrong checkKey[%v] and signature[%v] for update user account, user account status will be set to failed!", currentHashValue, signature)
		return errors.New("invalid signature for update user account")
	}

	logrus.Infoln("1. update useraccount status")
	selector := bson.M{}
	selector["_id"] = bson.ObjectIdHex(id)
	ua := entity.UserAccount{}
	err = dao.HandleQueryOne(&ua, dao.QueryStruct{p.collectionName, selector, 0, 0, ""})
	if err != nil {
		logrus.Errorf("query user account [objectId=%v] error is %v", id, err)
		return
	}

	if ua.Transaction_type != account_RechargeType {
		logrus.Errorf("only support recharge type, current user account [%v] type is %v", id, ua.Transaction_type)
		return
	}

	if ua.Transaction_status == status {
		logrus.Infof("the user account record is in [%v] stauts, no need to update it again", status)
		return
	}

	ua.Transaction_status = status
	ua.TimeUpdate = dao.GetCurrentTime()

	_, err = dao.HandleUpdateOne(&ua, dao.QueryStruct{p.collectionName, selector, 0, 0, ""})
	if err != nil {
		logrus.Errorf("update user account [objectId=%v] error is %v", id, err)
		return
	}

	if ua.Transaction_type == account_RechargeType && status == account_Status_success {
		logrus.Infoln("2. update user account balance for recharge operation")
		err = GetUserAccountBalanceService().updateBalanceByUserid(ua.User_id, ua.Price)
		if err != nil {
			logrus.Errorf("update user account balance error %v", err)
			return
		}
	}

	return
}

func (p *UserAccountService) createUserAccount(userAccount entity.UserAccount) (newua entity.UserAccount, err error) {
	userAccount.ObjectId = bson.NewObjectId()

	err = dao.HandleInsert(p.collectionName, &userAccount)
	if err != nil {
		logrus.Errorf("insert user account to db error is %v", err)
		return
	}

	newua = userAccount

	return
}

func (p *UserAccountService) ListUserAccount(token string, from string, to string, transactionType string, skip int, limit int, sort string) (account []entity.UserAccount, count int, errorCode string, err error) {
	code, err := TokenValidation(token)
	if err != nil {
		return nil, 0, code, err
	}

	authQuery, err := GetAuthService().BuildQueryByAuth("list_account", token)
	if err != nil {
		logrus.Errorf("get auth query by token [%v] error is %v", token, err)
		return nil, 0, COMMON_ERROR_UNAUTHORIZED, err
	}

	query := buildQuery(from, to, transactionType)

	mergeQuery(query, authQuery)

	return p.queryByQuery(query, skip, limit, sort)

}

func (p *UserAccountService) queryByQuery(query bson.M, skip int, limit int, sort string) (account []entity.UserAccount, count int, errorCode string, err error) {
	logrus.Debugln("query user account by query:", query)
	queryStruct := dao.QueryStruct{
		CollectionName: p.collectionName,
		Selector:       query,
		Skip:           skip,
		Limit:          limit,
		Sort:           sort}
	account = []entity.UserAccount{}
	count, err = dao.HandleQueryAll(&account, queryStruct)
	if err != nil {
		logrus.Errorf("query user account error %v", err)
		return nil, 0, COMMON_ERROR_INTERNAL, err
	}

	return
}

func buildQuery(from string, to string, transactionType string) bson.M {
	ret := make(map[string]interface{})

	timePart := make(map[string]interface{})
	fromTime, err := parseTime(from)
	if err == nil {
		timePart["$gte"] = fromTime
	}
	toTime, err := parseTime(to)
	if err == nil {
		timePart["$lte"] = toTime
	}

	if len(timePart) > 0 {
		ret["date"] = timePart
	}

	if strings.EqualFold(transactionType, "incoming") {
		ret["transaction_type"] = account_InComingType
	} else if strings.EqualFold(transactionType, "consume") {
		ret["transaction_type"] = account_ConsumeType
	}

	querybson, err := mejson.Unmarshal(ret)
	if err != nil {
		logrus.Errorln("format querymap for user account error %v", err)
		return ret
	}

	return querybson

}

func parseTime(date string) (int64, error) {
	if len(date) <= 0 {
		logrus.Infoln("date parameter is null")
		return 0, errors.New("date parameter is null")
	}

	newexec := date + " 00:00:00"
	execTime, err := time.ParseInLocation(f_datetime, newexec, time.Now().Location())
	if err != nil {
		logrus.Warnf("error to parse exec check time: %v , error: %v", newexec, err)

		return 0, errors.New("error to parse time")
	}

	return execTime.Unix(), nil
}

func mergeQuery(basequery bson.M, plusquery bson.M) {
	if plusquery == nil || len(plusquery) <= 0 {
		return
	}
	for key, value := range plusquery {
		basequery[key] = value
	}

	return
}
