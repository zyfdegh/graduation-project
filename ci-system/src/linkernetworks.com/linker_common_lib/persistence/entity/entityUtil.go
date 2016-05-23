package entity

import (
	"errors"
	"strings"
)

func findApp(rg *RefinedGroup, appId string) (returnApp *RefinedApp) {
	appPath := strings.Split(appId, "/")
	if rg.RefinedGroups != nil {
		for j := range rg.RefinedGroups {
			group := &rg.RefinedGroups[j]
			if group.Id == appPath[1] {
				returnApp = findApp(group, strings.Join(appPath[1:], "/"))
				return
			}
		}
	}
	if rg.RefinedApps != nil {
		for k := range rg.RefinedApps {
			app := &rg.RefinedApps[k]
			if app.Id == appPath[1] {
				returnApp = app
				return
			}
		}
	}
	return
}

// FindAppInSgi parses ServiceGroupInstance and returns,
// If successful, a RefinedApp
func FindAppInSgi(sgi *ServiceGroupInstance, appId string) (returnApp *RefinedApp,
	err error) {

	appPath := strings.Split(appId, "/")[1:]
	if appLen := len(appPath); appLen < 3 {
		//TODO: what is the meaning of this?
		// fmt.Println("TODO: what is the meaning of this?")
	} else {
		groupId := appPath[1]
		for i := range sgi.Groups {
			group := &sgi.Groups[i]
			if group.Id == groupId {
				returnApp = findApp(group, strings.Join(appPath[1:], "/"))
				return
			}
		}
	}
	err = errors.New("can't find app in service group instance!")
	return
}

// GetAppInstanceIdsFromGroupInstance parses ServiceGroupInstance and returns,
// If successful,get instancesIds.
func GetAppInstanceIdsFromGroupInstance(sgi *ServiceGroupInstance,
	appId string) (instancesIds []string) {

	app, err := FindAppInSgi(sgi, appId)
	if err != nil {
		instancesIds = []string{}
		return
	}
	instancesIds = app.InstanceIds
	return
}

func getAppFromGroup(gp *Group, appId string) (app *App) {
	appPath := strings.Split(appId, "/")
	if gp.Groups != nil {
		for i := range gp.Groups {
			group := &gp.Groups[i]
			if group.Id == appPath[1] {
				app = getAppFromGroup(group, strings.Join(appPath[1:], "/"))
				return
			}
		}
	}
	if gp.Apps != nil {
		for i := range gp.Apps {
			a := &gp.Apps[i]
			if a.Id == appPath[1] {
				app = a
				return
			}
		}
	}
	return
}

// GetAppFromServiceGroup parses ServiceGroup and returns,
// If successful,get App.
func GetAppFromServiceGroup(sg *ServiceGroup, appId string) (app *App, err error) {
	appPath := strings.Split(appId, "/")[1:]
	if appLen := len(appPath); appLen < 3 {
		//TODO: what is the meaning of this?
		// fmt.Println("TODO: what is the meaning of this?")
	} else {
		groupId := appPath[1]
		for i := range sg.Groups {
			group := &sg.Groups[i]
			if group.Id == groupId {
				app = getAppFromGroup(group, strings.Join(appPath[1:], "/"))
				if app == nil {
					err = errors.New("can't find app in service group")
				}
				return
			}
		}
	}
	err = errors.New("can't find app in service group")
	return
}

// GetMarathonAppIdFromSgo parses ServiceGroupOrder and returns,
// If successful,get marathonAppId.
func GetMarathonAppIdFromSgo(sgo *ServiceGroupOrder,
	oldAppPath string) (marathonAppId string) {
	// reset appContainerId
	appPath := strings.Split(oldAppPath, "/")
	appId := append(strings.Split(sgo.MarathonGroupId, "/"), appPath[2:]...)
	marathonAppId = strings.Join(appId, "/")
	return
}
