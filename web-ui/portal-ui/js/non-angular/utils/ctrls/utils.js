//for draw tree

function idToSimple(id) {
	return id.substring(id.lastIndexOf("/") + 1);
}

function idToPath(id,pid){
	return pid + "/" + id;
}

function idToSGID(id){
	var paths = id.split("/");
	var sgid = paths[0] + "/" + paths[1]; 
	return sgid;
}

function idPrefix(id) {
	return id.substring(0,id.indexOf("-") );
}

function idSuffix(id) {
	return id.substring(id.indexOf("-")+1 );
}

//modal
function clearModalMask(){
	$(".modal-backdrop").remove();
}

//allocate image
function allocateImageToApp(group){
	var self = this;
	if(_.isUndefined(group.groups) || group.groups == null){
		_.each(group.apps,function(app){
		    if(app.id == "nginx"){
		    	app.imageSrc = "images/products/designer/app/nginx.png";
		    }else if(app.id == "zookeeper"){
		    	app.imageSrc = "images/products/designer/app/zookeeper.png";
		    }else if(app.id == "haproxy"){
		    	app.imageSrc = "images/products/designer/app/haproxy.png";
		    }else if(app.id.indexOf("mysql")>=0){
		    	app.imageSrc = "images/products/designer/app/mysql.png";
		   	}else{
		    	app.imageSrc = "images/products/designer/app/default.png";
		    }
		});
	}else{
		_.each(group.groups,function(subgroup){
			self.allocateImageToApp(subgroup);
		});
	}	
}

function transferModelToTree(target) {
	var self = this;
	
	self.rootPool = [];
	self.groupPool = [];
	
	var groups = target.groups;
	var rootPathID;
	if (_.isUndefined(target.service_group_id)) {
		rootPathID = target.id;
	} else {
		rootPathID = target.service_group_id;
	}
	
	_.each(groups, function(group) {
		if(_.isUndefined(group.groups) || group.groups == null){
			var groupid = group.id;
			var group_path_id = self.idToPath(groupid,rootPathID);
			var deps = group.dependencies;
			var apps = group.apps;
			
			var thisGroup = self.generateGroup(self.groupPool, self.rootPool, groupid, group_path_id, false, apps, true);
			_.each(deps, function(dep) {
				var depid = self.idToSimple(dep);
				var dep_path_id = self.idToPath(depid,rootPathID);
				self.generateGroup(self.groupPool, self.rootPool, depid, dep_path_id, false,[], false, thisGroup);
			});
		}else{
			var groupid = group.id;
			var group_path_id = self.idToPath(groupid,rootPathID);
			
			var thisGroup = self.generateGroup(self.groupPool,self.rootPool,groupid, group_path_id, true, [], true);
			self.transferTemplateToSubTree(group.groups,thisGroup);
		}	
	});

	var root_sid, root_id, treetype;
	if (_.isUndefined(target.service_group_id)) {
		root_sid = self.idToSimple(rootPathID);
		root_id = rootPathID;
		treetype = "model";
	} else {
		root_sid = self.idToSimple(target.service_group_id);
		root_id = target.service_group_id;
		treetype = "instance";
	}
	var root = {
		"id": root_id,
		"data": {
			"pathid": root_id,
			"id": root_sid,
			"apps": [],
			"isTemplate": true,
			"imageSrc" : target.imageSrc
		},
		"children": []
	};
	_.each(self.rootPool, function(group) {
		if (group.flag) {
			var groupid = group.id;
			var groupindex = _.pluck(self.groupPool, "id").indexOf(groupid);
			var _group = self.groupPool[groupindex];
			root.children.push(_group);
		}
	});


	self.setLevel_getLevelHeight(root, treetype);
	self.setCanvasHeight();
	self.calPositionY(root, treetype);
	return root;
}

