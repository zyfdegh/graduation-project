login.controller('LoginController', ['$scope', '$modal','loginService', function($scope,$modal,loginService) {
	var storage = $scope.$translate.storage();
    var key = $scope.$translate.storageKey();
    var lang = storage.get(key)||'zh';
    $scope.$translate.use(lang);
	$scope.userInfo = {"email":"","password":""};	
	$scope.userNameResult = {"status":true,"message":""};
	$scope.passwordResult = {"status":true,"message":""};
	$scope.doLogin = function(){
        $scope.userNameResult = loginService.checkFunc("email", $scope.userInfo.email);
		$scope.passwordResult = loginService.checkFunc("password", $scope.userInfo.password);

		if ($scope.userNameResult.status == true && $scope.passwordResult.status == true) {			
			loginService.doLogin($scope.userInfo).then(function(response){				         	   
	 	    		sessionStorage.username=$scope.userInfo.email;
	 	    		// sessionStorage.token=response.data.id;
	 	    		sessionStorage.namespace=response.data.alias;
	 	    		window.location="/portal-ui/index.html";	 	    	
	 	    },function(errorMessage){	 	    	
                $modal.open({
				        templateUrl: 'templates/login/fail.html',
				        controller: 'OperateFailedCtrl',
				        size: 'sm',
				        resolve: {
				          model: function () {
				            return {				              
				              message: errorMessage
				            };
				          }
				        }
			      })			                       
	 	    });

		}else{

			return false;
		} 
	};
	
}])
.controller('SignUpController', ['$scope','$modal','loginService', function($scope,$modal,loginService){
	 $scope.policy={"agree":false};
	 $scope.userInfo = {"alias":"","email":"","password":"","confirmpassword":"","company":"","address":"","phonenumber":"","infosource":"community"};	
	 $scope.emailResult = {"status":true,"message":""};
	 $scope.passwordResult = {"status":true,"message":""};
	 $scope.confirmPwdResult = {"status":true,"message":""};
	 $scope.confirmCoPwdResult = {"status":true,"message":""};
	 $scope.namespaceResult = {"status":true,"message":""};
	 $scope.namespacePatternResult = {"status":true,"message":""};
	 $scope.register = {"step" : "step1"};
	 $scope.toggleStatus = function(){
         $scope.policy.agree = !$scope.policy.agree;
	 };
	

	 $scope.doSignUp = function(){   
	    $scope.namespaceResult = loginService.checkFunc("namespace", $scope.userInfo.alias);
	    $scope.namespacePatternResult = loginService.checkNamespacePattern($scope.userInfo.alias);
	    $scope.emailResult = loginService.checkFunc("email", $scope.userInfo.email);   
		$scope.passwordResult = loginService.checkFunc("password", $scope.userInfo.password);
		$scope.confirmPwdResult = loginService.checkConfirmPwd($scope.userInfo.password, $scope.userInfo.confirmpassword);
		$scope.confirmCoPwdResult = loginService.checkWithPwd($scope.userInfo.password, $scope.userInfo.confirmpassword);
		

		if ($scope.namespaceResult.status == true && $scope.namespacePatternResult.status == true && $scope.passwordResult.status == true && $scope.confirmPwdResult.status == true && $scope.confirmCoPwdResult.status == true && $scope.emailResult.status == true) {			
			loginService.doSignUp($scope.userInfo).then(function(response){	 	   
	 	    	
                 $scope.register.step = "step2";
	 	    		    	
	 	    },function(errorMessage){	 	    	
                $modal.open({
				        templateUrl: 'templates/login/fail.html',
				        controller: 'OperateFailedCtrl',
				        size: 'sm',
				        resolve: {
				          model: function () {
				            return {				              
				              message:errorMessage
				            };
				          }
				        }
			      })
			     
	 	    });

		}else{

			return false;
		} 
	};

}])
.controller('OperateSuccessCtrl', ['$scope', '$modalInstance', 'model',
	 function ($scope, $modalInstance, model) {
	      $scope.message = model.message;
	      $scope.close = function (res) {
	         $modalInstance.close(res);
	      };
    }
])
.controller('OperateFailedCtrl', ['$scope', '$modalInstance', 'model',
	 function ($scope, $modalInstance, model) {
	      $scope.message = model.message;
	      $scope.close = function (res) {
	         $modalInstance.close(res);
	      };
    }
])
.controller('ActiveSuccessController', ['$scope', 
	 function ($scope) {
	      
    }
])
.controller('ActiveFailedController', ['$scope', '$location','loginService','$modal',
	 function ($scope,$location,loginService,$modal) {	      
	      $scope.params=$location.search();
	      $scope.reactive = function () {
	          loginService.reactive($scope.params.uid).then(function(response){
                  // location.path="/login.html#/activeSuccess";
                   $modal.open({
				        templateUrl: 'templates/login/success.html',
				        controller: 'ReactiveSuccessCtrl',
				        size: 'sm',
				        resolve: {
				          model: function () {
				            return {				              
				              message:"We have send an active email to you, please active your account again."
				            };
				          }
				        }
			      });
	          },function(errorMessage){
                  location.path="/login.html#/activeFailed?uid="+$scope.params.uid;
	          })
	      };
    }
])
.controller('ReactiveSuccessCtrl',  ['$scope', '$modalInstance', 'model',
    function ($scope, $modalInstance,model) {   
      $scope.message = model.message;
      $scope.close = function (result) {
          $modalInstance.close();        
      };
}]);
