package documents

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/bndr/gojenkins"
	"github.com/emicklei/go-restful"
	"github.com/jmoiron/jsonq"

	"gopkg.in/mgo.v2/bson"
	//	"gopkg.in/mgo.v2"

	//	"linkernetworks.com/linker_cicd_platform/persistence/dao"
	"linkernetworks.com/linker_cicd_platform/api/response"
	entity "linkernetworks.com/linker_cicd_platform/persistence/entity"
)

type Body struct {
	body string
}

const (
	jobCollection = "jobs"
)

func (u Resource) JobsWebService() *restful.WebService {
	ws := new(restful.WebService)
	ws.Path("/v1/jobs")
	ws.Consumes("*/*")
	ws.Produces(restful.MIME_JSON)

	paramID := "{" + ParamID + "}"

	// id := ws.PathParameter(ParamID, "Storage identifier of the app configuration")
	// paramID := "{" + ParamID + "}"

	ws.Route(ws.POST("").To(u.JobsCreateHandler).
		Param(ws.QueryParameter("jobname", "Jenkins job name")).
		Param(ws.QueryParameter("version", "project version")).
		Param(ws.QueryParameter("projectid", "Project ID")).
		Param(ws.QueryParameter("branch", "Branch of gerrit")).
		Param(ws.QueryParameter("autodeploy", "Flag for autodeploy")))

	ws.Route(ws.GET("").To(u.ListJobsHandler).
		Param(ws.QueryParameter("projectid", "Project ID")))

	ws.Route(ws.PUT("/" + paramID).To(u.JobsVersionUpdateHandler).
		Param(ws.QueryParameter("version", "project version")).
		Param(ws.QueryParameter("projectid", "Project ID")))

	ws.Route(ws.DELETE("/" + paramID).To(u.JobsDeleteHandler).
		Param(ws.QueryParameter("projectid", "Project ID")))

	ws.Route(ws.PUT("/" + paramID + "/build").To(u.DeployJob).
		Param(ws.PathParameter(paramID, "Jenkins job id")).
		Param(ws.QueryParameter("projectid", "Project ID")))

	ws.Route(ws.POST("/" + paramID + "/jobenvs").To(u.JobEnvsCreateHandler).
		Param(ws.PathParameter(paramID, "Job ID")))

	ws.Route(ws.POST("/notifications").To(u.JobNotificationHandler).
		Param(ws.QueryParameter("jobid", "Jenkins job id")))

	ws.Route(ws.DELETE("/" + paramID + "/jobenvs").To(u.JobEnvsDeleteHandler).
		Param(ws.PathParameter(paramID, "Job ID")))

	return ws
}

func (u *Resource) JobsCreateHandler(req *restful.Request, resp *restful.Response) {
	logrus.Debugf("JobsCreateHandler is called")
	var jobname string = req.QueryParameter("jobname")
	var projectid string = req.QueryParameter("projectid")
	var version string = req.QueryParameter("version")
	var branch string = req.QueryParameter("branch")
	var autodeploy string = req.QueryParameter("autodeploy")

	autoDeploy, _ := strconv.ParseBool(autodeploy)

	newId := bson.NewObjectId()

	project, err := u.createJob(jobname, projectid, newId.Hex(), branch)
	if err != nil {
		logrus.Errorf("Create Jenkins Job error %s", err.Error())
		response.WriteError(response.ErrDBQuery, resp)
	}

	jobObj := entity.Job{Id: newId.Hex(), Name: jobname, GitUrl: project.GitUrl, Project: project.Id, Version: version, Branch: branch, AutoDeploy: autoDeploy}
	selector := make(bson.M)
	selector[ParamID] = newId

	u.Dao.HandleInsert(jobCollection, selector, ConvertJobToBson(jobObj))
	response.WriteResponse(jobObj, resp)
	return
}

func (u *Resource) ListJobsHandler(req *restful.Request, resp *restful.Response) {
	// query by project id
	u.handleListByProjectId(jobCollection, req, resp)
}

