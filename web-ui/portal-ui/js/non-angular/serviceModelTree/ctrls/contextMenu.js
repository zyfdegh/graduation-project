function generateContext(event){
	var self = this;
	// hide other contextMenus
	$("#model-menu,#group-menu,#model-app-menu").hide();
	var source = $(event.currentTarget);
	var target = source.data("groupid").split("/").length == 2 ? "#model-menu" : "#group-menu";
	$(target).show();
	$(target).find("li:eq(0)").find("a").html($.i18n.prop("add_dependency"));
	$(target).find("li:eq(1)").find("a").html($.i18n.prop("rename_id"));
	$(target).find("li:eq(2)").find("a").html($.i18n.prop("remove_group"));
	source.contextmenu({
		target:target, 
		onItem: function(e,ev) {
			if(ev.target.tabIndex == 0){
				self.inputDepID(e.data("groupid"),e.data("grouptype"));
			}else if(ev.target.tabIndex == 1){
				self.inputNewID(e.data("groupid"),e.data("grouptype"));
			}else if(ev.target.tabIndex == 2){
				self.confirmDeleteGroup(e.data("groupid"),e.data("grouptype"));
			}
			$(target).hide();
		}
	});
}

function generateAppContext(event){
	var self = this;
	// hide other contextMenus
	$("#model-menu,#group-menu,#model-app-menu").hide();
	var source = $(event.currentTarget);
	var target = "#model-app-menu";
	$(target).show();
	$(target).find("li:eq(0)").find("a").html($.i18n.prop("app_details"));
	$(target).find("li:eq(1)").find("a").html($.i18n.prop("remove_app_form_group"));
	$(target).find("li:eq(2)").find("a").html($.i18n.prop("configuration"));
	source.contextmenu({
		target:target, 
		onItem: function(e,ev) {
			if(ev.target.tabIndex == 0){
				self.showGroupAppDetail(e);
			}else if(ev.target.tabIndex == 1){
				var appid = e.data("appid");
				var groupid = appid.substring(0,appid.lastIndexOf("/"));
				var realappid = self.idToSimple(appid);
				self.confirmDeleteAppFromGroup(groupid,realappid);
			}else if(ev.target.tabIndex == 2){
				var appid = e.data("appid");
				self.getCP(appid);
			}
			$(target).hide();
		}
	});
}

// add dependency
function inputDepID(p_gid,p_gtype){
	var self = this;
	
	var parentgroup = {
		"pgid" : p_gid,
		"pgtype" : p_gtype
	}
	var compiledTemplate = _.template(inputDepIDTemplate,{"parentgroup" :parentgroup});
	var dialogOption = $.extend({}, {
		width : 220,
		height: 150,
		modal : true,
		buttons : null,
		beforeclose : function(event, ui) {
			// reset the content.
			$(this).empty();
		},
		close: function(event){
			self.closeSecondDialog();
		}
	});
	$("#linker-dialog").empty();
	$("#linker-dialog").append(compiledTemplate).dialog(dialogOption);
	$("#linker-dialog").height("auto");
	$(".ui-dialog-titlebar").hide();
	
	//i18n
	$("#depid").attr("placeholder",$.i18n.prop("input_dependency_group_id"));
}

function addDependency(event){
	if(event.keyCode == 13){
		var self = this;

		var p_gid = $(event.currentTarget).data("pgid");
		var p_gtype = $(event.currentTarget).data("pgtype");
		var gid = $("#depid").val().trim();
		
		if(!self.groupIDIsValid(gid)){
			var alert = {
				"title" : $.i18n.prop("error_title"),
				"type" : "failed",
				"message" : $.i18n.prop("invalid_group_id"),
				"modaltype" : "notify"
			}
			self.nonAngularAlert(alert);
			return false;
		}
		
		if(gid.length > 0){
			if(self.isDepGroupIDDuplicated(p_gid,p_gtype,gid)){
				var alert = {
								"title" : $.i18n.prop("duplicated_group_id"),
								"type" : "failed",
								"message" : $.i18n.prop("group_id") + " '" + gid + "' "+ $.i18n.prop("is_duplicated_please_reinput_one"),
								"modaltype" : "notify"
							}
				self.nonAngularAlert(alert);
				return;
			}
			if(p_gid.split("/").length > 2){
				var dependencyid = "../" + gid;
				var _group = self.findGroupByPath(p_gid);
				if(_.isUndefined(_group.dependencies)){
					_group.dependencies = [];
				}
				// check dependency exist
				if (!_.contains(_group.dependencies, dependencyid)) {
					_group.dependencies.push(dependencyid);
				}
			}
			var newgroup = {
				"id": gid,
				"apps": []
			}
			var parentgroup = self.findParentGroup(p_gid,p_gtype,true);
			parentgroup.groups.push(newgroup);
			self.closeSecondDialog();
			self.drawGroupRelations();
		}
	}
}
//add dependency end

