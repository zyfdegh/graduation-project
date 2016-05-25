linkerCloud.controller('ServicesController', ['$scope','$location','productsService','servicesService','responseService','confirmService',
	function($scope,$location,productsService,servicesService,responseService,confirmService) {

	$scope.showDetail = false;
	
	$scope.listMyServices = function(){
	    servicesService.listMyServices().then(function(data){
	    	if(responseService.successResponse(data)){
	    		_.each(data.data,function(order){
	    			order.displayName = idToSimple(order.service_group_id);
	    			order.imageSrc = "images/products/Appserver.png";
	    		})
	    		$scope.runningOrders = data.data;
	    	}
	    },
	    function(errorMessage){
	    		responseService.errorResp(errorMessage);
	    });
  	};
  	
  	$scope.getServiceDetail = function(order){
  		servicesService.getSGIOperations(order._id).then(function(data){
  		 	if(responseService.successResponse(data)){
  		 		allow_scaleapp_sgo = data.data.scaleapp_sgo == 1 ? true : false;
  		 		allow_metering_sgo = data.data.metering_sgo == 1 ? true : false;
  		 		$scope.allow_delete_sgo = data.data.delete_sgo == 1 ? true : false;
  		 		
  		 		$scope.showDetail = true;
//				$scope.detailTitle = "Service Details";
				selectedOrder = order;
				$scope.getInstanceByOrder();
  		 	}
	    },
	    function(errorMessage){
	    		responseService.errorResp(errorMessage);
	    });
	}
  	
  	$scope.getInstanceByOrder = function(){
  		servicesService.getService(selectedOrder.service_group_instance_id).then(function(data){
		    	if(responseService.successResponse(data)){
		    		showServiceDetail(data.data);
		    	}
	    },
	    function(errorMessage){
	   	 	responseService.errorResp(errorMessage);
	    });
  	}
  	
  	var showServiceDetail = function(service){
  		service.displayName = idToSimple(service.id);
	    service.imageSrc = "images/products/Appserver.png";
  		_.each(service.groups,function(group){
			allocateImageToApp(group);
		});
		selectedService = service;
		setTimeout(function(){
			drawServiceTree();
		},200);
  	}
  	
  	$scope.goBackToServiceList = function(){
		$scope.showDetail = false;
		$("body").scrollTop(0);
	}
  	
  	$scope.confirmDeleteServiceInstance = function(){
  		$scope.$translate(['rightContent.serviceSubscriptions.terminateConfirm', 'rightContent.serviceSubscriptions.terminateMessage', 'rightContent.serviceSubscriptions.terminateBtn']).then(function (translations) {
			$scope.confirm = {
				"title" : translations['rightContent.serviceSubscriptions.terminateConfirm'],
				"message" : translations['rightContent.serviceSubscriptions.terminateMessage'],
				"buttons" : [
					{
						"text" : translations['rightContent.serviceSubscriptions.terminateBtn'],
						"action" : $scope.deleteServiceInstance
					}
				]
			};
			confirmService.deleteConfirm($scope);
		});
	}
  	
  	$scope.deleteServiceInstance = function(){
		servicesService.deleteServiceInstance(selectedOrder.order_id).then(function(data){
	    		$scope.goBackToServiceList();
	    		$scope.listMyServices();
	    },
	    function(errorMessage){
	    		responseService.errorResp(errorMessage);
	    });
	}
  	
  	var initialize = function(){
		$scope.listMyServices();
	}
	initialize();
}]);