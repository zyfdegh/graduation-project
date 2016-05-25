function logoutService(http,q){
	return {
				
		doLogout : function(){
			var url="/logout";
			var request = {
			    "url" : url,
			    "method" : "GET"
			}
			var deferred = q.defer();
			http(request).success(function(response){
				deferred.resolve(response);
			}).error(function(error){
				deferred.reject(error);
			});
			return deferred.promise;
		}
		
	}
}

   
linkerCloud.factory('logoutService', ['$http','$q',logoutService]);