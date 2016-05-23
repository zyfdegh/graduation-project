package main

import (
	"flag"
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/emicklei/go-restful"
	"github.com/emicklei/go-restful/swagger"
	"github.com/magiconair/properties"
	"github.com/samuel/go-zookeeper/zk"
	"linkernetworks.com/linker_common_lib/persistence/dao"
	"linkernetworks.com/linker_common_lib/persistence/session"
	"linkernetworks.com/linker_controller/api/documents"
	"linkernetworks.com/linker_controller/common"
	"linkernetworks.com/linker_controller/services"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

var (
	Props          *properties.Properties
	PropertiesFile = flag.String("config", "controller.properties", "the configuration file")
	ZkFlag         = flag.String("zk", "", "zk url")
	HostnameFlag   = flag.String("hostname", "", "hostname")
	MongoAlias     string
	SwaggerPath    string
	LinkerIcon     string
	ZK             string
	Hostname       string
	// TODO:  Mongo          string
)

func init() {
	// get configuration
	flag.Parse()
	ZK = *ZkFlag
	Hostname = *HostnameFlag
	fmt.Printf("PropertiesFile is %s\n", *PropertiesFile)
	var err error
	if Props, err = properties.LoadFile(*PropertiesFile, properties.UTF8); err != nil {
		fmt.Printf("[error] Unable to read properties:%v\n", err)
	}

	// set log configuration
	// Log as JSON instead of the default ASCII formatter.
	switch Props.GetString("logrus.formatter", "") {
	case "text":
		logrus.SetFormatter(&logrus.TextFormatter{})
	case "json":
		logrus.SetFormatter(&logrus.JSONFormatter{})
	default:
		logrus.SetFormatter(&logrus.TextFormatter{})
	}
	// Use the Airbrake hook to report errors that have Error severity or above to
	// an exception tracker. You can create custom hooks, see the Hooks section.
	// log.AddHook(airbrake.NewHook("https://example.com", "xyz", "development"))

	// Output to stderr instead of stdout, could also be a file.
	logrus.SetOutput(os.Stderr)

	// Only log the warning severity or above.
	level, err := logrus.ParseLevel(Props.GetString("logrus.level", "info"))
	if err != nil {
		fmt.Printf("parse log level err is %v\n", err)
		fmt.Printf("using default level is %v \n", logrus.InfoLevel)
		level = logrus.InfoLevel
	}
	logrus.SetLevel(level)

}

func main() {
	// Swagger configuration
	SwaggerPath = Props.GetString("swagger.path", "")
	LinkerIcon = filepath.Join(SwaggerPath, "images/mora.ico")

	// TODO: Check mongo flag, if mongo is set by flag, using mongo flag

	// New, shared session manager, seprate DAO layer
	MongoAlias = Props.GetString("db.alias", "dev")
	sessMng := session.NewSessionManager(Props.FilterPrefix("mongod."), MongoAlias)
	defer sessMng.CloseAll()
	dao.DAO = &dao.Dao{SessMng: sessMng, MongoAlias: MongoAlias}
	fmt.Println(dao.DAO.MongoAlias)

	// Check zk flag, if zk is set by flag, using zk flag
	zkClient := common.ZkClient{
		Url:   ZK,
		Props: Props,
		Conn:  nil,
	}
	common.UTIL = &common.Util{ZkClient: &zkClient, Props: Props}

	// accept and respond in JSON unless told otherwise
	restful.DefaultRequestContentType(restful.MIME_JSON)
	restful.DefaultResponseContentType(restful.MIME_JSON)
	// gzip if accepted
	restful.DefaultContainer.EnableContentEncoding(true)
	// faster router
	restful.DefaultContainer.Router(restful.CurlyRouter{})
	// no need to access body more than once
	restful.SetCacheReadEntity(false)
	// API Cross-origin requests
	apiCors := Props.GetBool("http.server.cors", false)
	// Documents API
	documents.Register(restful.DefaultContainer, apiCors)

	// register to zookeeper
	zkConn := zkClient.GetZkConnection()
	defer zkConn.Close()
	// Check hostname flag, if hostname is set by flag, using hostname
	if strings.TrimSpace(Hostname) == "" {
		hostname, err := os.Hostname()
		if err != nil {
			logrus.Errorf("get hostname err is %+v", err)
		}
		Hostname = hostname
	}
	endpoint := Hostname + ":" + Props.MustGet("http.server.port")
	registerToZk(zkConn, Hostname, endpoint)

	// watch marathon
	zkClient.WatchMarathon(zkConn)

	//init user management endpoints
	userMgmtHA := Props.GetBool("usermgmt.ha", true)
	if userMgmtHA == true {
		//user management is true, watch the zookeeper
		//create usermgmt node here, or controller will failed to start due to no usermgmt node
		createUserMgmtNode(zkConn)
		//watch usermgmt
		zkClient.WatchUserMgmt(zkConn)
	} else {
		//user management is false, just use the endpoint in propertiey file
		userEndpoint := Props.MustGet("usermgmt.endpoint")
		services.SetUserMgmtEndpoint(userEndpoint)
	}

	//init cluster management endpoint
	createClusterMgmtNode(zkConn)
	//watch clustermgmt
	zkClient.WatchClusterMgmt(zkConn)

	basePath := "http://" + Props.MustGet("http.server.host") + ":" + Props.MustGet("http.server.port")
	// Register Swagger UI
	swagger.InstallSwaggerService(swagger.Config{
		WebServices:     restful.RegisteredWebServices(),
		WebServicesUrl:  "http://" + endpoint,
		ApiPath:         "/apidocs.json",
		SwaggerPath:     SwaggerPath,
		SwaggerFilePath: Props.GetString("swagger.file.path", ""),
	})

	// If swagger is not on `/` redirect to it
	if SwaggerPath != "/" {
		http.HandleFunc("/", index)
	}
	// Serve favicon.ico
	http.HandleFunc("/favion.ico", icon)
	logrus.Infof("ready to serve on %s", basePath)

	logrus.Fatal(http.ListenAndServe(Props.MustGet("http.server.host")+":"+Props.MustGet("http.server.port"), nil))

	// router := NewRouter().StrictSlash(true)
	// logrus.Fatal(http.ListenAndServe(Props["http.server.host"]+":"+Props["http.server.port"], router))
}

// If swagger is not on `/` redirect to it
func index(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, SwaggerPath, http.StatusMovedPermanently)
}
func icon(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, LinkerIcon, http.StatusMovedPermanently)
}

