package services

import (
	"encoding/json"
	"errors"
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

var f_date = "2006-01-02"
var f_datetime = "2006-01-02 15:04:05"

var tokenService *TokenService = nil
var tokenOnce sync.Once

var TOKEN_ERROR_CREATE = "E10051"
var TOKEN_ERROR_NOEXIST = "E10050"
var TOKEN_ERROR_EXPIRE = "E10052"
var TOKEN_ERROR_GET = "E10053"

type Ids struct {
	Tokenid  string `json:"tokenid"`
	Userid   string `json:"userid"`
	Email    string `json:"email"`
	Tenantid string `json:"tenantid"`
	Role     string `json:"role"`
}

type TokenId struct {
	Id string `json:"id"`
}

type TokenService struct {
	collectionName string
}

func GetTokenService() *TokenService {
	tokenOnce.Do(func() {
		tokenService = &TokenService{"token"}

		tokenService.initialize()
	})

	return tokenService
}

func (p *TokenService) initialize() {
	logrus.Infoln("initialize token check and clean process")
	interval := common.UTIL.Props.GetString("token_check_interval", "259200")
	if len(interval) <= 0 {
		interval = "259200"
	}
	exec := common.UTIL.Props.GetString("token_check_time", "02:00:00")
	if len(exec) <= 0 {
		exec = "02:00:00"
	}

	formatdate := time.Now().Format(f_date)

	newexec := formatdate + " " + exec
	execTime, err := time.ParseInLocation(f_datetime, newexec, time.Now().Location())
	if err != nil {
		logrus.Warnln("failed to parse exec check time: ", newexec)
		execTime, _ = time.ParseInLocation(f_datetime, formatdate+" 02:00:00", time.Now().Location())
	}

	intervalInt, err := strconv.ParseInt(interval, 10, 64)
	if err != nil {
		logrus.Warnln("failed to parse intervalTime: ", interval)
		intervalInt, _ = strconv.ParseInt("259200", 10, 64)
	}

	waitTime := common.GetWaitTime(execTime)

	go p.startTokenTimer(waitTime, intervalInt)

}

func (p *TokenService) startTokenTimer(waitTime int64, intervalTime int64) {
	logrus.Infoln("waiting for check token process to start...")
	t := time.NewTimer(time.Second * time.Duration(waitTime))
	<-t.C

	logrus.Infoln("begin to do token's expiration check process")
	p.checkAndRemoveToken()

	logrus.Infoln("set ticker for interval check")
	ticker := time.NewTicker(time.Second * time.Duration(intervalTime))
	go p.run(ticker)
}

func (p *TokenService) checkAndRemoveToken() {
	if !common.IsFirstNodeInZK() {
		logrus.Infoln("current node is not first node in zk, will skip expired token clean process")
		return
	}

	currentTime := time.Now().Unix()
	queryMap := make(map[string]interface{})
	valueMap := make(map[string]int64)

	valueMap["$lte"] = currentTime
	queryMap["expiretime"] = valueMap

	bytesValue, err := json.Marshal(queryMap)
	if err != nil {
		logrus.Warnln("marshal query object error %v", err)
		return
	}

	selector := make(bson.M)
	err = json.Unmarshal(bytesValue, &selector)
	if err != nil {
		logrus.Warnln("unmarshal querymap error %v", err)
		return
	}
	selector, err = mejson.Unmarshal(selector)
	if err != nil {
		logrus.Warnln("unmarshal querymap error %v", err)
		return
	}

	logrus.Debugln("delete expired tokens from database")
	dao.HandleDelete(p.collectionName, false, selector)
}

func (p *TokenService) run(ticker *time.Ticker) {
	for t := range ticker.C {
		logrus.Debugln("ticker ticked: ", t)
		p.checkAndRemoveToken()
	}
}

func (p *TokenService) TokenCreate(email string, password string, tenantname string) (ret TokenId, errorCode string, err error) {
	if len(email) == 0 || len(password) == 0 || len(tenantname) == 0 {
		logrus.Errorln("parameter can not be null!")
		return ret, TOKEN_ERROR_CREATE, errors.New("invalid parameter!")
	}

	token, err := p.checkAndGenerateToken(email, password, tenantname, true)
	if err != nil {
		logrus.Errorf("failed to generate token, error is %s", err)
		return ret, TOKEN_ERROR_CREATE, err
	}

	tokenId := TokenId{Id: token}

	return tokenId, "", nil
}

func (p *TokenService) TokenReGenerate(token string, userId string, tenantId string) (ret TokenId, errorCode string, err error) {
	if len(userId) == 0 || len(tenantId) == 0 {
		logrus.Errorf("user and tenant id can not be null!")
		return ret, "E12002", errors.New("invalid parameter!user and tenant id can not be null")
	}

	code, err := p.TokenValidate(token)
	if err != nil {
		return ret, code, err
	}

	if authorized := GetAuthService().Authorize("regenerate_token", token, "", p.collectionName); !authorized {
		logrus.Errorln("required opertion is not allowed!")
		return ret, "E12004", errors.New("required opertion is not authorized!")
	}

	user_name, tenant_name, err := getNamesById(userId, tenantId)
	if err != nil {
		logrus.Errorln("failed to get user and tenant by id, error is %v", err)
		return ret, TOKEN_ERROR_CREATE, err
	}

	newtoken, err := p.checkAndGenerateToken(user_name, "", tenant_name, false)
	if err != nil {
		logrus.Errorf("failed to generate token, error is %s", err)
		return ret, TOKEN_ERROR_CREATE, err
	}

	tokenId := TokenId{Id: newtoken}

	return tokenId, "", nil
}

func (p *TokenService) TokenDetail(token string, id string) (currentToken *entity.Token, errorCode string, err error) {
	code, err := GetTokenService().TokenValidate(token)
	if err != nil {
		return nil, code, err
	}

	if authorized := GetAuthService().Authorize("get_token", token, id, p.collectionName); !authorized {
		logrus.Errorln("required opertion is not allowed!")
		return nil, common.COMMON_ERROR_UNAUTHORIZED, errors.New("Required opertion is not authorized!")
	}

	currentToken, err = p.GetTokenById(id)
	if err != nil {
		logrus.Errorln("get token by id err %v", err)
		errorCode = TOKEN_ERROR_GET
		return
	}

	return

}

func getNamesById(userId string, tenantId string) (string, string, error) {
	user, err := GetUserService().GetUserByUserId(userId)
	if err != nil {
		return "", "", err
	}

	tenant, err := GetTenantService().GetTenantByTenantId(tenantId)
	if err != nil {
		return "", "", err
	}

	return user.Email, tenant.Tenantname, nil
}

func (p *TokenService) TokenValidate(token string) (errorCode string, err error) {
	if len(token) <= 0 {
		logrus.Errorf("no token for specific operation")
		return "E12002", errors.New("no token for operation")
	}
	currentToken, err := p.GetTokenById(token)
	if err != nil {
		return TOKEN_ERROR_NOEXIST, err
	}

	//check expire time
	expireTime := currentToken.Expire
	currentTime := float64(time.Now().Unix())

	if currentTime >= expireTime {
		logrus.Infoln("token expire!")
		return TOKEN_ERROR_EXPIRE, errors.New("token expire!")
	}

	return "", nil
}

func (p *TokenService) GetTokenById(token string) (currentToken *entity.Token, err error) {
	validId := bson.IsObjectIdHex(token)
	if !validId {
		return nil, errors.New("invalid token!")
	}

	selector := bson.M{}
	selector["_id"] = bson.ObjectIdHex(token)

	currentToken = new(entity.Token)
	queryStruct := dao.QueryStruct{
		CollectionName: p.collectionName,
		Selector:       selector,
		Skip:           0,
		Limit:          0,
		Sort:           ""}

	err = dao.HandleQueryOne(currentToken, queryStruct)
	if err != nil {
		logrus.Infoln("token does not exist! %v", err)
		return nil, err
	}

	return
}

func (p *TokenService) GetIdsFromToken(tokenid string) (ret Ids, errorCode string, err error) {
	code, err := p.TokenValidate(tokenid)
	if err != nil {
		return ret, code, err
	}

	token, err := p.GetTokenById(tokenid)
	if err != nil {
		logrus.Errorln("failed to get token, error: %v", err)
		return ret, "E12003", err
	}

	ids := Ids{
		Tokenid:  tokenid,
		Userid:   token.User.Id,
		Email:    token.User.Username,
		Tenantid: token.Tenant.Id,
		Role:     token.Role.Rolename,
	}
	return ids, "", nil
}

// CheckAndGenerateToken check the token and creat the new token.
func (p *TokenService) checkAndGenerateToken(user string, passwd string, tenant string, needPasswd bool) (result string, err error) {
	//get user by name
	userobj, err := GetUserService().getUserByName(user)
	if err != nil {
		logrus.Errorln("user does not exist! username:", user)
		return result, errors.New("user does not exist!")
	}

	//get tenant by name
	tenantobj, err := GetTenantService().getTenantByName(tenant)
	if err != nil {
		logrus.Errorln("tenant does not exist! name:", tenant)
		return result, errors.New("tenant does not exist!")
	}

	tenantid := tenantobj.ObjectId.Hex()
	tenantname := tenantobj.Tenantname

	userid := userobj.ObjectId.Hex()
	password := userobj.Password
	rolename := userobj.Rolename

	if needPasswd {
		encryPassword := common.HashString(passwd)

		if !strings.EqualFold(encryPassword, password) {
			logrus.Errorln("invalid password!")
			return result, errors.New("invalid password!")
		}
	}

	//get role
	roleobj, err := GetRoleService().getRoleByName(rolename)
	if err != nil {
		logrus.Errorln("role does not exist! rolename:", rolename)
		return result, errors.New("role does not exist!")
	}

	roleid := roleobj.ObjectId.Hex()

	time := common.UTIL.Props.GetString("expiration_time", "21600")
	newtime, err := strconv.ParseInt(strings.TrimSpace(time), 10, 64)
	if err != nil {
		logrus.Warnln("invalid expire time configured %v", err)
		newtime = int64(21600)
	}
	expireTime := common.GenerateExpireTime(newtime)
	currentTime := common.GetCurrentTime()

	userpart := entity.UserPart{Id: userid, Username: user}
	tenantpart := entity.TenantPart{Id: tenantid, Tenantname: tenantname}
	rolepart := entity.RolePart{Id: roleid, Rolename: rolename}

	objectId := bson.NewObjectId()
	newtoken := entity.Token{
		ObjectId:   objectId,
		Expire:     expireTime,
		User:       userpart,
		Tenant:     tenantpart,
		Role:       rolepart,
		TimeCreate: currentTime,
		TimeUpdate: currentTime,
	}

	err = dao.HandleInsert(p.collectionName, &newtoken)
	result = newtoken.ObjectId.Hex()

	if err != nil {
		logrus.Errorf("save token err is %v", err)
		return result, err
	}

	return
}
