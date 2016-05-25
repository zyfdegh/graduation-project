function appModelDesignService(http,q){
	var getApps = function(){
		var deferred = q.defer();
		var url = "/apps";
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
	
	var newApp = function(app){
		var deferred = q.defer();
		var url = "/apps";
		var request = {
			"url": url,
			"dataType": "json",
			"method": "POST",
			"data" : angular.toJson(app)
		}
			
		http(request).success(function(data){
			deferred.resolve(data);
		}).error(function(error){
			deferred.reject(error);
		});
		return deferred.promise;
	};
	
	var updateApp = function(app){
		var deferred = q.defer();
		var url = "/apps/" + app._id;
		var request = {
			"url": url,
			"dataType": "json",
			"method": "PUT",
			"data" : angular.toJson(app)
		}
			
		http(request).success(function(data){
			deferred.resolve(data);
		}).error(function(error){
			deferred.reject(error);
		});
		return deferred.promise;
	};
	
	var deleteApp = function(app){
		var deferred = q.defer();
		var url = "/apps/" + app._id;
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
	
	var getAppOperations = function(appid){
		var deferred = q.defer();
		var url = "/apps/operations/" + appid;
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
	
	return {
		'getApps' : getApps,
		'newApp' : newApp,
		'updateApp' : updateApp,
		'deleteApp' : deleteApp,
		'getAppOperations' : getAppOperations
	}
}


linkerCloud.factory('appModelDesignService',['$http','$q',appModelDesignService]);