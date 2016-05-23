package documents

import (
	"github.com/emicklei/go-restful"
	"linkernetworks.com/linker_cicd_platform/persistence/dao"
	"linkernetworks.com/linker_cicd_platform/util"
)

var ParamID = dao.ParamID // mongo id parameter

// Creates and adds documents resource to container
func Register(dao *dao.Dao, container *restful.Container, cors bool, util *linker_util.Util) {
	dc := Resource{Dao: dao, Util: util}
	dc.Register(container, cors)
}

// Adds documents resource to container
func (d Resource) Register(container *restful.Container, cors bool) {
	wss := []*restful.WebService{}
	wss = append(wss,
		d.JobsWebService(),
		d.OpsEnvWebService(),
		d.ProjectsWebService(),
		d.ProjectEnvWebService(),
		d.DockerWebService(),
		d.AccountsWebService())

	//ws := d.WebService()
	for _, ws := range wss {
		// Cross Origin Resource Sharing filter
		if cors {
			corsRule := restful.CrossOriginResourceSharing{ExposeHeaders: []string{"Content-Type"}, CookiesAllowed: false, Container: container}
			ws.Filter(corsRule.Filter)
		}
		// Add webservice to container
		container.Add(ws)
	}
}
