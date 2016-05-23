package services

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/compose/mejson"
	"gopkg.in/mgo.v2/bson"
	"linkernetworks.com/linker_common_lib/persistence/dao"
	"linkernetworks.com/linker_common_lib/persistence/entity"
	"linkernetworks.com/linker_usermgmt/common"
)

var sysadmin_user = "sysadmin"
var sysadmin_alias = "linker"
var sysadmin_pass = "password"
var sys_tenant = "sysadmin"
var sys_admin_role = "sysadmin"
var admin_role = "admin"
var common_role = "common"

var USER_ERROR_REG = "E10000"
var USER_ERROR_EXCEED = "E10001"
var USER_ERROR_ACTIVE = "E10002"
var USER_ERROR_CREATE = "E10003"
var USER_ERROR_NOEXIST = "E10004"
var USER_ERROR_WRONGPW = "E10005"
var USER_ERROR_INACTIVE_USER = "E10006"
var USER_ERROR_UPDATE = "E10007"
var USER_ERROR_DUP_NAMESPACE = "E10008"
var USER_ERROR_EXIST = "E10009"
var USER_ERROR_DELETE = "E10010"
var USER_ERROR_GET = "E10011"
var USER_ERROR_LOGIN = "E10012"

var f_rfc3339 = "2006-01-02T15:04:05Z07:00"

var key = "Pa55w0rd"
var activeCode_expire = 60 * 60 * 72

var emailBody = `NEWUSER, 这封邮件是由领科云发送的。
您收到这封邮件，是由于在Linker云进行了新用户注册。
如果您没有访问过Linker云或没有进行上述操作，请忽略这封邮件。

-------------------------------------------------------------------------------
	               用户注册说明
------------------------------------------------------------------------------- 
	         
如果您是领科云的新用户，我们需要对您的地址有效性进行验证以避免垃圾邮件或地址被滥用。

您只需点击下面的链接即可进行用户注册，以下链接有效期为3天。过期可以重新请求发送一封新的邮件验证：
NEWURL

如果有任何问题，请发送邮件到 support@linkernetworks.com 
请勿回复该邮件

感谢您的访问，祝您使用愉快！

此致
领科云管理团队





NEWUSER, This email is sent by Linker Cloud Platform.
You have registered an new account on Liner Cloud Platform,
Please ignore this email if you have not done above operations.

-------------------------------------------------------------------------------
	               usage instruction
------------------------------------------------------------------------------- 
	         
Please click below link to complete your registry. 
To protect your account, the verification link is only valid for 3 days.
NEWURL

Any problems, please send mail to support@linkernetworks.com
Please DO NOT reply this mail

Thanks & BestRegards!

Linker Cloud Platform Team`

var userService *UserService = nil
var userOnce sync.Once

type UserService struct {
	userCollectionName string
}

func GetUserService() *UserService {
	userOnce.Do(func() {
		userService = &UserService{"user"}

		userService.initialize()
	})

	return userService
}

func (p *UserService) initialize() bool {
	logrus.Infoln("UserMgmt initialize...")

	logrus.Infoln("check sysadmin tenant")

	_, tenantErr := GetTenantService().createAndInsertTenant(sys_tenant, "system admin tenant")
	if tenantErr != nil {
		logrus.Errorf("create and insert sys admin tenant error,  err is %v", tenantErr)
		return false
	}

	logrus.Infoln("check sysadmin role")
	_, roleErr := GetRoleService().createAndInsertRole(sys_admin_role, "sysadmin role")
	if roleErr != nil {
		logrus.Errorf("create and insert sys admin role error,  err is %v", roleErr)
		return false
	}

	logrus.Infoln("check admin role")
	_, roleErr = GetRoleService().createAndInsertRole(admin_role, "admin role")
	if roleErr != nil {
		logrus.Errorf("create and insert admin role error,  err is %v", roleErr)
		return false
	}

	logrus.Infoln("check common role")
	_, roleErr = GetRoleService().createAndInsertRole(common_role, "common role")
	if roleErr != nil {
		logrus.Errorf("create and insert common role error,  err is %v", roleErr)
		return false
	}

	logrus.Infoln("check sysadmin user")
	encryPassword := common.HashString(sysadmin_pass)
	_, userErr := p.createAndInsertUser(sysadmin_user, sysadmin_alias, encryPassword, sysadmin_user, 1, sys_tenant, sys_admin_role, "", "", "", "")
	if userErr != nil {
		logrus.Errorf("create and insert sysadmin user error,  err is %v", userErr)
		return false
	}

	p.userTimerCheck()

	return true
}