func (u *Resource) JobsVersionUpdateHandler(req *restful.Request, resp *restful.Response) {
	logrus.Debugf("JobsCreateHandler is called")
	jobid := req.PathParameter(ParamID)
	var projectid string = req.QueryParameter("projectid")
	var version string = req.QueryParameter("version")

	selector := make(bson.M)
	selector[ParamID] = bson.ObjectIdHex(jobid)
	selector["project"] = projectid
	logrus.Debugf("Selector: ", selector)

	change := bson.M{"version": version}

	err := u.Dao.HandleUpdateByQueryPartial(jobCollection, selector, change)
	if err != nil {
		logrus.Errorf("Update job err is %v", err)
	}

	return
}

func (u *Resource) DeployJob(req *restful.Request, resp *restful.Response) {
	jobid := req.PathParameter(ParamID)
	var projectid string = req.QueryParameter("projectid")
	project, _ := u.findProjectById(projectid)
	opsenv, _ := u.findOpsEnvById(project.OpsEnvId)
	jobobj := u.findJobByID(jobid)

	baseurl := "http://" + opsenv.JenkinsDockerIP + ":" + opsenv.JenkinsHttpPort
	jenkins, err := gojenkins.CreateJenkins(baseurl, u.Util.Props.GetString("jenkins.admin.name", ""), u.Util.Props.GetString("jenkins.admin.password", "")).Init()
	if err != nil {
		logrus.Errorf("Create Jenkins Client error %s", err.Error())
		response.WriteError(response.ErrCreateJob, resp)
	}

	_, err = jenkins.BuildJob(jobobj.Name)
	if err != nil {
		logrus.Errorf("Build job error %s", err.Error())
		response.WriteError(response.ErrBuildJob, resp)
	}
}

func (u *Resource) JobsDeleteHandler(req *restful.Request, resp *restful.Response) {
	// Be careful, the id is job name!!!
	jobname := req.PathParameter(ParamID)
	var projectid string = req.QueryParameter("projectid")
	logrus.Debugf("JobsDeleteHandler is called: " + jobname)
	logrus.Debugf("Project is : " + projectid)

	selector := make(bson.M)
	selector["name"] = jobname
	selector["project"] = projectid
	logrus.Debugf("Selector: ", selector)

	//	project, _ := u.findProjectById(projectid)

	//err := u.Dao.HandleDelete(jobCollection, true, selector)
	//get job id
	_, _, document, err := u.Dao.HandleQuery(jobCollection, selector, true, bson.M{}, 0, 1, "", "true")
	if err != nil {
		logrus.Errorf("query job err is %v", err)
		return
	}
	job := new(entity.Job)
	data, err := json.Marshal(document)
	if err != nil {
		logrus.Errorf("marshal job err is %v", err)
	}
	err = json.Unmarshal(data, &job)
	if err != nil {
		logrus.Errorf("ummarshal job err is %v", err)
	}

	//terminate project envs and this job
	err = u.terminateJob(job.Id)
	if err != nil {
		logrus.Debugf("Error terminating job,reason:%v,job id: %s", err, job.Id)
	}

	//	err = u.deleteJenkinsJob(project.OpsEnvId, jobname)
	//	if err != nil {
	//		response.WriteError(err, resp)
	//	}

	return
}

func (u *Resource) deleteJenkinsJob(opsenvId string, jobname string) (err error) {
	opsenv, _ := u.findOpsEnvById(opsenvId)

	baseurl := "http://" + opsenv.JenkinsDockerIP + ":" + opsenv.JenkinsHttpPort
	jenkins, err := gojenkins.CreateJenkins(baseurl, u.Util.Props.GetString("jenkins.admin.name", ""), u.Util.Props.GetString("jenkins.admin.password", "")).Init()
	if err != nil {
		logrus.Errorf("Create Jenkins Client error %s", err.Error())
		return err
	}

	_, err = jenkins.DeleteJob(jobname)
	if err != nil {
		logrus.Errorf("Delete Jenkins Client error %s", err.Error())
		return err
	}
	return nil
}

