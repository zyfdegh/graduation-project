package documents

import (
	"github.com/emicklei/go-restful"
	"linkernetworks.com/linker_common_lib/persistence/dao"
	"linkernetworks.com/linker_controller/services"
)

var ParamID = dao.ParamID // mongo id parameter

// Creates and adds documents resource to container
func Register(container *restful.Container, cors bool) {
	dc := Resource{}
	dc.Register(container, cors)
}

// Adds documents resource to container
func (p Resource) Register(container *restful.Container, cors bool) {
	wss := []*restful.WebService{}
	wss = append(wss,
		p.ServiceGroupOrderWebService(),
		p.ServiceGroupWebService(),
		p.ServiceGroupInstanceWebService(),
		p.AppWebService(),
		p.AppInstanceWebService(),
		p.AppPackageWebService(),
		p.IPWebService(),
		p.IPPoolWebService(),
		p.ServiceOfferingWebService(),
		p.BillingWebService(),
		p.MetrixWebService(),
		p.NotificationWebService(),
		p.UserAccountWebService(),
		p.UserAccountBalanceWebService(),
		p.BundleWebService(),
		p.AlertWebService(),
		p.RepairPolicyWebService())

	//ws := d.WebService()
	for _, ws := range wss {
		// Cross Origin Resource Sharing filter
		if cors {
			corsRule := restful.CrossOriginResourceSharing{ExposeHeaders: []string{"Content-Type"},
				CookiesAllowed: false, Container: container}
			ws.Filter(corsRule.Filter)
		}
		// Add webservice to container
		container.Add(ws)
	}

	services.GetBillingService()

}