func (p *UserService) userTimerCheck() {
	logrus.Infoln("initialize expire user check and clean process")
	interval := common.UTIL.Props.GetString("user_check_interval", "259200") //default 3 days
	if len(interval) <= 0 {
		interval = "259200"
	}
	exec := common.UTIL.Props.GetString("user_check_time", "01:00:00")
	if len(exec) <= 0 {
		exec = "01:00:00"
	}

	formatdate := time.Now().Format(f_date)

	newexec := formatdate + " " + exec
	execTime, err := time.ParseInLocation(f_datetime, newexec, time.Now().Location())
	if err != nil {
		logrus.Warnln("error to parse exec check time: ", newexec)
		execTime, _ = time.ParseInLocation(f_datetime, formatdate+" 01:00:00", time.Now().Location())
	}

	intervalInt, err := strconv.ParseInt(interval, 10, 64)
	if err != nil {
		logrus.Warnln("error to parse intervalTime: ", interval)
		intervalInt, _ = strconv.ParseInt("259200", 10, 64)
	}

	waitTime := common.GetWaitTime(execTime)

	go p.startUserTimer(waitTime, intervalInt)
}

func (p *UserService) startUserTimer(waitTime int64, intervalTime int64) {
	logrus.Infoln("waiting for check user process to start...")
	t := time.NewTimer(time.Second * time.Duration(waitTime))
	<-t.C

	logrus.Infoln("begin to do user's expiration check process")
	p.checkAndRemoveUser()

	logrus.Infoln("set ticker for interval check")
	ticker := time.NewTicker(time.Second * time.Duration(intervalTime))

	go func() {
		for t := range ticker.C {
			logrus.Debugln("ticker ticked: ", t)
			p.checkAndRemoveUser()
		}
	}()
}

func (p *UserService) checkAndRemoveUser() {
	if !common.IsFirstNodeInZK() {
		logrus.Infoln("current node is not first node in zk, will skip inactived user clean process")
		return
	}

	selector := make(bson.M)
	selector["state"] = 0

	queryStruct := dao.QueryStruct{
		CollectionName: p.userCollectionName,
		Selector:       selector,
		Skip:           0,
		Limit:          0,
		Sort:           ""}

	// jsonDocuments := &[]interface{}{}
	users := []entity.User{}
	_, err := dao.HandleQueryAll(&users, queryStruct)

	if err != nil {
		logrus.Errorln("query user by state error %v", err)
		return
	}

	// jsonarray := jsonDocuments
	currenttime := time.Now()
	location := time.Now().Location()
	for i := 0; i < len(users); i++ {
		// record := jsonarray[i].(map[string]interface{})
		record := users[i]

		createtime := record.TimeCreate
		createTime, err := time.ParseInLocation(f_rfc3339, createtime, location)
		if err != nil {
			logrus.Warnln("convert create time to date format error %v", err)
			continue
		}

		dur, _ := time.ParseDuration("+72h")
		expireTime := createTime.Add(dur)
		//unactive users will be remove after 3 days
		if expireTime.Before(currenttime) {
			id := record.ObjectId.Hex()
			logrus.Debugln("delete expired user from user, id:", id)
			p.deleteUserById(id)

			tenantname := record.Tenantname
			logrus.Debugln("delete expired user from tenant, name:", tenantname)
			GetTenantService().deleteTenantByName(tenantname)

		}
	}
}

func (p *UserService) deleteUserById(userId string) (err error) {
	if !bson.IsObjectIdHex(userId) {
		logrus.Errorln("invalid object id for deleteUserById: ", userId)
		err = errors.New("invalid object id for deleteUserById")
		return
	}

	selector := bson.M{}
	selector["_id"] = bson.ObjectIdHex(userId)

	err = dao.HandleDelete(p.userCollectionName, true, selector)
	return
}

