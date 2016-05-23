package documents

import (
	"encoding/json"
	"fmt"
	"strings"

	"linkernetworks.com/linker_cicd_platform/persistence/entity"
	conentity "linkernetworks.com/linker_common_lib/rest/response"

	"github.com/Sirupsen/logrus"
	"github.com/compose/mejson"
	"gopkg.in/mgo.v2/bson"
)

func ConvertJobToBson(job entity.Job) (document bson.M) {
	document, _ = ConvertToBson(job)
	return
}

func ConvertOpsEnvToBson(env entity.OpsEnv) (document bson.M) {
	document, _ = ConvertToBson(env)
	return
}

func ConvertDockerFileToBson(dockfile entity.DockerFile) (document bson.M) {
	document, _ = ConvertToBson(dockfile)
	return
}

func ConvertJobEnvToBson(jobenv entity.JobEnv) (document bson.M) {
	document, _ = ConvertToBson(jobenv)
	return
}

func ConvertProjectToBson(projInfo entity.Project) (document bson.M) {
	document, _ = ConvertToBson(projInfo)
	return
}

func ConvertProjectEnvToBson(projEnv entity.ProjectEnv) (document bson.M) {
	document, _ = ConvertToBson(projEnv)
	return
}

func ConvertAccountsToBson(accounts entity.Accounts) (document bson.M) {
	document, _ = ConvertToBson(accounts)
	return
}

func ConvertToBson(obj interface{}) (document bson.M, err error) {
	b, err := json.Marshal(obj)
	if err != nil {
		return
	}
	reader := strings.NewReader(string(b))
	decoder := json.NewDecoder(reader)
	err = decoder.Decode(&document)
	if err != nil {
		return
	}
	document, err = mejson.Unmarshal(document)
	return
}

func GetDataFromResponse(data []byte) (jsonout []byte, err error) {
	// create serviceGroupOrder
	var response *conentity.Response
	response = new(conentity.Response)
	err = json.Unmarshal(data, &response)
	if err != nil {
		return
	}

	var jmap map[string]interface{}

	switch response.Data.(type) {
	case []interface{}:
		jmaps := response.Data.([]interface{})
		jmap = jmaps[0].(map[string]interface{})
	case interface{}:
		jmap = response.Data.(map[string]interface{})
	}
	jsonout, err = json.Marshal(jmap)
	if err != nil {
		fmt.Println(err)
		return
	}
	return
}

//get mapping
func (u *Resource) getDfAtfsMapping(proj_id string) (result map[string][]entity.Artifact) {
	//query artifacts[] from projectCollection by proj_id
	selector := bson.M{}
	selector[ParamID] = bson.ObjectIdHex(proj_id)
	_, _, document, err := u.Dao.HandleQuery(projectCollection, selector, true, bson.M{}, 0, 0, "", "true")
	if err != nil {
		logrus.Debugf("Cannot query artifacts,this maybe caused by nonexistance proj_id")
		return
	}
	gpi := new(entity.Project)
	data, err := json.Marshal(document)
	if err != nil {
		logrus.Debugf("Cannot marshal document")
		return
	}
	err = json.Unmarshal(data, &gpi)
	if err != nil {
		logrus.Debugf("Cannot unmarshal document")
		return
	}
	//array of artifact
	var artifacts []entity.Artifact = gpi.Artifacts

	//load data to a map
	result = make(map[string][]entity.Artifact)

	//	logrus.Debugf("proj_id:%v" + proj_id)
	for _, artifact := range artifacts {
		var dockerfileArray []string = artifact.DockerfileIds
		//		logrus.Debugf("Len dockerfileArray:%s", len(dockerfileArray))
		for _, dockerfileid := range dockerfileArray {
			artifactArr := result[dockerfileid][:]
			artifactArr = append(artifactArr, artifact)
			result[dockerfileid] = artifactArr[:]
		}
	}

	//print
	fmt.Println("Printing map:")
	logrus.Debugf("Printing map(dfId-AtfIds):")
	for id, artifactArr := range result {
		fmt.Println("\n" + id)
		for _, artifact := range artifactArr {
			fmt.Printf(artifact.Id + "\t")
		}
	}
	fmt.Print("\n")
	return
}

func checkZipFile(filename string) (result bool) {
	result = strings.HasSuffix(filename, "zip")
	return result
}

func checkTarGzFile(filename string) (result bool) {
	result = strings.HasSuffix(filename, "tar.gz")
	return result
}
