function allowDrop(event) {
    event.preventDefault();
}

function dropHere(event){
	var self = this;
	var dragtype = event.dataTransfer.getData("dragType");
	if(dragtype == "app"){
		self.dropAppHere(event);
	}else if(dragtype == "template"){
		self.dropTemplateHere(event);
	}
}

function dropAppHere(event) {
	var appid = event.dataTransfer.getData("dragAppID");
	if(appid.length>0){
		var self = this;
	    event.preventDefault();
	    var groupid = $(event.currentTarget).data("groupid");
	   
	    var _group = self.findGroupByPath(groupid);
	    var existedApp = _.find(_group.apps,function(_app){
	    		return _app.id == appid;
	    });
	    
	    if(!_.isUndefined(existedApp)){
	    	existedApp.instances += 1;
	    }else{
	    	var index = _.pluck(apps,"id").indexOf(appid);
	    	var app = $.extend({},apps[index]);
	   	 	_group.apps.push(app);
	    }
	
	    self.drawGroupRelations();
	}
}

function dropTemplateHere(event) {
	var templateid = event.dataTransfer.getData("dragTemplateID");
	if(templateid.length>0){
		var self = this;
	    event.preventDefault();
	    var index = _.pluck(availableTemplates,"id").indexOf(templateid);
	    var template = $.extend(true,{},availableTemplates[index]);
	    var realtemplateid = self.idToSimple(templateid);
	    template.id = realtemplateid;
	    var groupid = $(event.currentTarget).data("groupid");
	    var grouptype = $(event.currentTarget).data("grouptype");
	   	
	   	if(self.isDepGroupIDDuplicated(groupid,grouptype,realtemplateid)){
	   		var changetid = realtemplateid;
			for(var i=0;i>=0;i++){
				changetid = realtemplateid + i;
				if(!self.isDepGroupIDDuplicated(groupid,grouptype,changetid)){
					break;
				}
			}
		   	template.id = changetid;
	   	}
	   	
	   	//set billing to false for all sub templates in this template
	   	template.billing = true;
	   	self.notAllowBillingForSubTemplate(template);
	   	
	    //add template to groups
	    var parentgroup = self.findParentGroup(groupid,grouptype,true);
	    parentgroup.groups.push(template);
	    self.allocateImageToApp(parentgroup);
	    
	    //add depedency
	    var _group = self.findGroupByPath(groupid);
	    if(_.isUndefined(_group.groups) || _group.groups == null){
		   	var depid = "../" + template.id;
		    if(_.isUndefined( _group.dependencies) ||  _group.dependencies == null){
		    	 	_group.dependencies = [];
		    }
		    _group.dependencies.push(depid);
		}
	    
		self.loadTemplateCP(groupid,grouptype,templateid,template.id);
	    self.drawGroupRelations();
	}
}

function notAllowBillingForSubTemplate(group){
	var self = this;
	_.each(group.groups,function(_group){
		if(!_.isUndefined(_group.groups) && _group.groups != null){
		   _group.billing = false;
		   self.notAllowBillingForSubTemplate(_group);
		}
	});
}

function loadTemplateCP(groupid,grouptype,templateid,newtemplateid){
	var self = this;
	newtemplateid = "/" + newtemplateid;
	
	var parentGroupPathID;
	if(grouptype == "template"){
		parentGroupPathID = groupid;
	}else{
		parentGroupPathID = groupid.substring(0,groupid.lastIndexOf("/"));
	}
	
	var index = _.pluck(availableTemplates,"id").indexOf(templateid);
	var template = $.extend(true,{},availableTemplates[index]);
	_.each(template.groups,function(group){
		self.getAppsInTemplate(group,templateid,newtemplateid,parentGroupPathID);
	});
	
//	var restURL = '/appConfigs?query={"service_group_id":"' + templateid + '"}';
//	$.ajax({
//		url: restURL,
//		type: 'GET',
//		dataType: 'json',
//		success: function(data) {
//			_.each(data.data,function(cp){
//				self.saveTemplateCP(parentGroupPathID,cp);
//			});
//		},
//		error: function(error) {
//			console.log(error.responseText);
//		}
//	});
}

