package documents

import (
	"io"
	"net/http"
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/emicklei/go-restful"
	"gopkg.in/mgo.v2/bson"
	"launchpad.net/xmlpath"
	"linkernetworks.com/linker_cicd_platform/api/response"
	command "linkernetworks.com/linker_cicd_platform/command"
	"linkernetworks.com/linker_cicd_platform/persistence/entity"
	linker_util "linkernetworks.com/linker_cicd_platform/util"
)

var dockerCollection = "dockers"

func (u Resource) DockerWebService() *restful.WebService {
	ws := new(restful.WebService)
	ws.Path("/v1/dockers")
	ws.Consumes("*/*")
	ws.Produces(restful.MIME_JSON)

	paramID := "{" + ParamID + "}"

	// id := ws.PathParameter(ParamID, "Storage identifier of the app configuration")
	// paramID := "{" + ParamID + "}"

	ws.Route(ws.GET("/").To(u.ListDockerFilesHandler).
		Param(ws.QueryParameter("count", "Counts total items and return the result in X-Object-Count header").DataType("boolean")).
		Param(ws.QueryParameter("query", "Query in json format")).
		Param(ws.QueryParameter("fields", "Comma separated list of field names")).
		Param(ws.QueryParameter("skip", "Number of items to skip in the result set, default=0")).
		Param(ws.QueryParameter("limit", "Maximum number of items in the result set, default=0")).
		Param(ws.QueryParameter("sort", "Comma separated list of field names to sort")).
		Param(ws.QueryParameter("status", "Status of DockFile")))

	ws.Route(ws.POST("/").To(u.UploadDockerFilesHandler))

	ws.Route(ws.PUT("/" + paramID).To(u.UpdateDockerFileStatusHandler).
		Param(ws.QueryParameter("passed", "Is passed?(true for passed, false for failed)")))

	ws.Route(ws.DELETE("/" + paramID).To(u.DeleteDockerFileHandler))

	ws.Route(ws.GET("/" + paramID + "/download").To(u.DownloadDockerFilesHandler))

	return ws
}

func (u *Resource) ListDockerFilesHandler(req *restful.Request, resp *restful.Response) {
	// query by userid & tenantid
	u.handleDockerFileListByUser(dockerCollection, req, resp)
}

func (u *Resource) UploadDockerFilesHandler(req *restful.Request, resp *restful.Response) {
	request := req.Request

	request.ParseMultipartForm(32 << 20)
	file, _, err := request.FormFile("file")

	newId := bson.NewObjectId()

	dockerfilename := request.FormValue("dockerfile")
	imagename := request.FormValue("imagename")
	fmt.Println(dockerfilename)
	zipfileName := request.FormValue("zipfilename")
	version := request.FormValue("version")
	email := request.FormValue("email")
	buildnow := request.FormValue("buildnow")
	buildFlag, flagerr := strconv.ParseBool(buildnow)
	
	if flagerr != nil {
		buildFlag = false
	}

	namespace := parseEmail(email)
	imageName := namespace + "/" + imagename
	
	selector := make(bson.M)
	selector["imagename"] = imageName
	selector["version"] = version
	
	total, _, _, err := u.Dao.HandleQuery(dockerCollection, selector, false, bson.M{}, 0, 0, "", "true")
	if err != nil {
		response.WriteError(response.ErrDBQuery, resp)
		return
	}
	
	if(total == 1) {
		// existed dockerfile 
		// will return the error to ui.
		response.WriteStatusError(http.StatusConflict, response.ErrDockerFileExisted, resp)
		return
	}


	userid, tenantid, _, _ := u.getUserAndTenantId(req.HeaderParameter("X-Auth-Token"))

	//	fmt.Fprintf(w, "%v", handler.Header)
	os.MkdirAll(u.Util.Props.GetString("dockerfile.store.path", "")+"/"+newId.Hex(), 0777)
	f, err := os.OpenFile(u.Util.Props.GetString("dockerfile.store.path", "")+"/"+newId.Hex()+"/"+zipfileName, os.O_WRONLY|os.O_CREATE, 0666)

	if err != nil {
		logrus.Errorf("Create file err is %v", err)
		response.WriteError(response.ErrCreateFile, resp)
		return
	}
	defer f.Close()
	io.Copy(f, file)

	// Insert data to db.
	dockfileObj := entity.DockerFile{Id: newId.Hex(), Version: version, DockerFile: dockerfilename, ZipFile: zipfileName, ImageName: imageName, Status: entity.DOCKER_FILE_STATUS_UPLOADED, UserId: userid, TenantId: tenantid}
	selector = make(bson.M)
	selector[ParamID] = newId
	u.Dao.HandleInsert(dockerCollection, selector, ConvertDockerFileToBson(dockfileObj))

	if !buildFlag {
		go u.BuildDockerImage(dockfileObj.Id, false, nil, nil, "")
	}

	// write response to client
	response.WriteResponse(dockfileObj, resp)
	return
}