//rename
function inputNewID(gid,gtype){
	var self = this;

	var compiledTemplate,height;
	if(gtype != "template"){
		compiledTemplate = _.template(inputNewGroupIDTemplate,{"gid" :gid,"s_gid":self.idToSimple(gid),"gtype":gtype});
	}else{
		compiledTemplate = _.template(inputNewTemplateIDTemplate,{"gid" :gid,"s_gid":self.idToSimple(gid),"gtype":gtype});
	}
	var dialogOption = $.extend({}, {
		width : 220,
		height: 150,
		modal : true,
		buttons : null,
		beforeclose : function(event, ui) {
			// reset the content.
			$(this).empty();
		},
		close: function(event){
			self.closeSecondDialog();
		}
	});
	$("#linker-dialog").empty();
	$("#linker-dialog").append(compiledTemplate).dialog(dialogOption);
	$("#linker-dialog").height("auto");
	$(".ui-dialog-titlebar").hide();
	
	//i18n
	$("#newid").attr("placeholder",$.i18n.prop("input_new_group_id"));
}

function renameGroupID(event){
	if(event.keyCode == 13){
		var self = this;
		var gid = $(event.currentTarget).data("gid");
		var realgid = self.idToSimple(gid);
		var gtype = $(event.currentTarget).data("gtype");
		var newid;
		if(gtype != "template"){
			newid = $("#newid").val().trim();
		}else{
			newid = self.idPrefix(realgid) + "-" + $("#newid").val().trim();
		}
		
		if(!self.groupIDIsValid(newid)){
			var alert = {
				"title" : $.i18n.prop("error_title"),
				"type" : "failed",
				"message" : $.i18n.prop("invalid_group_id"),
				"modaltype" : "notify"
			}
			self.nonAngularAlert(alert);
			return false;
		}
		
		if(newid.length > 0){
			self.prepareToFindRelatedCP_Rename(gid,newid);
			
			if(gid.split("/").length == 2){
				self.selectedModel.id = self.idToPath(newid,"");
			}else{
				if(self.isNewGroupIDDuplicated(gid,gtype,newid)){
					var alert = {
									"title" : $.i18n.prop("duplicated_group_id"),
									"type" : "failed",
									"message" : $.i18n.prop("new_input_group_id") + " '" + newid + "' " + $.i18n.prop("is_duplicated_please_reinput_one"),
									"modaltype" : "notify"
								}
					self.nonAngularAlert(alert);
					return;
				}
				var pid = gid.substring(0,gid.lastIndexOf("/"));
				var targetgroup = self.findGroupByPath(gid);
				var parentgroup = self.findGroupByPath(pid);
				targetgroup.id = newid;
				_.each(parentgroup.groups,function(group){
					_.each(group.dependencies,function(dep,index){
						var old_dep_id = "../" + realgid;
						var new_dep_id = "../" + newid;
						if(dep == old_dep_id){
							group.dependencies.splice(index,1,new_dep_id); 
						}
					});
				});
			}
		
			self.closeSecondDialog();	
			self.drawGroupRelations();
		}	
	}
}

function prepareToFindRelatedCP_Rename(gid,newid){
	var self = this;
	
	var sgid = self.idToSGID(gid);
	var restURL = '/appConfigs?query={"service_group_id":"' + sgid + '"}';
	$.ajax({
		url: restURL,
		type: 'GET',
		dataType: 'json',
		async : false,
		success: function(data) {
			if(data.data.length > 0){
				self.findRelatedCP_Rename(gid,newid,data.data);
			}
		},
		error: function(error) {
			console.log(error.responseText);
		}
	});	
}