func (p *UserService) UserRegistry(userParam common.UserParam) (errorCode string, userId string, err error) {

	//get max user registration
	maxUser := common.UTIL.Props.GetInt("max_user", -1)
	ok, err := p.checkMaxUserCount(maxUser)
	if err != nil {
		return common.COMMON_ERROR_INTERNAL, "", err
	}
	if !ok {
		return USER_ERROR_EXCEED, "", errors.New("The registered user has execced maximum limitation!")
	}

	email := userParam.Email
	password := userParam.Password
	confirmPassword := userParam.ConfirmPassword
	alias := userParam.Alias
	address := userParam.Address
	company := userParam.Company
	phonenum := userParam.Phonenum
	infoSource := userParam.InfoSource

	if !strings.EqualFold(password, confirmPassword) {
		logrus.Errorln("inconsistence password!")
		return USER_ERROR_REG, "", errors.New("Inconsistent password")
	}

	count, err := p.getUserByAlias(alias)
	if err != nil {
		logrus.Errorln("get user by alias error %v", err)
		return USER_ERROR_REG, "", err
	}
	if count != 0 {
		logrus.Errorln("the alias has already been used, please specify another one!")
		return USER_ERROR_DUP_NAMESPACE, "", errors.New("The alias has already been registered, please specified another one!")
	}

	_, err = GetTenantService().getTenantByName(email)
	if err == nil {
		logrus.Errorln("user already exist!")
		return USER_ERROR_EXIST, "", errors.New("The email has already been registered, please specified another one!")
	}

	encryPassword := common.HashString(password)

	_, err = GetTenantService().createAndInsertTenant(email, email)
	if err != nil {
		logrus.Errorf("create and insert new tenant error,  err is %v", err)
		return USER_ERROR_REG, "", err
	}

	userId, err = p.createAndInsertUser(email, alias, encryPassword, email, 1, email, admin_role, address, company, phonenum, infoSource)
	if err != nil {
		logrus.Errorf("create and insert new user error,  err is %v", err)
		return USER_ERROR_REG, "", err
	}

	return "", userId, nil
}

func (p *UserService) UserCreate(token string, username string, password string, confirmpassword string, alias string, tenantname string) (userId string, errorCode string, err error) {
	if len(username) == 0 || len(password) == 0 || len(confirmpassword) == 0 || len(tenantname) == 0 || len(alias) == 0 {
		logrus.Error("invalid parameter for user create!")
		return "", common.COMMON_ERROR_INVALIDATE, errors.New("invalid parameter! parameter should not be null")
	}
	code, err := GetTokenService().TokenValidate(token)
	if err != nil {
		return "", code, err
	}

	if authorized := GetAuthService().Authorize("create_user", token, "", p.userCollectionName); !authorized {
		logrus.Errorln("required opertion is not allowed!")
		return "", common.COMMON_ERROR_UNAUTHORIZED, errors.New("Required opertion is not authorized!")
	}

	if !strings.EqualFold(password, confirmpassword) {
		logrus.Errorln("inconsistence password!")
		return "", USER_ERROR_WRONGPW, errors.New("inconsistence password!")
	}

	count, err := p.getUserByAlias(alias)
	if err != nil {
		logrus.Errorln("the user by alias error %v", err)
		return "", USER_ERROR_CREATE, err
	}
	if count != 0 {
		logrus.Errorln("the alias has already been used, please specify another one!")
		return "", USER_ERROR_DUP_NAMESPACE, errors.New("The alias has already been registered, please specified another one!")

	}

	_, err = p.getUserByName(username)
	if err == nil {
		logrus.Errorln("user already exist!")
		return "", USER_ERROR_EXIST, errors.New("The email has already been registered, please specified another one!")
	}

	encryPassword := common.HashString(password)

	userId, err = p.createAndInsertUser(username, alias, encryPassword, tenantname, 1, username, common_role, "", "", "", "")
	if err != nil {
		logrus.Errorf("create and insert new user error,  err is %v", err)
		return "", USER_ERROR_CREATE, err
	}

	return userId, "", nil
}

