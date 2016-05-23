package documents

import (
	"bytes"
	"encoding/json"
	"net/http"

	"linkernetworks.com/linker_cicd_platform/api/response"
	"linkernetworks.com/linker_cicd_platform/persistence/entity"
	"linkernetworks.com/linker_cicd_platform/util"

	"github.com/Sirupsen/logrus"
	"github.com/emicklei/go-restful"
	"gopkg.in/mgo.v2/bson"
)

var (
	gerritHost    string
	gerritPort    string
	gerritSSHPort string
	adminName     string
	adminPswd     string
)

const (
	projectCollection = "gerrit_project_info"
	ParamAtfId        = "ParamAtfId"
)

func (u *Resource) initConfig() {
	//read config
	prop := u.Util.Props
	gerritPort = prop.MustGetString("gerrit.port")
	adminName = prop.MustGetString("gerrit.admin.name")
	adminPswd = prop.MustGetString("gerrit.admin.password")
	gerritSSHPort = prop.GetString("gerrit.ssh.port", "29418")

	//gerrit host will be inited dynamically
}

func (u Resource) ProjectsWebService() *restful.WebService {
	u.initConfig()

	ws := new(restful.WebService)
	ws.Path("/v1/projects")
	ws.Consumes("*/*")
	ws.Produces(restful.MIME_JSON)

	id := ws.PathParameter(ParamID, "Storage identifier of the collection Project")
	paramID := "{" + ParamID + "}"
	atf_id := ws.PathParameter(ParamAtfId, "Storage identifier of the artifact")
	paramAtfId := "{" + ParamAtfId + "}"

	//####  /projects

	//TODO mv params to request body
	//usage:POST  /projects?name=<NAME>&opsenv_id=<ID>&sm_id=<ID>
	ws.Route(ws.POST("").To(u.ProjectCreateHandler).
		Doc("Create a project").
		Operation("ProjectCreateHandler").
		Param(ws.QueryParameter("name", "Project name")).
		Param(ws.QueryParameter("opsenv_id", "Ops Env Id")).
		Param(ws.QueryParameter("sm_id", "ServiceModel ID")).
		Produces(restful.MIME_JSON).
		Reads(""))

	//usage:GET  /projects
	ws.Route(ws.GET("/").To(u.ProjectListHandler).
		Doc("List all projects").
		Operation("ProjectListHandler").
		Param(ws.QueryParameter("id", "Project Id")).
		Param(ws.QueryParameter("name", "Project Name")).
		Param(ws.QueryParameter("opsenv_id", "OpsEnv Id")).
		Param(ws.QueryParameter("sm_id", "ServiceModel Id")).
		Param(ws.QueryParameter("status", "Project status")).
		Produces(restful.MIME_JSON).
		Reads(""))

	//usage:GET  /projects/<ID>
	ws.Route(ws.GET("/" + paramID).To(u.ProjectListHandler).
		Doc("List a project by id").
		Operation("ProjectListHandler").
		Produces(restful.MIME_JSON).
		Reads(id))

	//usage:PUT  /projects/<ID>
	ws.Route(ws.PUT("/" + paramID).To(u.ProjectUpdateHandler).
		Doc("Update a project").
		Operation("ProjectUpdateHandler").
		Produces(restful.MIME_JSON).
		Reads(id))

	//usage:DELETE  /projects/<ID>
	ws.Route(ws.DELETE("/" + paramID).To(u.ProjectDeleteHandler).
		Doc("Delete a project").
		Operation("ProjectDeleteHandler").
		Reads(id))

	//#### /projects/artifacts

	//usage:POST  /projects/<PROJ_ID>/artifacts
	ws.Route(ws.POST("/" + paramID + "/artifacts").To(u.ArtifactAddHandler).
		Doc("Add an artifact").
		Operation("ArtifactAddHandler").
		Produces(restful.MIME_JSON).
		Reads(id))

	//usage:GET  /projects/<PROJ_ID>/artifacts
	ws.Route(ws.GET("/" + paramID + "/artifacts").To(u.ArtifactListHandler).
		Doc("List all artifacts").
		Operation("ArtifactListHandler").
		Produces(restful.MIME_JSON).
		Reads(""))

	//usage:GET  /projects/<PROJ_ID>/artifacts/<ATF_ID>
	ws.Route(ws.GET("/" + paramID + "/artifacts/" + paramAtfId).
		To(u.ArtifactListHandler).
		Doc("List an artifact").
		Operation("ArtifactListHandler").
		Produces(restful.MIME_JSON).
		Reads(id).
		Reads(atf_id))

	//usage:PUT  /projects/<PROJ_ID>/artifacts
	ws.Route(ws.PUT("/" + paramID + "/artifacts").To(u.ArtifactUpdateHandler).
		Doc("Update all artifacts").
		Operation("ArtifactUpdateHandler").
		Produces(restful.MIME_JSON).
		Reads(id))

	//usage:PUT  /projects/<PROJ_ID>/artifacts/<ATF_ID>
	ws.Route(ws.PUT("/" + paramID + "/artifacts/" + paramAtfId).
		To(u.ArtifactUpdateHandler).
		Doc("Update all artifacts").
		Operation("ArtifactUpdateHandler").
		Produces(restful.MIME_JSON).
		Reads(id).
		Reads(atf_id))

	//usage:DELETE  /projects/<PROJ_ID>/artifacts
	ws.Route(ws.DELETE("/" + paramID + "/artifacts").To(u.ArtifactDeleteHandler).
		Doc("Delete all artifacts").
		Operation("ArtifactDeleteHandler").
		Reads(""))

	//usage:DELETE  /projects/<PROJ_ID>/artifacts/<ATF_ID>
	ws.Route(ws.DELETE("/" + paramID + "/artifacts/" + paramAtfId).
		To(u.ArtifactDeleteHandler).
		Doc("Delete an artifact").
		Operation("ArtifactDeleteHandler").
		Produces(restful.MIME_JSON).
		Reads(id).
		Reads(atf_id))

	return ws
}

