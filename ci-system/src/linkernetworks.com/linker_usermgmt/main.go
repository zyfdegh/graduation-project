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
	"linkernetworks.com/linker_usermgmt/api/usermgmt"
	"linkernetworks.com/linker_usermgmt/common"
	"net/http"
	"os"
	"path/filepath"
)

var (
	PROPS          *properties.Properties
	PROPERTIESFILE = flag.String("config", "usermgmt.properties", "the configuration file")
	MONGOALIAS     string
	SWAGGERPATH    string
	LINKERICON     string
)

func init() {
	// get configuration
	flag.Parse()
	fmt.Printf("propertiesFile is %s\n", *PROPERTIESFILE)
	var err error
	if PROPS, err = properties.LoadFile(*PROPERTIESFILE, properties.UTF8); err != nil {
		fmt.Printf("[error] Unable to read properties:%v\n", err)
	}

	// set log configuration
	// Log as JSON instead of the default ASCII formatter.
	switch PROPS.GetString("logrus.formatter", "") {
	case "text":
		logrus.SetFormatter(&logrus.TextFormatter{})
	case "json":
		logrus.SetFormatter(&logrus.JSONFormatter{})
	default:
		logrus.SetFormatter(&logrus.TextFormatter{})
	}

	// Output to stderr instead of stdout, could also be a file.
	logFile := PROPS.GetString("logrus.file", "/var/log/linker_userMgmt.log")
	f, err := os.OpenFile(logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0777)
	if err != nil {
		fmt.Println("error opening file %v", err)
		f, err = os.OpenFile("linker_userMgmt.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0777)
		if err != nil {
			fmt.Println("still failed to open log file linker_userMgmt.log %v", err)
		}
	}
	logrus.SetOutput(f)

	// Only log the warning severity or above.
	level, err := logrus.ParseLevel(PROPS.GetString("logrus.level", "info"))
	if err != nil {
		fmt.Printf("parse log level err is %v\n", err)
		fmt.Printf("using default level is %v \n", logrus.InfoLevel)
		level = logrus.InfoLevel
	}
	logrus.SetLevel(level)

}

func main() {
	// Swagger configuration
	SWAGGERPATH = PROPS.GetString("swagger.path", "")
	LINKERICON = filepath.Join(SWAGGERPATH, "images/mora.ico")
	zkClient := common.ZkClient{PROPS, nil}

	// New, shared session manager, seprate DAO layer
	MONGOALIAS = PROPS.GetString("db.alias", "dev")
	sessMng := session.NewSessionManager(PROPS.FilterPrefix("mongod."), MONGOALIAS)
	defer sessMng.CloseAll()
	dao.DAO = &dao.Dao{SessMng: sessMng, MongoAlias: MONGOALIAS}
	common.UTIL = &common.Util{ZkClient: &zkClient, Props: PROPS}

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
	apiCors := PROPS.GetBool("http.server.cors", false)

	//UserMgmt API
	usermgmt.Register(restful.DefaultContainer, apiCors)

	// register to zookeeper
	zkConn := zkClient.GetZkConnection()
	defer zkConn.Close()
	hostname, err := os.Hostname()
	if err != nil {
		logrus.Errorf("get hostname err is %+v", err)
	}
	endpoint := hostname + ":" + PROPS.MustGet("http.server.port")
	registerToZk(zkConn, hostname, endpoint)

	// watch controller
	// zkClient.WatchController(zkConn)

	basePath := "http://" + PROPS.MustGet("http.server.host") + ":" + PROPS.MustGet("http.server.port")
	// Register Swagger UI
	swagger.InstallSwaggerService(swagger.Config{
		WebServices:     restful.RegisteredWebServices(),
		WebServicesUrl:  "http://" + endpoint,
		ApiPath:         "/apidocs.json",
		SwaggerPath:     SWAGGERPATH,
		SwaggerFilePath: PROPS.GetString("swagger.file.path", ""),
	})

	// If swagger is not on `/` redirect to it
	if SWAGGERPATH != "/" {
		http.HandleFunc("/", index)
	}
	// Serve favicon.ico
	http.HandleFunc("/favion.ico", icon)
	logrus.Infof("ready to serve on %s", basePath)

	logrus.Fatal(http.ListenAndServe(PROPS.MustGet("http.server.host")+":"+PROPS.MustGet("http.server.port"), nil))

}

// If swagger is not on `/` redirect to it
func index(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, SWAGGERPATH, http.StatusMovedPermanently)
}
func icon(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, LINKERICON, http.StatusMovedPermanently)
}

func registerToZk(conn *zk.Conn, hostname, endpoing string) {
	rootPath := "/userMgmt"
	exists, _, err := conn.Exists(rootPath)
	if err != nil {
		// TODO: error handling
		logrus.Errorln("check node exist err is %+v", err)
	}
	if !exists {
		// create root path first
		logrus.Infoln("create rootpath")
		_, err := conn.Create(rootPath, []byte(""), 0, zk.WorldACL(zk.PermAll))
		if err != nil {
			// TODO: error handling
			logrus.Errorln("create rootpath err is %+v", err)
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