func (p *UserService) UserReactive(userId string) (errorCode string, err error) {
	if len(userId) <= 0 {
		logrus.Error("userid should not be null")
		return common.COMMON_ERROR_INVALIDATE, errors.New("user id shoule not be null")
	}

	user, err := p.GetUserByUserId(userId)
	if err != nil {
		logrus.Errorln("user does exist %v", err)
		return USER_ERROR_ACTIVE, err
	}

	if user.State == 1 {
		logrus.Infoln("user has already been actived!")
		return "", nil
	}

	sendEmailForRegistry(user)

	return "", nil
}

func (p *UserService) GetUserByUserId(userId string) (user *entity.User, err error) {
	if !bson.IsObjectIdHex(userId) {
		logrus.Errorln("invalid object id for getUseerById: ", userId)
		err = errors.New("invalid object id for getUserById")
		return nil, err
	}
	selector := bson.M{}
	selector["_id"] = bson.ObjectIdHex(userId)

	user = new(entity.User)
	queryStruct := dao.QueryStruct{
		CollectionName: p.userCollectionName,
		Selector:       selector,
		Skip:           0,
		Limit:          0,
		Sort:           ""}

	err = dao.HandleQueryOne(user, queryStruct)

	if err != nil {
		logrus.Warnln("failed to get user by id %v", err)
		return
	}

	return
}

func (p *UserService) UserLogin(username string, password string) (errorCode string, login *entity.LoginResponse, err error) {
	currentUser, err := p.getUserByName(username)
	if err != nil {
		return USER_ERROR_NOEXIST, nil, err
	}

	if currentUser.State == 0 {
		logrus.Errorln("user is not actived!")
		return USER_ERROR_INACTIVE_USER, nil, errors.New("User is not actived!")
	}

	encryPassword := common.HashString(password)
	if !strings.EqualFold(encryPassword, currentUser.Password) {
		logrus.Errorln("invalid password!")
		return USER_ERROR_WRONGPW, nil, errors.New("Invalid password!")

	}

	tenantname := currentUser.Tenantname
	token, err := GetTokenService().checkAndGenerateToken(username, password, tenantname, true)
	if err != nil {
		logrus.Errorf("failed to generate token, error is %s", err)
		return USER_ERROR_LOGIN, nil, err
	}

	var loginRes *entity.LoginResponse
	loginRes = new(entity.LoginResponse)
	loginRes.Id = token
	loginRes.Alias = currentUser.Alias
	loginRes.Rolename = currentUser.Rolename
	loginRes.UserId = currentUser.ObjectId.Hex()

	return "", loginRes, nil
}

func (p *UserService) UserUpdate(token string, newuser entity.User, userId string) (created bool, id string, errorCode string, err error) {
	code, err := GetTokenService().TokenValidate(token)
	if err != nil {
		return false, userId, code, err
	}

	if authorized := GetAuthService().Authorize("update_user", token, userId, p.userCollectionName); !authorized {
		logrus.Errorln("required opertion is not allowed!")
		return false, userId, common.COMMON_ERROR_UNAUTHORIZED, errors.New("Required opertion is not authorized!")
	}

	// doc["time_update"] = common.GetCurrentTime()

	if !bson.IsObjectIdHex(userId) {
		logrus.Errorf("invalid user id format for user update %v", userId)
		return false, "", common.COMMON_ERROR_INVALIDATE, errors.New("Invalid object Id for user update")
	}

	selector := bson.M{}
	selector["_id"] = bson.ObjectIdHex(userId)

	queryStruct := dao.QueryStruct{
		CollectionName: p.userCollectionName,
		Selector:       selector,
		Skip:           0,
		Limit:          0,
		Sort:           ""}

	user := new(entity.User)
	err = dao.HandleQueryOne(user, queryStruct)
	if err != nil {
		logrus.Errorf("get user by id error %v", err)
		return false, "", USER_ERROR_UPDATE, err
	}

	//set new values
	if len(newuser.Address) > 0 {
		user.Address = newuser.Address
	}
	if len(newuser.Company) > 0 {
		user.Company = newuser.Company
	}
	if len(newuser.PhoneNum) > 0 {
		user.PhoneNum = newuser.PhoneNum
	}

	user.TimeUpdate = common.GetCurrentTime()

	created, err = dao.HandleUpdateOne(user, queryStruct)
	return created, userId, "", nil
}

