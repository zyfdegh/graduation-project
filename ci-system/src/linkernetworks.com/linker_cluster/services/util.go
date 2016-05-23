package services

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"github.com/Sirupsen/logrus"
	"io/ioutil"
	"linkernetworks.com/linker_cluster/common"
	"linkernetworks.com/linker_common_lib/httpclient"
	"linkernetworks.com/linker_common_lib/persistence/entity"
	"linkernetworks.com/linker_common_lib/rest/response"
	"strings"
)

var (
	COMMON_ERROR_INVALIDATE   = "E12002"
	COMMON_ERROR_UNAUTHORIZED = "E12004"
	COMMON_ERROR_UNKNOWN      = "E12001"
	COMMON_ERROR_INTERNAL     = "E12003"
)

func getErrorFromResponse(data []byte) (errorCode string, err error) {
	var resp *response.Response
	resp = new(response.Response)
	err = json.Unmarshal(data, &resp)
	if err != nil {
		return COMMON_ERROR_INTERNAL, err
	}

	errorCode = resp.Error.Code
	err = errors.New(resp.Error.ErrorMsg)
	return
}

func TokenValidation(tokenId string) (errorCode string, err error) {
	userUrl, err := common.UTIL.ZkClient.GetUserMgmtEndpoint()
	if err != nil {
		logrus.Errorf("get userMgmt endpoint err is %v", err)
		return COMMON_ERROR_INTERNAL, err
	}
	url := strings.Join([]string{"http://", userUrl, "/v1/token/?", "token=", tokenId}, "")
	logrus.Debugln("token validation url=" + url)

	resp, err := httpclient.Http_get(url, "",
		httpclient.Header{"Content-Type", "application/json"})
	defer resp.Body.Close()
	if err != nil {
		logrus.Errorf("http get token validate error %v", err)
		return COMMON_ERROR_INTERNAL, err
	}

	data, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		logrus.Errorf("token validation failed %v", string(data))
		errorCode, err = getErrorFromResponse(data)
		return
	}

	return "", nil
}
func GetTokenById(token string) (currentToken *entity.Token, err error) {
	userUrl, err := common.UTIL.ZkClient.GetUserMgmtEndpoint()
	if err != nil {
		logrus.Errorf("get userMgmt endpoint err is %v", err)
		return nil, err
	}
	url := strings.Join([]string{"http://", userUrl, "/v1/token/", token}, "")
	logrus.Debugln("get token url=" + url)

	resp, err := httpclient.Http_get(url, "",
		httpclient.Header{"Content-Type", "application/json"}, httpclient.Header{"X-Auth-Token", token})
	defer resp.Body.Close()
	if err != nil {
		logrus.Errorf("http get token error %v", err)
		return nil, err
	}

	data, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		logrus.Errorf("get token by id failed %v", string(data))
		return nil, errors.New("get token by id failed")
	}

	currentToken = new(entity.Token)
	err = getRetFromResponse(data, currentToken)
	return
}

func copyBundle(token string, sgid string, subConstrollerUrl string) (err error) {
	logrus.Infoln("1. export bundle from primary controller")
	controllerURL, err := common.UTIL.ZkClient.GetControllerEndpoint()
	if err != nil {
		logrus.Warnf("get controller url error %v", err)
		return err
	}

	exportURL := strings.Join([]string{"http://", controllerURL, "/v1/bundle/?", "sgid=", sgid}, "")
	logrus.Debugln("export bundle url:", exportURL)

	resp, err := httpclient.Http_get(exportURL, "",
		httpclient.Header{"Content-Type", "application/json"}, httpclient.Header{"X-Auth-Token", token})
	defer resp.Body.Close()
	if err != nil {
		logrus.Errorf("http export bundle error %v", err)
		return err
	}

	data, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		logrus.Errorf("export bundle failed %v", string(data))
		return errors.New("export bundle failed")
	}

	bundle := new(entity.Bundle)
	err = getRetFromResponse(data, bundle)
	if err != nil {
		logrus.Warnf("get bundle from response error %v", err)
		return
	}

	logrus.Infoln("2. import bundle to sub controller")
	importURL := strings.Join([]string{"http://", subConstrollerUrl, "/v1/bundle/"}, "")
	logrus.Debugln("import bundle url:", importURL)

	body, err := json.Marshal(bundle)
	if err != nil {
		logrus.Errorf("marshal bundle error %v", err)
		return err
	}

	reqbody := string(body)
	logrus.Debugln("body=" + reqbody)
	resp, err = httpclient.Http_post(importURL, reqbody,
		httpclient.Header{"Content-Type", "application/json"}, httpclient.Header{"X-Auth-Token", token})
	defer resp.Body.Close()
	if err != nil {
		logrus.Errorf("http import bundle error %v", err)
		return err
	}

	data, _ = ioutil.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		logrus.Errorf("import bundle failed %v", string(data))
		return errors.New("import bundle failed")
	}

	return

}