function findRelatedCP_Rename(gid,newid,allcps){
	var self = this;
	
	var group = self.findGroupByPath(gid);
	var newgid = _.initial(gid.split("/")).join("/") + "/" + newid;
	
	var relatedApps = [];
	if(newgid.split("/").length == 2){
		 var relatedApp = {
		    	"oldAppId" : gid,
		    	"newAppId" : newgid
		 }
		 relatedApps.push(relatedApp);
	}
	self.recordRelatedApps(group,gid,newgid,relatedApps);
	self.replaceAllCps(allcps,relatedApps);
}

function recordRelatedApps(group,gid,newgid,relatedApps){
	var self = this;
	if(_.isUndefined(group.groups) || group.groups == null){
		_.each(group.apps,function(app){
		    var appid = gid + "/" + app.id;
		    var newappid = newgid + "/" + app.id;
		    var relatedApp = {
		    		"oldAppId" : appid,
		    		"newAppId" : newappid
		    }
		    relatedApps.push(relatedApp);
		});
	}else{
		_.each(group.groups,function(subgroup){
			var old_subgid = gid + "/" + subgroup.id;
			var new_subgid = newgid + "/" + subgroup.id;
			self.recordRelatedApps(subgroup,old_subgid,new_subgid,relatedApps);
		});
	}
}

function replaceAllCps(allcps,relatedApps){
	var self = this;
	
	_.each(allcps,function(cp){
		var needUpdate = false;
		
		//check app_container_id
		var acid = cp.app_container_id;
		var index_acid = _.pluck(relatedApps,"oldAppId").indexOf(acid);
		if(index_acid>=0){
			cp.app_container_id = relatedApps[index_acid].newAppId;
			needUpdate = true;
		}
		
		//check service_group_id
		var sgid = cp.service_group_id;
		var index_sgid = _.pluck(relatedApps,"oldAppId").indexOf(sgid);
		if(index_sgid>=0){
			cp.service_group_id = relatedApps[index_sgid].newAppId;
			needUpdate = true;
		}
		
		//check configs
		var configs = cp.configurations;
		_.each(configs,function(config){
			var preconditions = config.preconditions;
			_.each(preconditions,function(precondition){
				var cons = precondition.condition.split(" ");
				var condition = cons[0];
				var index_condition = _.pluck(relatedApps,"oldAppId").indexOf(condition);
				if(index_condition>=0){
					cons[0] = relatedApps[index_condition].newAppId;
					precondition.condition = cons.join(" ");
					needUpdate = true;
				}
			});
			
			var steps = config.steps;
			_.each(steps,function(step){
				var execute = step.execute;
				_.each(relatedApps,function(app){
					var separator = "%" + app.oldAppId + ".";
					var newSep = "%" + app.newAppId + ".";
					step.execute = execute.split(separator).join(newSep);
					needUpdate = true;
				})
			});
		});
		
		//check notifies
		var notifies = cp.notifies;
		_.each(notifies,function(notify){
			var notify_path = notify.notify_path;
			var index_notify_path = _.pluck(relatedApps,"oldAppId").indexOf(notify_path);
			if(index_notify_path>=0){
				notify.notify_path = relatedApps[index_notify_path].newAppId;
				needUpdate = true;
			}
		});
		
		if(needUpdate){
			var restURL = '/appConfigs/'+ cp._id;
			$.ajax({
				url: restURL,
				type: "PUT",
				dataType: 'json',
				data: cp,
				success: function() {
					console.log("configuration package of "+cp.app_container_id+" has been updated.");
				},
				error: function(error) {
					console.log(error.responseText);
				}
			});
		}
	})
}
//rename end

//delete group
function confirmDeleteGroup(groupid,grouptype){
	var self = this;
	var messagetype = $.i18n.prop("group") + " ";
	if(grouptype == "template"){
		messagetype = $.i18n.prop("template") + " ";
	}
	var realgid = self.idToSimple(groupid);
	var message = messagetype + "'" + realgid + "' " + $.i18n.prop("and_all_dep_groups_will_be_del") + $.i18n.prop("are_you_sure");
	
	if(groupid.split("/").length > 3 && grouptype != "template"){
		var parentgroupid = groupid.substring(0,groupid.lastIndexOf("/"));
		var templateid = self.idToSimple(parentgroupid);
		var isOnlyOne = true;
		_.each(self["rootPool"+parentgroupid], function(group) {
			if (!group.flag && group.id == groupid) {
				isOnlyOne = false;
			}else if(group.flag && group.id != groupid){
				isOnlyOne = false;
			}
		});
		if(isOnlyOne){
			message = messagetype+ "'" + realgid + "' " 
					+  $.i18n.prop("and_all_dep_groups_will_be_del") 
					+ $.i18n.prop("btw") + " ' "+ messagetype + "'" + realgid + "' " 
					+ $.i18n.prop("is_the_only_one_dep_of_temp") 
					+ ", " + $.i18n.prop("if_you_del_this_group_the_template_will_be_also_deleted") 
					+ $.i18n.prop("are_you_sure");
		}
	}
	var alert = {
				"title" : $.i18n.prop("delete_confirm"),
				"type" : "confirm",
				"message" : message,
				"modaltype" : "dconfirm",
				"actions" : [
					"deleteGroup('" +groupid + "','" +grouptype + "')"
				]
	}
	self.nonAngularAlert(alert);
}

