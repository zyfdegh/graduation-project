function mppService(http,q){
	var listDockerHubImage = function(query, url){
			var deferred = q.defer();
			var request = {
				"url": "/dockerHub",
				"dataType": "json",
				"method": "POST",
				"data" : angular.toJson({
					"imageName" : query,
					"url" : url
				})
			}
				
			http(request).success(function(data){
				deferred.resolve(data);
			}).error(function(error){
				deferred.reject(error);
			});
			return deferred.promise;
	}
	
	var prepareServiceToOrder = function(service){
		var serviceToOrder = {
//			"_id" : service._id,
			"id" : service.displayName,
			"apps" : []
		}
		var serviceid = service.id;
		_.each(service.groups,function(group){
			var groupid = serviceid + "/" + group.id;
			if(_.isUndefined(group.groups) || group.groups == null){
				_.each(group.apps,function(app){
					var appid = groupid + "/" + app.id;
					_.each(app.container.docker.parameters,function(parameter){
						if(parameter.editable){
							var index = _.pluck(serviceToOrder.apps,"id").indexOf(appid);
							if(index < 0){
								var pname;
								if(_.isUndefined(parameter.description) || parameter.description.trim().length==0){
									pname = parameter.value.substring(0,parameter.value.indexOf("="));
								}else{
									pname = parameter.description;
								}
								var app = {
									"id" : appid,
									"parameters" : [
										{
											"key" : parameter.value.substring(0,parameter.value.indexOf("=")),
											"name" : pname,
											"value" : parameter.value.substring(parameter.value.indexOf("=")+1)
										}
									]
								};
								serviceToOrder.apps.push(app);
							}else{
								var pname;
								if(_.isUndefined(parameter.description) || parameter.description.trim().length==0){
									pname = parameter.value.substring(0,parameter.value.indexOf("="));
								}else{
									pname = parameter.description;
								}
								
								var par = {
									"key" : parameter.value.substring(0,parameter.value.indexOf("=")),
									"name" : pname,
									"value" : parameter.value.substring(parameter.value.indexOf("=")+1)
								};
								serviceToOrder.apps[index].parameters.push(par);
							}
						}
					});
				});
			}else{
				prepareTemplateInSerivceToOrder(serviceToOrder,groupid,group);
			}	
		});
		
		return serviceToOrder;
	};
	
	var prepareTemplateInSerivceToOrder = function(serviceToOrder,parentid,template){
		_.each(template.groups,function(group){
			var groupid = parentid + "/" + group.id;
			if(_.isUndefined(group.groups) || group.groups == null){
				_.each(group.apps,function(app){
					var appid = groupid + "/" +app.id;
					_.each(app.container.docker.parameters,function(parameter){
						if(parameter.editable){
							var index = _.pluck(serviceToOrder.apps,"id").indexOf(appid);
							if(index < 0){
								var pname;
								if(_.isUndefined(parameter.description) || parameter.description.trim().length==0){
									pname = parameter.value.substring(0,parameter.value.indexOf("="));
								}else{
									pname = parameter.description;
								}
								var app = {
									"id" : appid,
									"parameters" : [
										{
											"key" : parameter.value.substring(0,parameter.value.indexOf("=")),
											"name" : pname,
											"value" : parameter.value.substring(parameter.value.indexOf("=")+1)
										}
									]
								};
								serviceToOrder.apps.push(app);
							}else{
								var pname;
								if(_.isUndefined(parameter.description) || parameter.description.trim().length==0){
									pname = parameter.value.substring(0,parameter.value.indexOf("="));
								}else{
									pname = parameter.description;
								}
								
								var par = {
									"key" : parameter.value.substring(0,parameter.value.indexOf("=")),
									"name" : pname,
									"value" : parameter.value.substring(parameter.value.indexOf("=")+1)
								};
								serviceToOrder.apps[index].parameters.push(par);
							}
						}
					});
				});
			}else{
				prepareTemplateInSerivceToOrder(serviceToOrder,groupid,group);
			}	
		});
	};
	
	var runServiceModel = function(service){
		var deferred = q.defer();
		var serviceid = "/" + service.id;
		var data = {
			"service_group_id":serviceid,
//			"service_group_obj_id":service._id,
			"parameters": []
		};
		_.each(service.apps,function(app){
			_.each(app.parameters,function(par){
				var parameter = {
				    "appId": app.id,
				    "paramName": par.key,
				    "paramValue": par.value
				}
				data.parameters.push(parameter);
			})
		});
		
		var url = "/serviceGroupOrders";
		var request = {
			"url": url,
			"dataType": "json",
			"method": "POST",
			"data" : angular.toJson(data)
		}
			
		http(request).success(function(data){
			deferred.resolve(data);
		}).error(function(error){
			deferred.reject(error);
		});
		return deferred.promise;
	};
	
	return{	
		'listDockerHubImage' : listDockerHubImage,
		'prepareServiceToOrder' : prepareServiceToOrder,
		'runServiceModel' : runServiceModel
    }
}


linkerCloud.factory('mppService',['$http','$q',mppService]);