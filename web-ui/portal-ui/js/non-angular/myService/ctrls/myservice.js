//context menu
function generateServiceContext(event) {
	if(allow_scaleapp_sgo || allow_metering_sgo){
		var self = this;
		var source = $(event.currentTarget);
		var target = "#instance-group-menu";
	
		source.contextmenu({
			target: target,
			onItem: function(e, ev) {
				if (ev.target.tabIndex == 0) {
					self.scale(e.data("appid"), "In");
				} else if (ev.target.tabIndex == 1) {
					self.scale(e.data("appid"), "Out");
				} else if (ev.target.tabIndex == 2) {
					self.scale(e.data("appid"), "To");
				}else if (ev.target.tabIndex == 3) {
					self.goToCadvisorUI(e.data("meteringURL"),e.data("appid"));
				}
				$(target).hide();
			}
		});
		
		if(!allow_scaleapp_sgo || self.selectedService.life_cycle_status == "MODIFYING"){
			$("#instance-group-menu ul li:eq(0)").hide();
			$("#instance-group-menu ul li:eq(1)").hide();
			$("#instance-group-menu ul li:eq(2)").hide();
		}else{
			$("#instance-group-menu ul li:eq(0)").show();
			$("#instance-group-menu ul li:eq(1)").show();
			$("#instance-group-menu ul li:eq(2)").show();
			$("#instance-group-menu ul li:eq(0)").find("a").html($.i18n.prop("scale_in"));
			$("#instance-group-menu ul li:eq(1)").find("a").html($.i18n.prop("scale_out"));
			$("#instance-group-menu ul li:eq(2)").find("a").html($.i18n.prop("scale_to"));
		}
		
		if(!allow_metering_sgo){
			$("#instance-group-menu ul li:eq(3)").hide();
		}else{
			$("#instance-group-menu ul li:eq(3)").show();
			$("#instance-group-menu ul li:eq(3)").find("a").html($.i18n.prop("metering"));
		}
	}
}

function scale(appPath, operate) {
	var serviceInstanceId = this.selectedService.id;
	var compiledTemplate = _.template(serviceScaleTemplate, {
		"gid": appPath,
		"operate": operate,
		"placeholder" : operate == "In" ? $.i18n.prop("placeholder_scale_in") : (operate == "Out" ? $.i18n.prop("placeholder_scale_out") : $.i18n.prop("placeholder_scale_to")),
		"number": 1
	});
	var dialogOption = $.extend({}, {
		width: 220,
		height: 150,
		modal: true,
		buttons: null,
		beforeclose: function(event, ui) {
			// reset the content.
			$(this).empty();
		},
		close: function(event) {
			self.closeSecondDialog();
		}
	});
	$("#linker-dialog").empty();
	$("#linker-dialog").append(compiledTemplate).dialog(dialogOption);
	$("#linker-dialog").height("auto");
	$(".ui-dialog-titlebar").hide();
	
	$("#scaleform").validate({
		rules:{
			scaleText : {
				"number" : true,
				"min" : 1
			}
		}
	});
}

function doScale(event) {
	if (event.keyCode == 13) {
		if(!$("#scaleText").valid()){
			return false;
		}
		var scaleNum = parseInt(event.target.value, 10);
		var appId = $(event.target).data("gid");
		currentNum = this.getCurrentAppNum(appId);
		var operate = $(event.target).data("operate");
		toNum = 1;
		switch (operate) {
			case 'In':
				toNum = currentNum - scaleNum;
				break;
			case 'Out':
				toNum = currentNum + scaleNum;
				break;
			case 'To':
				toNum = scaleNum;
				break;
		}
		
		if(toNum <= 0){
			self.closeSecondDialog();
			var alert = {
				"title" : $.i18n.prop("operation_fail"),
				"type" : "failed",
				"message" : $.i18n.prop("scale_invalid_message"),
				"modaltype" : "notify"
			}
			self.nonAngularAlert(alert);
			return false;
		}
		
		this._doScaleTo(this.selectedOrder.order_id, appId, toNum,operate);
		self.closeSecondDialog();
	}
}

function _doScaleTo(orderId, appId, num,operate) {
	console.log("scale " + appId + " to " + num);
	var self = this;
	var restURL = "/serviceGroupOrders/"+ orderId + "/scaleApp?appId=" + appId + "&num=" + num;
	$.ajax({
		url: restURL,
		type: 'PUT',
		dataType: 'json',
		success: function() {
			var alert = {
				"title" : $.i18n.prop("operation_success"),
				"type" : "success",
				"message" : operate == "In" ? $.i18n.prop("scale_in_success_message") : (operate == "Out" ? $.i18n.prop("scale_out_success_message") : $.i18n.prop("scale_to_success_message")),
				"modaltype" : "notify"
			}
			self.nonAngularAlert(alert);
			self.refreshInstance();
		},
		error: function(error) {
			var alert = {
				"title" : $.i18n.prop("operation_fail"),
				"type" : "failed",
				"message" : operate == "In" ? $.i18n.prop("scale_in_fail_message") : (operate == "Out" ? $.i18n.prop("scale_out_fail_message") : $.i18n.prop("scale_to_fail_message")),
				"modaltype" : "notify"
			}
			self.nonAngularAlert(alert);
			console.log(error.responseText);
		}
	});
}