function transferTemplateToSubTree(templateGroups,templateNode){
	var self = this;
	
	var rootPathID = templateNode.id;
	
	var tempRootPoolName = "rootPool"+rootPathID;
	var tempGroupPoolName = "groupPool"+rootPathID;
	
	self[tempRootPoolName] = [];
	self[tempGroupPoolName] = [];
	
	_.each(templateGroups, function(group) {
		if(_.isUndefined(group.groups) || group.groups == null){
			var groupid = group.id;
			var group_path_id = self.idToPath(groupid,rootPathID);
			var deps = group.dependencies;
			var apps = group.apps;
			
			var thisGroup = self.generateGroup(self[tempGroupPoolName],self[tempRootPoolName],groupid, group_path_id, false, apps, true);
			_.each(deps, function(dep) {
				var depid = self.idToSimple(dep);
				var dep_path_id = self.idToPath(depid,rootPathID);
				self.generateGroup(self[tempGroupPoolName],self[tempRootPoolName],depid, dep_path_id, false,[], false, thisGroup);
			});
		}else{
			var groupid = group.id;
			var group_path_id = self.idToPath(groupid,rootPathID);
			
			var thisGroup = self.generateGroup(self[tempGroupPoolName],self[tempRootPoolName],groupid, group_path_id, true, [], true);
			self.transferTemplateToSubTree(group.groups,thisGroup);
		}	
	});

	_.each(self[tempRootPoolName], function(group) {
		if (group.flag) {
			var groupid = group.id;
			var groupindex = _.pluck(self[tempGroupPoolName], "id").indexOf(groupid);
			var _group = self[tempGroupPoolName][groupindex];
			templateNode.children.push(_group);
		}
	});
}

function generateGroup(groupPool,rootPool, id, pathid, isTemplate, apps, isParent, parentGroup) {
	var thisGroup;
	var index = _.pluck(groupPool, "id").indexOf(pathid);
	if (index >= 0) {
		thisGroup = groupPool[index];
		if(thisGroup.data.isTemplate != isTemplate){
			thisGroup.data.isTemplate = thisGroup.data.isTemplate || isTemplate;
		}
		if(thisGroup.data.isTemplate){
			thisGroup.data.imageSrc = "images/products/Appserver.png";
		}
	} else {
		thisGroup = {
			"id": pathid,
			"data": {
				"pathid": pathid,
				"id": id,
				"isTemplate" : isTemplate,
				"apps": []
			},
			"children": []
		};
		if(thisGroup.data.isTemplate){
			thisGroup.data.imageSrc = "images/products/Appserver.png";
		}
		groupPool.push(thisGroup);
	}
	if (isParent) {
		thisGroup.data.apps = apps;
	}else{
		parentGroup.children.push(thisGroup);
	}
	self.insertToRootPool(rootPool,thisGroup.id, isParent);
	
	return thisGroup;
}

function insertToRootPool(rootPool,id, maybeRoot) {
	var self = this;
	var index = _.pluck(rootPool, "id").indexOf(id);
	if (index < 0) {
		var item = {
			"id": id,
			"flag": maybeRoot
		};
		rootPool.push(item);
	} else {
		var thisGroup = rootPool[index];
		thisGroup.flag = thisGroup.flag && maybeRoot;
	}
}

//re-cal position
function setLevel_getLevelHeight(root, treetype) {
	root.data.level = 1;
	var h_node = root.data.apps.length * 104 > 0 ? root.data.apps.length * 104 : 104;
	self.levelHeights = [];
	var _height = {
		"level": root.data.level,
		"height": h_node,
		"drawnFrom": 0
	}
	self.levelHeights.push(_height);

	self.setNextLevel(root.children, root.data.level + 1, treetype);
}

function setNextLevel(children, level, treetype) {
	var self = this;
	_.each(children, function(child) {
		child.data.level = level;
		var appslen;
		if (treetype == "model") {
			appslen = child.data.apps.length > 0 ? child.data.apps.length : 1;
		} else {
			appslen = 0;
			_.each(child.data.apps, function(app) {
				appslen = appslen + app.instances;
			});
			appslen = appslen > 0 ? appslen : 1;
		}

		var h_node = appslen * 104;
		var index = _.pluck(self.levelHeights, "level").indexOf(level);
		if (index >= 0) {
			self.levelHeights[index].height = self.levelHeights[index].height + 100 + h_node;
		} else {
			var _height = {
				"level": level,
				"height": h_node,
				"drawnFrom": 0
			}
			self.levelHeights.push(_height);
		}
		self.setNextLevel(child.children, level + 1, treetype);
	});
}

function calPositionY(root, treetype) {
	var self = this;
	self.y_positions = [];
	
	var y = 0;
	var pos = {
		"id": root.id,
		"y": y,

	};
	self.y_positions.push(pos);
	self.calChildPos(root.children, treetype,y);
}

