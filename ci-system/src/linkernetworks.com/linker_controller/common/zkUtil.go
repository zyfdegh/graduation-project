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
	marathonPath      string = "/marathon/leader"
	marathonEndpoints []string

	userMgmtPath      string = "/userMgmt"
	userMgmtEndpoints []string

	clusterMgmtPath      string = "/cluster"
	clusterMgmtEndpoints []string
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

func (p *ZkClient) SetMarathonEndpoints(children []string) {
	marathonEndpoints = []string{}
	for _, child := range children {
		endpoint, err := p.getPathData(marathonPath + "/" + child)
		if err == nil {
			marathonEndpoints = append(marathonEndpoints, endpoint)
		}
	}
	logrus.Debugf("reset alived marathon endpoints, now are %+v", marathonEndpoints)
}

func (p *ZkClient) SetUserMgmtEndpoints(children []string) {
	userMgmtEndpoints = []string{}
	for _, child := range children {
		endpoint, err := p.getPathData(userMgmtPath + "/" + child)
		if err == nil {
			userMgmtEndpoints = append(userMgmtEndpoints, endpoint)
			logrus.Infoln("userMgmtendpoints", userMgmtEndpoints)
		}
	}
	logrus.Debugf("reset alived user endpoints, now are %+v", userMgmtEndpoints)
}

func (p *ZkClient) SetClusterMgmtEndpoints(children []string) {
	clusterMgmtEndpoints = []string{}
	for _, child := range children {
		endpoint, err := p.getPathData(clusterMgmtPath + "/" + child)
		if err == nil {
			clusterMgmtEndpoints = append(clusterMgmtEndpoints, endpoint)
			logrus.Infoln("clusterMgmtEndpoints", clusterMgmtEndpoints)
		}
	}
	logrus.Debugf("reset alived cluster endpoints, now are %+v", clusterMgmtEndpoints)
}

func (p *ZkClient) GetMarathonEndpoints() []string {
	return marathonEndpoints
}

func (p *ZkClient) getPathData(path string) (data string, err error) {
	conn := p.getZkConnection()
	content, _, err := conn.Get(path)
	return string(content), err
}

func (p *ZkClient) WatchMarathon(conn *zk.Conn) {
	snapshots, errors := mirror(conn, marathonPath)
	go func() {
		for {
			select {
			case snapshot := <-snapshots:
				logrus.Debugf("%+v\n", snapshot)
				p.SetMarathonEndpoints(snapshot)
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
				p.SetUserMgmtEndpoints(snapshot)
			case err := <-errors:
				panic(err)
			}
		}
	}()
}

func (p *ZkClient) WatchClusterMgmt(conn *zk.Conn) {
	snapshots, errors := mirror(conn, clusterMgmtPath)
	go func() {
		for {
			select {
			case snapshot := <-snapshots:
				logrus.Debugf("%+v\n", snapshot)
				p.SetClusterMgmtEndpoints(snapshot)
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

func (p *ZkClient) GetMarathonEndpoint() (url string, err error) {
	endpointLen := len(marathonEndpoints)
	if endpointLen > 0 {
		url = marathonEndpoints[rand.Intn(endpointLen)]
	} else {
		err = errors.New("Can not get alived marathon endpoint from zookeeper")
	}
	return
}

func (p *ZkClient) GetUserMgmtEndpoint() (url string, err error) {
	endpointLen := len(userMgmtEndpoints)
	if endpointLen > 0 {
		url = userMgmtEndpoints[rand.Intn(endpointLen)]
	} else {
		err = errors.New("Can not get alived usermgmt endpoint")
	}
	return
}

func (p *ZkClient) GetClusterMgmtEndpoint() (url string, err error) {
	endpointLen := len(clusterMgmtEndpoints)
	if endpointLen > 0 {
		url = clusterMgmtEndpoints[rand.Intn(endpointLen)]
	} else {
		err = errors.New("Can not get alived cluster endpoint")
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
