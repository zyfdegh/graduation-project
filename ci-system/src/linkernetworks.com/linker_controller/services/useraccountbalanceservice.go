package services

import (
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/Sirupsen/logrus"
	"gopkg.in/mgo.v2/bson"
	"linkernetworks.com/linker_common_lib/persistence/dao"
	"linkernetworks.com/linker_common_lib/persistence/entity"
	"linkernetworks.com/linker_controller/common"
)

var (
	userAccountBalanceService *UserAccountBalanceService = nil
	onceUserAccountBalance    sync.Once
)

var noBalanceBody = `NEWUSER, 这封邮件是由领科云发送的。
您收到这封邮件，是由于您在领科云的账户余额不足，您在领科
云上的所有服务将被停止并删除，请悉知。

如果有任何问题，请发送邮件到 support@linkernetworks.com 
请勿回复该邮件


此致
领科云管理团队





NEWUSER, This email is sent by Linker Cloud Platform.
You don't have sufficent account balance in Linker Cloud
Platform, all your services will be terminated and removed.

Any problems, please send mail to support@linkernetworks.com
Please DO NOT reply this mail

Thanks & BestRegards!

Linker Cloud Platform Team`

var currentBalanceBody = `NEWUSER, 这封邮件是由领科云发送的。
当前您在领科云的账户余额为:CURRENTVALUE,请注意充值。

如果有任何问题，请发送邮件到 support@linkernetworks.com 
请勿回复该邮件


此致
领科云管理团队





NEWUSER, This email is sent by Linker Cloud Platform.
Your current account balance is:CURRENTVALUE, please
recharge.

Any problems, please send mail to support@linkernetworks.com
Please DO NOT reply this mail

Thanks & BestRegards!

Linker Cloud Platform Team`

type UserAccountBalanceService struct {
	collectionName string
}

func GetUserAccountBalanceService() *UserAccountBalanceService {
	onceUserAccountBalance.Do(func() {
		logrus.Debugf("Once called from userAccountBalanceService ......................................")
		userAccountBalanceService = &UserAccountBalanceService{"user_account_balance"}
	})
	return userAccountBalanceService
}

func (p *UserAccountBalanceService) QueryByUserId(token string, userId string) (uabs []entity.UserAccountBalance, errorCode string, err error) {

	code, err := TokenValidation(token)
	if err != nil {
		return nil, code, err
	}

	authQuery, err := GetAuthService().BuildQueryByAuth("list_account_balance", token)
	if err != nil {
		logrus.Errorf("get auth query by token [%v] error is %v", token, err)
		return nil, COMMON_ERROR_INTERNAL, err
	}

	if len(userId) > 0 && !bson.IsObjectIdHex(userId) {
		logrus.Errorln("not valid user id for query user account balance operation:", userId)
		return nil, COMMON_ERROR_INVALIDATE, errors.New("no valid object id parameter!")
	}

	query := bson.M{}
	if len(userId) > 0 {
		query["user_id"] = userId
	}

	mergeQuery(query, authQuery)

	uabs, _, errorCode, err = p.queryByQuery(query, 0, 0, "")

	if len(uabs) <= 0 {
		uab := entity.UserAccountBalance{User_id: userId, Balance: 0.0, Consume: 0.0, Income: 0.0}
		uabs = append(uabs, uab)
	}

	return
}

func (p *UserAccountBalanceService) queryByQuery(query bson.M, skip int, limit int, sort string) (account []entity.UserAccountBalance, count int, errorCode string, err error) {
	logrus.Debugln("query user account balance by query:", query)
	queryStruct := dao.QueryStruct{
		CollectionName: p.collectionName,
		Selector:       query,
		Skip:           skip,
		Limit:          limit,
		Sort:           sort}
	account = []entity.UserAccountBalance{}
	count, err = dao.HandleQueryAll(&account, queryStruct)
	if err != nil {
		logrus.Errorln("query user account balance error %v", err)
		return nil, 0, COMMON_ERROR_INTERNAL, err
	}

	return
}