func (u *Resource) UpdateDockerFileStatusHandler(req *restful.Request, resp *restful.Response) {
	// query by userid & tenantid
	// Compose a selector from request
	selector, _, err := getSelector(req)
	if err != nil {
		logrus.Errorf("get docker file by id err is %v", err)
		//		response.WriteError(err, resp)
		//TODO return 500? or 400?
		response.WriteStatusError(http.StatusBadRequest,
			response.ErrBadRequestGeneric, resp)
		return
	}

	passed := req.QueryParameter("passed")
	var status string
	if "true" == passed {
		status = entity.DOCKER_FILE_STATUS_PUBLISHED
	} else {
		status = entity.DOCKER_FILE_STATUS_REJECTED
	}

	change := bson.M{"status": status}

	err = u.Dao.HandleUpdateByQueryPartial(dockerCollection, selector, change)
	if err != nil {
		logrus.Errorf("Update dockerfile status err is %v", err)
	}
}

func (u *Resource) DeleteDockerFileHandler(req *restful.Request, resp *restful.Response) {
	// query by userid & tenantid
	// Compose a selector from request
	selector, one, err := getSelector(req)
	if err != nil {
		logrus.Errorf("delete docker file err is %v", err)
		response.WriteStatusError(http.StatusBadRequest,
			response.ErrBadRequestGeneric, resp)
		return
	}

	id := req.PathParameter(ParamID)
	os.RemoveAll(u.Util.Props.GetString("dockerfile.store.path", "") + "/" + id)

	u.handleDelete(dockerCollection, one, selector, req, resp)
}