/*
Handler to create project in gerrit.
In this method, the http request received will be parsed ,checked and at last
resent to gerrit.
*/
func (u *Resource) ProjectCreateHandler(req *restful.Request,
	resp *restful.Response) {
	logrus.Debugf("ProjectCreateHandler is called")

	proj_name := req.QueryParameter("name")
	opsenv_id := req.QueryParameter("opsenv_id")
	sm_id := req.QueryParameter("sm_id")

	//query gerrit host form env collection and
	//set gerritHost
	u.setGerritHostByEnvId(opsenv_id)

	//check host
	if len(gerritHost) == 0 {
		logrus.Errorln("Gerrit host not found")
		response.WriteError(response.ErrGenericInternal, resp)
		return
	}

	//call gerrit to create project
	http_resp, err := u.createProjectInGerrit(proj_name)
	if err != nil {
		logrus.Errorf("Cannot creating project in gerrit.%v", err.Error())
		response.WriteError(response.ErrCreateProject, resp)
		return
	}

	//if failed
	if http_resp.StatusCode != 201 {
		logrus.Debugf("Failed to create project:" + proj_name)
		switch http_resp.StatusCode {
		case 400:
			logrus.Errorln("Check if opsenv_id exists in opsenvCollection")
			response.WriteStatusError(http_resp.StatusCode,
				response.ErrCreateProject, resp)
		case 401:
			logrus.Errorln("Check admin name and password")
			response.WriteStatusError(http_resp.StatusCode,
				response.ErrCreateProject, resp)
		case 403:
			logrus.Errorln("Authorization failed,or no capability")
			response.WriteStatusError(http_resp.StatusCode,
				response.ErrCreateProject, resp)
		case 404:
			logrus.Errorln("Check if opsenv_id exists in opsenvCollection")
			response.WriteStatusError(http_resp.StatusCode,
				response.ErrCreateProject, resp)
		case 409:
			logrus.Errorln("Project name already exists")
			response.WriteStatusError(http_resp.StatusCode,
				response.ErrProjectConflict, resp)
		default:
			logrus.Errorln("Other reason:%s", http_resp.Status)
			response.WriteStatusError(http_resp.StatusCode,
				response.ErrCreateProject, resp)
		}
		return
	}

	//created
	//combine git url
	git_url := u.combineGitUrl(proj_name)
	//new entity Project
	objId := bson.NewObjectId()
	project := entity.Project{
		Id:             objId.Hex(),
		Name:           proj_name,
		GitUrl:         git_url,
		OpsEnvId:       opsenv_id,
		ServiceModelId: sm_id,
		Status:         entity.PROJECT_STATUS_RUNNING,
	}
	//insert into database
	selector := make(bson.M)
	selector[ParamID] = objId
	_, _, err = u.Dao.HandleInsert(projectCollection, selector,
		ConvertProjectToBson(project))
	if err != nil {
		logrus.Errorf("insert projInfo err is %v", err)
		response.WriteError(response.ErrDBInsert, resp)
		return
	}

	//response
	logrus.Debugf("Success: Created project:" + proj_name)
	response.WriteResponseStatus(http.StatusCreated, project, resp)
	return
}

