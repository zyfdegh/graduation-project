linkerCloud.controller('AppDesignerController',  ['$scope','$location','productsService','appModelDesignService', 'responseService', 'confirmService',"dockerImageService",
	function($scope,$location,productsService,appModelDesignService,responseService,confirmService,dockerImageService) {
	$scope.showDetail = false;
	$scope.linkerRepoPrefix = "linkerrepository:5000/";
	$scope.dockerhubPrefix = "docker.io/"
	$scope.listApps = function(){
	    appModelDesignService.getApps().then(function(data){
		    	if(responseService.successResponse(data)){
		    		_.each(data.data,function(app){
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
		    		$scope.apps = data.data;
		    	}
	    },
	    function(errorMessage){
	    		responseService.errorResp(errorMessage);
	    });
  	};
  	
	$scope.showAppDetail = function(app){
		getDockerImages();
		
		appModelDesignService.getAppOperations(app._id).then(function(data){
		    	if(responseService.successResponse(data)){
		    		$scope.allow_update_app = data.data.update_app == 1 ? true : false;
		    		$scope.allow_delete_app = data.data.delete_app == 1 ? true : false;
		    		
		    		$scope.showDetail = true;
		    		$scope.detailTitle = "App Details";
				transferToJsonString(app);
				parseAppImage(app);
				$scope.selectedApp = app;
		    	}
	    },
	    function(errorMessage){
	   	 	responseService.errorResp(errorMessage);
	    });
	}
	
	$scope.newAppDetail = function(){
		getDockerImages(true);
		var app = {
			"id": "",
			"cpus": 0.1,
			"mem": 128,
			"instances" : 1,
			"cmd": "echo linker",
			"executor" : "/usr/local/bin/startExecutor.sh",
			"container": {
			    "type": "DOCKER",
			    "docker": {
			      "network": "BRIDGE",
			      "image": "",
			      "privileged": true,
			      "forcePullImage": true,
			      "parameters": []
			    },
			    "volumes":[]
			},
			"env": {},
			"imageSrc" : "images/products/designer/app/default.png",
			"scale" : {
			    "enabled": false,
			    "min_num": 1,
			    "max_num": 1,
			    "scale_step": 0
			}
		};
		transferToJsonString(app);
		$scope.selectedApp = app;
		$scope.showDetail = true;
		$scope.detailTitle = "New App";		
		$scope.allow_update_app = true;
       
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
    $scope.getImageTags = function(init){
        dockerImageService.getImageTags($scope.dockerImage.fromLinker).then(function(data){
            $scope.imageTags = data.tags;       	
        	if(init){
        		$scope.imageTag.tag = $scope.imageTags[0] || "";	
        	}else{
        		if($scope.radio.repoType=="dockerhub"){
                   $scope.imageTag.tag = $scope.imageTags[0];
        		}
        	}
        },function(errorMessage){
            responseService.errorResp(errorMessage);
        })
    }
	function getDockerImages(init){
        $scope.radio = {"repoType":"linker"};
		$scope.dockerImage = {			
				fromLinker : "",
			    fromDockerhub : "image from dockerhub"						
		};
		$scope.imageTag = {			
				tag : ""			
		};
        dockerImageService.getDockerImages().then(function(data){
        	var prefix = parsePrefix();
        	var allImages = _.map(data.results,function(result){return result.name;});
        	$scope.dockerImages = _.filter(allImages,function(image){return image.indexOf(prefix)!=-1||image.indexOf("linker\/")!=-1});
        	
        	if(init){
        		$scope.dockerImage.fromLinker = $scope.dockerImages[0] || "";
        		if($scope.dockerImage.fromLinker != ""){
        			$scope.getImageTags(init);	
        		}
        		
        	}else{
        		if($scope.radio.repoType=="dockerhub"){
                    $scope.dockerImage.fromLinker = $scope.dockerImages[0];
                   if($scope.dockerImage.fromLinker != ""){
                     $scope.getImageTags(init);
                   }
        		}
        	}
            

        },function(errorMessage){
            responseService.errorResp(errorMessage);
        });
	};
	function generateAppImage(app){
		if($scope.radio.repoType == "linker"){
			app.container.docker.image =$scope.linkerRepoPrefix + $scope.dockerImage.fromLinker+":"+$scope.imageTag.tag;
		}else{
			app.container.docker.image = $scope.dockerhubPrefix + $scope.dockerImage.fromDockerhub;
		}
		
	};
	function parseAppImage(app){
        // var prefix = parsePrefix();
        if(app.container.docker.image.indexOf($scope.linkerRepoPrefix) !=-1 ){
			$scope.radio = {"repoType":"linker"};
			var tempImage = app.container.docker.image.replace($scope.linkerRepoPrefix,"").trim();	
			var imageInfo = tempImage.split(":");
			$scope.dockerImage.fromLinker = imageInfo[0];
			$scope.imageTag.tag = imageInfo[1];
			if($scope.dockerImage.fromLinker != ""){
                   $scope.getImageTags();	
            }
								
		}else{
            $scope.radio = {"repoType":"dockerhub"};
            $scope.dockerImage.fromDockerhub = app.container.docker.image.replace($scope.dockerhubPrefix,"").trim();
		}
	};
	$scope.saveApp = function(app){
		if(!formIsValid()){
			return false;
		}
		delete app.imageSrc;
		transferToJson(app);
		generateAppImage(app);
		if(!_.isUndefined(app._id)){
			updateApp(app);
		}else{
			newApp(app);
		}
	}

	function newApp(app){
		appModelDesignService.newApp(app).then(function(data){
		    	$scope.goBackToAppList(app);
		    	$scope.listApps();
	    },
	    function(errorMessage){
	    		responseService.errorResp(errorMessage);
	    });
	}

	function updateApp(app){
		appModelDesignService.updateApp(app).then(function(data){
		    	$scope.goBackToAppList(app);
		    	$scope.listApps();
	    },
	    function(errorMessage){
	    		responseService.errorResp(errorMessage);
	    });
	}
	
	$scope.confirmDeleteApp = function(app){
		$scope.$translate(['rightContent.app.deleteConfirm', 'rightContent.app.deleteMessage', 'rightContent.app.deleteBtn']).then(function (translations) {
		    $scope.confirm = {
				"title" : translations['rightContent.app.deleteConfirm'],
				"message" : translations['rightContent.app.deleteMessage'],
				"buttons" : [
					{
						"text" : translations['rightContent.app.deleteBtn'],
						"action" : function(){
							$scope.deleteApp(app);
						}
					}
				]
			};
			confirmService.deleteConfirm($scope);
		  });
	}
	
	$scope.deleteApp = function(app){
		appModelDesignService.deleteApp(app).then(function(data){
	    		$scope.goBackToAppList(app);
	   	 	$scope.listApps();
	    },
	    function(errorMessage){
	    		responseService.errorResp(errorMessage);
	    });
	}
	
	$scope.goBackToAppList = function(app){
		transferToJson(app);
		$scope.showDetail = false;
		$("body").scrollTop(0);
	}
	
	function transferToJsonString(app){
		_.each(app.container.docker.parameters,function(par){
			var value = par.value;
			par.vKey = value.substring(0,value.indexOf("="));
			par.vValue = value.substring(value.indexOf("=")+1);
		});
		
		if(app.env.LINKER_EXPOSE_PORTS == "true"){
			app.exposePorts = "yes";
		}else{
			app.exposePorts = "no";
		}
		
		app.env = angular.toJson(app.env);
		if(!_.isUndefined(app.constraints)){
			app.constraints = angular.toJson(app.constraints);
		}else{
			app.constraints = undefined;
		}

		if (_.isUndefined(app.scale)) {
			app.scale = {
				enabled: false,
				min_num: 1,
				max_num:1,
				scale_step:0
			}
		}
		
	}
	
	function transferToJson(app){
		_.each(app.container.docker.parameters,function(par){
			par.key = "env";
			par.value = par.vKey + "=" + par.vValue;
			delete par.vKey;
			delete par.vValue;
		});
		
		try{
			app.env = angular.fromJson(app.env);
		}catch(e){
			app.env = {};
		}
		
		if(app.exposePorts == "yes"){
			app.env.LINKER_EXPOSE_PORTS = "true";
		}else{
			delete app.env.LINKER_EXPOSE_PORTS;
		}
		delete app.exposePorts;
		
		try{
			if(!_.isUndefined(app.constraints)){
				app.constraints = angular.fromJson(app.constraints);
			}else{
				delete app.constraints;
			}
		}catch(e){
			delete app.constraints;
		}	
		
	}
	
	$scope.newParameter = function(){
		var par = {
			"key" : "",
			"value" : "",
			"editable" : false
		};
		$scope.selectedApp.container.docker.parameters.push(par);
	}
	
	$scope.removeParameter = function(index){
		$scope.selectedApp.container.docker.parameters.splice(index,1);
	}
	
	$scope.newVolume = function(){
		var volume =  {
        		"containerPath": "",
            "hostPath": "",
            "mode": "RO"
        };
		$scope.selectedApp.container.volumes.push(volume);
	}
	
	$scope.removeVolume = function(index){
		$scope.selectedApp.container.volumes.splice(index,1);
	}
	
	var formIsValid = function(){
		var basic = appform.appid.validity.valid && $("#invalidID").css("display") == "none" && appform.cpus.validity.valid && appform.memory.validity.valid && appform.instances.validity.valid;
		var docker_image = $scope.repoType== "linker" ? true: appform.docker_image.validity.valid;
		var docker_pars = true;
		_.each($scope.selectedApp.container.docker.parameters,function(par,index){
			docker_pars = docker_pars && eval("appform.docker_par_name_"+index+".validity.valid");
			docker_pars = docker_pars && eval("appform.docker_par_value_"+index+".validity.valid");
		});
		var container_volumes = true;
		_.each($scope.selectedApp.container.volumes,function(volume,index){
			container_volumes = container_volumes && eval("appform.volume_containerpath_"+index+".validity.valid");
			container_volumes = container_volumes && eval("appform.volume_hostpath_"+index+".validity.valid");
		});
		return basic&&docker_image && docker_pars && container_volumes;	
	}
	
	var initialize = function(){
		$scope.listApps();
	}
	
	initialize();
}]);
