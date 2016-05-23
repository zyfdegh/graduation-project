package documents

import (
	"encoding/json"
	"github.com/Sirupsen/logrus"
	"github.com/emicklei/go-restful"
	"linkernetworks.com/linker_common_lib/persistence/entity"
	"linkernetworks.com/linker_common_lib/rest/response"
	"linkernetworks.com/linker_controller/services"
	"sort"
	"strconv"
	"strings"
)

func (p Resource) UserAccountWebService() *restful.WebService {
	ws := new(restful.WebService)
	ws.Path("/v1/userAccounts")
	ws.Consumes("*/*")
	ws.Produces(restful.MIME_JSON)

	// id := ws.PathParameter(ParamID, "Storage identifier of service group")
	// paramID := "{" + ParamID + "}"

	ws.Route(ws.GET("/").To(p.UserAccountListHandler).
		Doc("Returns all useraccount items").
		Operation("UserAccountsListHandler").
		Param(ws.HeaderParameter("X-Auth-Token", "A valid authentication token")).
		Param(ws.QueryParameter("count", "Count total items and return the result in X-Object-Count header").DataType("boolean")).
		Param(ws.QueryParameter("from", "from date of billing record. YYYY-MM-DD")).
		Param(ws.QueryParameter("to", "to date of billing record. YYYY-MM-DD")).
		Param(ws.QueryParameter("type", "billing record type, default=All")).
		Param(ws.QueryParameter("skip", "Number of items to skip in the result set, default=0")).
		Param(ws.QueryParameter("limit", "Maximum number of items in the result set, default=0")).
		Param(ws.QueryParameter("sort", "Comma separated list of field names to sort")))

	ws.Route(ws.POST("/").To(p.UserRechargeAccountCreateHandler).
		Doc("Store a user account").
		Operation("UserRechargeAccountCreateHandler").
		Param(ws.HeaderParameter("X-Auth-Token", "A valid authentication token")).
		Param(ws.BodyParameter("body", "User account body in json format,for example {\"price\":\"...\", \"transaction_type\":\"...\", \"transaction_desc\":\"...\"}")))

	ws.Route(ws.PUT("/").To(p.UserAccountNotifyRechargeHandler).
		Doc("update user account status").
		Operation("UserAccountNotifyRechargeHandler").
		Param(ws.QueryParameter("userId", "User id of user account")).
		Param(ws.QueryParameter("payNo", "User account object id")).
		Param(ws.QueryParameter("amount", "Transaction amount.  Integer type")).
		Param(ws.QueryParameter("payChannelName", "the channel name of payment")).
		Param(ws.QueryParameter("tradeName", "Transaction name")).
		Param(ws.QueryParameter("tradeDescription", "Transaction description")).
		Param(ws.QueryParameter("status", "Transaction status")).
		Param(ws.QueryParameter("signature", "The signature of request operations")))

	return ws
}

// UserAccountListHandler parses the http request and returns the useraccount items.
// Usage :
//		GET /v1/userAccounts
// If successful,response code will be set to 201.
func (p *Resource) UserAccountListHandler(req *restful.Request, resp *restful.Response) {
	logrus.Infof("UserAccountListHandle is called!")
	token := req.HeaderParameter("X-Auth-Token")
	skip := queryIntParam(req, "skip", 0)
	limit := queryIntParam(req, "limit", 0)
	sort := req.QueryParameter("sort")

	from := req.QueryParameter("from")
	to := req.QueryParameter("to")
	transactionType := req.QueryParameter("type")

	useraccounts, total, errorCode, err := services.GetUserAccountService().ListUserAccount(token, from, to, transactionType, skip, limit, sort)
	if err != nil {
		response.WriteStatusError(errorCode, err, resp)
		return
	}

	res := response.QueryStruct{Success: true, Data: useraccounts}
	if c, _ := strconv.ParseBool(req.QueryParameter("count")); c {
		res.Count = total
		resp.AddHeader("X-Object-Count", strconv.Itoa(total))
	}
	resp.WriteEntity(res)
	return
}

func (p *Resource) UserRechargeAccountCreateHandler(req *restful.Request, resp *restful.Response) {
	logrus.Infoln("UserRechargeAccountCreateHandler is called!")

	token := req.HeaderParameter("X-Auth-Token")

	ua := entity.UserAccount{}
	// Populate the user data
	err := json.NewDecoder(req.Request.Body).Decode(&ua)
	if err != nil {
		logrus.Errorf("convert body to user account failed, error is %v", err)
		response.WriteStatusError(services.COMMON_ERROR_INVALIDATE, err, resp)
		return
	}

	newua, errorCode, err := services.GetUserAccountService().CreateRechargeUserAccount(token, ua)
	if err != nil {
		response.WriteStatusError(errorCode, err, resp)
		return
	}

	res := response.QueryStruct{Success: true, Data: newua.ObjectId.Hex()}
	resp.WriteEntity(res)
	return
}

func (p *Resource) UserAccountNotifyRechargeHandler(req *restful.Request, resp *restful.Response) {
	logrus.Infoln("UserAccountNotifyRechargeHandler is called")

	payNo := req.QueryParameter("payNo")
	status := req.QueryParameter("status")
	signature := req.QueryParameter("signature")

	keylist := []string{"userId", "payNo", "amount", "payChannelName", "tradeName", "tradeDescription", "status"}
	sort.Strings(keylist)
	ret := []string{}
	for i := 0; i < len(keylist); i++ {
		key := keylist[i]
		value := req.QueryParameter(key)
		ret = append(ret, key+"="+value+"&")
	}

	v := strings.Join(ret, "")
	rs := []rune(v)
	reqkey := string(rs[0 : len(v)-1])
	checkKey := "linker_cloud_pay2015" + reqkey

	logrus.Infof("the final checkkey[%v] and sigature [%v]:", checkKey, signature)

	err := services.GetUserAccountService().NotifyRechargeStatusById(status, payNo, checkKey, signature)
	if err != nil {
		response.WriteStatusError(services.COMMON_ERROR_INTERNAL, err, resp)
		return
	}

	response.WriteSuccess(resp)

}