function getCurrentAppNum(appPathId) {
	var groupPathId = _.initial(appPathId.split("/")).join("/");
	var appid = appPathId.substring(appPathId.lastIndexOf("/")+1);
	var app = this.findAppInSelectedModel(groupPathId,appid,true);
	var currentNum = app.instances;
	return currentNum;
}

function drawServiceTree() {
	var self = this;
	self.modelDesignerHeight = 560;
	var treejson = self.transferModelToTree(self.selectedService);

	var offsetx = 0,offsety = 0;
	if (treejson.children.length > 0) {
		offsetx = self.$('#serviceTreeArea').width() / 2 - 100;
	}
	
	self.setRealOffset();
	
	$("#serviceTreeArea").empty();
	$("#serviceTreeArea").height(self.modelDesignerHeight);
	//Create a new ST instance  
	var st = new $jit.ST({
		//id of viz container element  
		siblingOffset: 50,
		injectInto: 'serviceTreeArea',
		//set duration for the animation  
		constrained : false,
		levelsToShow: 50,
		offsetX : offsetx,
	    offsetY : offsety,
		duration: 0,
		//set animation transition type  
		transition: $jit.Trans.Quart.easeInOut,
		//set distance between node and its children  
		levelDistance: 80,

		//enable panning  
		Navigation: {
			enable: true,
			panning: true
		},
		//set node and edge styles  
		//set overridable=true for styling individual  
		//nodes or edges  
		Node: {
			height: 104,
			width: 153,
			overridable: true
		},

		Edge: {
			type: 'arrow',
			lineWidth: 2,
			color: '#787878',
			dim: '8',
			overridable: true
		},

		//This method is called on DOM label creation.  
		//Use this method to add event handlers and styles to  
		//your node.  
		onCreateLabel: function(label, node) {
			label.id = node.id;
//			var isroot = node.data.isRoot;
			var istemplate = node.data.isTemplate;
			if (istemplate) {
				label.innerHTML = _.template(serviceTreeRootTemplate, {
					'node': node
				});
			} else {
				label.innerHTML = _.template(serviceTreeItemTemplate, {
					'node': node
				});
			}
		},
		onPlaceLabel: function(label, node, controllers) {
			//override label styles
			var style = label.style;
			// show the label and let the canvas clip it
			style.display = '';
		},
		onBeforePlotNode: function(node) {
			node.data.$color = "#fff";
			var appslen = 0;
			_.each(node.data.apps, function(app) {
				appslen = appslen + app.instances;
			});
			appslen = appslen > 0 ? appslen : 1;
			node.data.$height = appslen * 104;
//			var realID = node.data.realID;
			var realID = node.id;
			var offsetindex = _.pluck(self.y_positions, "id").indexOf(realID);
			node.pos.y = self.y_positions[offsetindex].y;
		},
		onBeforePlotLine: function(adj) {
			adj.data.$color = "#787878";
		}
	});
	//load json data  
	st.loadJSON(treejson);
	//compute node positions and layout  
	st.compute();
	//optional: make a translation of the tree  
	st.geom.translate(new $jit.Complex(-200, 0), "current");
	//emulate a click on the root node.  
	st.onClick(st.root,{
		 Move: {
            enable: true,
            offsetX: _.isUndefined(self.st) ? offsetx : offsetx - self.canvasX,
            offsetY: _.isUndefined(self.st) ? offsety : offsety - self.canvasY
        }
	});
	
	self.st = st;
	
	setTimeout(self.showInstanceStatus,500);
	//refresh instance if pending
	if(self.selectedService.life_cycle_status != "DEPLOYED"){
		setTimeout(refreshInstance,10000);
	}
}

function readIPAddress(instanceid) {
	var self = this;
	var restURL = location.protocol + "//" + location.host + "/appInstances/" + instanceid;

	$.ajax({
		url: restURL,
		type: 'GET',
		dataType: 'json',
		success: function(data) {
			self.showAppInstanceDetail(instanceid,data);
		},
		error: function(error) {
			console.log(error.responseText);
		}
	});
}