func (u *Resource) BuildDockerImage(dockerfileId string, downloadArtifact bool, artifactArray []entity.Artifact,
	env *entity.OpsEnv, version string) (err error) {
	// first delete the temp folder.
	err = os.RemoveAll(u.Util.Props.GetString("dockerfile.store.path", "") + "/" + dockerfileId + "/temp")
	if err != nil {
		logrus.Errorf("Failed to delete docker file temp folder %s err is %v", err)
		return
	}

	logrus.Debugln("docker file id is " + dockerfileId)
	selector := make(bson.M)
	selector[ParamID] = bson.ObjectIdHex(dockerfileId)
	_, _, dockerfileJson, err := u.Dao.HandleQuery(dockerCollection, selector, true, bson.M{}, 0, 0, "", "true")

	if err != nil {
		logrus.Errorf("Failed to find dockerfile in db with id: %s, and err is %v", dockerfileId, err)
		return
	}

	if dockerfileJson == nil {
		logrus.Errorf("Can't find dockerfile obj in db with id: %s", dockerfileId)
		return
	}

	var dockerfileObj *entity.DockerFile
	dockerfileObj = new(entity.DockerFile)
	dockerfileout, _ := json.Marshal(dockerfileJson)
	json.Unmarshal(dockerfileout, &dockerfileObj)

	logrus.Debugln("docker file name is " + dockerfileObj.ImageName)

	// create a temp folder to zip file.
	tempfolder := u.Util.Props.GetString("dockerfile.store.path", "") + "/" + dockerfileObj.Id + "/temp"
	logrus.Debugf("Temp folder is %s", tempfolder)

	os.MkdirAll(tempfolder, 0775)

	if checkZipFile(dockerfileObj.ZipFile) {
		// extract
		err = Unzip(u.Util.Props.GetString("dockerfile.store.path", "")+"/"+dockerfileObj.Id+"/"+dockerfileObj.ZipFile, tempfolder)
		if err != nil {
			logrus.Errorf("Unzip file error is %v", err)
			return
		}
	} else if checkTarGzFile(dockerfileObj.ZipFile) {
		// extract tar.gz file.
		err = Untar(u.Util.Props.GetString("dockerfile.store.path", "")+"/"+dockerfileObj.Id+"/"+dockerfileObj.ZipFile, tempfolder)
		if err != nil {
			logrus.Errorf("Untar file error is %v", err)
			return
		}
	} else {
		logrus.Errorln("Can't extact file: " + dockerfileObj.ZipFile + ", nothing can be done.")
	}

	if downloadArtifact {
		// TODO: download artifacts files.
		baseurl := strings.Join([]string{"http://", env.NexusDockerIP, ":" , env.NexusHttpPort ,"/nexus/service/local/repositories/releases/content/"}, "")
		for _, artifact := range artifactArray {
			groupPath := strings.Replace(artifact.GroupId, ".", "/", -1)
			downloadurl := ""
			// get metadata file
			if strings.Contains(version, "-SNAPSHOT") {
				baseurl = strings.Join([]string{"http://", env.NexusDockerIP, ":" , env.NexusHttpPort ,"/nexus/service/local/repositories/snapshots/content/"}, "")
				//get metadata file
				metaDatadownloadurl := strings.Join([]string{baseurl, groupPath, "/", artifact.Name, "/", version, "/maven-metadata.xml"}, "")
	
				logrus.Debugf("Downloading file: ", metaDatadownloadurl)
				err := linker_util.DownloadFile(metaDatadownloadurl, tempfolder, "maven-metadata.xml")
				if err != nil {
					logrus.Errorf("Downloading file failed %s err is %v", err)
				}
	
				logrus.Debugln("Parse metadata: maven-metadata.xml")
				timestamp, buildnumber, err := parseMetadataXML(tempfolder + "/maven-metadata.xml")
				if err != nil {
					logrus.Errorf("Parse metadata file failed %s err is %v", err)
				}
	
				logrus.Debugln("Build number is " + buildnumber)
	
				versions := strings.Split(version, "-")
				versionNumber := versions[0]
				fullname := strings.Join([]string{artifact.Name, "-", versionNumber, "-", timestamp, "-", buildnumber, ".", artifact.Type}, "")
	
				downloadurl = strings.Join([]string{baseurl, groupPath, "/", artifact.Name, "/", version, "/", fullname}, "")
			} else {
				baseurl = strings.Join([]string{"http://", env.NexusDockerIP, ":" , env.NexusHttpPort ,"/nexus/service/local/repositories/releases/content/"}, "")
				fullname := strings.Join([]string{artifact.Name, "-", version, ".", artifact.Type}, "")
				downloadurl = strings.Join([]string{baseurl, groupPath, "/", artifact.Name, "/", version, "/", fullname}, "")
			}
	
			logrus.Debugf("Downloading file: ", downloadurl)
			err = linker_util.DownloadFile(downloadurl, tempfolder, artifact.Name+"."+artifact.Type)
			if err != nil {
				logrus.Errorf("Downloading file failed %s err is %v", err)
			}
		}
	}
	
	buildcmd := "docker build -t " + u.Util.Props.GetString("docker.repo.prefix", "") + "/" + dockerfileObj.ImageName + ":" + dockerfileObj.Version + " -f " + tempfolder + "/" + dockerfileObj.DockerFile + " " + tempfolder
	logrus.Debugf("Build cmd is ", buildcmd)
	_, _, errcmd := command.ExecCommand(buildcmd)
	if errcmd != nil {
		logrus.Errorf("Build docker image failed %s err is %v", err)
		u.updateBuildStatus(dockerfileId, "Failed")
		return errcmd
	}

	pushimagecmd := "docker push " + u.Util.Props.GetString("docker.repo.prefix", "") + "/" + dockerfileObj.ImageName + ":" + dockerfileObj.Version
	_, _, errcmd = command.ExecCommand(pushimagecmd)
	logrus.Debugf("Push cmd is ", pushimagecmd)
	if errcmd != nil {
		logrus.Errorf("Push docker image failed %s err is %v", err)
		u.updateBuildStatus(dockerfileId, "Failed")
		return errcmd
	}

	u.updateBuildStatus(dockerfileId, "Success")

	return 
}

