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