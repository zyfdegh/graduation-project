package services

import (
	// "errors"
	"github.com/Sirupsen/logrus"
	"gopkg.in/mgo.v2/bson"
	"linkernetworks.com/linker_common_lib/persistence/dao"
	"linkernetworks.com/linker_common_lib/persistence/entity"
	"sync"
	"time"
)

var (
	lockService *LockService = nil
	onceLock    sync.Once
	RetryTime   = 5
)

type LockService struct {
	collectionName string
}

func GetLockService() *LockService {
	onceLock.Do(func() {
		logrus.Debugf("Once called from lockService ......................................")
		lockService = &LockService{"service_group_lock"}
	})
	return lockService
}

func (p *LockService) ReleaseInstanceLock(sgiId string) {
	logrus.Debugf("remove lock with sgiId [%v]", sgiId)
	selector := bson.M{}
	selector["instance_id"] = sgiId
	err := dao.HandleDelete(p.collectionName, false, selector)
	if err != nil {
		logrus.Warnf("release instance lock with sgiId [%v] failed, error is %v", sgiId, err)
	}
	return
}

func (p *LockService) getLockBySgi(sgiId string) (locks []entity.InstanceLock, err error) {
	var selector = bson.M{}
	selector["instance_id"] = sgiId
	locks = []entity.InstanceLock{}
	_, err = dao.HandleQueryAll(&locks, dao.QueryStruct{p.collectionName, selector, 0, 0, ""})
	if err != nil {
		logrus.Errorf("query lock [instance_id=%v] error is %v", sgiId, err)
	}
	return
}

func (p *LockService) createLockBySgi(sgiId string) (lock entity.InstanceLock, err error) {
	lock = entity.InstanceLock{}
	lock.GroupInstanceId = sgiId
	lock.ObjectId = bson.NewObjectId()
	err = dao.HandleInsert(p.collectionName, lock)
	if err != nil {
		logrus.Errorf("create lock [instance_id=%v] error is %v", sgiId, err)
	}
	return
}

func (p *LockService) CreateInstanceLock(sgiId string) bool {
	logrus.Debugf("check lock with sgiId [%v]", sgiId)
	for i := 0; i < RetryTime; i++ {
		// get lock by service group instance id
		locks, err := p.getLockBySgi(sgiId)
		if err != nil {
			logrus.Errorf("find lock err is %v", err)
			return false
		}

		if len(locks) == 0 {
			// no lock exist, lock this instance
			_, err := p.createLockBySgi(sgiId)
			if err != nil {
				logrus.Errorf("create lock err is %v", err)
				return false
			}
			return true
		} else {
			// instance is locked, will retry
			logrus.Debugf("instance already locked with sgiId [%v]", sgiId)
			time.Sleep(5000 * time.Millisecond)
		}

	}
	return false
}
