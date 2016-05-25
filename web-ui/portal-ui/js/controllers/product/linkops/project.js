linkerCloud.controller('ProjectController', ['$scope','$location','$modal','projectService','responseService','confirmService',function($scope,$location,$modal,projectService,responseService,confirmService) {
      $scope.projects = [];
      $scope.envs = [];      
      $scope.jobs = []; 
      $scope.projectEnvs = [];    
      $scope.selectedEnv = {};
      $scope.selectedProject = {};
      $scope.selectedJob = {};
      $scope.show = {
      	 "content" : "project"
      };
      $scope.getLinkOpsEnvs = function(init){
          projectService.getLinkOpsEnvs().then(function(data){
          	  if(responseService.successResponse(data)){	    		
		    		$scope.envs = data.data;
		    		if(init){
		    			$scope.selectedEnv = data.data[0];
		    		}		    		
		      }
          },
            function(error){
		    	  	responseService.errorResp(error);  
		    });
      };
      $scope.showDetail = function(env){
          $scope.selectedEnv = env;         
      };
      $scope.refresh = function(){
          projectService.getEnvDetail($scope.selectedEnv.service_order_id).then(function(data){
           if(responseService.successResponse(data)){	    		
		    	  $scope.selectedEnv.gerrit_info = data.data.gerrit_info;
		    	  $scope.selectedEnv.jenkins_info = data.data.jenkins_info;
		    	  $scope.selectedEnv.nexus_info = data.data.nexus_info;
		    	  $scope.selectedEnv.ldap_info = data.data.ldap_info;
		    	  $scope.selectedEnv.status = data.data.status;
		       }
          },
          function(error){
          	   responseService.errorResp(error);  
          });         
      };
       $scope.orderLinkOpsEnvs = function(){
           	$modal.open({
	            templateUrl: 'templates/product/linkops/project/envCreate.html',
	            controller: 'CreateEnvController'
	          })
	          .result
	          .then(function (response) {
		            if (response.operation === 'execute') {
		                  projectService.createLinkOpsEnv(response.data).then(function(data){
						      if(responseService.successResponse(data)){	    		
						    		$scope.getLinkOpsEnvs(false);
						    		$scope.selectedEnv = data.data;
						      }
					        },
						    function(error){
						    	responseService.errorResp(error);  
						    });
		            }
	          });

      };
       $scope.confirmDeleteEnv = function(event,env){
    		event.preventDefault();
    		$scope.$translate(['rightContent.app.deleteConfirm', 'rightContent.linkops.deleteMessage', 'rightContent.app.deleteBtn']).then(function (translations) {
    		    $scope.confirm = {
    				"title" : translations['rightContent.app.deleteConfirm'],
    				"message" : translations['rightContent.linkops.deleteMessage'],
    				"buttons" : [
    					{
    						"text" : translations['rightContent.app.deleteBtn'],
    						"action" : function(){
    							$scope.deleteEnv(env);
    						}
    					}
    				]
    			};
    			confirmService.deleteConfirm($scope);
    		  });
    	};
      $scope.deleteEnv = function(env){
          projectService.deleteEnv(env.id).then(function(response){
              if(responseService.successResponse(response)){                          
                var init = true;
                if($scope.selectedEnv.id != env.id){
                	init = false;               	
                }
                $scope.getLinkOpsEnvs(init);    
                $scope.show.content = "project";                           
            }
          },
	       function(error){
	            responseService.errorResp(error);  
	       });
      };
      $scope.getProjects = function(envId){
           projectService.getProjects(envId).then(function(data){
		      if(responseService.successResponse(data)){	    		
		    		$scope.projects = data.data;
		      }
	        },
		    function(error){
		    	responseService.errorResp(error); 
		    });
      };
      $scope.createProject = function(){
           	$modal.open({
	            templateUrl: 'templates/product/linkops/project/projectCreate.html',
	            controller: 'CreateProjectController'
	          })
	          .result
	          .then(function (response) {
		            if (response.operation === 'execute') {
		                  var args = angular.merge({}, response.data, {"envId" : $scope.selectedEnv.id});
		                  projectService.createProject(args).then(function(data){
						      if(responseService.successResponse(data)){	    		
						    		 $scope.getProjects($scope.selectedEnv.id);
						    	}
					        },
						    function(error){
						    	responseService.errorResp(error);  
						    });
		            }
	          });
      };
     $scope.confirmDeleteProject = function(projectId){
    		$scope.$translate(['common.terminateConfirm', 'rightContent.linkops.terminateMessage', 'common.terminateConfirm']).then(function (translations) {
    		    $scope.confirm = {
    				"title" : translations['common.terminateConfirm'],
    				"message" : translations['rightContent.linkops.terminateMessage'],
    				"buttons" : [
    					{
    						"text" : translations['common.terminateConfirm'],
    						"action" : function(){
    							$scope.deleteProject(projectId);
    						}
    					}
    				]
    			};
    			confirmService.deleteConfirm($scope);
    		  });
    	};
      $scope.deleteProject = function(projectId){
          projectService.deleteProject(projectId).then(function(response){
              if(responseService.successResponse(response)){                          
                $scope.getProjects($scope.selectedEnv.id);                               
            }
          },
	       function(error){
	            responseService.errorResp(error);  
	       });
      };
      $scope.getJobs = function(project){                   
         projectService.getJobs(project.id).then(function(data){
		      if($scope.selectedProject.id != project.id){
           	     $scope.selectedProject = project;
              }   
		      $scope.show.content = "job";
		      if(responseService.successResponse(data)){	    		
		    		$scope.jobs = data.data;
		      }
	        },
		    function(error){
		    	responseService.errorResp(error); 
		    });
      };
      $scope.createJob = function(){
           	$modal.open({
	            templateUrl: 'templates/product/linkops/project/jobCreate.html',
	            controller: 'CreateJobController',
	            resolve: {
			          model: function () {
			            return {
			               project:$scope.selectedProject
			            };
			          }
		        }
	          })
	          .result
	          .then(function (response) {
		            if (response.operation === 'execute') {
		                  projectService.createJob(response.data).then(function(data){
      						      if(responseService.successResponse(data)){	    		
      						    		$scope.getJobs($scope.selectedProject);
      						      }
					        },
						    function(error){
						    		responseService.errorResp(error);
						    });
		            }
	          });

      };
      $scope.confirmDeleteJob = function(job){
    		$scope.$translate(['rightContent.app.deleteConfirm', 'rightContent.linkops.deleteMessage', 'rightContent.app.deleteBtn']).then(function (translations) {
    		    $scope.confirm = {
    				"title" : translations['rightContent.app.deleteConfirm'],
    				"message" : translations['rightContent.linkops.deleteMessage'],
    				"buttons" : [
    					{
    						"text" : translations['rightContent.app.deleteBtn'],
    						"action" : function(){
    							$scope.deleteJob(job);
    						}
    					}
    				]
    			};
    			confirmService.deleteConfirm($scope);
    		  });
    	};
      $scope.deleteJob = function(job){
          projectService.deleteJob(job,$scope.selectedProject).then(function(response){
              if(responseService.successResponse(response)){                          
                $scope.getJobs($scope.selectedProject);                               
            }
          },
	       function(error){
	            responseService.errorResp(error);  
	       });
      };
      $scope.confirmTerminateProjectEnv = function(job){
        $scope.$translate(['common.terminateConfirm', 'rightContent.linkops.terminateMessage', 'common.terminate']).then(function (translations) {
            $scope.confirm = {
            "title" : translations['common.terminateConfirm'],
            "message" : translations['rightContent.linkops.terminateMessage'],
            "buttons" : [
              {
                "text" : translations['common.terminate'],
                "action" : function(){
                  $scope.terminateProjectEnv(job);
                }
              }
            ]
          };
          confirmService.deleteConfirm($scope);
          });
      };
      $scope.terminateProjectEnv = function(projectEnvs){
          projectService.terminateProjectEnv(projectEnvs).then(function(response){
              if(responseService.successResponse(response)){                          
                $scope.getProjectEnvs($scope.selectedJob);                               
            }
          },
         function(error){
              responseService.errorResp(error);  
         });
      };
      $scope.getProjectEnvs = function(job){
          if($scope.selectedJob.id != job.id){
                 $scope.selectedJob = job;
              }   
          $scope.show.content = "projectEnvs";
          projectService.getProjectEnvs(job).then(function(response){
              if(responseService.successResponse(response)){                          
                    $scope.projectEnvs = response.data;                               
            }
          },
         function(error){
              responseService.errorResp(error);  
         });
      };
      $scope.deployJob = function(){
          projectService.deployJob($scope.selectedJob,$scope.selectedProject).then(function(response){
              if(responseService.successResponse(response)){  
                 // $scope.getJobs($scope.selectedProject);                               
                 $scope.getProjectEnvs($scope.selectedJob);                                             
          }},
         function(error){
              responseService.errorResp(error);  
         });
      };
      $scope.buildJob = function(job){
          projectService.buildJob(job,$scope.selectedProject).then(function(response){
              if(responseService.successResponse(response)){  
                 $scope.getJobs($scope.selectedProject);                               
                 // $scope.getProjectEnvs($scope.selectedJob); 
                  $modal.open({
                    templateUrl: 'templates/common/success.html',
                    controller: 'BuildSuccessController',               
                    size: "sm"
                  });                                           
               }},
             function(error){
                  responseService.errorResp(error);  
             });
      };
      $scope.getArtifacts = function(project){      
        projectService.getArtifacts(project.id).then(function(data){
          if($scope.selectedProject.id != project.id){
                 $scope.selectedProject = project;
              }   
          $scope.show.content = "artifact";
          if(responseService.successResponse(data)){          
             $scope.artifacts = data.data;
          }
          },
        function(error){
          responseService.errorResp(error); 
        });
      };
       $scope.createArtifact = function(project){
          $modal.open({
              templateUrl: 'templates/product/linkops/project/artifactCreate.html',
              controller: 'CreateArtifactController',
              resolve: {
                model: function () {
                  return {
                     project:$scope.selectedProject
                  };
                }
            }
            })
            .result
            .then(function (response) {
                if (response.operation === 'execute') {
                      projectService.createArtifact(response.data).then(function(data){
                  if(responseService.successResponse(data)){          
                      $scope.getArtifacts($scope.selectedProject);
                  }
                  },
                function(error){
                    responseService.errorResp(error);
                });
                }
            });

      };
      $scope.updateArtifact = function(artifact){
            $modal.open({
	            templateUrl: 'templates/product/linkops/project/artifactUpdate.html',
	            controller: 'UpdateArtifactController',
	            resolve: {
			          model: function () {
			            return {
			               project: $scope.selectedProject,
                     artifact : artifact
			            };
			          }
		        }
	          })
	          .result
	          .then(function (response) {
		            if (response.operation === 'execute') {
		                  projectService.updateArtifact(response.data).then(function(data){
						      if(responseService.successResponse(data)){	    		
						    		$scope.getArtifacts($scope.selectedProject);
						      }
					        },
						    function(error){
						    		responseService.errorResp(error);
						    });
		            }
	          });
      };
       $scope.confirmDeleteArtifact = function(artifact){
         $scope.$translate(['common.deleteConfirm', 'rightContent.linkops.deleteMessage', 'common.delete']).then(function (translations) {
            $scope.confirm = {
            "title" : translations['common.deleteConfirm'],
            "message" : translations['rightContent.linkops.deleteMessage'],
            "buttons" : [
              {
                "text" : translations['common.delete'],
                "action" : function(){
                  $scope.deleteArtifact(artifact);
                }
              }
            ]
          };
          confirmService.deleteConfirm($scope);
          });
      };
      $scope.deleteArtifact = function(artifact){
          projectService.deleteArtifact(artifact,$scope.selectedProject).then(function(response){
              if(responseService.successResponse(response)){                          
                $scope.getArtifacts($scope.selectedProject);                               
            }
          },
         function(error){
              responseService.errorResp(error);  
         });
      };

      $scope.gotoProjectPage = function(){
          $scope.show.content = "project";
      };
      $scope.gotoJobPage = function(){
          $scope.show.content = "job";
      };
       $scope.gotoArtifactPage = function(){
          $scope.show.content = "artifact";
      };
      $scope.getLinkOpsEnvs(true);

}])
.controller('CreateEnvController',  ['$scope', '$modalInstance', 
    function ($scope, $modalInstance) {
      $scope.envInfo = {"name" : ""};
      $scope.close = function (result) {
         	$modalInstance.close({"operation":result,"data":$scope.envInfo});        
      };
}])
.controller('CreateProjectController',  ['$scope', '$modalInstance', 'serviceModelDesignService','responseService',
    function ($scope, $modalInstance,serviceModelDesignService,responseService) {
      $scope.serviceGroups = [];
      $scope.selectedServiceGroup = {};
      serviceModelDesignService.listServiceGroup().then(function(response){
          if(responseService.successResponse(response)){
             $scope.serviceGroups = response.data;
             $scope.selectedServiceGroup = response.data[0] || {};
          }},
      function(errorMessage){
          responseService.errorResp(errorMessage);
      });
      $scope.projectInfo = {"name" : ""};
      $scope.close = function (result) {
         	var data = _.extend({},$scope.projectInfo,{"selectedServiceGroup":$scope.selectedServiceGroup.id});
          $modalInstance.close({"operation":result,"data":data});        
      };
}])
.controller('CreateJobController',  ['$scope', '$modalInstance', 'model',
    function ($scope, $modalInstance, model) {
      $scope.projectInfo = model.project;
      $scope.autoDeploy = false;
      $scope.toggleStatus = function(){
          $scope.autoDeploy = !$scope.autoDeploy;
      };
      $scope.jobInfo = {"name" : "","version":"","branch":"","projectId":model.project.id};
      $scope.close = function (result) {
         	var data = angular.merge({},$scope.jobInfo,{"autoDeploy":$scope.autoDeploy});
          $modalInstance.close({"operation":result,"data":data});        
      };
}])
.controller('CreateArtifactController',  ['$scope', '$modalInstance', 'model','contentService','responseService',
    function ($scope, $modalInstance, model,contentService,responseService) {
      $scope.dockerfiles = [];
      $scope.selectedDockerfiles = [];
      $scope.selectAll = false;
      var getDockerfiles = function(){
          contentService.getUploadedContents().then(function(response){
              if(responseService.successResponse(response)){          
                $scope.dockerfiles = response.data;
            }
          },
          function(error){
            responseService.errorResp(error);  
         });
      }();
      $scope.modifySelectedFile = function(event,dockefileId){
             if(event.currentTarget.checked){
               $scope.selectedDockerfiles.push(dockefileId);
             }else{
               $scope.selectedDockerfiles = _.without($scope.selectedDockerfiles,dockefileId);
             }
      };
     
      $scope.selectAllOperation = function(){
        $scope.selectAll = !$scope.selectAll;
        if($scope.selectAll){
           $scope.selectedDockerfiles = _.map($scope.dockerfiles,function(dockerfile){return dockerfile.id});
        }else{
           $scope.selectedDockerfiles = [];
        }
      };
      $scope.projectInfo = model.project;
      $scope.artifactInfo = {"name" : "", "groupId":"","type":"" ,"projectId":model.project.id};
      $scope.close = function (result) {
          var data = _.extend({},$scope.artifactInfo,{"selectedDockerfiles":$scope.selectedDockerfiles});
          $modalInstance.close({"operation":result,"data":data});        
      };
}])
.controller('UpdateArtifactController',  ['$scope', '$modalInstance', 'model','contentService','responseService',
    function ($scope, $modalInstance, model,contentService,responseService) {
      $scope.dockerfiles = [];
      $scope.selectedDockerfiles = model.artifact.df_ids || [];
      // $scope.operation = {"selectAll":false};
     
      var getDockerfiles = function(){
          contentService.getUploadedContents().then(function(response){
              if(responseService.successResponse(response)){          
                
                var initObj = {};
                _.each(response.data,function(dockerfile){
                    if($scope.selectedDockerfiles.indexOf(dockerfile.id) != -1){
                        initObj = {"initSelected" : true};
                       
                    }else{
                        initObj = {"initSelected" : false};
                    }
                    _.extend(dockerfile,initObj);
                });
               $scope.dockerfiles = response.data; 
            }
          },
          function(error){
            responseService.errorResp(error);  
         });
      }();
      $scope.modifySelectedFile = function(event,dockerfile){
              // _.each($scope.dockerfiles,function(dockerfile){
              //      dockerfile.initSelected = false;
              // });
             if(event.currentTarget.checked){
               
               $scope.selectedDockerfiles.push(dockerfile.id);
                _.each($scope.dockerfiles,function(df){
                  if(df.id == dockerfile.id)
                   dockerfile.initSelected = true;
                });
             }else{
               $scope.selectedDockerfiles = _.without($scope.selectedDockerfiles,dockerfile.id);
                 _.each($scope.dockerfiles,function(df){
                  if(df.id == dockerfile.id)
                   dockerfile.initSelected = false;
                });
             }
             // $scope.$apply();
      };
     
      $scope.selectAllOperation = function(event){
            // _.each($scope.dockerfiles,function(dockerfile){
            //       dockerfile.initSelected = false;
            // });
            // $scope.operation.selectAll = !$scope.operation.selectAll;
            if(event.currentTarget.checked){
               $scope.selectedDockerfiles = _.map($scope.dockerfiles,function(dockerfile){return dockerfile.id});
               angular.forEach($scope.dockerfiles,function(dockerfile){
                  dockerfile.initSelected = true;
                });
            }else{
               $scope.selectedDockerfiles = [];
               angular.forEach($scope.dockerfiles,function(dockerfile){
                  dockerfile.initSelected = false;
                });
            }
            // $scope.$apply();
      };
      $scope.projectInfo = model.project;
      $scope.artifactInfo = {"id":model.artifact.id,"name" : model.artifact.name, "groupId":model.artifact.group_id,"type":model.artifact.type,"projectId":model.project.id};
      $scope.close = function (result) {
          var data = _.extend({},$scope.artifactInfo,{"selectedDockerfiles":$scope.selectedDockerfiles});
          $modalInstance.close({"operation":result,"data":data});        
      };
}])
.controller('BuildSuccessController',  ['$scope', '$modalInstance', 
    function ($scope, $modalInstance) {   
      $scope.close = function (result) {
          $modalInstance.close();        
      };
}]);