func (u *Resource) updateBuildStatus(id, status string) (err error){
	change := bson.M{"build_status": status}
	selector := make(bson.M)
	selector[ParamID] = bson.ObjectIdHex(id)
	err = u.Dao.HandleUpdateByQueryPartial(dockerCollection, selector, change)
	if err != nil {
		logrus.Errorf("Update dockerfile build status err is %v", err)
	}
	return
}

func Unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer func() {
		if err := r.Close(); err != nil {
			panic(err)
		}
	}()

	os.MkdirAll(dest, 0755)

	// Closure to address file descriptors issue with all the deferred .Close() methods
	extractAndWriteFile := func(f *zip.File) error {
		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer func() {
			if err := rc.Close(); err != nil {
				panic(err)
			}
		}()

		path := filepath.Join(dest, f.Name)

		if f.FileInfo().IsDir() {
			os.MkdirAll(path, f.Mode())
		} else {
			f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return err
			}
			defer func() {
				if err := f.Close(); err != nil {
					panic(err)
				}
			}()

			_, err = io.Copy(f, rc)
			if err != nil {
				return err
			}
		}
		return nil
	}

	for _, f := range r.File {
		err := extractAndWriteFile(f)
		if err != nil {
			return err
		}
	}

	return nil
}

func Untar(file, dest string) error {
	os.MkdirAll(dest, 0755)
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()
	gr, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gr.Close()
	tr := tar.NewReader(gr)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}
		if hdr.FileInfo().IsDir() {
			os.MkdirAll(dest+string(os.PathSeparator)+hdr.Name, hdr.FileInfo().Mode())
		} else {
			fw, err := os.OpenFile(dest+string(os.PathSeparator)+hdr.Name, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, hdr.FileInfo().Mode())
			if err != nil {
				return err
			}
			defer fw.Close()
			_, err = io.Copy(fw, tr)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func parseMetadataXML(filepath string) (timestamp, buildnumber string, err error) {

	buildNumberPath := xmlpath.MustCompile("metadata/versioning/snapshot/buildNumber")
	timestampPath := xmlpath.MustCompile("metadata/versioning/snapshot/timestamp")

	file, err := os.Open(filepath)
	if err != nil {
		return
	}

	root, err := xmlpath.Parse(file)
	if err != nil {
		return
	}

	buildnumber, _ = buildNumberPath.String(root)
	timestamp, _ = timestampPath.String(root)

	return
}

func (u *Resource) handleDockerFileListByUser(collectionName string, req *restful.Request, resp *restful.Response) {
	// Get selector from `_id` path parameter and `query` query parameter
	selector, one, err := getSelector(req)
	if err != nil {
		return
	}

	// parse request
	var fields bson.M = getFields(req)
	var skip int = queryIntParam(req, "skip", 0)
	var limit int = queryIntParam(req, "limit", 0)
	var sort string = req.QueryParameter("sort")
	var extended_json = req.QueryParameter("extended_json")

	status := req.QueryParameter("status")
	if status != "" {
		selector["Status"] = status
	}

	userid, tenantid, role, _ := u.getUserAndTenantId(req.HeaderParameter("X-Auth-Token"))

	if role != "sysadmin" {
		selector["user_id"] = userid
		selector["tenant_id"] = tenantid
	}

	if one {
		_, _, jsonDocument, err := u.Dao.HandleQuery(collectionName, selector, one, fields, skip, limit, sort, extended_json)

		if err != nil {
			logrus.Errorf("handle query err is %v", err)
			response.WriteError(response.ErrDBQuery, resp)
			return
		}

		response.WriteResponse(jsonDocument, resp)
		return
	}

	_, lenth, jsonDocuments, err := u.Dao.HandleQuery(collectionName, selector, one, fields, skip, limit, sort, extended_json)
	if err != nil {
		logrus.Errorf("handle query err is %v", err)
		response.WriteError(response.ErrDBQuery, resp)
		return
	}

	//currently add auth check here! TODO: need refactor
	var res QueryStruct
	var filterCount int
	res = QueryStruct{Success: true, Data: jsonDocuments}

	// Get limit amount
	limitnum := 10
	if limit > 0 {
		limitnum = limit
	}

	// If got full limit set next link
	if lenth == limitnum {
		res.Prev, res.Next = u.prevnexturl(req)
	}

	// Count documents if count parameter is included in query
	if c, _ := strconv.ParseBool(req.QueryParameter("count")); c {
		res.Count = filterCount
		resp.AddHeader("X-Object-Count", strconv.Itoa(filterCount))
	}

	// Write result back to client
	resp.WriteEntity(res)
}

func parseEmail(email string) string {

	if email != "" {
		if "sysadmin" == email {
			return "linker"
		}

		array := strings.Split(email, "@")

		if len(array) == 2 {
			fmt.Println("len = 2")
			value1 := strings.Replace(array[0], ".", "_", -1)
			value2 := strings.Replace(array[1], ".", "_", -1)

			result := value1 + "_at_" + value2
			return result
		} else {

			logrus.Errorf("Invalid email address is %s can't parse it!", email)
			return ""
		}
	} else {
		return email
	}

}

func (u *Resource) DownloadDockerFilesHandler(req *restful.Request, resp *restful.Response) {
	id := req.PathParameter(ParamID)

	logrus.Debugln("Downloading docker file... with id: " + id)

	selector := make(bson.M)
	selector[ParamID] = bson.ObjectIdHex(id)
	_, _, dockerfileJson, err := u.Dao.HandleQuery(dockerCollection, selector, true, bson.M{}, 0, 0, "", "true")

	if err != nil {
		logrus.Errorf("Get docker ifle by id is %v", err)
		response.WriteError(response.ErrDBQuery, resp)
		return
	}

	var dockerfileObj *entity.DockerFile
	dockerfileObj = new(entity.DockerFile)
	dockerfileout, _ := json.Marshal(dockerfileJson)
	json.Unmarshal(dockerfileout, &dockerfileObj)

	// read file content to response
	fullPath := u.Util.Props.GetString("dockerfile.store.path", "") + "/" + dockerfileObj.Id + "/" + dockerfileObj.ZipFile
	file, err := os.Open(fullPath)
	if err != nil {
		logrus.Errorf("Read docker file error %v", err)
		response.WriteError(response.ErrReadFile, resp)
		return
	}

	resp.AddHeader("Content-type", "application/octet-stream")
	resp.AddHeader("Content-disposition", " attachment; filename="+dockerfileObj.ZipFile+"")
	resp.AddHeader("Content-Transfer-Encoding", "binary")
	io.Copy(resp.ResponseWriter, file)

}
