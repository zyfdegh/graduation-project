function openAppTree(event,from){
	var self = this;
	event.stopPropagation();
	$("#apptreearea").remove();
	var top,left,width;
	if(from == "notify"){
		top = $(event.currentTarget).next().offset().top + 35;
		left = $(event.currentTarget).next().offset().left;
		width = $(event.currentTarget).next().width() + 30;
	}else{
		top = $(event.currentTarget).prev().offset().top + 35;
		left = $(event.currentTarget).prev().offset().left;
		width = $(event.currentTarget).prev().width() + 30;
	}
	var dom = _.template(cp_app_tree_template,{"top":top,"left":left,"width":width});
	$(".modal").append(dom);
	self.drawAppTree(event,from);
}

function hideAppTree(){	
	$("#apptreearea").remove();
}

function drawAppTree(evt,from){
	var self = this;
	var data = self.generateAppTreeData();
	var setting = {
					view: {
						selectedMulti: false,
						dblClickExpand: false,
						showLine : false
					},
					callback: {
						onClick: function(event, treeId, treeNode) {
							if(!treeNode.isParent){
								var value;
								if(from == "condition" || from == "notify"){
									value = treeNode.id;
								}else if(from == "step"){
									value = "%" + treeNode.id + ".[docker_container_ip]%";
								}		
								if(from == "condition" || from == "step"){
									$(evt.target).prev().val(value);
									$(evt.target).prev().trigger("change");
								}else if(from == "notify"){
									$(evt.target).next().val(value);
									$(evt.target).next().trigger("change");
								}
								self.hideAppTree();
							}else{
								var zTree = $.fn.zTree.getZTreeObj("apptree");
								zTree.expandNode(treeNode);
							}
						}
					}
				};
		
	var zTreeObj = $.fn.zTree.init($("#apptree"), setting, data);	
}

function generateAppTreeData(){
	var self = this;
	var json = {
		id : self.selectedModel.id,
		name : self.selectedModel.displayName,
		open : true
	};
	if(self.selectedModel.groups != null){
		json.children = [];
		self.pushSubGroups(json,self.selectedModel.groups);
	}
	return json;
}

function pushSubGroups(parentJson,subgroups){
	var self = this;
	_.each(subgroups,function(group){
		var id = parentJson.id + "/" + group.id;
		var displayName = group.id;
		var json = {
			id : id,
			name : displayName,
			open : true
		};
		if(group.groups == null){
			if(group.apps.length>0){
				json.children = [];
				_.each(group.apps,function(app){
					var id = json.id + "/" + app.id;
					var displayName = app.id;
					var appjson = {
						id : id,
						name : displayName
					};
					json.children.push(appjson);
				});
			}
		}else{
			json.children = [];
			self.pushSubGroups(json,group.groups);
		}
		parentJson.children.push(json);
	});
}