function billingService(http,q){
	var getAllBillings = function(){
		var deferred = q.defer();
		var url="/billing";	
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
	};
	
	var newBilling = function(billing){
		var deferred = q.defer();
		var url = "/billing";
		var request = {
			"url": url,
			"dataType": "json",
			"method": "POST",
			"data" : angular.toJson(billing)
		}
			
		http(request).success(function(data){
			deferred.resolve(data);
		}).error(function(error){
			deferred.reject(error);
		});
		return deferred.promise;
	};
	
	var updateBilling = function(billing){
		var deferred = q.defer();
		var url = "/billing/" + billing._id;
		var request = {
			"url": url,
			"dataType": "json",
			"method": "PUT",
			"data" : angular.toJson(billing)
		}
			
		http(request).success(function(data){
			deferred.resolve(data);
		}).error(function(error){
			deferred.reject(error);
		});
		return deferred.promise;
	};
	
	var getBillRecords = function(query,skip,limit){
		var deferred = q.defer();
		var url="/userAccounts?count=true&query="+ query + "&skip="+skip+"&limit="+limit;	
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
	};
	
	var getAllBillingsNoAuth = function(){
		var deferred = q.defer();
		var url="/billing/all";	
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
	};
	
	return {
		'getAllBillings' : getAllBillings,
		'newBilling' : newBilling,
		'updateBilling' : updateBilling,
		'getBillRecords' : getBillRecords,
		'getAllBillingsNoAuth' : getAllBillingsNoAuth
	}
}


linkerCloud.factory('billingService',['$http','$q',billingService]);