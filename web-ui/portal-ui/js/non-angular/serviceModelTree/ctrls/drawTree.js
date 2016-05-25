function drawGroupRelations(){
	var self = this;
	self.modelDesignerHeight = 560;
	var treejson = self.transferModelToTree(self.selectedModel);
	
	var offsetx = self.$('#modelDesignerArea').width() / 2 - 100,offsety = 0;
	if(treejson.children.length > 0){
    		offsetx = self.$('#modelDesignerArea').width() / 2 - 100;
	}
	
	self.setRealOffset();
	
	$("#modelDesignerArea").empty();
	$("#modelDesignerArea").height(self.modelDesignerHeight);
	//Create a new ST instance  
	var st = new $jit.ST({  
	    //id of viz container element  
	    siblingOffset : 50,
	    injectInto: 'modelDesignerArea',  
	    //set duration for the animation  
	    constrained : false,
	    levelsToShow : 50,
	    offsetX : offsetx,
	    offsetY : offsety,
	    duration: 0,  
	    //set animation transition type  
	    transition: $jit.Trans.Quart.easeInOut,  
	    //set distance between node and its children  
	    levelDistance: 80,  
		
	    //enable panning  
	    Navigation: {  
	      enable:true,  
	      panning:true  
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
	  		dim : '8',
	        overridable: true  
	    },  
	      
	    //This method is called on DOM label creation.  
	    //Use this method to add event handlers and styles to  
	    //your node.  
	    onCreateLabel: function(label, node){  
	        label.id = node.id;   
	        var istemplate = node.data.isTemplate;
	        if(istemplate){
	        	 label.innerHTML = _.template(serviceGroupTemplate, {
									'node' : node,
									'allow_update' : self.allow_update_sg
								});
	        }else{
	        	 label.innerHTML = _.template(groupItemTemplate, {
									'node' : node,
									'allow_update' : self.allow_update_sg
								});
	        }
	    },  
	    onPlaceLabel: function(label, node, controllers){          
            //override label styles
            var style = label.style;  
            // show the label and let the canvas clip it
            style.display = '';
       },
	    onBeforePlotNode: function(node){  
	         node.data.$color = "#fff"; 
	         node.data.$height = node.data.apps.length * 104 > 0 ? node.data.apps.length * 104 : 104;
	         var realID = node.id;
	         var offsetindex = _.pluck(self.y_positions,"id").indexOf(realID);
	         node.pos.y = self.y_positions[offsetindex].y;
	    }, 
	    onBeforePlotLine: function(adj){  
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
}

function setRealOffset(){
	var self = this;
	if(_.isUndefined(self.canvasX)){
		self.canvasX = 0;
	}
	if(_.isUndefined(self.canvasY)){
		self.canvasY = 0;
	}
	if(!_.isUndefined(self.st)){
		self.canvasX += self.st.canvas.translateOffsetX;
		self.canvasY += self.st.canvas.translateOffsetY;
	}
}

function showGroupAppDetail(target){	
	
	var self = this;
	var apppathid = target.data("appid");
	var groupid = apppathid.substring(0,apppathid.lastIndexOf("/"));
	var appid = apppathid.substring(apppathid.lastIndexOf("/")+1);
	var app = self.findAppInSelectedModel(groupid,appid);
	self.openedApp = $.extend(true,{},app);
	
	self.openedApp.exposePorts = self.openedApp.env.LINKER_EXPOSE_PORTS == "true" ? "yes" : "no";
	
	self.openedApp.env = self.openedApp.env != ""?JSON.stringify(self.openedApp.env):"";
	self.openedApp.constraints = self.openedApp.constraints != ""?JSON.stringify(self.openedApp.constraints):"";
	
	_.each(self.openedApp.container.docker.parameters,function(par){
		var value = par.value;
		par.vKey = value.substring(0,value.indexOf("="));
		par.vValue = value.substring(value.indexOf("=")+1);
	});
	
	if(_.isUndefined(self.openedApp.container.volumes) || self.openedApp.container.volumes == null){
    		self.openedApp.container.volumes = [];
    }
	
	var buttons = {};
	buttons[$.i18n.prop("remove_app_form_group")] = function(){
		self.confirmDeleteAppFromGroup(groupid,appid);
		$(this).dialog('close');
	};
	buttons[$.i18n.prop("save_changes")] = function(){
		if(self.formisvalid()){
			if(self.saveAppInGroup(app)){
				$(this).dialog('close');
			}
		}
	};
	buttons[$.i18n.prop("cancel")] = function() {
		$(this).dialog('close');
	}
	var compiledTemplate = _.template(appDetailTemplate);
	var dialogOption = $.extend({}, {
		draggable : true,
		buttons : buttons,
		width : 780,
		height: 620,
		modal : true,
		beforeclose : function(event, ui) {
			// reset the content.
			$(this).empty();
		},
		close : function(event){
			self.closeSecondDialog();
			$(".group-app-item-selected").addClass("group-app-item");
			$(".group-app-item-selected").removeClass("group-app-item-selected");
		}
	});
	$("#linker-dialog").empty();
	$("#linker-dialog").append(compiledTemplate).dialog(dialogOption);
	$(".ui-dialog-titlebar").hide(); 
	
//	if(_.isUndefined(self.openedApp.container.docker.parameters)){
//		self.openedApp.container.docker.parameters = [];
//	}
    getDockerImages();
	parseAppImage(self.openedApp);
	new Vue({
	  el: '#appdetailform',
	  data: {
	  	"openedApp" : self.openedApp,
	  	"radio" : self.radio,
	  	"dockerImage":self.dockerImage,
	  	"dockerImages" : self.dockerImages,
        "imageTag":self.imageTag,
	  	"imageTags" : self.imageTags
	  },
	  methods : {
	  	  "getImageTags" : self.getImageTags
 	  }
	});
	
	$("#appdetailform").validate({
		rules:{
			cpus : {
				"number" : true,
				"min" : 0.1
			},
			memory : {
				"number" : true,
				"min" : 1
			},
			instances : {
				"number" : true,
				"min" : 1
			}	
		}
	});
}
function parsePrefix(){
        var currentUserName = sessionStorage.username;
        var parsedPrefix = "";
        if(currentUserName != "sysadmin"){
        	parsedPrefix = currentUserName.replace(/@/g, "_at_");
        	parsedPrefix = parsedPrefix.replace(/\./g, "_");
        }
        return parsedPrefix;
}
function getImageTags(){
        var self = this;
        var restURL = '/docker/imageTag?imageName=' + encodeURIComponent(self.dockerImage.fromLinker);
        $.ajax({
				url: restURL,
				type: 'GET',
				dataType: 'json',	
				async :false,			
				success: function(data) {
					if(data){
						self.imageTags = data.tags;   
                        self.imageTag.tag = self.imageTags[0] || "";  
        	            }
				},
				error: function(error) {
					console.log(error.responseText);
				}
		});	      
       
}
function getDockerImages(){
        var self = this;
        this.radio = {"repoType":"linker"};
		this.dockerImage = {			
				fromLinker : "",
			    fromDockerhub : "image from dockerhub"						
		};
		this.imageTag = {			
				tag : ""			
		};
		var restURL = '/docker/image';
			$.ajax({
				url: restURL,
				type: 'GET',
				dataType: 'json',	
				async :false,			
				success: function(data) {
					if(data.results.length > 0){
						var prefix = self.parsePrefix();
						var allImages = _.map(data.results,function(result){return result.name;});
        	            self.dockerImages = _.filter(allImages,function(image){return image.indexOf(prefix)!=-1||image.indexOf("linker\/")!=-1});			
        	            self.dockerImage.fromLinker = self.dockerImages[0] || "";
        	            // self.getImageTags();       	           
        	            }
				},
				error: function(error) {
					console.log(error.responseText);
				}
			});	      
};
function parseAppImage(app){
        // var prefix = this.parsePrefix();
        var linkerRepoPrefix = "linkerrepository:5000/";
	    var dockerhubPrefix = "docker.io/" 
        if( app.container.docker.image.indexOf(linkerRepoPrefix) !=-1){
			this.radio = {"repoType":"linker"};
			var tempImage = app.container.docker.image.replace(linkerRepoPrefix,"").trim();	
			var imageInfo = tempImage.split(":");
			this.dockerImage.fromLinker = imageInfo[0];
			this.getImageTags();   
			this.imageTag.tag = imageInfo[1];

		}else{
            this.radio = {"repoType":"dockerhub"};
            this.dockerImage.fromDockerhub = app.container.docker.image.replace(dockerhubPrefix,"").trim();;
            this.getImageTags();   
        }
		
};
function newContainerParameter(){
	var par = {
		"key" : "",
		"value" : "",
		"editable" : false
	}; 
	if(_.isUndefined(this.openedApp.container.docker.parameters) || this.openedApp.container.docker.parameters == null){
    		this.openedApp.container.docker.parameters = [];
    }
	this.openedApp.container.docker.parameters.push(par);
}

function removeContainerParameter(event){
	var index = $(event.currentTarget).parent().data("index");
	this.openedApp.container.docker.parameters.splice(index,1);
}

function newContainerVolume(){
	var volume = {
        		"containerPath": "",
            "hostPath": "",
            "mode": "RO"
        }; 
    if(_.isUndefined(this.openedApp.container.volumes) || this.openedApp.container.volumes == null){
    		this.openedApp.container.volumes = [];
    }
	this.openedApp.container.volumes.push(volume);
}

function removeContainerVolume(event){
	var index = $(event.currentTarget).parent().data("index");
	this.openedApp.container.volumes.splice(index,1);
}

function formisvalid(){
	var formIsValid = true;
	var cpuIsValid = $("#cpus").valid();
	var memoryIsValid = $("#memory").valid();
	var instanceIsValid = $("#instances").valid();
	var dockerImageIsValid = this.radio.repoType == "linker" ? true : $("#docker_image").valid();
	formIsValid = formIsValid && cpuIsValid && memoryIsValid && instanceIsValid && dockerImageIsValid;
	if(formIsValid){
		_.each(this.openedApp.container.docker.parameters,function(par,index){
			var nameIsValid = $("#docker_par_name_"+index).valid();
			var valueIsValid = $("#docker_par_value_"+index).valid();
			formIsValid = formIsValid && nameIsValid && valueIsValid;
		});
		
		_.each(this.openedApp.container.volumes,function(volume,index){
			var cpathIsValid = $("#volume_containerpath_"+index).valid();
			var hpathIsValid = $("#volume_hostpath_"+index).valid();
			formIsValid = formIsValid && cpathIsValid && hpathIsValid;
		});
	}
	return formIsValid;
}
