package main

import (
	"flag"
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/emicklei/go-restful"
	"github.com/emicklei/go-restful/swagger"
	"github.com/magiconair/properties"
	"linkernetworks.com/linker_cicd_platform/api/documents"
	"linkernetworks.com/linker_cicd_platform/util"
	"linkernetworks.com/linker_cicd_platform/persistence/dao"
	"linkernetworks.com/linker_cicd_platform/persistence/session"
	"log"
	"net/http"
	"os"
)

var (
	props          *properties.Properties
	propertiesFile = flag.String("config", "cicd.properties", "the configuration file")
	SwaggerPath    string
	mongoAlias     string
)

func init() {

	// get configuration
	flag.Parse()
	fmt.Printf("propertiesFile is %s\n", *propertiesFile)
	var err error
	if props, err = properties.LoadFile(*propertiesFile, properties.UTF8); err != nil {
		fmt.Printf("[error] Unable to read properties:%v\n", err)
	}

	// set log configuration
	// Log as JSON instead of the default ASCII formatter.
	switch props.GetString("logrus.formatter", "") {
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
	level, err := logrus.ParseLevel(props.GetString("logrus.level", "info"))
	if err != nil {
		fmt.Printf("parse log level err is %v\n", err)
		fmt.Printf("using default level is %v \n", logrus.InfoLevel)
		level = logrus.InfoLevel
	}
	logrus.SetLevel(level)

}

func main() {
	// Swagger configuration
	SwaggerPath = props.GetString("swagger.path", "")
	zkClient := linker_util.ZkClient{props, nil}
	conn := zkClient.GetZkConnection()
	defer conn.Close()
	zkClient.WatchController(conn)

	util := linker_util.Util{
		ZkClient: &zkClient,
		Props:    props,
	}

	basePath := "http://" + props.MustGet("http.server.host") + ":" + props.MustGet("http.server.port")

	mongoAlias = props.GetString("db.alias", "dev")
	sessMng := session.NewSessionManager(props.FilterPrefix("mongod."), mongoAlias)
	defer sessMng.CloseAll()
	dao := &dao.Dao{SessMng: sessMng, MongoAlias: mongoAlias}
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
	apiCors := props.GetBool("http.server.cors", false)
	// Documents API
	documents.Register(dao, restful.DefaultContainer, apiCors, &util)

	// Optionally, you can install the Swagger Service which provides a nice Web UI on your REST API
	// You need to download the Swagger HTML5 assets and change the FilePath location in the config below.
	// Open http://localhost:8080/apidocs and enter http://localhost:8080/apidocs.json in the api input field.
	config := swagger.Config{
		WebServices:    restful.RegisteredWebServices(), // you control what services are visible
		WebServicesUrl: basePath,
		ApiPath:        "/apidocs.json",

		// Optionally, specifiy where the UI is located
		SwaggerPath:     SwaggerPath,
		SwaggerFilePath: props.GetString("swagger.file.path", "")}

	swagger.InstallSwaggerService(config)

	log.Printf("start listening on " + props.MustGet("http.server.host") + ":" + props.MustGet("http.server.port"))
	log.Fatal(http.ListenAndServe(":"+props.MustGet("http.server.port"), nil))
}