func OrderSGO(token string, sgo entity.ServiceGroupOrder, controllerURL string) (newsgo entity.ServiceGroupOrder, err error) {
	url := strings.Join([]string{"http://", controllerURL, "/v1/serviceGroupOrders/"}, "")
	logrus.Debugln("order service group order url=" + url)

	body, err := json.Marshal(sgo)
	if err != nil {
		logrus.Errorf("marshal sgo error %v", err)
		return newsgo, err
	}

	reqbody := string(body)
	logrus.Debugln("body=" + reqbody)
	resp, err := httpclient.Http_post(url, reqbody,
		httpclient.Header{"Content-Type", "application/json"}, httpclient.Header{"X-Auth-Token", token})
	defer resp.Body.Close()
	if err != nil {
		logrus.Errorf("http order sgo error %v", err)
		return newsgo, err
	}

	data, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		logrus.Errorf("create sg order failed %v", string(data))
		return newsgo, errors.New("create service group order failed")
	}

	newsgo = entity.ServiceGroupOrder{}
	err = getRetFromResponse(data, &newsgo)
	if err != nil {
		logrus.Warnf("get sgo from response error %v", err)
		return
	}

	return
}

func QuerySGO(token string, sgoId string, controllerURL string) (newsgo entity.ServiceGroupOrder, err error) {
	url := strings.Join([]string{"http://", controllerURL, "/v1/serviceGroupOrders/", sgoId}, "")
	logrus.Debugln("get service group order url=" + url)

	resp, err := httpclient.Http_get(url, "",
		httpclient.Header{"Content-Type", "application/json"}, httpclient.Header{"X-Auth-Token", token})
	defer resp.Body.Close()
	if err != nil {
		logrus.Errorf("http get sgo error %v", err)
		return newsgo, err
	}

	data, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		logrus.Errorf("get sg order failed %v", string(data))
		return newsgo, errors.New("get service group order failed")
	}

	newsgo = entity.ServiceGroupOrder{}
	err = getRetFromResponse(data, &newsgo)
	if err != nil {
		logrus.Warnf("get sgo from response error %v", err)
		return
	}

	return
}

func QueryAuthOrder(token string, sgoId string, controllerURL string) (operations map[string]int, err error) {
	url := strings.Join([]string{"http://", controllerURL, "/v1/serviceGroupOrders/operations/", sgoId}, "")
	logrus.Debugln("auth operations service group order url=" + url)

	resp, err := httpclient.Http_get(url, "",
		httpclient.Header{"Content-Type", "application/json"}, httpclient.Header{"X-Auth-Token", token})
	defer resp.Body.Close()
	if err != nil {
		logrus.Errorf("http get auth operations sgo error %v", err)
		return operations, err
	}

	data, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		logrus.Errorf("get sg order failed %v", string(data))
		return operations, errors.New("get service group order failed")
	}

	operations = map[string]int{}
	err = getRetFromResponse(data, &operations)
	if err != nil {
		logrus.Warnf("get auth operations sgo from response error %v", err)
		return
	}

	return
}