func registerToZk(conn *zk.Conn, hostname, endpoing string) {
	rootPath := "/controller"
	exists, _, err := conn.Exists(rootPath)
	if err != nil {
		// TODO: error handling
		logrus.Errorf("get hostname err is %+v", err)
	}
	if !exists {
		// create root path first
		logrus.Infof("create rootpath")
		_, err := conn.Create(rootPath, []byte(""), 0, zk.WorldACL(zk.PermAll))
		if err != nil {
			// TODO: error handling
			logrus.Errorf("create rootpath err is %+v", err)
		}
	}

	// register to rootpath as a child
	logrus.Infof("create nodepath: " + rootPath + "/" + hostname)
	_, err = conn.Create(rootPath+"/"+hostname, []byte(endpoing), zk.FlagEphemeral+zk.FlagSequence, zk.WorldACL(zk.PermAll))
	if err != nil {
		// TODO: error handling
		logrus.Errorf("create nodepath err is %+v", err)
	}
}

func createUserMgmtNode(conn *zk.Conn) {
	exists, _, err := conn.Exists("/userMgmt")
	if err != nil {
		// TODO: error handling
		logrus.Errorln("check node exist err is %+v", err)
	}
	if !exists {
		// create root path first
		logrus.Infoln("create userMgmt node")
		_, err := conn.Create("/userMgmt", []byte(""), 0, zk.WorldACL(zk.PermAll))
		if err != nil {
			// TODO: error handling
			logrus.Errorln("create userMgmt node err is %+v", err)
		}
	}
}

func createClusterMgmtNode(conn *zk.Conn) {
	exists, _, err := conn.Exists("/cluster")
	if err != nil {
		// TODO: error handling
		logrus.Errorln("check node exist err is %+v", err)
	}
	if !exists {
		// create root path first
		logrus.Infoln("create cluster node")
		_, err := conn.Create("/cluster", []byte(""), 0, zk.WorldACL(zk.PermAll))
		if err != nil {
			// TODO: error handling
			logrus.Errorln("create cluster node err is %+v", err)
		}
	}
}