function deleteGroup(groupid,grouptype){
	var self = this;
	
	self.findRelatedAppToDeleteCP(groupid);
	
	self.deleteGroup_self(groupid,grouptype);
	var parentgroup = self.findParentGroup(groupid,grouptype,false);
	_.each(parentgroup.groups,function(group,index){
		if(!_.isUndefined(group.dependencies)){
			_.each(group.dependencies,function(dep,index){
				var realgid = self.idToSimple(groupid);
				var depid = "../"+realgid;
				if(dep == depid){
					group.dependencies.splice(index,1);
				}
			});
		}
	});
	if(groupid.split("/").length > 3 && parentgroup.groups.length == 0){
		var pgid = groupid.substring(0,groupid.lastIndexOf("/"));
		self.deleteGroup(pgid,"template");
	}
	self.drawGroupRelations();
}

function deleteGroup_self(groupid,grouptype){
	var self = this;
	var thisgroup = self.findGroupByPath(groupid);
	
	if(!_.isUndefined(thisgroup)){
		if(!_.isUndefined(thisgroup.dependencies)){
			_.each(thisgroup.dependencies,function(dep){
				var suffix = groupid.substring(0,groupid.lastIndexOf("/"));
				var depid = self.idToSimple(dep);
				var deppathid = suffix + "/" + depid;
				self.deleteGroup_self(deppathid,grouptype);
			});
		}
		var parentgroup = self.findParentGroup(groupid,grouptype,false);
		_.each(parentgroup.groups,function(group,index){
			if(group.id == self.idToSimple(groupid)){
				parentgroup.groups.splice(index,1);
			}
		});
	}
}

function findRelatedAppToDeleteCP(groupid){
	var self = this;
	
	var group = self.findGroupByPath(groupid);
	if(_.isUndefined(group.groups) || group.groups == null){
		_.each(group.apps,function(app){
		    var appid = groupid + "/" + app.id;
		   	self.deleteCP(appid);
		});
		
		if(!_.isUndefined(group.dependencies)){
			_.each(group.dependencies,function(dep){
				var suffix = groupid.substring(0,groupid.lastIndexOf("/"));
				var depid = self.idToSimple(dep);
				var deppathid = suffix + "/" + depid;
				self.findRelatedAppToDeleteCP(deppathid);
			});
		}
	}else{
		_.each(group.groups,function(subgroup){
			var subgid = groupid + "/" + subgroup.id;
			self.findRelatedAppToDeleteCP(subgid);
		});
	}
}
//delete group end