/*
Handler to  delete project in gerrit.
Note that project in gerrit will not be actually removed, it will be hidden.
*/
func (u *Resource) ProjectDeleteHandler(req *restful.Request,
	resp *restful.Response) {
	logrus.Debugf("ProjectDeleteHandler is called")
	//get project id
	var projId string = req.PathParameter(ParamID)
	if len(projId) == 0 {
		projId = req.QueryParameter("id")
	}
	if len(projId) == 0 {
		logrus.Errorln("Project id not provided.")
		response.WriteStatusError(http.StatusBadRequest,
			response.ErrBadRequestURL, resp)
		return
	}

	//query project
	project, err := u.findProjectById(projId)
	if err != nil {
		logrus.Debugf("Error query project. %v", err)
		response.WriteStatusError(http.StatusNotFound,
			response.ErrDBQuery, resp)
		return
	}
	logrus.Debugf("Project to delete Id,name:%s,%s", project.Id, project.Name)

	//query gerrit host
	if len(gerritHost) == 0 {
		err = u.setGerritHostByProjId(project.Id)
		if err != nil {
			logrus.Errorf("Error query gerrit host. %v", err.Error())
			return
		}
	}

	if len(project.Name) == 0 {
		logrus.Errorf("Project not found or it was not created "+
			"via linker_cicd_platform.%v", err)
		response.WriteStatusError(http.StatusNotFound,
			response.ErrProjectNotFound, resp)
		return
	} else {
		//hide project in gerrit if proj_namd is found
		http_resp, err := u.hideProjectInGerrit(project.Name)
		defer http_resp.Body.Close()

		if err != nil {
			logrus.Debugf(err.Error())
			response.WriteError(response.ErrHideProject, resp)
			return
		}
		//If gerrit has hidden project
		if http_resp.StatusCode == 200 {
			//All OK
			//terminate jobs and this project
			err = u.terminateProject(project.Id)
			if err != nil {
				logrus.Errorf("Error terminate project,reason:%v,project id:%s",
					err, project.Id)
			}

			logrus.Debugf("Deleted projects:" + project.Name)
			response.WriteSuccess(resp)
			return
		} else {
			logrus.Debugf("Failed to hide project.:" + project.Name)
			logrus.Errorln("Fail to hide project in gerrit.Does it exist?")
			response.WriteStatusError(http_resp.StatusCode,
				response.ErrHideProject, resp)
			return
			//TODO handle unexpected errors here if necessary
		}
	}
}

//ProjectUpdateHandler
func (u *Resource) ProjectUpdateHandler(req *restful.Request,
	resp *restful.Response) {
	//convert request body to obj
	body := bson.M{}
	decoder := json.NewDecoder(req.Request.Body)
	err := decoder.Decode(&body)
	if err != nil {
		logrus.Errorln("Error decoding request body as entity.Project")
		response.WriteStatusError(http.StatusBadRequest,
			response.ErrBadRequestBody, resp)
		return
	}
	// Compose a selector from request
	proj_id := req.PathParameter(ParamID)
	if len(proj_id) > 0 {
		selector := bson.M{}
		selector[ParamID] = bson.ObjectIdHex(proj_id)

		//update
		//u.handleUpdate(projectCollection, one, selector, document, req, resp)
		_, newDoc, _, err := u.Dao.HandleUpdateById(projectCollection, selector,
			body)
		if err != nil {
			response.WriteError(response.ErrDBQuery, resp)
			return
		}
		response.WriteResponseStatus(http.StatusOK, newDoc, resp)
		return
	} else {
		response.WriteStatusError(http.StatusBadRequest,
			response.ErrBadRequestURL, resp)
		return
	}
}

