linkerCloud.controller('TenantIdentityController', ['$scope','$location','tenantService','responseService', function($scope,$location,tenantService, responseService) {
     $scope.tenants = [];

      $scope.getTenants = function(){
           tenantService.getTenants().then(function(response){
		      if(responseService.successResponse(response)){	    		
		    		$scope.tenants = response.data;
		      }
	        },
		    function(error){
		    	responseService.errorResp(error);		    
		    });
      };
      $scope.getTenants();
}]);