function saveAppInGroup(app) {
	var self = this;
    var linkerRepoPrefix = "linkerrepository:5000/";
	var dockerhubPrefix = "docker.io/" 
	app.cpus = Number(self.openedApp.cpus);
	app.mem = Number(self.openedApp.mem);
	app.instances = Number(self.openedApp.instances);
	app.cmd = self.openedApp.cmd;
	// app.container.docker.image = self.openedApp.container.docker.image;
	app.container.docker.image = self.radio.repoType=="linker"? linkerRepoPrefix + self.dockerImage.fromLinker+":"+self.imageTag.tag : dockerhubPrefix+self.dockerImage.fromDockerhub;
	app.container.docker.parameters = [];
	_.each(self.openedApp.container.docker.parameters,function(par){
		var parameter = {
			"key" : "env",
			"value" : par.vKey + "=" + par.vValue,
			"editable" : par.editable,
			"description" : par.description
		}
		app.container.docker.parameters.push(parameter);
	})
	
	app.container.volumes = [];
	_.each(self.openedApp.container.volumes,function(v){
		var volume = {
        		"containerPath": v.containerPath,
            "hostPath": v.hostPath,
            "mode": v.mode
        }
		app.container.volumes.push(volume);
	})
	
	if (self.openedApp.env.trim().length == 0) {
		$("#env-error").append($.i18n.prop("error_empty_env"));
		$("#env-error").show();
		return;
	}
	if (!self.isJson(self.openedApp.env)) {
		$("#env-error").append($.i18n.prop("error_invalid_env"));
		$("#env-error").show();
		return;
	}
	app.env = JSON.parse(self.openedApp.env);
	
	if(self.openedApp.exposePorts == "yes"){
		app.env.LINKER_EXPOSE_PORTS = "true";
	}else{
		delete app.env.LINKER_EXPOSE_PORTS;
	}
	
	var constraints = "";
	if ($("#constraints").val().trim().length > 0) {
		if (!self.isJson($("#constraints").val())) {
			$("#constraints-error").append($.i18n.prop("error_invalid_constraints"));
			$("#constraints-error").show();
			return;
		} else {
			constraints = JSON.parse($("#constraints").val());
			app.constraints = constraints;
		}
	}else {
		if (!_.isUndefined(app.constraints)) {
			delete app.constraints;
		}
	}

	self.drawGroupRelations();

	return true;
}

function confirmDeleteAppFromGroup(gid, aid){
	var self = this;
	
	var realgid = gid.substring(gid.lastIndexOf("/")+1);
	var alert = {
				"title" : $.i18n.prop("delete_confirm"),
				"type" : "confirm",
				"message" : $.i18n.prop("app") + " '" + aid + "'" + $.i18n.prop("will_be_del_from_group") + ", " + $.i18n.prop("cp_also_del") + $.i18n.prop("are_you_sure"),
				"modaltype" : "dconfirm",
				"actions" : [
					"deleteAppFromGroup('" +gid + "','" +aid + "')"
				]
	}
	self.nonAngularAlert(alert);
}

function deleteAppFromGroup(gid, aid) {
	var self = this;
	var _group = self.findGroupByPath(gid);
	var index = _.pluck(_group.apps, "id").indexOf(aid);
	_group.apps.splice(index, 1);
	self.deleteCP(gid+"/"+aid);
	self.drawGroupRelations();
}

//configuration package
function getCP(appid){
	var self = this;

	var restURL = '/appConfigs?query={"app_container_id":"' + appid + '"}';
		$.ajax({
			url: restURL,
			type: 'GET',
			dataType: 'json',
			success: function(data) {
				if(data.data.length>0){
					self.openedCP = data.data[0];
					self.openedCP.method = "PUT";
				}else{
					self.openedCP = {
						"app_container_id" : appid,
						"configurations": [],
//						"image": "",
						"notifies" : [],
						"service_group_id": self.idToSGID(appid),
						"method" : "POST"
					}
				}
				self.openCPDialog();
			},
			error: function(error) {
				var alert = {
					"title" : $.i18n.prop("get_data_fail"),
					"type" : "failed",
					"message" : $.i18n.prop("fail_to_get_cp"),
					"modaltype" : "notify"
				}
				self.nonAngularAlert(alert);
				console.log(error.responseText);
			}
		});
}

function showCPButtons(){
	var self = this;
	
	var buttonTemplate = _.template(cpButtonsTemplate,{currentstep:self.cp_current_step});
	$("#cp_buttons").empty();
	$("#cp_buttons").append(buttonTemplate);
}

function showCPHeader(){
	var self = this;
	
	var stepname;
	switch(self.cp_current_step){
		case 1 : stepname = $.i18n.prop("basic_info"); break;
		case 2 : stepname = $.i18n.prop("configurations"); break;
		case 3 : stepname = $.i18n.prop("notifies"); break;
	}
	var headerTemplate = _.template(cpHeaderTemplate,{stepname:stepname});
	$("#cp_header").empty();
	$("#cp_header").append(headerTemplate);
}

function wizardControl(currentstep,stepchange){
	this.cp_current_step = currentstep + stepchange;
	this.showCPSteps();
}

