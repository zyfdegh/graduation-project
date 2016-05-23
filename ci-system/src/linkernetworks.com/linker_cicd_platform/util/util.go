package linker_util

import (
	"errors"
	"github.com/Sirupsen/logrus"
	"github.com/magiconair/properties"
	"github.com/samuel/go-zookeeper/zk"
	"math/rand"
	"strings"
	"time"
)

type Util struct {
	ZkClient *ZkClient
	Props    *properties.Properties
}

type ZkClient struct {
	Props *properties.Properties
	Conn  *zk.Conn
}

var (
	controllerPath      string = "/controller"
	controllerEndpoints []string
)

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

func (z *ZkClient) SetControllerEndpoints(children []string) {
	controllerEndpoints = []string{}
	for _, child := range children {
		endpoint, err := z.getPathData(controllerPath + "/" + child)
		if err == nil {
			controllerEndpoints = append(controllerEndpoints, endpoint)
		}
	}
	logrus.Debugf("reset alived controller endpoints, now are %+v", controllerEndpoints)
}

func (z *ZkClient) GetControllerEndpoints() []string {
	return controllerEndpoints
}

func (z *ZkClient) getPathData(path string) (data string, err error) {
	conn := z.getZkConnection()
	content, _, err := conn.Get(path)
	return string(content), err
}

func (z *ZkClient) WatchController(conn *zk.Conn) {
	snapshots, errors := mirror(conn, controllerPath)
	go func() {
		for {
			select {
			case snapshot := <-snapshots:
				logrus.Debugf("%+v\n", snapshot)
				z.SetControllerEndpoints(snapshot)
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

func (z *ZkClient) GetControllerEndpoint() (url string, err error) {
	endpointLen := len(controllerEndpoints)
	if endpointLen > 0 {
		url = controllerEndpoints[rand.Intn(endpointLen)]
	} else {
		err = errors.New("Can not get alived controller endpoint from zookeeper")
	}
	return
}