func (p *UserAccountBalanceService) updateBalanceByUserid(userId string, price float64) (err error) {
	logrus.Infof("update user account balance by user id: %v", userId)
	uabs, err := p.queryByUserId(userId)
	if err != nil {
		return err
	}

	var uab entity.UserAccountBalance
	if len(uabs) > 1 {
		logrus.Warnln("exist more than one user account balance record for userId: ", userId)
		uab = uabs[0]
	} else if len(uabs) <= 0 {
		record, _ := p.buildRecord(userId, price, false)
		_, err = p.create(record)
		return
	} else {
		uab = uabs[0]
	}

	value := formatValue(price)
	uab.Balance += value

	//reset user's notifylevel
	uab.NotifyLevel = 0

	//update
	id := uab.ObjectId.Hex()
	err = p.update(uab, id)
	return
}

func (p *UserAccountBalanceService) updateConsumeByUserid(userId string, value float64) (err error) {
	logrus.Infof("update user account consume by userid: %v", userId)
	uabs, err := p.queryByUserId(userId)
	if err != nil {
		return err
	}

	var uab entity.UserAccountBalance
	if len(uabs) > 1 {
		logrus.Warnf("exist more than one user account balance record for userId: %v", userId)
		uab = uabs[0]
	} else if len(uabs) <= 0 {
		logrus.Errorf("no user account balance record for userId: %v", userId)
		// return errors.New("no user account balance record for userId")
		DeleteClusterByUserId(userId)
		return
	} else {
		uab = uabs[0]
	}

	if uab.Balance == 0 {
		logrus.Infof("user [%v] account balance is zero, can not continue consuming", userId)
		return
	}

	uab.Consume += value
	uab.Balance -= value
	if uab.Balance < 0.0 {
		logrus.Warnf("user [%v] account balance is zero!", userId)
		uab.Balance = float64(0.0)
	}

	notifyEmail(&uab)

	//update
	id := uab.ObjectId.Hex()
	err = p.update(uab, id)
	if err != nil {
		return
	}

	if uab.Balance <= float64(0.0) {
		DeleteClusterByUserId(userId)
		return
	}

	return

}

func (p *UserAccountBalanceService) updateIncomeByUserid(userId string, value float64) (err error) {
	logrus.Infof("update user account Incoming by user id: %v", userId)
	uabs, err := p.queryByUserId(userId)
	if err != nil {
		return err
	}

	var uab entity.UserAccountBalance
	if len(uabs) > 1 {
		logrus.Warnln("exist more than one user account balance record for userId: ", userId)
		uab = uabs[0]
	} else if len(uabs) <= 0 {
		record, _ := p.buildRecord(userId, value, true)
		_, err = p.create(record)
		return
	} else {
		uab = uabs[0]
	}

	newvalue := formatValue(value)
	uab.Income += newvalue
	uab.Balance += newvalue

	//update
	id := uab.ObjectId.Hex()
	err = p.update(uab, id)

	return
}

func (p *UserAccountBalanceService) buildRecord(userId string, price float64, isIncoming bool) (uab entity.UserAccountBalance, err error) {
	logrus.Infoln("build new user account balance record by userId and price")
	tenantId := ""
	token, err := GenerateToken()
	if err != nil {
		logrus.Warnf("generate token error %v", err)
	} else {
		id, err := GetTenantIdByUserId(token, userId)
		if err != nil {
			logrus.Warnf("get tenantId by userId error %v", err)
		} else {
			tenantId = id
		}
	}

	value := formatValue(price)
	currenttime := dao.GetCurrentTime()
	uab = entity.UserAccountBalance{
		ObjectId:   bson.NewObjectId(),
		User_id:    userId,
		Tenant_id:  tenantId,
		Balance:    value,
		Consume:    float64(0.0),
		Income:     float64(0.0),
		TimeCreate: currenttime,
		TimeUpdate: currenttime}

	if isIncoming {
		uab.Income = value
	}

	return uab, nil

}

