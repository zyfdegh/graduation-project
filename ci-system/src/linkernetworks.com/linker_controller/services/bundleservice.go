package services

import (
	"errors"
	"sync"

	"github.com/Sirupsen/logrus"
	"gopkg.in/mgo.v2/bson"
	"linkernetworks.com/linker_common_lib/persistence/entity"
)

var bundleService *BundleService = nil
var oncebundle sync.Once

var BUNDLE_ERROR_EXPORT = "E11081"
var BUNDLE_ERROR_IMPORT = "E11082"

type BundleService struct {
}

func GetBundleService() *BundleService {
	oncebundle.Do(func() {
		logrus.Debugf("Once called from bundleService ......................................")
		bundleService = &BundleService{}
	})
	return bundleService
}

func (p *BundleService) ExportBundle(token string, sgid string) (bundle *entity.Bundle, errorCode string, err error) {
	code, err := TokenValidation(token)
	if err != nil {
		return nil, code, err
	}

	if authorized := GetAuthService().Authorize("export", token, "", ""); !authorized {
		logrus.Errorln("required opertion is not allowed!")
		return nil, COMMON_ERROR_UNAUTHORIZED, errors.New("required opertion is not authorized!")
	}

	logrus.Infoln("begin to get sg and configuration package by sgid:", sgid)

	if len(sgid) <= 0 {
		logrus.Errorln("invalid sgid for export, sgid:", sgid)
		return nil, COMMON_ERROR_INVALIDATE, errors.New("invalid parameter for export")
	}

	sgSelector := bson.M{}
	if !bson.IsObjectIdHex(sgid) {
		sgSelector["id"] = sgid
	} else {
		sgSelector["_id"] = bson.ObjectIdHex(sgid)
	}

	_, sgs, _, err := GetSgService().queryByQuery(sgSelector, 0, 1, token, true)
	if err != nil {
		logrus.Errorf("get service group failed, error is %v", err)
		errorCode = BUNDLE_ERROR_EXPORT
		return
	}

	if len(sgs) < 1 {
		logrus.Errorf("can not find service group by query [%v]", sgSelector)
		err = errors.New("invalidate service group")
		errorCode = BUNDLE_ERROR_EXPORT
		return
	}

	sg := sgs[0]

	_, acps, _, err := GetAcpService().QueryAllByName(sg.Id, "", 0, 0, token)
	if err != nil {
		logrus.Errorf("get configuration package error %v", err)
		return nil, BUNDLE_ERROR_EXPORT, err
	}

	bundle = new(entity.Bundle)
	bundle.ServiceGroupItem = sg
	if len(acps) > 0 {
		bundle.ConfigrationPackagesItem = acps
	}

	logrus.Infoln("get bundle by sgid success: %v", *bundle)

	return bundle, "", nil

}

func (p *BundleService) ImportBundle(token string, bundle entity.Bundle) (errorCode string, err error) {
	code, err := TokenValidation(token)
	if err != nil {
		return code, err
	}

	if authorized := GetAuthService().Authorize("import", token, "", ""); !authorized {
		logrus.Errorln("required opertion is not allowed!")
		return COMMON_ERROR_UNAUTHORIZED, errors.New("required opertion is not authorized!")
	}

	logrus.Debugln("begin to import bundle:", bundle)

	logrus.Infoln("1. handle service group")
	sg := bundle.ServiceGroupItem
	errorCode, err = p._importServiceGroup(sg, token)
	if err != nil {
		return
	}

	logrus.Infoln("2. handle configuration packages")
	acps := bundle.ConfigrationPackagesItem
	// maps := buildMap(acps)
	if acps != nil && len(acps) > 0 {
		errorCode, err = p._importConfigurationPackage(sg.Id, acps, token)
		if err != nil {
			return
		}

	} else {
		logrus.Infoln("no configuration will be imported")
		return
	}

	return
}

func (p *BundleService) _importServiceGroup(sg entity.ServiceGroup, token string) (errorCode string, err error) {
	_, err = GetSgService().DeleteBySgId(sg.Id, token)
	if err != nil {
		logrus.Errorf("delete sg by id[%v] error %v", sg.Id, err)
		return BUNDLE_ERROR_IMPORT, errors.New("delete service group by id error!")
	}

	_, _, err = GetSgService().Create(sg, token)
	if err != nil {
		logrus.Errorf("create sg[%v] error %v", sg, err)
		return BUNDLE_ERROR_IMPORT, errors.New("create service group error!")
	}

	return

}

func (p *BundleService) _importConfigurationPackage(sgid string, newacps []entity.AppContainerPackage, token string) (errorCode string, err error) {
	_, err = GetAcpService().DeleteBySgOrApp(sgid, "", token)
	if err != nil {
		logrus.Errorf("delete acps by gsid[%v] error %v", sgid, err)
		return BUNDLE_ERROR_IMPORT, errors.New("delete configuration package by sgid error!")
	}

	for i := 0; i < len(newacps); i++ {
		_, _, err = GetAcpService().Create(newacps[i], token)
		if err != nil {
			logrus.Errorf("create acp[%v] error %v", newacps[i], err)
			return BUNDLE_ERROR_IMPORT, errors.New("create new configuration package error!")
		}
	}

	return

}

// func buildMap(acps []entity.AppContainerPackage) map[string][]entity.AppContainerPackage {
// 	ret := make(map[string][]entity.AppContainerPackage)
// 	if acps != nil && len(acps) > 0 {
// 		for i := 0; i < len(acps); i++ {
// 			acp := acps[i]
// 			sgid := acp.ServiceGroupId
// 			acparrays, exist := ret[sgid]
// 			if exist {
// 				acparrays = append(acparrays, acp)
// 				ret[sgid] = acparrays
// 			} else {
// 				acparrays = []entity.AppContainerPackage{acp}
// 				ret[sgid] = acparrays
// 			}
// 		}
// 	}

// 	return ret
// }