func (p *UserService) UserDelete(token string, userId string) (errorCode string, err error) {
	if !bson.IsObjectIdHex(userId) {
		logrus.Errorln("invalid object id for UserDelete: ", userId)
		err = errors.New("invalid object id for UserDelete")
		return USER_ERROR_DELETE, err
	}

	code, err := GetTokenService().TokenValidate(token)
	if err != nil {
		return code, err
	}

	if authorized := GetAuthService().Authorize("delete_user", token, userId, p.userCollectionName); !authorized {
		logrus.Errorln("required opertion is not allowed!")
		return common.COMMON_ERROR_UNAUTHORIZED, errors.New("Required opertion is not authorized!")
	}

	selector := bson.M{}
	selector["_id"] = bson.ObjectIdHex(userId)

	err = dao.HandleDelete(p.userCollectionName, true, selector)
	if err != nil {
		logrus.Warnln("delete user error %v", err)
		return USER_ERROR_DELETE, err
	}

	return "", nil
}

func (p *UserService) UserActive(code string, id string) (errorCode string, err error) {
	if len(code) <= 0 || len(id) <= 0 {
		logrus.Errorln("paramter should be null for user active ")
		return USER_ERROR_ACTIVE, errors.New("parameter should be null for user active")
	}

	activeCode, expire, err := parseActiveCodeAndExpireTime(code)
	if err != nil {
		return common.COMMON_ERROR_INVALIDATE, err
	}

	user, err := p.GetUserByUserId(id)
	if err != nil {
		logrus.Errorln("active user error %v", err)
		return USER_ERROR_ACTIVE, err
	}

	currentTime := time.Now().Unix()
	expireTime, err := strconv.ParseInt(expire, 10, 0)
	if err != nil {
		logrus.Errorln("convert expire time to string error, expire time is :", expire)
		return USER_ERROR_ACTIVE, err

	}

	if currentTime >= expireTime {
		logrus.Errorln("expired active code!")
		return USER_ERROR_ACTIVE, errors.New("Active code is expired!")

	}

	if user.ActiveCode == activeCode {
		user.State = 1
		user.TimeUpdate = common.GetCurrentTime()

		selector := bson.M{}
		selector["_id"] = bson.ObjectIdHex(id)

		queryStruct := dao.QueryStruct{
			CollectionName: p.userCollectionName,
			Selector:       selector,
			Skip:           0,
			Limit:          0,
			Sort:           ""}

		_, err = dao.HandleUpdateOne(user, queryStruct)
		if err != nil {
			logrus.Errorln("active user error %v", err)
			return USER_ERROR_ACTIVE, err
		}

		return "", nil

	} else {
		logrus.Errorln("invalid active code!")
		return USER_ERROR_ACTIVE, errors.New("Invalid active code")
	}
}

func (p *UserService) UserChangePassword(token string, id string, password string, newpassword string, confirm_newpassword string) (created bool, errorCode string, err error) {
	code, err := GetTokenService().TokenValidate(token)
	if err != nil {
		return false, code, err
	}

	if authorized := GetAuthService().Authorize("change_password", token, id, p.userCollectionName); !authorized {
		logrus.Errorln("required opertion is not allowed!")
		return false, common.COMMON_ERROR_UNAUTHORIZED, errors.New("Required opertion is not authorized!")
	}

	user, err := p.GetUserByUserId(id)
	if err != nil {
		logrus.Errorln("user does exist %v", err)
		return false, common.COMMON_ERROR_INTERNAL, errors.New("User does not exist!")
	}

	pwdEncry := common.HashString(password)
	if !strings.EqualFold(pwdEncry, user.Password) {
		logrus.Errorln("incorrect password!")
		return false, USER_ERROR_WRONGPW, errors.New("Incorrect password!")
	}

	if !strings.EqualFold(newpassword, confirm_newpassword) {
		logrus.Errorln("inconsistence new password!")
		return false, USER_ERROR_WRONGPW, errors.New("Inconsistent new password!")
	}

	newpasswordEncry := common.HashString(newpassword)
	user.Password = newpasswordEncry

	user.TimeUpdate = common.GetCurrentTime()

	// userDoc := ConvertToBson(*user)
	selector := bson.M{}
	selector["_id"] = bson.ObjectIdHex(id)

	queryStruct := dao.QueryStruct{
		CollectionName: p.userCollectionName,
		Selector:       selector,
		Skip:           0,
		Limit:          0,
		Sort:           ""}

	created, err = dao.HandleUpdateOne(user, queryStruct)
	if err != nil {
		logrus.Error("update user password error! %v", err)
		return created, USER_ERROR_UPDATE, err
	}

	return created, "", nil
}

