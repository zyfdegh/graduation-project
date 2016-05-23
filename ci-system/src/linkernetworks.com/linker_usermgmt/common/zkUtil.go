package common

import (
	"errors"
	"github.com/Sirupsen/logrus"
	"github.com/magiconair/properties"
	"github.com/samuel/go-zookeeper/zk"
	// "math/rand"
	"strings"
	"time"
)

var UTIL *Util

type Util struct {
	ZkClient *ZkClient
	Props    *properties.Properties
}

type ZkClient struct {
	Props *properties.Properties
	Conn  *zk.Conn
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func (z *ZkClient) getZkConnection() *zk.Conn {
	if z.Conn == nil {
		zksStr := z.Props.GetString("zookeeper.url", "")
		zks := strings.Split(zksStr, ",")
		conn, _, err := zk.Connect(zks, time.Second)
		must(err)
		z.Conn = conn
	}
	return z.Conn
}

func (z *ZkClient) GetZkConnection() *zk.Conn {
	return z.getZkConnection()
}

func (z *ZkClient) GetFirstUserMgmtPath() (path string, err error) {
	userPath := "/userMgmt"
	conn := z.getZkConnection()
	children, _, err := conn.Children(userPath)
	if err != nil {
		logrus.Errorln("get chilren of root path error!", err)
		return
	}
	if len(children) > 0 {
		path = children[0]
	} else {
		err = errors.New("Can not get alived userMgmt endpoint from zookeeper")
	}

	return
}
