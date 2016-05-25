function contentService(http,q){
	return{

		getUploadedContents : function(){
			var deferred = q.defer();
			var url = "/linkops/content";
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
		
		deleteUploadedContents : function(contentId){
			var deferred = q.defer();
			var url = "/linkops/content/" + contentId;
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
		},
	    changeStatus : function(contentId){
			var deferred = q.defer();
			var url = "/linkops/content/" + contentId;
			var request = {
				"url": url,
				"dataType": "json",
				"method": "PUT"
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


linkerCloud.factory('contentService',['$http','$q',contentService]);