func (u *Resource) JobNotificationHandler(req *restful.Request, resp *restful.Response) {
	logrus.Debugf("JobNotificationHandler is called")

	var jobid string = req.QueryParameter("jobid")

	body, _ := ioutil.ReadAll(req.Request.Body)
	logrus.Debugf(string(body))
	fmt.Println(string(body))

	jobObj := u.findJobByID(jobid)

	jsondata := map[string]interface{}{}

	result := json.NewDecoder(strings.NewReader(string(body)))
	result.Decode(&jsondata)

	jq := jsonq.NewQuery(jsondata)

	status, _ := jq.String("build", "status")
	buildNumber, _ := jq.Int("build", "number")

	selector := make(bson.M)
	selector["_id"] = bson.ObjectIdHex(jobid)

	change := bson.M{"buildnumber": buildNumber, "status": status}

	err := u.Dao.HandleUpdateByQueryPartial(jobCollection, selector, change)
	if err != nil {
		logrus.Errorf("Update job err is %v", err)
	}

	if "SUCCESS" == status {
		// asyc to create docker instances in mesos
		if jobObj.AutoDeploy == true {
			go u.LaunchAllTasks(jobid)
		}
	}

	return
}

func (u *Resource) JobEnvsDeleteHandler(req *restful.Request, resp *restful.Response) {
	jobid := req.PathParameter(ParamID)
	logrus.Debugf("JobEnvsDeleteHandler is called: " + jobid)

	selector := make(bson.M)
	selector["job_id"] = jobid
	logrus.Debugf("Selector: ", selector)

	_, _, jobEnvJson, err := u.Dao.HandleQuery(projectenvCollection, selector, true, bson.M{}, 0, 1, "", "true")
	if err != nil {
		logrus.Errorf("Query JobEnv error when delete job env %s", err.Error())
		response.WriteError(response.ErrDBQuery, resp)
	}

	var jobEnv *entity.ProjectEnv
	jobEnv = new(entity.ProjectEnv)
	envout, _ := json.Marshal(jobEnvJson)
	json.Unmarshal(envout, &jobEnv)

	logrus.Debugf("Get token by Userid %s & Tenantid %s ", jobEnv.UserId, jobEnv.TenantId)
	token, err := u.getUserToken(jobEnv.UserId, jobEnv.TenantId)

	logrus.Debugf("Delete customer env %s ", jobEnv.ServiceOrderId)
	err = u.deleteEnv(jobEnv.ServiceOrderId, token)
	if err != nil {
		logrus.Errorf("Delete customer error %s", err.Error())
		response.WriteError(response.ErrDeleteOpsenv, resp)
	}

	err = u.Dao.HandleDelete(projectenvCollection, true, selector)
	if err != nil {
		logrus.Errorf("Delete JobEnv in db error %s", err.Error())
		response.WriteError(response.ErrDBDelete, resp)
	}
	return
}

func (u *Resource) LaunchAllTasks(jobid string) {
	logrus.Debugf("Find job by id %s", jobid)
	job := u.findJobByID(jobid)

	logrus.Debugf("Find project by id %s", job.Project)
	project, _ := u.findProjectById(job.Project)

	logrus.Debugf("Find env by id %s", project.OpsEnvId)
	env, _ := u.findOpsEnvById(project.OpsEnvId)

	dockerMap := u.getDfAtfsMapping(job.Project)
	for dockerfileid, artifactArray := range dockerMap {
		logrus.Debugf("Try to build docker image id: %s", dockerfileid)
		u.BuildDockerImage(dockerfileid, true, artifactArray, env, job.Version)
	}

	logrus.Debugf("Get token by Userid %s & Tenantid %s ", env.UserId, env.TenantId)
	token, err := u.getUserToken(env.UserId, env.TenantId)
	if err != nil {
		logrus.Errorf("Get user Token err is %v", err)
		return
	}

	// Create customer ENV, order it in Controller.
	orderString := "{\"service_group_id\":\"" + project.ServiceModelId + "\",\"parameters\": []}"
	err, serviceGroupInstanceId, serviceOrderId, serviceOfferingInstanceId, _, _ := u.createEnv(orderString, token)

	if err != nil {
		logrus.Errorf("Create customer env err is %v", err)
		return
	}

	// Create JobEnv obj to store in DB
	newId := bson.NewObjectId()
	projectEnvObj := entity.ProjectEnv{Id: newId.Hex(), OpsEnvId: project.OpsEnvId, ProjectId: job.Project,
		JobId: jobid, UserId: env.UserId, TenantId: env.TenantId, ServiceGroupInstanceId: serviceGroupInstanceId,
		ServiceOrderId: serviceOrderId, Status: "Ordered", ServiceOfferingInstanceId: serviceOfferingInstanceId}
	selector := make(bson.M)
	selector[ParamID] = newId

	document, err := ConvertToBson(projectEnvObj)
	if err != nil {
		logrus.Errorf("convert to bson error is %v", err)
	} else {
		_, _, err = u.Dao.HandleInsert(projectenvCollection, selector, document)
		if err != nil {
			logrus.Errorf("insert project env err is %v", err)
			return
		}
	}

	return
}

