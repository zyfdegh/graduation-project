function responseService($window,$modal){
	var successResponse = function(response){
//		if(_.isString(response) && response.indexOf("<!DOCTYPE html>") != -1 ){
//	 	    $window.location = "/portal-ui/login.html";
//	 	    return false;
//	 	}else if(response.reply != 1){
//	 	   	layer.alert(response.replyDesc, {    	 
//			 	icon: 2
//			});
//			return false;
//	 	}else {
	 	    return true;
//	 	}
	};
	
	var errorResponse = function(message,statusCode){
		  
		  $modal.open({
		        templateUrl: 'templates/common/fail.html',
		        controller: 'ActionSuccessBoxCtrl',
		        size: 'sm',
		        resolve: {
		          model: function () {
		            return {
		              id:message
		            };
		          }
		        }
	      })
	      .result
	      .then(function (result) {
	           if(statusCode == 401 || statusCode == 402){
                   sessionStorage.clear();
                   $window.location = "/portal-ui/login.html";
	           }
	      });
	};
	var errorResp = function(error){
		
		error = error || {"name" : "Controller Exception"};
		$modal.open({
		        templateUrl: 'templates/common/fail.html',
		        controller: 'ActionSuccessBoxCtrl',
		        size: 'sm',
		        resolve: {
		          model: function () {
		            return {
		              id:error.name
		            };
		          }
		        }
	      })
	      .result
	      .then(function (result) {
	           if(error.code == 401 || error.code == 402){
                   sessionStorage.clear();
                   $window.location = "/portal-ui/login.html";
	           }
	      });
	}
	var checkSession = function(){
		if(_.isUndefined(sessionStorage.username) || _.isEmpty(sessionStorage.username)){
			return false;
		}else{
		    return true;	   
		}
	};
	
	return {
		"successResponse" : successResponse,
		"errorResponse" : errorResponse,
		"checkSession" : checkSession,
		"errorResp" : errorResp
	}
}
   
linkerCloud.factory('responseService', ['$window', '$modal',responseService]);