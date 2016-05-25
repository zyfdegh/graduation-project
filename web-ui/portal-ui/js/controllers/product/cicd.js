linkerCloud.controller('CICDController', ['$scope','$location','$modal','CICDService','responseService',function($scope,$location,$modal,CICDService,responseService) {
      
      $scope.projects = [];

      $scope.getProjects = function(){
           CICDService.getProjects().then(function(data){
		      if(responseService.successResponse(data)){	    		
		    		$scope.projects = data.data;
		      }
	        },
		    function(errorMessage){
		    	// responseService.errorResponse("Failed to get projects.");
		    	$scope.projects = errorMessage;
		    });
      };
      
      $scope.createProject = function(){
           	$modal.open({
	            templateUrl: 'templates/product/cicd/projectCreate.html',
	            controller: 'CreateProjectController'
	          })
	          .result
	          .then(function (response) {
		            if (response.operation === 'execute') {
		                  CICDService.createProject(response.data).then(function(data){
						      if(responseService.successResponse(data)){	    		
						    		$scope.getProjects();
						      }
					        },
						    function(errorMessage){
						    	// responseService.errorResponse("Failed to get projects.");
						    	// $scope.projects = errorMessage;
						    	$scope.getProjects();
						    });
		            }
	          });

      };
      $scope.projectSource = "choose";
      $scope.getProjects();

}])
.controller('CreateProjectController',  ['$scope', '$modalInstance', 
    function ($scope, $modalInstance) {
      $scope.projectInfo = {"name" : "", "description":""};
      $scope.close = function (result) {
         	$modalInstance.close({"operation":result,"data":$scope.projectInfo});        
      };
}]);