func (p *UserAccountBalanceService) create(uab entity.UserAccountBalance) (newuab entity.UserAccountBalance, err error) {
	logrus.Infoln("create a user account balance")

	uab.ObjectId = bson.NewObjectId()
	currenttime := dao.GetCurrentTime()
	uab.TimeCreate = currenttime
	uab.TimeUpdate = currenttime

	err = dao.HandleInsert(p.collectionName, &uab)
	if err != nil {
		logrus.Errorf("create new user account balance %v error %v", uab, err)
		return
	}

	newuab = uab

	return
}

func (p *UserAccountBalanceService) update(uab entity.UserAccountBalance, objectId string) (err error) {
	logrus.Infoln("update user account balance")

	uab.TimeUpdate = dao.GetCurrentTime()

	query := bson.M{}
	query["_id"] = uab.ObjectId

	_, err = dao.HandleUpdateOne(&uab, dao.QueryStruct{p.collectionName, query, 0, 0, ""})
	if err != nil {
		logrus.Errorln("update user account balance %v error %v", uab, err)
		return
	}

	return
}

func (p *UserAccountBalanceService) queryByUserId(userId string) (uabs []entity.UserAccountBalance, err error) {
	logrus.Infoln("query user account balance by userid")

	query := bson.M{}
	query["user_id"] = userId

	uabs, _, _, err = p.queryByQuery(query, 0, 0, "")
	if err != nil {
		logrus.Errorf("query user account balance by userid error: %v", err)
		return
	}

	return

}

func notifyEmail(uab *entity.UserAccountBalance) {
	balance := uab.Balance
	level := uab.NotifyLevel
	if balance <= float64(0.0) {
		if level == 3 {
			return
		} else {
			uab.NotifyLevel = 3
			logrus.Infoln("send notification email to user for insufficient balance")
			subject := "[Linker Cloud Platform] Account Balance Notification"
			body := strings.Replace(noBalanceBody, "CURRENTVALUE", fmt.Sprintf("%.3f", balance), -1)
			sendEmail(uab.User_id, subject, body)
			return
		}
	}
	if balance < float64(50.0) {
		if level == 2 {
			return
		} else {
			uab.NotifyLevel = 2
			logrus.Infoln("send notification email to user for current balance")
			subject := "[Linker Cloud Platform] Account Balance Notification"
			body := strings.Replace(currentBalanceBody, "CURRENTVALUE", fmt.Sprintf("%.3f", balance), -1)
			sendEmail(uab.User_id, subject, body)
			return
		}
	}
	if uab.Balance <= float64(100.0) {
		if level == 1 {
			return
		} else {
			uab.NotifyLevel = 1
			logrus.Infoln("send notification email to user for current balance")
			subject := "[Linker Cloud Platform] Account Balance Notification"
			body := strings.Replace(currentBalanceBody, "CURRENTVALUE", fmt.Sprintf("%.3f", balance), -1)
			sendEmail(uab.User_id, subject, body)
			return
		}
	}
}

func sendEmail(userId string, subject string, body string) {

	token, err := GenerateToken()
	if err != nil {
		logrus.Errorf("generate token error for sending notify email %v", err)
		return
	}
	user, err := GetUserById(token, userId)
	if err != nil {
		logrus.Errorf("get user by id error for sending notify email %v", err)
	}

	emailHost := common.UTIL.Props.GetString("email.host", "")
	emailUsername := common.UTIL.Props.GetString("email.username", "")
	emailPasswd := common.UTIL.Props.GetString("email.password", "")

	newbody := strings.Replace(body, "NEWUSER", user.Username, -1)

	go SendMail(emailHost, emailUsername, emailPasswd, user.Email, subject, newbody)

}