function saveCP(){
	var self = this;
	var cp = self.parseToCP();
	var method = self.openedCP.method;
	
	var restURL = '/appConfigs';
	if(method == "PUT"){
		restURL += "/" + self.openedCP._id;
	};
		$.ajax({
			url: restURL,
			type: method,
			dataType: 'json',
			data: cp,
			success: function() {
				var alert = {
					"title" : $.i18n.prop("success_to_save"),
					"type" : "success",
					"message" : $.i18n.prop("save_cp_success"),
					"modaltype" : "notify"
				}
				self.nonAngularAlert(alert);
			},
			error: function(error) {
				var alert = {
					"title" : $.i18n.prop("fail_to_save"),
					"type" : "failed",
					"message" : $.i18n.prop("save_cp_fail"),
					"modaltype" : "notify"
				}
				self.nonAngularAlert(alert);
				console.log(error.responseText);
			}
		});
}

function parseToCP(){
	var self = this;
	var cp = {
		  "service_group_id": "",
		  "app_container_id": "",
//		  "image": "",
		  "configurations": [],
		  "notifies": []
	};
	cp.service_group_id = self.openedCP.service_group_id;
	cp.app_container_id = self.openedCP.app_container_id;
//	cp.image = self.openedCP.image;
	cp.configurations = self.parseCPConfigs(self.openedCP.configurations);
	cp.notifies = self.parseCPNotifies(self.openedCP.notifies);
	return cp;
}

function parseCPConfigs(cps){
	var self = this;
	var configs = [];
	_.each(cps,function(cp){
		var config =  {
	      "name": cp.name,
	      "preconditions": [],
	      "steps": []
	    }
		config.preconditions = self.parseCPConfigCondition(cp.my_preconditions);
		config.steps = self.parseCPConfigStep(cp.my_steps,cp.hasConfigIPStep);
		configs.push(config);
	});
	return configs;
}

function parseCPConfigCondition(pres){
	var self = this;
	var preconditions = [];
	_.each(pres,function(pre){
		var value = pre.targetapp + " " + pre.operator + " " +pre.value;
		var precondition = {
          "condition": value
       };
       preconditions.push(precondition);
	});
	return preconditions;
}

function parseCPConfigStep(mysteps,hasConfigIPStep){
	var self = this;
	var steps = [];
	if(hasConfigIPStep){
		var restURL = location.protocol + '//' + location.host + '/portal-ui/conf/cp.json';
		$.ajax({
			url: restURL,
			type: 'GET',
			dataType: 'json',
			async : false,
			success: function(data) {
				var execute = data.pipework.split("*app_container_id*").join(self.openedCP.app_container_id);
				var firstStep = {
		          "config_type": "command",
		          "execute": execute,
		          "scope": "ONLYME"
		       	};
		       	steps.push(firstStep);
			},
			error: function(error) {
				console.log(error.responseText);
			}
		});	
	}
	_.each(mysteps,function(mystep){
		var execute = mystep.execute.command + " " + self.parseStepParams(mystep.execute.params);
		var step = {
          "config_type": mystep.config_type,
          "execute": execute,
          "scope": mystep.scope
        }
		steps.push(step);
	});
	return steps;
}

function parseStepParams(pars){
	var self = this;
	var tmps = [];
	_.each(pars,function(par){
		if(par.name.trim().length>0&&par.value.trim().length>0){
			var name = "--" + par.name;
			var value = par.value;
			tmps.push(name);
			tmps.push(value);
		}
	});
	var params = tmps.join(" ");
	return params;
}

function parseCPNotifies(mynotifies){
	var self = this;
	var notifies = [];
	_.each(mynotifies,function(mynotify){
		var notify ={
	      "notify_path": mynotify.notify_path,
	      "scope": mynotify.scope
	    };
       notifies.push(notify);
	});
	return notifies;
}

function deleteCP(appid){
	var self = this;

	var restURL = '/appConfigs?query={"app_container_id":"' + appid + '"}';
		$.ajax({
			url: restURL,
			type: 'DELETE',
			dataType: 'json',
			success: function(data) {
//				var alert = {
//					"title" : "Delete Success",
//					"type" : "success",
//					"message" : "Success to delete configuration package of " + appid,
//					"modaltype" : "notify"
//				}
//				self.nonAngularAlert(alert);
				console.log("success to delete cp for "+appid);
			},
			error: function(error) {
//				var alert = {
//					"title" : "Delete Failed",
//					"type" : "failed",
//					"message" : "Failed to delete configuration package of " + appid,
//					"modaltype" : "notify"
//				}
//				self.nonAngularAlert(alert);
				console.log(error.responseText);
			}
		});
}
//configuration package end