//metering
function goToCadvisorUI(meteringURL,appid){
	var self = this;
	var compiledTemplate = _.template(cadvisorTemplate,{"appid":appid});
	$("#linker-alert").empty();
	$("#linker-alert").append(compiledTemplate);
	$("#linker-alert").modal("show");
	
    $('#embed_cadvisor iframe').attr('src', meteringURL);
}

//instance status
function showInstanceStatus(){
	var content= '<font class="popover-font">' + this.selectedService.life_cycle_status +'</font>';
	var title = _.template('<%=$.i18n.prop("instance_status")%>');
	$(".node[id='"+ this.selectedService.service_group_id +"']").popover({content:content,html:true,trigger:'hover',viewport:{"selector": "#viewport", "padding": 0 }});
	var statusstyle;
	if(this.selectedService.life_cycle_status == "DEPLOYED"){
		statusstyle = "instance-active";
	}else if(this.selectedService.life_cycle_status == "TERMINATED"){
		statusstyle = "instance-fail";
	}else{
		statusstyle = "instance-pending";
	}
	$(".node[id='"+ this.selectedService.service_group_id +"']").find(".service-group").addClass(statusstyle);
}

function showAppInstanceDetail(instanceid,data){
	var docker_container_ip = data.data.docker_container_ip == "" ? '&nbsp;' : data.data.docker_container_ip;
	var docker_container_port = data.data.docker_container_port == "" ? '&nbsp;' : data.data.docker_container_port;
	var docker_container_name = data.data.docker_container_name == "" ? '&nbsp;' : data.data.docker_container_name;
	var lifecycle_status = data.data.lifecycle_status == "" ? '&nbsp;' : data.data.lifecycle_status;
	var mesos_slave = data.data.mesos_slave == "" ? '&nbsp;' : data.data.mesos_slave;
	var mesos_slave_host_port = data.data.mesos_slave_host_port == "" ? '&nbsp;' : data.data.mesos_slave_host_port;
	var content= _.template( '<font class="popover-font"><%=$.i18n.prop("docker_container_ip")%></font><hr class="popover-hr"/>'+ docker_container_ip
			     +'<hr class="popover-hr-1"/>'+'<font class="popover-font"><%=$.i18n.prop("docker_container_port")%></font><hr class="popover-hr"/>'+docker_container_port
			     +'<hr class="popover-hr-1"/>'+'<font class="popover-font"><%=$.i18n.prop("docker_container_name")%></font><hr class="popover-hr"/>'+docker_container_name
			     +'<hr class="popover-hr-1"/>'+'<font class="popover-font"><%=$.i18n.prop("lifecycle_status")%></font><hr class="popover-hr"/>'+lifecycle_status 			           
			     +'<hr class="popover-hr-1"/>'+'<font class="popover-font"><%=$.i18n.prop("mesos_slave")%></font><hr class="popover-hr"/>'+mesos_slave
			     +'<hr class="popover-hr-1"/>'+'<font class="popover-font"><%=$.i18n.prop("mesos_slave_port")%></font><hr class="popover-hr"/>'+mesos_slave_host_port);
			     
	var title = _.template('<%=$.i18n.prop("detail_info")%>');
	$("#" + instanceid).popover({title:title,content:content,html:true,trigger:'click hover'});
//	$("#" + instanceid).click(function (e) {
//	    e.stopPropagation();
//	});
//	
//	$(document).click(function (e) {
//	    if (($('.popover').has(e.target).length == 0) || $(e.target).is('.close')) {
//	        $("#" + instanceid).popover('hide');
//	    }
//	});
	
	var meteringURL = "http://" + data.data.mesos_slave + ":10000/docker/" + data.data.docker_container_long_id;
	$("#" + instanceid).data("meteringURL",meteringURL);
	
	var statusstyle;
	if(data.data.lifecycle_status == "CONFIGED"){
		statusstyle = "instance-active";
	}else if(data.data.lifecycle_status == "UNCONFIGED"){
		statusstyle = "instance-fail";
	}else{
		statusstyle = "instance-pending";
	}
	$("#" + instanceid).addClass(statusstyle);
}

function refreshInstance() {
	var self = this;
	var restURL = location.protocol + "//" + location.host + "/groupInstances/" + self.selectedOrder.service_group_instance_id;

	$.ajax({
		url: restURL,
		type: 'GET',
		dataType: 'json',
		success: function(data) {
			self.showServiceDetail(data.data);
		},
		error: function(error) {
			console.log(error.responseText);
		}
	});
}

function showServiceDetail(service){
	var self = this;
	service.displayName = self.idToSimple(service.id);
	service.imageSrc = "images/products/Appserver.png";
  	_.each(service.groups,function(group){
		self.allocateImageToApp(group);
	});
	self.selectedService = service;
	setTimeout(function(){
		self.drawServiceTree();
	},200);
}
