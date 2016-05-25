function servicesService(http,q){
	var listMyServices = function(){
		var deferred = q.defer();
		var url = "/serviceGroupOrders?query={\"life_cycle_status\": {\"$ne\":\"TERMINATED\"}}";
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
	
	var getService = function(serviceid){
		var deferred = q.defer();
		var url = "/groupInstances/" + serviceid;
		var request = {
			"url": url,
			"dataType": "json",
			"method": "GET"
		}
			
		http(request).success(function(data){
			deferred.resolve(data);
		}).error(function(error){
			deferred.reject(error.responseText);
		});
		return deferred.promise;
	};
	
//	var doScaleTo = function(groupInstanceId, appId, num){
//		var deferred = q.defer();
//		var url = "/groupInstances/scaleApp/" + groupInstanceId + "?appId=" + appId + "&num=" + num;
//		var request = {
//			"url": url,
//			"dataType": "json",
//			"method": "PUT"
//		}
//			
//		http(request).success(function(data){
//			deferred.resolve(data);
//		}).error(function(error){
//			deferred.reject(error.responseText);
//		});
//		return deferred.promise;
//	};
	
	var deleteServiceInstance = function(orderid){
		var deferred = q.defer();
		var url = "/serviceGroupOrders/" + orderid;
		var request = {
			"url": url,
			"dataType": "json",
			"method": "DELETE"
		}
			
		http(request).success(function(data){
			deferred.resolve(data);
		}).error(function(error){
			deferred.reject(error.responseText);
		});
		return deferred.promise;
	};
	
//	var readIPAddress = function(instanceid){
//		var deferred = q.defer();
//		var url = "/appInstances/" + instanceid;
//		var request = {
//			"url": url,
//			"dataType": "json",
//			"method": "GET"
//		}
//			
//		http(request).success(function(data){
//			deferred.resolve(data);
//		}).error(function(error){
//			deferred.reject(error.responseText);
//		});
//		return deferred.promise;
//	};
	
	var getSGIOperations = function(sgi_id){
		var deferred = q.defer();
		var url = "/serviceGroupOrders/operations/" + sgi_id;
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
		'listMyServices' : listMyServices,
		'getService' : getService,
//		'doScaleTo' : doScaleTo,
		'deleteServiceInstance' : deleteServiceInstance,
//		'readIPAddress' : readIPAddress
		'getSGIOperations' : getSGIOperations
	}
}


linkerCloud.factory('servicesService',['$http','$q',servicesService]);