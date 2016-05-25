function tenantService(http,q){
	return{

		getTenants : function(){
			var deferred = q.defer();
			var url = "/tenant";
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
		}
	

    }

}


linkerCloud.factory('tenantService',['$http','$q',tenantService]);