function getAppsInTemplate(group,parentid,newpid,parentGroupPathID){
	var self = this;
	if(_.isUndefined(group.groups) || group.groups == null){
		_.each(group.apps,function(app){
		    var apppathid = parentid + "/" + group.id + "/" + app.id;
		    
		    var restURL = '/appConfigs?query={"app_container_id":"' + apppathid + '"}';
		    	$.ajax({
				url: restURL,
				type: 'GET',
				dataType: 'json',
				success: function(data) {
					_.each(data.data,function(cp){
						self.saveTemplateCP(parentGroupPathID,cp,newpid);
					});
				},
				error: function(error) {
					console.log(error.responseText);
				}
			});
		});
	}else{
		parentid = parentid + "/" + group.id;
		_.each(group.groups,function(subgroup){
			self.getAppsInTemplate(subgroup,parentid,newpid,parentGroupPathID);
		});
	}	
}

function saveTemplateCP(parentGroupPathID,cp,newpid){
	var self = this;
	
	var sgid = self.idToSGID(parentGroupPathID);
	var appid = parentGroupPathID + cp.app_container_id.split(cp.service_group_id).join(newpid);
	
	var newcp = {
		"app_container_id" : appid,
		"configurations": [],
		"image": cp.image,
		"notifies" : [],
		"service_group_id": sgid,
	};
			
	_.each(cp.configurations,function(configuration){
		var config =  {
	      "name": configuration.name,
	      "preconditions": [],
	      "steps": []
	    }
		
		_.each(configuration.preconditions,function(precondition){
			var value = precondition.condition;
			if(value.indexOf(cp.service_group_id) == 0){
				value = parentGroupPathID + newpid + value.substring(cp.service_group_id.length);
			};
			var condition = {
				condition: value
			};
			config.preconditions.push(condition);
		});
		
		_.each(configuration.steps,function(step){
			var value = step.execute;
			value = value.split("%"+cp.service_group_id).join("%"+parentGroupPathID+newpid);
			var newstep = {
				config_type: step.config_type,
				execute: value,
				scope: step.scope
			};
			config.steps.push(newstep);
		});
		
		newcp.configurations.push(config);
	});
	
	_.each(cp.notifies,function(notify){
		var value = notify.notify_path;
		if(value.indexOf(cp.service_group_id) == 0){
			value = parentGroupPathID + newpid + value.substring(cp.service_group_id.length);
		};
		var newnotify = {
			notify_path: value,
			scope: notify.scope
		};
		newcp.notifies.push(newnotify);
	});
	
	var restURL = '/appConfigs';
	$.ajax({
		url: restURL,
		type: "POST",
		dataType: 'json',
		data: newcp,
		success: function() {
			console.log("configuration package of "+newcp.app_container_id+" has been saved.");
		},
		error: function(error) {
			console.log(error.responseText);
		}
	});
}

function finishDeleteApp(event) {
	var appid = event.dataTransfer.getData("dropAppID");
	if(appid.length>0){
		var self = this;
	    event.preventDefault();
	    var groupid = appid.substring(0,appid.lastIndexOf("/"));    
	    var realappid = appid.substring(appid.lastIndexOf("/")+1);
	    var _group = self.findGroupByPath(groupid);
	    var existedApp = _.find(_group.apps,function(_app){
	    		return _app.id == realappid;
	    });
	    
	    if(existedApp.instances > 1){
	    	existedApp.instances -= 1;
	    	self.drawGroupRelations();
	    }else{
	    	self.confirmDeleteAppFromGroup(groupid,realappid);
	    }
	}
}

function startDeleteApp(event){
	simulate($(event.target)[0], 'mouseup')
	var appid = $(event.target).data("appid");
	event.dataTransfer.setData("dropAppID", appid);
}