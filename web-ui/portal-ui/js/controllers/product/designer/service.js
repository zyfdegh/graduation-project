linkerCloud.controller('ServiceDesignerController', ['$scope','$location','$modal','productsService','serviceModelDesignService', 'billingService',
	'appModelDesignService','responseService','confirmService',
	function($scope,$location,$modal,productsService,serviceModelDesignService,billingService,appModelDesignService,responseService,confirmService) {
	$scope.showDetail = false;
	$scope.openAppModels = true;
	
	$scope.listServiceGroup = function(){
		availableTemplates = [];
	    serviceModelDesignService.listServiceGroup().then(function(data){
		    	if(responseService.successResponse(data)){
		    		_.each(data.data,function(model){
		    			model.displayName = idToSimple(model.id);
		    			model.imageSrc = "images/products/Appserver.png";
					switch(model.state){
						case "published" : model.stateicon=""; model.statetitle=''; break;
						case "unpublished" : model.stateicon="glyphicon-lock"; model.statetitle='rightContent.serviceDesign.unpublished'; break;
						case "verifying" : model.stateicon="glyphicon-hourglass"; model.statetitle='rightContent.serviceDesign.verifying'; break;
					}
					
					var billingmodel = _.find(allBillings,function(billing){
						return billing.modelid == model._id;
					});
					if(model.state == "published" && !_.isUndefined(billingmodel)){
						model.price = billingmodel.totalPrice;
					}else{
						model.price = 0;
					}
			
		    			var modelcopy = $.extend(true,{},model);
		    			availableTemplates.push(modelcopy);

		    		})
		    		$scope.serviceGroups = data.data;
		    	}
	    },
	    function(errorMessage){
	    		responseService.errorResp(errorMessage);
	    });
  	};
  	
  	$scope.getAllBillings = function(){
  		allBillings = [];
	    billingService.getAllBillings().then(function(data){
		    	if(responseService.successResponse(data)){
		    		$scope.listServiceGroup();
		    		allBillings = data.data;
		    	}
	    },
	    function(errorMessage){
	    		responseService.errorResp(errorMessage);
	    });
  	}
  	
  	$scope.showAvailableApps = function(){
	    appModelDesignService.getApps().then(function(data){
		    	if(responseService.successResponse(data)){
		    		_.each(data.data,function(app){
		    			if(app.id == "nginx"){
		    				app.imageSrc = "images/products/designer/app/nginx.png";
		    			}else if(app.id == "zookeeper"){
		    				app.imageSrc = "images/products/designer/app/zookeeper.png";
		    			}else if(app.id == "haproxy"){
		    				app.imageSrc = "images/products/designer/app/haproxy.png";
		    			}else if(app.id.indexOf("mysql")>=0){
		    				app.imageSrc = "images/products/designer/app/mysql.png";
		    			}else{
		    				app.imageSrc = "images/products/designer/app/default.png";
		    			}
		    		})
		    		$scope.availableApps = data.data;
		    		apps = data.data;
		    	}
	    },
	    function(errorMessage){
	    		responseService.errorResp(errorMessage);
	    });
  	};
  	
  	$scope.showServiceGroupDetail = function(model){
  		 serviceModelDesignService.getSGOperations(model._id).then(function(data){
  		 	if(responseService.successResponse(data)){
  		 		allow_update_sg = data.data.update_sg == 1 ? true : false;
  		 		$scope.allow_delete_sg = data.data.delete_sg == 1 ? true : false;
  		 		$scope.allow_publish_sg = data.data.publish_sg == 1 ? true : false;
  		 		$scope.allow_submit_sg = data.data.submit_sg == 1 ? true : false;
  		 		$scope.allow_update_sg = allow_update_sg;
  		 		
			 	$scope.showDetail = true;
			 	$scope.isUpdate = true;
//				$scope.detailTitle = "Service Details";
				$scope.state = model.state;
				_.each(model.groups,function(group){
					allocateImageToApp(group);		
				});
				selectedModel = model;
				relatedBilling = _.find(allBillings,function(billing){
					return billing.modelid == model._id;
				});
				
				setTimeout(function(){
					drawGroupRelations();
					$scope.showAvailableApps();
				},200);
  		 	}
	    },
	    function(errorMessage){
	    		responseService.errorResp(errorMessage);
	    });
	}
	
	$scope.newServiceGroupDetail = function(){
		var model = {
			"id" : "",
			"displayName" : "",
			"groups" : [],
			"imageSrc" : "images/products/Appserver.png"
		}
//		var displayName = sessionStorage.namespace +"-newservice";
//		var model = {
//			"id" : "/"+ displayName,
//			"displayName" : displayName,
//			"groups" : [],
//			"imageSrc" : "images/products/Appserver.png"
//		};
		$scope.selectedModel = model;
		relatedBilling = {
			"modelid":"",
			"refs":[],
			"price":0,
			"description" : ""
		};
		
		$scope.showDetail = true;
		$scope.isUpdate = false;
		$scope.state = "";
//		$scope.detailTitle = "New Service Design";
//		allow_update_sg = true;
//		$scope.allow_update_sg = allow_update_sg;
  		 		
//		setTimeout(function(){
//			drawGroupRelations();
//			$scope.showAvailableApps();
//		},200);
	}
	
	$scope.dragApp = function(event) {
		var appid = $(event.target).data("appid");
	    event.dataTransfer.setData("dragAppID", appid);
	    event.dataTransfer.setData("dragType", "app");
	}
	
	$scope.dragTemplate = function(event) {
		var groupid = $(event.target).data("groupid");
		var objectid = $(event.target).data("objectid");
	    event.dataTransfer.setData("dragTemplateID", groupid);
	    event.dataTransfer.setData("dragType", "template");
	}
	
	$scope.goBackToServiceList = function(){
		$scope.showDetail = false;
		$("body").scrollTop(0);
		delete st;
		delete canvasX;
		delete canvasY;
	}
	
	//api invocation
	var clearServiceModel = function(servicemodel){
		delete servicemodel.displayName;
		delete servicemodel.imageSrc;
		_.each(servicemodel.groups,function(group){
			_.each(group.apps,function(app){
	    		delete app.imageSrc;
	    	});
	    	if(!_.isUndefined(group.groups)){
	    		clearServiceModel(group);
	    	}
		});
	}
	
	var generateBillingModel = function(servicemodel){
		_.each(servicemodel.groups,function(group){
			if(!_.isUndefined(group.groups) && group.groups != null){
				var billing = group.billing;
				var state = group.state;
				if(billing && state == "published"){
					var billingmodel = _.find(allBillings,function(b){
						return b.modelid == group._id;
					});
					relatedBilling.refs.push(billingmodel._id);
				}
				generateBillingModel(group);
			}
		})
	}
	
	$scope.saveServiceModel = function(){
		clearServiceModel(selectedModel);
		relatedBilling.refs = [];
		generateBillingModel(selectedModel);
		
		serviceModelDesignService.updateServiceModel(selectedModel).then(function(data){
			figureOutTotalPrice();
			$scope.saveBillingModel(true);
		},
		function(errorMessage){
		    responseService.errorResp(errorMessage);
		});
	}
	
	$scope.saveNewService = function(){
		if(!newServiceformIsValid()){
			return false;
		}
		
		$scope.selectedModel.id = "/" + sessionStorage.namespace + "-" + $scope.selectedModel.displayName;
		delete $scope.selectedModel.displayName;
		delete $scope.selectedModel.imageSrc;
		serviceModelDesignService.newServiceModel($scope.selectedModel).then(function(data){
			var modelid = data.data.url.substring(data.data.url.lastIndexOf("/")+1);
			relatedBilling.modelid = modelid;
			figureOutTotalPrice();
			$scope.saveBillingModel(false);
		},
		function(errorMessage){
		    responseService.errorResp(errorMessage);
		});
	}
	
	$scope.saveBillingModel = function(update){
		if(!update){
			billingService.newBilling(relatedBilling).then(function(data){
				$scope.goBackToServiceList();
	    			$scope.getAllBillings();
				console.log("Service model and billing model saved.")
		    },
		    function(errorMessage){
		    		responseService.errorResp(errorMessage);
		    });
		}else{
			billingService.updateBilling(relatedBilling).then(function(data){
				$scope.goBackToServiceList();
	    			$scope.getAllBillings();
				console.log("Service model and billing model saved.")
		    },
		    function(errorMessage){
		    		responseService.errorResp(errorMessage);
		    });
		}
	}
	
	$scope.confirmDeleteServiceModel = function(){
		$scope.$translate(['rightContent.serviceDesign.deleteConfirm', 'rightContent.serviceDesign.deleteMessage', 'rightContent.serviceDesign.deleteBtn']).then(function (translations) {
		    $scope.confirm = {
				"title" : translations['rightContent.serviceDesign.deleteConfirm'],
				"message" : translations['rightContent.serviceDesign.deleteMessage'],
				"buttons" : [
					{
						"text" : translations['rightContent.serviceDesign.deleteBtn'],
						"action" : $scope.deleteServiceModel
					}
				]
			};
			confirmService.deleteConfirm($scope);
		  });
	}
    
	$scope.deleteServiceModel = function(){
		deleteAllCP(selectedModel,selectedModel.id);
		serviceModelDesignService.deleteServiceModel(selectedModel).then(function(data){
	    		$scope.goBackToServiceList();
	    		$scope.getAllBillings();
	    },
	    function(errorMessage){
	    		responseService.errorResp(errorMessage);
	    });
	}
	
	var deleteAllCP = function(sgroup,sgroupid){
		_.each(sgroup.groups,function(group){
			var groupid = sgroupid + "/" + group.id;
			_.each(group.apps,function(app){
	    		var appid = groupid + "/" + app.id;
	    		deleteOneCP(appid);
	    	});
	    	if(!_.isUndefined(group.groups)){
	    		deleteAllCP(group,groupid);
	    	}
		});
	}
	
	var deleteOneCP = function(appid){
		serviceModelDesignService.deleteCP(appid);
	}
	
	$scope.deleteServiceModel = function(){
		deleteAllCP(selectedModel,selectedModel.id);
		serviceModelDesignService.deleteServiceModel(selectedModel).then(function(data){
	    		$scope.goBackToServiceList();
	    		$scope.getAllBillings();
	    },
	    function(errorMessage){
	    		responseService.errorResp(errorMessage);
	    });
	}
	
	$scope.confirmPublishServiceModel = function(flag){
		if(flag == 0){
			$scope.$translate(['rightContent.serviceDesign.publishConfirm', 'rightContent.serviceDesign.publishMessage', 'rightContent.serviceDesign.publishBtn']).then(function (translations) {
			    $scope.confirm = {
					"title" : translations['rightContent.serviceDesign.publishConfirm'],
					"message" : translations['rightContent.serviceDesign.publishMessage'],
					"buttons" : [
						{
							"text" : translations['rightContent.serviceDesign.publishBtn'],
							"action" : $scope.publishServiceModel
						}
					]
				};
				confirmService.deleteConfirm($scope);
			  });
		}else{
			$scope.$translate(['rightContent.serviceDesign.unPublishConfirm', 'rightContent.serviceDesign.unPublishMessage', 'rightContent.serviceDesign.unPublishBtn']).then(function (translations) {
			    $scope.confirm = {
					"title" : translations['rightContent.serviceDesign.unPublishConfirm'],
					"message" : translations['rightContent.serviceDesign.unPublishMessage'],
					"buttons" : [
						{
							"text" : translations['rightContent.serviceDesign.unPublishBtn'],
							"action" : $scope.unpublishServiceModel
						}
					]
				};
				confirmService.deleteConfirm($scope);
			  });
		}
	}
	
	$scope.confirmSubmitServiceModel = function(){
		$scope.$translate(['rightContent.serviceDesign.submitConfirm', 'rightContent.serviceDesign.submitMessage', 'rightContent.serviceDesign.submitBtn']).then(function (translations) {
			$scope.confirm = {
				"title" : translations['rightContent.serviceDesign.submitConfirm'],
				"message" : translations['rightContent.serviceDesign.submitMessage'],
				"buttons" : [
					{
						"text" : translations['rightContent.serviceDesign.submitBtn'],
						"action" : $scope.setPriceAndDesc
					}
				]
			};
			confirmService.deleteConfirm($scope);
		});
	}
	
	$scope.setPriceAndDesc = function(){
		$modal.open({
	            templateUrl: 'templates/common/billingPrice.html',
	            controller: 'billingPriceCtrl',
	            size: 'lg',
	            resolve: {
	              model: function () {
	                return {
	                  price : relatedBilling.price,
	                  desc : relatedBilling.description
	                };
	              }
	            }
	          })
	          .result
	          .then(function (result) {
	            if (result === 'execute') {
	            		if(formIsValid()){
	            			relatedBilling.price = Number($(priceform.serviceprice).val());
	            			relatedBilling.description = $(priceform.servicedesc).val();
	            			$scope.submitServiceModel();
	            		} 
	            }
	          });
	}
	
	$scope.submitServiceModel = function(){
		serviceModelDesignService.submitServiceModel(selectedModel).then(function(data){
			figureOutTotalPrice();
			$scope.saveBillingModel(true);
	    },
	    function(errorMessage){
	    		responseService.errorResp(errorMessage);
	    });
	}
	
	var formIsValid = function(){
		return (priceform.serviceprice.validity.valid || priceform.serviceprice.validity.stepMismatch) && priceform.servicedesc.validity.valid;	
	}
	
	$scope.publishServiceModel = function(){
		serviceModelDesignService.publishServiceModel(selectedModel).then(function(data){
			$scope.goBackToServiceList();
	    		$scope.getAllBillings();
	    },
	    function(errorMessage){
	    		responseService.errorResp(errorMessage);
	    });
	}
	
	var figureOutTotalPrice = function(){
		relatedBilling.totalPrice = relatedBilling.price;
		_.each(relatedBilling.refs,function(ref){
			var refBilling = _.find(allBillings,function(billing){
				return billing._id == ref;
			});
			relatedBilling.totalPrice += refBilling.totalPrice;
		});
	}
	
	$scope.unpublishServiceModel = function(){
		serviceModelDesignService.unpublishServiceModel(selectedModel).then(function(data){
	    		$scope.goBackToServiceList();
	    		$scope.getAllBillings();
	    },
	    function(errorMessage){
	    		responseService.errorResp(errorMessage);
	    });
	}
	
	var newServiceformIsValid = function(){
		var basic = serviceform.serviceid.validity.valid && $("#invalidID").css("display") == "none" ;
		return basic;	
	}
	
	var initialize = function(){
		$scope.getAllBillings();
	}
	initialize();
}]);