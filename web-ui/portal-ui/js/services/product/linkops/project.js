function projectService(http,q){
	return{

		getLinkOpsEnvs : function(){
			var deferred = q.defer();
			var url = "/linkops/env";
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
		createLinkOpsEnv : function(data){
			var deferred = q.defer();
			var url = "/linkops/env";
			var request = {
				"url": url,
				"dataType": "json",
				"data" : angular.toJson(data),
				"method": "POST"
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
		getEnvDetail : function(serviceOrderId){
			var deferred = q.defer();
			var url = "/linkops/env/" + serviceOrderId;
			var request = {
				"url": url,
				"dataType": "json",
				"method": "POST"
			}
				
			http(request).success(function(data){
				deferred.resolve(data);
			}).error(function(error){
				deferred.reject(error);
			});
			return deferred.promise;
		},
		deleteEnv : function(envId){
            var deferred = q.defer();
			var url = "/linkops/env/" + envId;
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
        getProjects : function(envId){
			var deferred = q.defer();
			var url = "/linkops/project/"+envId;
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
		deleteProject : function(projectId){
			var deferred = q.defer();
			var url = "/linkops/project/"+projectId;
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
		createProject : function(args){
			var deferred = q.defer();
			var url = "/linkops/project" ;
			var request = {
				"url": url,
				"dataType": "json",
				"method": "POST",
				"data" : angular.toJson(args)
			}
				
			http(request).success(function(data){
				deferred.resolve(data);
			}).error(function(error){
				deferred.reject(error.responseText);
			});
			return deferred.promise;
		},
		getJobs: function(projectId){
            var deferred = q.defer();
			var url = "/linkops/job/"+projectId;
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
		createJob : function(args){
			var deferred = q.defer();
			var url = "/linkops/job" ;
			var request = {
				"url": url,
				"dataType": "json",
				"method": "POST",
				"data" : angular.toJson(args)
			}
				
			http(request).success(function(data){
				deferred.resolve(data);
			}).error(function(error){
				deferred.reject(error.responseText);
			});
			return deferred.promise;
		},
		deleteJob : function(job, project){
            var deferred = q.defer();
            var url = "/linkops/job/" + job.name+"?projectId=" + project.id;
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
		terminateProjectEnv : function(projectEnvs){
            var deferred = q.defer();
            var url = "/linkops/projectenvs/" + projectEnvs.id;
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
		getProjectEnvs : function(job){
            var deferred = q.defer();
            var url = "/linkops/projectenvs/" + job.id;
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
		},
		buildJob : function(job,project){
            var deferred = q.defer();
            var url = "/linkops/job/" + job.id + "?projectId=" + project.id;
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
		},
		deployJob : function(job,project){
            var deferred = q.defer();
            var url = "/linkops/job/" + job.id + "/jobenv";
			var request = {
				"url": url,
				"dataType": "json",
				"method": "POST"				
			}
				
			http(request).success(function(data){
				deferred.resolve(data);
			}).error(function(error){
				deferred.reject(error.responseText);
			});
			return deferred.promise;
		},
		getArtifacts : function(project){
            var deferred = q.defer();
			var url = "/linkops/artifact/"+project;
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
		createArtifact : function(args){
			var deferred = q.defer();
			var url = "/linkops/artifact" ;
			var request = {
				"url": url,
				"dataType": "json",
				"method": "POST",
				"data" : angular.toJson(args)
			}
				
			http(request).success(function(data){
				deferred.resolve(data);
			}).error(function(error){
				deferred.reject(error.responseText);
			});
			return deferred.promise;
		},
		updateArtifact : function(artifact){
 			var deferred = q.defer();
            var url = "/linkops/artifact/" + artifact.id+"?projectId="+artifact.projectId;
			var request = {
				"url": url,
				"dataType": "json",
				"method": "PUT",
				"data" : angular.toJson(artifact)
			}
				
			http(request).success(function(data){
				deferred.resolve(data);
			}).error(function(error){
				deferred.reject(error.responseText);
			});
			return deferred.promise;
		},
		deleteArtifact : function(artifact,project){
            var deferred = q.defer();
            var url = "/linkops/artifact/" + artifact.id+"?projectId=" + project.id;
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
		}

    }

}


linkerCloud.factory('projectService',['$http','$q',projectService]);