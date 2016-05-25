linkerCloud.controller('ContentController', ['$scope','$location','responseService', 'contentService','confirmService',function($scope,$location,responseService,contentService,confirmService) {
     $scope.contentCount = ["1"];
     $scope.contents = [];
     $scope.addContentComponent = function(){
        $scope.contentCount.push("1");
     };
     $scope.getContentList = function(){
          contentService.getUploadedContents().then(function(response){
              if(responseService.successResponse(response)){          
                $scope.contents = response.data;
              }
          },
          function(error){
            responseService.errorResp(error);  
         });
     };
     $scope.confirmDeleteContent = function(contentId){
    		$scope.$translate(['rightContent.app.deleteConfirm', 'rightContent.linkops.deleteMessage', 'rightContent.app.deleteBtn']).then(function (translations) {
    		    $scope.confirm = {
    				"title" : translations['rightContent.app.deleteConfirm'],
    				"message" : translations['rightContent.linkops.deleteMessage'],
    				"buttons" : [
    					{
    						"text" : translations['rightContent.app.deleteBtn'],
    						"action" : function(){
    							$scope.deleteContent(contentId);
    						}
    					}
    				]
    			};
    			confirmService.deleteConfirm($scope);
    		  });
    	};
     $scope.deleteContent = function(contentId){
     	 contentService.deleteUploadedContents(contentId).then(function(response){
              if(responseService.successResponse(response)){          
                $scope.getContentList();
            }
          },
          function(error){
            responseService.errorResp(error);  
         });
     };
     $scope.changeStatus = function(contentId){
          $modal.open({
                templateUrl: 'templates/product/linkops/content/contentUpdate.html',
                controller: 'UpdateContentController',
                resolve: {
                      model: function () {
                        return {
                           contentId: contentId
                        };
                      }
                }
              })
              .result
              .then(function (response) {
                    if (response.operation === 'execute') {
                          contentService.changeStatus(response.data).then(function(data){
                              if(responseService.successResponse(data)){                
                                    $scope.getContentList();
                              }
                            },
                            function(error){
                                    responseService.errorResp(error);
                            });
                    }
              });
     };

}])
.controller('UpdateContentController',  ['$scope', '$modalInstance', 'model',
    function ($scope, $modalInstance, model) {
      $scope.contentInfo = {"id" : model.content.id,"status":""};
      $scope.close = function (result) {
            $modalInstance.close({"operation":result,"data":$scope.contentInfo});        
      };
}]);