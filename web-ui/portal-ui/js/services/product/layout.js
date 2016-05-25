function layoutService(http,q){
	var getLayout = function(){
		var deferred = q.defer();
//		var url = "/apps";
//		var request = {
//			"url": url,
//			"dataType": "json",
//			"method": "GET"
//		}
			
//		http(request).success(function(data){
//			deferred.resolve(data);
//		}).error(function(error){
//			deferred.reject(error.responseText);
//		});
		var result = {
			"type" :sessionStorage.mpp_layout_type,
			"data" : sessionStorage.mpp_layout_data
		}
		deferred.resolve(result);
		return deferred.promise;
	};
	
	var saveLayout = function(layoutdata,layouttype){
		var deferred = q.defer();
//		var url = "/apps";
//		var request = {
//			"url": url,
//			"dataType": "json",
//			"method": "POST",
//			"data" : angular.toJson(app)
//		}
//			
//		http(request).success(function(data){
//			deferred.resolve(data);
//		}).error(function(error){
//			deferred.reject(error.responseText);
//		});
//		return deferred.promise;
		sessionStorage.mpp_layout_data = angular.toJson(layoutdata);
		sessionStorage.mpp_layout_type = layouttype;
		deferred.resolve("save success");
		return deferred.promise;
	};
	
	return {
		'getLayout' : getLayout,
		'saveLayout' : saveLayout
	}
}


linkerCloud.factory('layoutService',['$http','$q',layoutService]);