function calChildPos(children, treetype,base_y) {
	var self = this;
	_.each(children, function(child) {
		var level = child.data.level;
		var nodeH;
		if (treetype == "model") {
			nodeH = child.data.apps.length * 104 > 0 ? child.data.apps.length * 104 : 104;
		} else {
			var appslen = 0;
			_.each(child.data.apps, function(app) {
				appslen = appslen + app.instances;
			});
			appslen = appslen > 0 ? appslen : 1;
			nodeH = appslen * 104;
		}
		var levelH = _.find(self.levelHeights, function(lh) {
			return lh.level == level;
		});
		var wholeHeight = levelH.height;
		var drawnFrom = levelH.drawnFrom;
		var y = base_y - (wholeHeight / 2) + drawnFrom + (nodeH / 2);
		levelH.drawnFrom += nodeH + 100;
		var pos = {
			"id": child.id,
			"y": y,

		};
		self.y_positions.push(pos);
		self.calChildPos(child.children, treetype,base_y);
	});
}

function setCanvasHeight() {
	var self = this;
	var height = _.max(_.pluck(self.levelHeights, "height"));
	if (self.modelDesignerHeight < height + 300) {
		self.modelDesignerHeight = height + 300;
	}
}
//re-cal position end

//draw tree end

//for find group
function findGroupByPath(groupPathID,fromInstance){
	var self = this;
	var group_paths = groupPathID.split("/");
	var target_group = self.selectedModel;
	if(fromInstance == true){
		target_group = self.selectedService;
	}
	for(var i=2;i<group_paths.length;i++){
		var groupid = group_paths[i];
		target_group = self.findGroupByID(target_group,groupid);
	}
	return target_group;
}

function findGroupByID(parentgroup,groupid){
	return _.find(parentgroup.groups,function(group){
		return group.id == groupid;
	});
}

function isDepGroupIDDuplicated(p_gid,p_gtype,gid){
	var self = this;
	var parentgroup = self.findParentGroup(p_gid,p_gtype,true);
	return _.pluck(parentgroup.groups, "id").indexOf(gid) >= 0 ;
}

function isNewGroupIDDuplicated(p_gid,p_gtype,gid){
	var self = this;
	if(self.idToSimple(p_gid) == gid){
		return false;
	}
	var parentgroup = self.findParentGroup(p_gid,p_gtype,false);
	return _.pluck(parentgroup.groups, "id").indexOf(gid) >= 0 ;
}

function findParentGroup(p_gid,p_gtype,addDep){
	var self = this;
	var parentgroup;
	if(p_gid.split("/").length == 2){
		parentgroup = self.selectedModel;
	}else if(p_gtype == "normalgroup" || !addDep){
		p_gid = p_gid.substring(0,p_gid.lastIndexOf("/"));
		parentgroup = self.findGroupByPath(p_gid);
	}else if(addDep){
		parentgroup = self.findGroupByPath(p_gid);
	}
	return parentgroup;
}

//for find group end

//actions app and group
function findAppInSelectedModel(gid, aid,fromInstance) {
	var self = this;
	var _group = self.findGroupByPath(gid,fromInstance);
	var _app = _.find(_group.apps, function(app) {
		return app.id == aid;
	});
	return _app;
}

function closeSecondDialog() {
	$("#linker-dialog").dialog('close');
	$(".ui-widget-overlay").css("z-index", 100);
}

function replaceStr(source, oid, nid) {
	var length = oid.length;
	var suffix = source.substring(length);
	var target = nid + suffix;
	return target;
}

function isJson(str) {
	try {
		JSON.parse(str);
	} catch (e) {
		return false;
	}
	return true;
}

//alert bootstrap modal
function nonAngularAlert(alert){
	this.clearModalMask();
	var iconmap = {
		"success" : "ok",
		"failed" : "remove",
		"confirm" : "question"
	}
	
	alert.sign = iconmap[alert.type];
	var compiledTemplate = _.template(nonAngularAlertTemplate, {
		"alert": alert
	});
	$("#linker-alert").empty();
	$("#linker-alert").append(compiledTemplate);
	$("#linker-alert").modal("show");
}

//validate id
function groupIDIsValid(gid){
	var regExp = /^(([a-z0-9]|[a-z0-9][a-z0-9\\-]*[a-z0-9])\\.)*([a-z0-9]|[a-z0-9][a-z0-9\\-]*[a-z0-9])$/;
	return regExp.test(gid);
}