func (u *Resource) JobEnvsCreateHandler(req *restful.Request, resp *restful.Response) {
	jobid := req.PathParameter(ParamID)
	u.LaunchAllTasks(jobid)
	return
}

func (u *Resource) createJob(jobName string, projectid string, jobid string, branch string) (project *entity.Project, err error) {
	bytes, err := ioutil.ReadFile("/usr/local/bin/javaproj_job.xml")
	if err != nil {
		logrus.Errorf("Read template job file err is %v", err)
	}
	s := string(bytes)

	project, _ = u.findProjectById(projectid)
	logrus.Debugf("Find project by id %s", projectid)

	opsenv, _ := u.findOpsEnvById(project.OpsEnvId)
	logrus.Debugf("Find opsenv by id %s", project.OpsEnvId)

	url := "http://" + u.Util.Props.GetString("http.server.host", "") + ":" + u.Util.Props.GetString("http.server.port", "") + "/v1/jobs/notifications?jobid=" + jobid
	logrus.Debugf("url for notification: " + url)
	jobtext := strings.Replace(s, "@GIT_URL", project.GitUrl, -1)
	jobtext = strings.Replace(jobtext, "@CICD_CONTROLLER", url, -1)
	jobtext = strings.Replace(jobtext, "@BRANCH", branch, -1)

	baseurl := "http://" + opsenv.JenkinsDockerIP + ":" + opsenv.JenkinsHttpPort

	jenkins, err := gojenkins.CreateJenkins(baseurl, u.Util.Props.GetString("jenkins.admin.name", ""), u.Util.Props.GetString("jenkins.admin.password", "")).Init()

	if err != nil {
		logrus.Errorf("Create Jenkins Client error %s", err.Error())
	}

	job, err := jenkins.CreateJob(jobtext, jobName)
	if err != nil {
		logrus.Errorf("Create job error %s", err.Error())
	}
	logrus.Debugf("Job created, %s\n" + job.GetName())

	return
}

func (u *Resource) findJobByID(jobid string) (job *entity.Job) {
	selector := make(bson.M)
	selector["_id"] = bson.ObjectIdHex(jobid)
	_, _, jobJson, _ := u.Dao.HandleQuery(jobCollection, selector, true, bson.M{}, 0, 1, "", "true")
	job = new(entity.Job)
	jobOut, _ := json.Marshal(jobJson)
	json.Unmarshal(jobOut, &job)
	return job
}

func (u *Resource) terminateJobs(projId string) (err error) {
	//query jobs with projId
	selector := bson.M{}
	selector["project"] = projId
	_, _, jobsJson, _ := u.Dao.HandleQuery(jobCollection, selector, false, bson.M{}, 0, 0, "", "true")
	jobs := make([]entity.Job, 0)
	jobsOut, _ := json.Marshal(jobsJson)
	json.Unmarshal(jobsOut, &jobs)

	//loop
	for _, job := range jobs {
		//terminate jobs
		err = u.terminateJob(job.Id)
		if err != nil {
			logrus.Debugf("Error terminating job,reason:%v,job id: %s", err, job.Id)
		}
	}
	return
}

//terminate project envs ,if ok , terminate job
func (u *Resource) terminateJob(jobId string) (err error) {
	//terminate project envs
	err = u.terminateProjectEnvs(jobId)
	if err != nil {
		logrus.Debugf("Error terminating project envs,reason:%v,job id: %s", err, jobId)
		return
	} else {
		//delete job in jenkins
		job := u.findJobByID(jobId)
		proj, _ := u.findProjectById(job.Project)
		u.deleteJenkinsJob(proj.OpsEnvId, job.Name)
		//terminate job
		selector := bson.M{}
		selector[ParamID] = bson.ObjectIdHex(jobId)
		change := bson.M{"status": entity.JOB_STATUS_TERMINATED}
		err = u.Dao.HandleUpdateByQueryPartial(jobCollection, selector, change)
	}
	return
}