/*
This function can list gerrit projects created via cicd_platform.
It returns project names formatted ad JSON in http response body.
*/
func (u *Resource) ProjectListHandler(req *restful.Request,
	resp *restful.Response) {
	logrus.Debugf("ProjectListHandler is called")

	//get project id
	projId := req.PathParameter(ParamID)

	//Request format:
	//GET projects/<ID>
	if len(projId) > 0 {
		u.handleList(projectCollection, "list_projects", req, resp)
		return
	}
	//Request format:
	//GET projects?opsenv_id=<A1>&id=<A2>&sm_id=<A3>...
	selector := bson.M{}
	if len(req.QueryParameter("opsenv_id")) > 0 {
		selector["opsenv_id"] = req.QueryParameter("opsenv_id")
	}
	if len(req.QueryParameter("id")) > 0 {
		selector["id"] = req.QueryParameter("id")
	}
	if len(req.QueryParameter("sm_id")) > 0 {
		selector["sm_id"] = req.QueryParameter("sm_id")
	}
	if len(req.QueryParameter("name")) > 0 {
		selector["name"] = req.QueryParameter("name")
	}
	if len(req.QueryParameter("status")) > 0 {
		selector["status"] = req.QueryParameter("status")
	}
	//check map len
	if len(selector) == 0 {
		response.WriteStatusError(http.StatusBadRequest,
			response.ErrBadRequestURL, resp)
		return
	}
	//query
	_, _, document, err := u.Dao.HandleQuery(projectCollection, selector,
		false, bson.M{}, 0, 0, "", "true")
	if err != nil {
		response.WriteError(response.ErrDBQuery, resp)
		return
	}

	//parse query result as project array
	var projs []entity.Project
	data, err := json.Marshal(document)
	if err != nil {
		logrus.Errorln("Error marshal array of entity.Project")
		response.WriteError(response.ErrConvertJson, resp)
		return
	}
	err = json.Unmarshal(data, &projs)
	if err != nil {
		logrus.Errorln("Error unmarshal array of entity.Project")
		response.WriteError(response.ErrConvertJson, resp)
		return
	}
	//OK
	response.WriteResponseStatus(http.StatusOK, projs, resp)
	return
}

//CRUD on artifacts

//list all artifacts, or list a single artifact if atf_id is provided
func (u *Resource) ArtifactListHandler(req *restful.Request,
	resp *restful.Response) {
	logrus.Debugf("ArtifactListHandler is called")

	//get params
	proj_id := req.PathParameter(ParamID)
	atf_id := req.PathParameter(ParamAtfId)
	if len(proj_id) == 0 {
		response.WriteStatusError(http.StatusBadRequest,
			response.ErrBadRequestURL, resp)
		return
	}

	//query project from db
	var project entity.Project
	selector, _, _ := getSelector(req)
	_, _, document, err := u.Dao.HandleQuery(projectCollection, selector, true,
		bson.M{}, 0, 0, "", "true")
	if err != nil {
		logrus.Errorln("Cannot query artifacts,check proj_id")
		response.WriteError(response.ErrDBQuery, resp)
		return
	}
	data, err := json.Marshal(document)
	if err != nil {
		response.WriteError(response.ErrConvertJson, resp)
		return
	}
	err = json.Unmarshal(data, &project)
	if err != nil {
		response.WriteError(response.ErrConvertJson, resp)
		return
	}

	//get artifacts from project
	var artifacts []entity.Artifact = project.Artifacts

	if len(atf_id) == 0 {
		//query all
		response.WriteResponseStatus(http.StatusOK, artifacts, resp)
		return
	}

	//query one
	var found bool = false
	var artifact entity.Artifact
	for _, each := range artifacts {
		if each.Id == atf_id {
			found = true
			artifact = each
			break
		}
	}
	if !found {
		response.WriteStatusError(http.StatusNotFound,
			response.ErrArtifactNotFound, resp)
		return
	}
	response.WriteResponseStatus(http.StatusOK, artifact, resp)
	return
}

