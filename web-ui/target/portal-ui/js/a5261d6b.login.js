var login = angular.module('login',['ngRoute','ui.bootstrap','pascalprecht.translate','ngCookies']);

login.config(function($routeProvider, $locationProvider) {
	$routeProvider
		.when('/', {
    		templateUrl: 'templates/login/signin.html',
    		controller: 'LoginController'
  		})
  		.when('/signup', {
    		templateUrl: 'templates/login/signup.html',
    		controller: 'SignUpController'
  		})
      .when('/activeSuccess', {
        templateUrl: 'templates/login/active_success.html',
        controller: 'ActiveSuccessController'
      })
      .when('/activeFailed', {
        templateUrl: 'templates/login/active_failed.html',
        controller: 'ActiveFailedController'
      })
  		.otherwise({  
            redirectTo: '/'
       });
})
.config(['$translateProvider',function($translateProvider) {
     var lang="zh";
     $translateProvider.useStaticFilesLoader({
        prefix: '/portal-ui/locales/',
        suffix: '.json'
     });
     $translateProvider.useLocalStorage();
     $translateProvider.useSanitizeValueStrategy('escape');
     $translateProvider.preferredLanguage(lang);    
}])
.run(['$rootScope', '$translate',
    function ($rootScope, $translate) {
       $rootScope.$translate = $translate;      
    }
  ])
.run(['$rootScope',
    function ($rootScope) {
      // Wait for window load
      $(window).load(function () {
        $rootScope.loaded = true;
      });
    }
  ]);
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

function loginService(http,q){
	return {
		checkFunc : function(key,value){
			var message="",status=true;
			if (_.isEmpty(value)) {
				message = key + " is required";
				status = false;			
			} 
			return {"status":status,"message":message};
		},
		checkConfirmPwd : function(pwd,confirmPwd){
            var message="",status=true;
            if (_.isEmpty(confirmPwd)) {
				message = "confirm password is required";
				status = false;	
					
			}
//          else if(confirmPwd != pwd){
//              message = "inconsistent with the password";
//              status = false;    
//                    
//          } 
            return {"status":status,"message":message};	        
		},
		checkWithPwd : function(pwd,confirmPwd){
			var status=true;
			if(confirmPwd != pwd){
                status = false;                       
            } 
            return {"status":status};	   
		},
		checkNamespacePattern : function(namespace){
            var status = true;
            var reg = /^[0-9a-z]+$/;
         
            if(namespace!="" && !reg.test(namespace)){
				status = false;	
            }
            return {"status":status};
		},
	    doLogin : function(data){
			var url = "/user/login";
			var request = {
			    "url": url,
			    "dataType": "json",
			    "method": "POST",
			    "data": JSON.stringify(data)
			}			   			    
			var deferred = q.defer();
			http(request).success(function(response){					
				 deferred.resolve(response);
			}).error(function(error){
				deferred.reject(error);
			});
			return deferred.promise;
		},
		doSignUp : function(data){
			var url = "/user/registry";
			var request = {
			    "url": url,
			    "dataType": "json",
			    "method": "POST",
			    "data": JSON.stringify(data)
			}				   			    
			var deferred = q.defer();
			http(request).success(function(response){					
				 deferred.resolve(response);
			}).error(function(error){
				deferred.reject(error);
			});
			return deferred.promise;
		},
		doLogout : function(){
			var url="/logout";
			var request = {
			    "url" : url,
			    "method" : "GET"
			}
			var deferred = q.defer();
			http(request).success(function(response){
				deferred.resolve(response);
			}).error(function(error){
				deferred.reject(error);
			});
			return deferred.promise;
		},
        reactive : function(uid){
            var url="/user/reactive?uid=" + uid;
			var request = {
			    "url" : url,
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

// function isEmail(email){
// 	var filter  = /^([a-zA-Z0-9_\.\-])+\@(([a-zA-Z0-9\-])+\.)+([a-zA-Z0-9]{2,4})+$/;
// 	if (filter.test(email)) 
// 	return true;
// 	return false;
// };
   
login.factory('loginService', ['$http','$q',loginService]);
login.directive('ngEnter', function () {
    return function (scope, element, attrs) {
        element.bind("keydown keypress", function (event) {
            if(event.which === 13) {
                scope.$apply(function (){
                    scope.$eval(attrs.ngEnter);
                });

                event.preventDefault();
            }
        });
    };
});