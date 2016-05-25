linkerCloud.controller('HomeController', ['$scope','$location','$modal','$compile', 'productsService', 'serviceModelDesignService', 'responseService', 
						'layoutService','mppService','billingService','preloader','confirmService',
		function($scope,$location,$modal,$compile,productsService,serviceModelDesignService,responseService,layoutService,mppService,billingService,preloader,confirmService) {
    $scope.myInterval = 3000;
    $scope.recommend = {
    	title : "Linker MiFi Solution for Uber",
    	desc : "Linker MiFi Solution @ CLOUD Provide End2End Free WiFi Solution for Private Car. It Makes User Enjoy Stable and High Quality Free WiFi With Better Experience, Access to User Context Information Services, Free Access to Internet."
    };
    $scope.searchedImage = {"name":""};
	$scope.images=[];
	$scope.searchUrl = "";
	
	//preload images
	$scope.isLoading = true;
	$scope.isSuccessful = false;
	$scope.percentLoaded = 0;
	$scope.slides = [
		    {
		      image: '/portal-ui/images/products/banner1.png'
		    },
		    {
		      image: '/portal-ui/images/products/banner2.png'
		    },
		    {
		      image: '/portal-ui/images/products/banner3.png'
		    },
		    {
		      image: '/portal-ui/images/products/banner4.png'
		    },
		    {
		      image: '/portal-ui/images/products/banner5.png'
		    },
		     {
		      image: '/portal-ui/images/products/arc.png'
		    }
	  ];
	  
	// Preload the images; then, update display when returned.
	preloader.preloadImages(  _.pluck($scope.slides,"image") ).then(
		function handleResolve( imageLocations ) {
		 
			// Loading was successful.
			$scope.isLoading = false;
			$scope.isSuccessful = true;
			 
			console.info( "Preload Successful" );
		 
		},
		function handleReject( imageLocation ) {
		 
			// Loading failed on at least one image.
			$scope.isLoading = false;
			$scope.isSuccessful = false;
			 
			console.error( "Image Failed", imageLocation );
			console.info( "Preload Failure" );
		 
		},
		function handleNotify( event ) {
		 
			$scope.percentLoaded = event.percent;
			 
			console.info( "Percent loaded:", event.percent );
			 
		}
	);
	//preload end
	
	  $scope.popularServices = [
				{
					image: 'images/products/hadoop.jpg',
					name :'Hadoop'
				},
				{
					image: 'images/products/mongo.jpg',
					name :'MongoDB'
				},
				{
					image: 'images/products/mysql.png',
					name :'Mysql'
				},
				{
					image: 'images/products/spark.png',
					name :'Spark'
				}
		];
	  
	   $scope.featuredServices = [
				{
					image: 'images/products/Appserver.png',
					name :'App Server'
				},
				{
					image: 'images/products/mae.png',
					name :'MAE'
				},
				{
					image: 'images/products/gateway.png',
					name :'Gateway'
				}
		];
	  
	  $scope.allServices = [
				{
					image: 'images/products/bigdata.png',
					name :'Big Data Solution based on OpenStack',
					desc :'This is a production-ready, single-tenant configured SAP HANA database instance. Perform real-time analysis, develop and …',
					price :'20'
				},
				{
					image: 'images/products/realtime.png',
					name :'Real-time Search for Mobile Industry',
					desc :'This is a production-ready, single-tenant configured SAP HANA database instance. Perform real-time analysis, develop and …',
					price :'50'
				}
		];
		
	  $scope.listServiceGroup = function(){
		    billingService.getAllBillingsNoAuth().then(function(data){
			    	if(responseService.successResponse(data)){
			    		$scope.allBillings = data.data;
			    		getServiceGroup();
			    	}
		    },
		    function(errorMessage){
		    		responseService.errorResp(errorMessage);
		    });
	  };
	  
	  var getServiceGroup = function(){
	  		serviceModelDesignService.listServiceGroup("published").then(function(data){
				if(responseService.successResponse(data)){
					_.each(data.data,function(model){
						model.displayName = idToSimple(model.id);
						model.imageSrc = "images/products/Appserver.png";
						var relatedBilling = _.find($scope.allBillings,function(billing){
							return billing.modelid == model._id;
						});
						if(!_.isUndefined(relatedBilling)){
						    model.price = relatedBilling.totalPrice;
						    model.desc = relatedBilling.description;
						}	
					});
					$scope.serviceGroups = data.data;
				}
			},
			function(errorMessage){
				responseService.errorResp(errorMessage);
			});	
	  };
	  
	  $scope.confirmOrder = function(serviceGroup){
	        var logged = responseService.checkSession();
	        if(logged){
	             $scope.openOrderPage(serviceGroup);
	        }else{
	          $scope.$translate(['common.orderConfirmQues', 'mainPage.orderConfirmMessage', 'signIn']).then(function (translations) {
		           $scope.confirm = {
	    				"title" : translations['common.orderConfirmQues'],
	    				"message" : translations['mainPage.orderConfirmMessage'],
	    				"buttons" : [
	    					{
	    						"text" : translations['signIn'],
	    						"action" : function(){
	    							window.location="/portal-ui/login.html";
	    						}
	    					}
	    				]
	    			};
	    			confirmService.deleteConfirm($scope);
	    	});	           	
	        }	    
	  };
	  
	  $scope.openOrderPage = function(serviceGroup){
	  	var serviceToBeOrdered = mppService.prepareServiceToOrder(serviceGroup);
	  	$modal.open({
	            templateUrl: 'templates/common/orderPage.html',
	            controller: 'OrderPageCtrl',
	            size: serviceToBeOrdered.apps.length>0?'lg':'sm',
	            resolve: {
	              model: function () {
	                return {
	                  apps : serviceToBeOrdered.apps,
	                  id : serviceToBeOrdered.id
	                };
	              }
	            }
	          })
	          .result
	          .then(function (result) {
	            if (result === 'execute') {
	                 $scope.orderServiceGroup(serviceToBeOrdered);
	            }
	          });
	  }
	  
	  $scope.orderServiceGroup = function(serviceToBeOrdered){
		    mppService.runServiceModel(serviceToBeOrdered).then(function(data){
			    	if(responseService.successResponse(data)){
			    	     $scope.showSuccess(data);
			    	}		    	
		    },
		    function(errorMessage){
		    		responseService.errorResp(errorMessage);
		    });
	  };
	  
	  $scope.showSuccess = function (res) {
        $modal.open({
            templateUrl: 'templates/common/success.html',
            controller: 'ActionSuccessBoxCtrl',
            size: 'sm',
            resolve: {
              model: function () {
                return {
                  id: idToSimple(res.data.service_group_id)
                };
              }
            }
          }).result
          .then(function (result) {
            if (result === 'refresh') {
              $scope.reload();
            }
          })
    };
    $scope.listDockerHubImage = function(){
		    mppService.listDockerHubImage($scope.searchedImage.name,$scope.searchUrl).then(function(data){
		    	if(responseService.successResponse(data)){
		    		$scope.images = data.results;
		    		$scope.searchUrl = data.next;
		    	}
		    },
		    function(errorMessage){
		    	responseService.errorResp(errorMessage);
		    });
	 };

    var idToSimple = function(id) {
			return id.substring(id.lastIndexOf("/") + 1);
	  };
	  
	  //for layout
	  var getLayout = function(){
			layoutService.getLayout().then(function(data){
				if(!_.isUndefined(data.data)){
					$scope.layoutData = angular.fromJson(data.data);
				}
				if(!_.isUndefined(data.type)){
					$scope.layoutTemplateType = Number(data.type);
				}
				if(!_.isUndefined($scope.layoutData)){
					transferToLayout();
				}
				showMppPage();
		  },
		  function(errorMessage){
		    responseService.errorResp(errorMessage);
		  });
		}
	  
	  $scope.layout = {
	  	"advertise" : {
	  		"width" : "100%",
	  		"display" : true,
	  		"title" : ""
	  	},
	  	"popular" : {
	  		"width" : "100%",
	  		"display" : true,
	  		"title" : "mainPage.popularProducts"
	  	},
	  	"featured" : {
	  		"width" : "100%",
	  		"display" : true,
	  		"title" : "mainPage.featuredProducts"
	  	},
	  	"new" : {
	  		"width" : "100%",
	  		"display" : true,
	  		"title" : "mainPage.newService"
	  	},
	  	"all" : {
	  		"width" : "100%",
	  		"display" : true,
	  		"title" : "mainPage.allProducts"
	  	}
	  }
	  
	  var transferToLayout = function(){
	  	if($scope.layoutTemplateType == 2){
	  		if($scope.layoutData[1].enable && !$scope.layoutData[2].enable){
	  			$scope.layoutData[1].width = "100%";
	  		}else if(!$scope.layoutData[1].enable && $scope.layoutData[2].enable){
	  			$scope.layoutData[2].width = "100%";
	  		}
	  		if($scope.layoutData[3].enable && !$scope.layoutData[4].enable){
	  			$scope.layoutData[3].width = "100%";
	  		}else if(!$scope.layoutData[3].enable && $scope.layoutData[4].enable){
	  			$scope.layoutData[4].width = "100%";
	  		}
	  	}
	  	
	  	if($scope.layoutTemplateType == 3){
	  		if($scope.layoutData[0].enable && !$scope.layoutData[1].enable){
	  			$scope.layoutData[0].width = "100%";
	  		}else if(!$scope.layoutData[0].enable && $scope.layoutData[1].enable){
	  			$scope.layoutData[1].width = "100%";
	  		}
	  		if($scope.layoutData[3].enable && !$scope.layoutData[4].enable){
	  			$scope.layoutData[3].width = "100%";
	  		}else if(!$scope.layoutData[3].enable && $scope.layoutData[4].enable){
	  			$scope.layoutData[4].width = "100%";
	  		}
	  	}
	  	
	  	if($scope.layoutTemplateType == 4){
	  		if($scope.layoutData[0].enable && !$scope.layoutData[1].enable){
	  			$scope.layoutData[0].width = "100%";
	  		}else if(!$scope.layoutData[0].enable && $scope.layoutData[1].enable){
	  			$scope.layoutData[1].width = "100%";
	  		}
	  		if($scope.layoutData[2].enable && !$scope.layoutData[3].enable){
	  			$scope.layoutData[2].width = "100%";
	  		}else if(!$scope.layoutData[2].enable && $scope.layoutData[3].enable){
	  			$scope.layoutData[3].width = "100%";
	  		}
	  	}
	  	
	  	$scope.mpp_page = "";
	  	_.each($scope.layoutData,function(block){
	  		if(block.type == "advertise"){
	  			$scope.layout.advertise.width = block.width;
	  			$scope.layout.advertise.display = block.enable;
	  			$scope.layout.advertise.title = block.title;
	  			$scope.mpp_page += "<mppadvertise></mppadvertise>";
	  		}else if(block.type == "popular"){
	  			$scope.layout.popular.width = block.width;
	  			$scope.layout.popular.display = block.enable;
	  			$scope.layout.popular.title = block.title;
	  			$scope.mpp_page += "<popularservice></popularservice>";
	  		}else if(block.type == "featured"){
	  			$scope.layout.featured.width = block.width;
	  			$scope.layout.featured.display = block.enable;
	  			$scope.layout.featured.title = block.title;
	  			$scope.mpp_page += "<featuredservice></featuredservice>";
	  		}else if(block.type == "new"){
	  			$scope.layout.new.width = block.width;
	  			$scope.layout.new.display = block.enable;
	  			$scope.layout.new.title = block.title;
	  			$scope.mpp_page += "<newservice></newservice>";
	  		}else if(block.type == "all"){
	  			$scope.layout.all.width = block.width;
	  			$scope.layout.all.display = block.enable;
	  			$scope.layout.all.title = block.title;
	  			$scope.mpp_page += "<allservice></allservice>";
	  		}
	  	});	
	  }
	  
//	  $scope.mpp_page = "<mppadvertise></mppadvertise><popularservice></popularservice><featuredservice></featuredservice><newservice></newservice><allservice></allservice>";
	$scope.mpp_page = "<mppadvertise></mppadvertise><popularservice></popularservice>";
	  var showMppPage = function(){
	  	var el = $compile( $scope.mpp_page )( $scope );
	  	$("#home-content").append(el);
	  }
	  //for layout end
	  // $scope.watch
	  var initialize = function(){
			$scope.listServiceGroup();
//			getLayout();
			showMppPage();
	  };
	  
	  initialize();
}])
.controller('ConfirmCtrl',  ['$scope', '$modalInstance', 
    function ($scope, $modalInstance) {     
      $scope.close = function (result) {
            $modalInstance.close();        
      };
}]);