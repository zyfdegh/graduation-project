function fileUploadService(http,q){
	return{

		uploadFile : function(data){
			var deferred = q.defer();
			var url = "/linkops/uploadFile";
            
            var formData = new FormData();
	        var dockerfileName = data.dockerfile == "" ? "Dockerfile" : data.dockerfile;
		     formData.append('imagename', data.imagename);
             formData.append('dockerfile', dockerfileName);
             formData.append('version', data.version);
             formData.append('file', data.file);
             formData.append('email', sessionStorage.username);
             var req = {
				 method: 'POST',
				 url: url,
				 headers: {
				   'Content-Type': undefined
				 },
				 data: formData
				}
			http(req)
			.success(function(data){
				deferred.resolve(data);
			}).error(function(error){
				deferred.reject(error);
			});
			return deferred.promise;
		}
	

    }

}


linkerCloud.factory('fileUploadService',['$http','$q',fileUploadService]);