func (p *UserService) UserList(token string, limit int, skip int, sort string) (ret []entity.User, count int, errorCode string, err error) {
	code, err := GetTokenService().TokenValidate(token)
	if err != nil {
		return nil, 0, code, err
	}

	query, err := GetAuthService().BuildQueryByAuth("list_users", token)
	if err != nil {
		logrus.Error("auth failed during query all user: %v", err)
		return nil, 0, USER_ERROR_GET, err
	}

	result := []entity.User{}
	queryStruct := dao.QueryStruct{
		CollectionName: p.userCollectionName,
		Selector:       query,
		Skip:           skip,
		Limit:          limit,
		Sort:           sort}
	count, err = dao.HandleQueryAll(&result, queryStruct)

	return result, count, "", err
}

func (p *UserService) UserDetail(token string, userId string) (ret interface{}, errorCode string, err error) {
	code, err := GetTokenService().TokenValidate(token)
	if err != nil {
		return nil, code, err
	}

	if authorized := GetAuthService().Authorize("get_user", token, userId, p.userCollectionName); !authorized {
		logrus.Errorln("required opertion is not allowed!")
		return nil, common.COMMON_ERROR_UNAUTHORIZED, errors.New("Required opertion is not authorized!")
	}

	selector := bson.M{}
	selector["_id"] = bson.ObjectIdHex(userId)

	ret = new(entity.User)
	queryStruct := dao.QueryStruct{
		CollectionName: p.userCollectionName,
		Selector:       selector,
		Skip:           0,
		Limit:          0,
		Sort:           ""}

	err = dao.HandleQueryOne(ret, queryStruct)
	logrus.Errorln(ret)
	return
}

func parseActiveCodeAndExpireTime(code string) (acviceCode string, expireTime string, err error) {
	if len(code) == 0 {
		logrus.Warnln("the active Code is null!")
		err = errors.New("Active code is null")
		return
	}

	//if exist "+"in code, it will because space after getting from url
	code = strings.Replace(code, " ", "+", -1)

	input, err := base64.StdEncoding.DecodeString(code)
	if err != nil {
		logrus.Warnln("decode active code error %v", err)
		err = errors.New("Invalid active code! ")
		return
	}
	// activeCode + "," + expireTime
	origData, err := common.DesDecrypt(input, []byte(key))
	if err != nil {
		logrus.Warnln("decrypt error %v", err)
		err = errors.New("invalid active code! ")
		return
	}

	result := string(origData)

	values := strings.Split(result, ",")
	if len(values) < 2 {
		logrus.Warnln("invalid parametr: ", values)
		err = errors.New("Active user failed: invalid active code")
		return
	}

	return values[0], values[1], nil
}

func (p *UserService) checkMaxUserCount(max int) (bool, error) {
	if max < 0 {
		return true, nil
	}

	doc := []interface{}{}
	queryStruct := dao.QueryStruct{
		CollectionName: p.userCollectionName,
		Selector:       bson.M{},
		Skip:           0,
		Limit:          0,
		Sort:           ""}
	count, err := dao.HandleQueryAll(&doc, queryStruct)
	if err != nil {
		logrus.Errorf("handle query err is %v", err)
		return false, err
	}

	if count >= max {
		logrus.Infoln("the registered user has execced the max limitation")
		return false, nil
	} else {
		return true, nil
	}
}