//add an artifact
func (u *Resource) ArtifactAddHandler(req *restful.Request,
	resp *restful.Response) {
	logrus.Debugf("ArtifactsAddHandler is called")
	proj_id := req.PathParameter(ParamID)
	if len(proj_id) > 0 {
		//query gerritProjectInfo
		selector := bson.M{}
		selector[ParamID] = bson.ObjectIdHex(proj_id)
		_, _, document, err := u.Dao.HandleQuery(projectCollection, selector,
			true, bson.M{}, 0, 0, "", "true")
		if err != nil {
			response.WriteError(response.ErrDBQuery, resp)
			return
		}
		project := new(entity.Project)
		data, _ := json.Marshal(document)
		json.Unmarshal(data, &project)

		//parse request body as artifact
		var artifact *entity.Artifact
		artifact = new(entity.Artifact)
		decoder := json.NewDecoder(req.Request.Body)
		err = decoder.Decode(&artifact)
		if err != nil {
			response.WriteStatusError(http.StatusBadRequest,
				response.ErrBadRequestBody, resp)
			return
		}

		//if this artifact has no field named "id", add one
		if artifact.Id == "" {
			artifact.Id = bson.NewObjectId().Hex()
		}

		//add this artifact
		slice := project.Artifacts[:]
		slice = append(slice, *artifact)

		project.Artifacts = slice

		//save to db
		_, newDoc, _, err := u.Dao.HandleUpdateById(projectCollection, selector,
			ConvertProjectToBson(*project))
		if err != nil {
			logrus.Errorln("Error inserting artifacts into db")
			response.WriteError(response.ErrDBUpdate, resp)
			return
		}
		response.WriteResponseStatus(http.StatusCreated, newDoc, resp)
		return
	}
}

//update artifacts
func (u *Resource) ArtifactUpdateHandler(req *restful.Request,
	resp *restful.Response) {
	logrus.Debugf("ArtifactsUpdateHandler is called")
	//fetch req body,transform it into obj
	//replace artifacts ,and save it at last
	proj_id := req.PathParameter(ParamID)
	atf_id := req.PathParameter(ParamAtfId)

	logrus.Debugf("params:%v", proj_id)
	if len(proj_id) == 0 {
		response.WriteStatusError(http.StatusBadRequest,
			response.ErrBadRequestURL, resp)
		return
	}

	//query collection first
	selector := bson.M{}
	selector[ParamID] = bson.ObjectIdHex(proj_id)
	_, _, document, err := u.Dao.HandleQuery(projectCollection, selector, true,
		bson.M{}, 0, 0, "", "true")

	if err != nil {
		response.WriteError(response.ErrDBQuery, resp)
		return
	}

	project := new(entity.Project)
	out, _ := json.Marshal(document)
	json.Unmarshal(out, &project)

	//parse req body
	if len(atf_id) > 0 {
		//parse json artifact
		var atf entity.Artifact
		decoder := json.NewDecoder(req.Request.Body)
		err = decoder.Decode(&atf)
		if err != nil {
			logrus.Errorf("Err decoding http body as artifact :%v", err)
			response.WriteStatusError(http.StatusBadRequest,
				response.ErrBadRequestBody, resp)
			return
		}
		var found bool = false

		for idx, each := range project.Artifacts {
			if each.Id == atf_id {
				found = true
				if atf.Id == "" {
					atf.Id = atf_id
				}
				project.Artifacts[idx] = atf
				break
			}
		}

		if !found {
			response.WriteStatusError(http.StatusNotFound,
				response.ErrArtifactNotFound, resp)
			return
		}
	} else {
		//parse json array of artifacts
		var atfs []entity.Artifact
		decoder := json.NewDecoder(req.Request.Body)
		err = decoder.Decode(&atfs)
		if err != nil {
			logrus.Errorf("Err decoding artifact array :%v", err)
			response.WriteStatusError(http.StatusBadRequest,
				response.ErrBadRequestBody, resp)
			return
		}
		//replace artifacts,and save to db
		project.Artifacts = atfs
	}

	_, newDoc, _, err := u.Dao.HandleUpdateById(projectCollection, selector,
		ConvertProjectToBson(*project))

	if err != nil {
		response.WriteError(response.ErrDBUpdate, resp)
		return
	}

	response.WriteResponseStatus(http.StatusOK, newDoc, resp)
	return
}

