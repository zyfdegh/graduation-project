function dashboardService(http,q){
	return{
		getServiceInstances : function(){
			var deferred = q.defer();
			var url = "/metrix/serviceInstance";
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
		},
		
		getResources : function(){
			var deferred = q.defer();
			var url = "/metrix/resource";
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
		}
	   	

    }

}


linkerCloud.factory('dashboardService',['$http','$q',dashboardService]);