func (p *UserService) createAndInsertUser(userName string, alias string, password string, email string, state int, tenanName string, roleName string, address string, company string, phonenum string, source string) (userId string, err error) {
	// var jsondocument interface{}
	currentUser, erro := p.getUserByName(userName)
	if erro == nil {
		logrus.Infoln("user already exist! username:", userName)
		userId = currentUser.ObjectId.Hex()
		return
	}

	currentTime := common.GetCurrentTime()
	user := new(entity.User)
	user.ObjectId = bson.NewObjectId()
	user.Username = userName
	user.Alias = alias
	user.Password = password
	user.Tenantname = tenanName
	user.Rolename = roleName
	user.Email = email
	user.State = state
	user.ActiveCode = common.GenerateActiveCode()
	user.Address = address
	user.Company = company
	user.PhoneNum = phonenum
	user.Source = source
	user.TimeCreate = currentTime
	user.TimeUpdate = currentTime

	err = dao.HandleInsert(p.userCollectionName, user)
	if err != nil {
		logrus.Warnln("create user error %v", err)
		return
	}
	userId = user.ObjectId.Hex()

	if state != 1 {
		sendEmailForRegistry(user)
	}

	return
}

func (p *UserService) getUserByName(username string) (user *entity.User, err error) {
	query := strings.Join([]string{"{\"email\": \"", username, "\"}"}, "")

	selector := make(bson.M)
	err = json.Unmarshal([]byte(query), &selector)
	if err != nil {
		return
	}
	selector, err = mejson.Unmarshal(selector)
	if err != nil {
		return
	}

	user = new(entity.User)
	queryStruct := dao.QueryStruct{
		CollectionName: p.userCollectionName,
		Selector:       selector,
		Skip:           0,
		Limit:          0,
		Sort:           ""}

	err = dao.HandleQueryOne(user, queryStruct)
	// _, _, jsonUser, err = p.Dao.HandleQuery(USERCOLLECTION, selector, true, fields, skip, limit, sort, extended_json)

	return
}

func (p *UserService) getUserByAlias(alias string) (count int, err error) {
	selector := bson.M{}
	selector["alias"] = alias

	// field := bson.M{}
	// field["id"] = 1

	documents := []interface{}{}
	queryStruct := dao.QueryStruct{
		CollectionName: p.userCollectionName,
		Selector:       selector,
		Skip:           0,
		Limit:          0,
		Sort:           ""}

	count, err = dao.HandleQueryAll(&documents, queryStruct)

	return
}

func sendEmailForRegistry(user *entity.User) {
	logrus.Infoln("send mail for new user registry")
	subject := "[Linker Cloud Platform] Account " + user.Username + " Email Verification"

	body := strings.Replace(emailBody, "NEWUSER", user.Username, -1)

	hostname, err := os.Hostname()
	if err != nil {
		logrus.Warnln("get hostname err is %+v", err)
		hostname = "localhost"
	}

	host := common.UTIL.Props.GetString("portalUI.host", hostname)
	port := common.UTIL.Props.GetString("portalUI.port", "8080")
	portalUrl := host + ":" + port
	code := buildCode(user.ActiveCode)
	url := strings.Join([]string{"http://", portalUrl, "/user/active?", "uid=", user.ObjectId.Hex(), "&activeCode=", code}, "")

	body = strings.Replace(body, "NEWURL", url, -1)

	emailHost := common.UTIL.Props.GetString("email.host", "")
	emailUsername := common.UTIL.Props.GetString("email.username", "")
	emailPasswd := common.UTIL.Props.GetString("email.password", "")

	go common.SendMail(emailHost, emailUsername, emailPasswd, user.Email, subject, body)
}

func buildCode(activeCode string) string {
	t := time.Now().Unix()
	t += int64(activeCode_expire)
	expireTime := strconv.FormatInt(t, 10)

	code := activeCode + "," + expireTime
	result, err := common.DesEncrypt([]byte(code), []byte(key))
	if err != nil {
		logrus.Warnln("encrypt code error %v", err)
		return code
	}

	return base64.StdEncoding.EncodeToString(result)
}
