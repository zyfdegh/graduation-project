package common

import (
	"errors"
	"math/rand"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/magiconair/properties"
	"github.com/samuel/go-zookeeper/zk"
)

var UTIL *Util

type Util struct {
	ZkClient *ZkClient
	Props    *properties.Properties
}

type ZkClient struct {
	Url   string
	Props *properties.Properties
	Conn  *zk.Conn
}

var (
	controllerPath      string = "/controller"
	controllerEndpoints []string

	userMgmtPath      string = "/userMgmt"
	userMgmtEndpoints []string
)

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func (p *ZkClient) getZkConnection() *zk.Conn {
	if p.Conn == nil {
		zksStr := p.Props.GetString("zookeeper.url", "")
		if strings.TrimSpace(p.Url) != "" {
			zksStr = p.Url
		}
		logrus.Infof("connect to zk: %v", zksStr)
		zks := strings.Split(zksStr, ",")
		conn, _, err := zk.Connect(zks, time.Second)
		must(err)
		p.Conn = conn
	}
	return p.Conn
}

func (p *ZkClient) GetZkConnection() *zk.Conn {
	return p.getZkConnection()
}

func (p *ZkClient) SetControllerEndpoints(children []string) {
	controllerEndpoints = []string{}
	for _, child := range children {
		endpoint, err := p.getPathData(controllerPath + "/" + child)
		if err == nil {
			controllerEndpoints = append(controllerEndpoints, endpoint)
		}
	}
	logrus.Debugf("reset alived controller endpoints, now are %+v", controllerEndpoints)
}

func (p *ZkClient) setUserMgmtEndpoints(children []string) {
	userMgmtEndpoints = []string{}
	for _, child := range children {
		endpoint, err := p.getPathData(userMgmtPath + "/" + child)
		if err == nil {
			userMgmtEndpoints = append(userMgmtEndpoints, endpoint)
			logrus.Infoln("userMgmtendpoints", userMgmtEndpoints)
		}
	}
	logrus.Debugf("reset alived controller endpoints, now are %+v", userMgmtEndpoints)
}

// func (p *ZkClient) GetMarathonEndpoints() []string {
// 	return marathonEndpoints
// }

func (p *ZkClient) getPathData(path string) (data string, err error) {
	conn := p.getZkConnection()
	content, _, err := conn.Get(path)
	return string(content), err
}

func (p *ZkClient) WatchController(conn *zk.Conn) {
	snapshots, errors := mirror(conn, controllerPath)
	go func() {
		for {
			select {
			case snapshot := <-snapshots:
				logrus.Debugf("%+v\n", snapshot)
				p.SetControllerEndpoints(snapshot)
			case err := <-errors:
				panic(err)
			}
		}
	}()
}

func (p *ZkClient) WatchUserMgmt(conn *zk.Conn) {
	snapshots, errors := mirror(conn, userMgmtPath)
	go func() {
		for {
			select {
			case snapshot := <-snapshots:
				logrus.Debugf("%+v\n", snapshot)
				p.setUserMgmtEndpoints(snapshot)
			case err := <-errors:
				panic(err)
			}
		}
	}()
}

func mirror(conn *zk.Conn, path string) (chan []string, chan error) {
	snapshots := make(chan []string)
	errors := make(chan error)
	go func() {
		for {
			snapshot, _, events, err := conn.ChildrenW(path)
			if err != nil {
				errors <- err
				return
			}
			snapshots <- snapshot
			evt := <-events
			if evt.Err != nil {
				errors <- evt.Err
				return
			}
		}
	}()
	return snapshots, errors
}

func (p *ZkClient) GetControllerEndpoint() (url string, err error) {
	endpointLen := len(controllerEndpoints)
	if endpointLen > 0 {
		url = controllerEndpoints[rand.Intn(endpointLen)]
	} else {
		err = errors.New("Can not get alived controller endpoint from zookeeper")
	}
	return
}

func (p *ZkClient) GetUserMgmtEndpoint() (url string, err error) {
	endpointLen := len(userMgmtEndpoints)
	if endpointLen > 0 {
		url = userMgmtEndpoints[rand.Intn(endpointLen)]
	} else {
		err = errors.New("Can not get alived usermgmt endpoint from zookeeper")
	}
	return
}

func (p *ZkClient) GetFirstControllerPath() (path string, err error) {
	controllerPath := "/controller"
	conn := p.getZkConnection()
	children, _, err := conn.Children(controllerPath)
	if err != nil {
		logrus.Errorln("get chilren of root path error!", err)
		return
	}
	if len(children) > 0 {
		path = children[0]
	} else {
		err = errors.New("Can not get alived controller endpoint from zookeeper")
	}

	return
}
