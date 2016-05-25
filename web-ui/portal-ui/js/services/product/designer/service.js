function serviceModelDesignService(http,q){
	var listServiceGroup = function(condition){
		var deferred = q.defer();
		var url="";
		if(condition == "published"){
             url = "/serviceGroups/published?query={\"state\":\"published\"}";
		}else{
			 url = "/serviceGroups";
		}		
		var request = {
			"url": url,
			"dataType": "json",
			"method": "GET"
		}
			
		http(request).success(function(data){
			deferred.resolve(data);
		}).error(function(error){
			deferred.reject(error);
		});
		return deferred.promise;
	};
	
	var newServiceModel = function(model){
		var deferred = q.defer();
		var url = "/serviceGroups";
		var request = {
			"url": url,
			"dataType": "json",
			"method": "POST",
			"data" : angular.toJson(model)
		}
			
		http(request).success(function(data){
			deferred.resolve(data);
		}).error(function(error){
			deferred.reject(error);
		});
		return deferred.promise;
	};
	
	var updateServiceModel = function(model){
		var deferred = q.defer();
		var url = "/serviceGroups/" + model._id;
		var request = {
			"url": url,
			"dataType": "json",
			"method": "PUT",
			"data" : angular.toJson(model)
		}
			
		http(request).success(function(data){
			deferred.resolve(data);
		}).error(function(error){
			deferred.reject(error);
		});
		return deferred.promise;
	};
	
	var deleteServiceModel = function(model){
		var deferred = q.defer();
		var url = "/serviceGroups/" + model._id;
		var request = {
			"url": url,
			"dataType": "json",
			"method": "DELETE"
		}
			
		http(request).success(function(data){
			deferred.resolve(data);
		}).error(function(error){
			deferred.reject(error);
		});
		return deferred.promise;
	};
	
	var deleteCP = function(app_path){
		var deferred = q.defer();
		var url = '/appConfigs?query={"app_container_id":"' + app_path + '"}';
		var request = {
			"url": url,
			"dataType": "json",
			"method": "DELETE"
		}
			
		http(request).success(function(data){
			deferred.resolve(data);
		}).error(function(error){
			deferred.reject(error);
		});
		return deferred.promise;
	};
	
	var getSGOperations = function(sgid){
		var deferred = q.defer();
		var url = "/serviceGroups/operations/" + sgid;
		var request = {
			"url": url,
			"dataType": "json",
			"method": "GET"
		}
			
		http(request).success(function(data){
			deferred.resolve(data);
		}).error(function(error){
			deferred.reject(error);
		});
		return deferred.promise;
	};
	
	var publishServiceModel = function(model){
		var deferred = q.defer();
		var url = "/serviceGroups/publish/" + model._id;
		var request = {
			"url": url,
			"dataType": "json",
			"method": "PUT",
			"data" : {}
		}
			
		http(request).success(function(data){
			deferred.resolve(data);
		}).error(function(error){
			deferred.reject(error);
		});
		return deferred.promise;
	};
	
	var unpublishServiceModel = function(model){
		var deferred = q.defer();
		var url = "/serviceGroups/unpublish/" + model._id;
		var request = {
			"url": url,
			"method": "PUT"
		}
			
		http(request).success(function(data){
			deferred.resolve(data);
		}).error(function(error){
			deferred.reject(error);
		});
		return deferred.promise;
	};
	
	var submitServiceModel = function(model){
		var deferred = q.defer();
		var url = "/serviceGroups/submit/" + model._id;
		var request = {
			"url": url,
			"dataType": "json",
			"method": "PUT",
			"data" : {}
		}
			
		http(request).success(function(data){
			deferred.resolve(data);
		}).error(function(error){
			deferred.reject(error);
		});
		return deferred.promise;
	};
	
	return {
		'listServiceGroup' : listServiceGroup,
		'newServiceModel' : newServiceModel,
		'updateServiceModel' : updateServiceModel,
		'deleteServiceModel' : deleteServiceModel,
		'deleteCP' : deleteCP,
		'getSGOperations' : getSGOperations,
		'publishServiceModel' : publishServiceModel,
		'unpublishServiceModel' : unpublishServiceModel,
		"submitServiceModel" : submitServiceModel
	}
}


linkerCloud.factory('serviceModelDesignService',['$http','$q',serviceModelDesignService]);