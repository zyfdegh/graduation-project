function dockerImageService(http,q){
	return {		
	    getDockerImages : function(){
			var url = "/docker/image";
			var request = {
			    "url": url,
			    "dataType": "json",
			    "method": "GET"
			    		   
			}			   			    
			var deferred = q.defer();
			http(request).success(function(response){					
				 deferred.resolve(response);
			}).error(function(error){
				deferred.reject(error);
			});
			return deferred.promise;
		},
		getImageTags : function(imageName){
			var url = "/docker/imageTag?imageName="+encodeURIComponent(imageName);
			var request = {
				"url" : url,
				"dataType" : "json",
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

linkerCloud.factory('dockerImageService', ['$http','$q',dockerImageService]);