func QueryAppInOrder(token string, sgoId string, appId string, controllerURL string) (app *entity.App, err error) {
	url := strings.Join([]string{"http://", controllerURL, "/v1/serviceGroupOrders/", sgoId, "/scaleInfo/?appId=", appId}, "")
	logrus.Debugln("query app in order url=" + url)

	resp, err := httpclient.Http_get(url, "",
		httpclient.Header{"Content-Type", "application/json"}, httpclient.Header{"X-Auth-Token", token})
	defer resp.Body.Close()
	if err != nil {
		logrus.Errorf("query app in order error %v", err)
		return nil, err
	}

	data, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		logrus.Errorf("query app in order failed %v", string(data))
		return nil, errors.New("query app in order failed")
	}

	app = new(entity.App)
	err = getRetFromResponse(data, app)
	if err != nil {
		logrus.Warnf("query app in order from response error %v", err)
		return
	}

	return
}

func ScaleAppByOrderId(token string, sgoId string, appId string, numStr string, controllerURL string) (err error) {
	url := strings.Join([]string{"http://", controllerURL, "/v1/serviceGroupOrders/", sgoId, "/scaleApp/?appId=", appId, "&num=", numStr}, "")
	logrus.Debugln("scale app in order url=" + url)

	resp, err := httpclient.Http_put(url, "",
		httpclient.Header{"Content-Type", "application/json"}, httpclient.Header{"X-Auth-Token", token})
	defer resp.Body.Close()
	if err != nil {
		logrus.Errorf("scale app in order error %v", err)
		return err
	}

	data, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		logrus.Errorf("scale app in order failed %v", string(data))
		return errors.New("query app in order failed")
	}

	return
}

func GetSGI(token string, sgiId string, controllerURL string) (sgi entity.ServiceGroupInstance, err error) {
	url := strings.Join([]string{"http://", controllerURL, "/v1/groupInstances/", sgiId}, "")
	logrus.Debugln("query sgi in order url=" + url)

	resp, err := httpclient.Http_get(url, "",
		httpclient.Header{"Content-Type", "application/json"}, httpclient.Header{"X-Auth-Token", token})
	defer resp.Body.Close()
	if err != nil {
		logrus.Errorf("query sgi in order error %v", err)
		return sgi, err
	}

	data, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		logrus.Errorf("query sgi in order failed %v", string(data))
		return sgi, errors.New("query sgi in order failed")
	}

	sgi = entity.ServiceGroupInstance{}
	err = getRetFromResponse(data, &sgi)
	if err != nil {
		logrus.Warnf("query sgi in order from response error %v", err)
		return
	}

	return
}

func TerminateSGO(token string, sgoId string, controllerURL string) (err error) {
	url := strings.Join([]string{"http://", controllerURL, "/v1/serviceGroupOrders/", sgoId}, "")
	logrus.Debugln("get order service group url=" + url)

	resp, err := httpclient.Http_delete(url, "",
		httpclient.Header{"Content-Type", "application/json"}, httpclient.Header{"X-Auth-Token", token})
	defer resp.Body.Close()
	if err != nil {
		logrus.Errorf("http terminate sgo error %v", err)
		return err
	}

	data, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		logrus.Errorf("terminate sgo failed %v", string(data))
		return errors.New("terminate sgo failed")
	}

	return
}

func getRetFromResponse(data []byte, obj interface{}) (err error) {
	var resp *response.Response
	resp = new(response.Response)
	err = json.Unmarshal(data, &resp)
	if err != nil {
		return err
	}

	jsonout, err := json.Marshal(resp.Data)
	if err != nil {
		return err
	}

	json.Unmarshal(jsonout, obj)

	return
}

func getCountFromResponse(data []byte) (count int, err error) {
	var resp *response.QueryStruct
	resp = new(response.QueryStruct)
	err = json.Unmarshal(data, &resp)
	if err != nil {
		return
	}

	jsonout, err := json.Marshal(resp.Count)
	if err != nil {
		return
	}

	json.Unmarshal(jsonout, &count)

	return
}

func HashString(password string) string {
	encry := sha256.Sum256([]byte(password))
	return hex.EncodeToString(encry[:])
}