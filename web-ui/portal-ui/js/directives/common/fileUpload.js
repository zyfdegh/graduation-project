linkerCloud.controller('fileUploadCtrl', ['$scope', 'fileUploadService', 'responseService', function($scope,fileUploadService,responseService){
    $scope.contentData = {
       imagename : "",
       dockerfile : "Dockerfile",
       file: "",
       version:""
    }; 
    $scope.uploadStatus = "";
    $scope.buttonAvailable = true;
    $scope.finished = false;
    $scope.ongoing = false;
    $scope.doUpload = function(){
       $scope.ongoing = true;
       $scope.buttonAvailable = false;
       fileUploadService.uploadFile($scope.contentData).then(function(response){
            $scope.ongoing = false;
            if(responseService.successResponse(response)){          
                $scope.contentData = {
                   imagename : "",
                   dockerfile : "Dockerfile",
                   file: {},
                   version: ""
                };
                $scope.buttonAvailable = true;
                $scope.uploadStatus = "Uploaded. Waitting for Linker's review."
                $scope.finished = true;
                
            }
          },
          function(error){
            responseService.errorResp(error);  
            $scope.uploadStatus = "Upload Failed."  
            $scope.buttonAvailable = true;  
            $scope.ongoing = false; 
          });
    }
}]);
 linkerCloud.directive('fileupload', function () {
    return {
       restrict : 'EA',
       scope : {},
       templateUrl : "templates/common/fileUpload.html",
       replace: true,
       controller : 'fileUploadCtrl',
       link : function postLink(scope, element, attrs) {
            element.bind('change', function (event) {
                  if(event.target.type == "file"){
                    scope.contentData.file = event.target.files[0];
                  }                 
            });
        }
    }
});