//delete an artifact
func (u *Resource) ArtifactDeleteHandler(req *restful.Request,
	resp *restful.Response) {
	logrus.Debugf("ArtifactDeleteHandler is called")
	//fetch req body,transform it into obj
	//remove artifacts ,and save it at last
	proj_id := req.PathParameter(ParamID)
	atf_id := req.PathParameter(ParamAtfId)
	logrus.Debugf("params:%v,%v", proj_id, atf_id)
	if len(proj_id) == 0 {
		response.WriteStatusError(http.StatusBadRequest,
			response.ErrBadRequestURL, resp)
		return
	}
	//query collection first
	selector := bson.M{}
	//	selector["id"] = proj_id
	selector[ParamID] = bson.ObjectIdHex(proj_id)
	_, _, document, err := u.Dao.HandleQuery(projectCollection, selector, true,
		bson.M{}, 0, 0, "", "true")

	if err != nil {
		response.WriteError(response.ErrDBQuery, resp)
		return
	}

	project := new(entity.Project)
	out, err := json.Marshal(document)
	if err != nil {
		logrus.Debugf("Error marshal document,%v", err.Error())
		response.WriteError(response.ErrConvertJson, resp)
		return
	}
	err = json.Unmarshal(out, &project)
	if err != nil {
		logrus.Debugf("Error unmarshal document,%v", err.Error())
		response.WriteError(response.ErrConvertJson, resp)
		return
	}

	if len(atf_id) > 0 {
		//delete the artifact with atf_id
		//replace artifact
		var found bool = false
		for idx, each := range project.Artifacts {
			if each.Id == atf_id {
				//Remove indexed element
				found = true
				project.Artifacts = append(project.Artifacts[:idx],
					project.Artifacts[idx+1:]...)
				break
			}
		}

		if !found {
			response.WriteStatusError(http.StatusNotFound,
				response.ErrArtifactNotFound, resp)
			return
		}
	} else {
		//Clear all artifacts
		//TODO check if exceptions occur
		project.Artifacts = nil
	}
	//save the modified project to db
	selector[ParamID] = bson.ObjectIdHex(proj_id)
	_, newDoc, _, err := u.Dao.HandleUpdateById(projectCollection, selector,
		ConvertProjectToBson(*project))

	if err != nil {
		logrus.Errorf("insert projInfo err is %v", err)
		response.WriteError(response.ErrDBUpdate, resp)
		return
	}

	response.WriteResponseStatus(http.StatusOK, newDoc, resp)
	return
}

//Combine git clone url
func (u *Resource) combineGitUrl(project_name string) (git_url string) {
	//eg:ssh://sxin@192.168.10.119:29418/myproj2
	git_url = "ssh://" + adminName + "@" + gerritHost + ":" +
		gerritSSHPort + "/" + project_name
	return
}

//Create project in gerrit
func (u *Resource) createProjectInGerrit(project_name string) (resp *http.Response,
	err error) {
	//eg: PUT http://192.168.10.119:8083/a/projects/proj_name
	url := "http://" + gerritHost + ":" + gerritPort + "/a/projects/" +
		project_name
	method := "PUT"

	logrus.Debugf("Resend request to gerrit.Url:%v" + url)

	//use default request body
	var body = []byte(`  {
		"description": "This project is created via linker cicd platform.",
		"submit_type": "CHERRY_PICK",
		"owners": ["Administrators"]
	}`)
	req, err := http.NewRequest(method, url, bytes.NewBuffer(body))
	if err != nil {
		logrus.Errorf("Error making new http request,reason:%v", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("charset", "UTF-8")

	//add digest authorization in Header
	digestStr, err := linker_util.CalcDigestHeader(adminName, adminPswd,
		method, url)
	if err != nil {
		logrus.Errorf("Error generating digest,reason:%v", err)
		return
	}

	req.Header.Set("Authorization", digestStr)

	//do http request
	client := http.DefaultClient
	resp, err = client.Do(req)
	defer resp.Body.Close()
	if err != nil {
		logrus.Errorf("Error do http request,reason:%v", err)
		return
	}
	return
}

//Hide project in gerrit
//If projects need to be deleted, install plugin on gerrit,then lookup REST API
func (u *Resource) hideProjectInGerrit(project_name string) (resp *http.Response,
	err error) {
	//PUT /projects/proj_name/config
	//eg: PUT http://192.168.10.119:8083/a/projects/proj_name/config
	url := "http://" + gerritHost + ":" + gerritPort + "/a/projects/" +
		project_name + "/config"
	method := "PUT"
	//this body is a ConfigInfo entity formatted as json
	var body = []byte(`{"state": "HIDDEN"}`)

	req, err := http.NewRequest(method, url, bytes.NewBuffer(body))
	if err != nil {
		logrus.Debugf("Error making a new http request")
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("charset", "UTF-8")

	//add digest authorization in Header
	digestStr, err := linker_util.CalcDigestHeader(adminName, adminPswd,
		method, url)
	if err != nil {
		logrus.Debugf("Error generating digest.%v", err.Error())
		return
	}
	req.Header.Set("Authorization", digestStr)

	//do http request
	client := http.DefaultClient
	resp, err = client.Do(req)
	defer resp.Body.Close()
	if err != nil {
		logrus.Debugf("Error do http request,check gerrit host and port.%v",
			err.Error())
		return
	}
	return
}

//This function will query opsenv_id from projectCollection
func (u *Resource) setGerritHostByProjId(proj_id string) (err error) {
	//query projectCollection and return opsenv_id
	selector := bson.M{}
	selector[ParamID] = bson.ObjectIdHex(proj_id)
	_, _, document, err := u.Dao.HandleQuery(projectCollection, selector, true,
		bson.M{}, 0, 0, "", "true")
	if err != nil {
		logrus.Errorf("Error querying opsenv_id from projectCollection %v", err)
		return
	}
	var projObj *entity.Project
	projObj = new(entity.Project)
	out, _ := json.Marshal(document)
	json.Unmarshal(out, &projObj)
	env_id := projObj.OpsEnvId

	if len(env_id) > 0 {
		err = u.setGerritHostByEnvId(env_id)
	} else {
		logrus.Errorf("OpsEnvId not found in projectCollection.")
		return
	}
	return
}

//This function will query Gerrit Host from opsenvCollection
//It be called before accessing gerrit,if not, gerrit host will be empty.
//Projects cannot be created nor hidden
func (u *Resource) setGerritHostByEnvId(env_id string) (err error) {
	//query opsenvCollection and return gerritInfo
	selector := bson.M{}
	selector[ParamID] = bson.ObjectIdHex(env_id)
	_, _, envJson, err := u.Dao.HandleQuery(opsenvCollection, selector, true,
		bson.M{}, 0, 0, "", "true")
	if err != nil {
		logrus.Errorf("Error querying opsenvCollection, %v", err)
		return
	}
	var envObj *entity.OpsEnv
	envObj = new(entity.OpsEnv)
	envout, _ := json.Marshal(envJson)
	json.Unmarshal(envout, &envObj)

	host := envObj.GerritDockerIP
	port := envObj.GerritHttpPort

	if len(host) > 0 {
		gerritHost = host
		gerritPort = port
		logrus.Infof("Gerrit host:%v", gerritHost)
	} else {
		//use default
		gerritHost = u.Util.Props.GetString("gerrit.host.default", "")
		logrus.Warningf("Gerrit host not found,use default:" + gerritHost)
	}
	return
}

func (u *Resource) findProjectById(projId string) (project *entity.Project,
	err error) {
	selector := bson.M{}
	selector[ParamID] = bson.ObjectIdHex(projId)
	_, _, document, err := u.Dao.HandleQuery(projectCollection, selector, true,
		bson.M{}, 0, 1, "", "true")
	project = new(entity.Project)
	data, err := json.Marshal(document)
	if err != nil {
		return
	}
	err = json.Unmarshal(data, &project)
	if err != nil {
		return
	}
	return
}

func (u *Resource) terminateProjects(opsEnvId string) (err error) {
	//query projects with specific opsEnvId
	selector := bson.M{}
	selector["opsenv_id"] = opsEnvId
	_, _, document, err := u.Dao.HandleQuery(projectCollection, selector, false,
		bson.M{}, 0, 0, "", "true")
	var projects []entity.Project = make([]entity.Project, 0)
	data, _ := json.Marshal(document)
	err = json.Unmarshal(data, &projects)

	//loop
	for _, project := range projects {
		//terminate projects
		err = u.terminateProject(project.Id)
		if err != nil {
			logrus.Debugf("Error terminating project,reason:%v,project id:%s",
				err, project.Id)
		}
	}
	return
}

//terminate jobs ,if ok , terminate project
func (u *Resource) terminateProject(projId string) (err error) {
	//terminate jobs
	err = u.terminateJobs(projId)
	if err != nil {
		logrus.Debugf("Error terminating jobs,reason:%v,project id:%s", err, projId)
		return
	} else {
		//hide project
		proj, _ := u.findProjectById(projId)
		_, err = u.hideProjectInGerrit(proj.Name)
		//terminate project
		change := bson.M{"status": entity.PROJECT_STATUS_TERMINATED}
		selector := bson.M{}
		selector[ParamID] = bson.ObjectIdHex(projId)
		err = u.Dao.HandleUpdateByQueryPartial(projectCollection, selector, change)
		return
	}
}
