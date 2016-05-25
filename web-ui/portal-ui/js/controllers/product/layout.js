linkerCloud.controller('LayoutController', ['$scope','$location','$modal','productsService','layoutService', 'responseService', 
	function($scope,$location,$modal,productsService,layoutService,responseService) {	
	$scope.selectedTemplateType = 1;
	$scope.layoutData = [
			{
				type : "advertise",
				title : 'rightContent.layoutManagement.advertiseArea',
				width : "100%",
				edit : false,
				enable : true
			},
			{
				type : "popular",
				title : 'mainPage.popularProducts',
				width : "100%",
				edit : false,
				enable : true
			},
			{
				type : "featured",
				title : 'mainPage.featuredProducts',
				width : "100%",
				edit : false,
				enable : true
			},
			{
				type : "new",
				title : 'mainPage.newService',
				width : "100%",
				edit : false,
				enable : true
			},
			{
				type : "all",
				title : 'mainPage.allProducts',
				width : "100%",
				edit : false,
				enable : true
			}
		];
	
	$scope.selectTemplate = function(event,type){
		$(".layout-template-active").addClass("layout-template");
		$(".layout-template-active").removeClass("layout-template-active");
		$(event.currentTarget).addClass("layout-template-active");
		$(event.currentTarget).removeClass("layout-template");
		
		$scope.selectedTemplateType = type;
		changeOrder(type);
		changeWidth(type);
	};
	
	var changeOrder = function(type){
		var index = _.pluck($scope.layoutData, 'type').indexOf("advertise");
				
		if(type == 2 && $scope.layoutData[0].type != "advertise"){
			var target = $scope.layoutData[index];
			$scope.layoutData.splice(index,1);
			$scope.layoutData = [].concat(target).concat($scope.layoutData);
		}
		if(type == 3 && $scope.layoutData[2].type != "advertise"){
			var target = $scope.layoutData[index];
			$scope.layoutData.splice(index,1);
			var item0 = $scope.layoutData[0];
			var item1 = $scope.layoutData[1];
			var item2 = $scope.layoutData[2];
			var item3 = $scope.layoutData[3];
			$scope.layoutData = [].concat(item0).concat(item1).concat(target).concat(item2).concat(item3);
		}
		if(type == 4 && $scope.layoutData[4].type != "advertise"){
			var target = $scope.layoutData[index];
			$scope.layoutData.splice(index,1);
			var item0 = $scope.layoutData[0];
			var item1 = $scope.layoutData[1];
			var item2 = $scope.layoutData[2];
			var item3 = $scope.layoutData[3];
			$scope.layoutData = [].concat(item0).concat(item1).concat(item2).concat(item3).concat(target);
		}
	}
	
	var changeWidth = function(type){
		_.each($scope.layoutData,function(item,index){
			item.width = templatewidths[$scope.selectedTemplateType-1][index];
		});
	}
	
	var templatebuttons = [
		[[1,0,0,0],[1,1,0,0],[1,1,0,0],[1,1,0,0],[0,1,0,0]],
		[[0,0,0,0],[1,0,0,1],[1,0,1,0],[0,1,0,1],[0,1,1,0]],
		[[1,0,0,1],[1,0,1,0],[0,0,0,0],[0,1,0,1],[0,1,1,0]],
		[[1,0,0,1],[1,0,1,0],[0,1,0,1],[0,1,1,0],[0,0,0,0]]
	];

	var templatewidths = [
		["100%","100%","100%","100%","100%"],
		["100%","50%","50%","50%","50%"],
		["50%","50%","100%","50%","50%"],
		["50%","50%","50%","50%","100%"]
	];
	
	$scope.isShowButton = function(layoutindex,buttontype){
		return templatebuttons[$scope.selectedTemplateType-1][layoutindex][buttontype-1];
	}
	
	var exchangePos = function(index1,index2){
		var target1 = $scope.layoutData[index1];
		var target2 = $scope.layoutData[index2];
		$scope.layoutData.splice(index1,1,target2);
		$scope.layoutData.splice(index2,1,target1);
	}
	
	$scope.downPos = function(layoutindex){
		switch($scope.selectedTemplateType){
			case 1 : exchangePos(layoutindex,layoutindex+1); break;
			case 2 : exchangePos(layoutindex,layoutindex+2); break;
			case 3 : exchangePos(layoutindex,layoutindex+3); break;
			case 4 : exchangePos(layoutindex,layoutindex+2); break;
		}
	}
	
	$scope.upPos = function(layoutindex){
		switch($scope.selectedTemplateType){
			case 1 : exchangePos(layoutindex-1,layoutindex); break;
			case 2 : exchangePos(layoutindex-2,layoutindex); break;
			case 3 : exchangePos(layoutindex-3,layoutindex); break;
			case 4 : exchangePos(layoutindex-2,layoutindex); break;
		}
	}
	
	$scope.leftPos = function(layoutindex){
		exchangePos(layoutindex-1,layoutindex);
	}
	
	$scope.rightPos = function(layoutindex){
		exchangePos(layoutindex,layoutindex+1);
	}
	
	$scope.saveLayout = function(){
		layoutService.saveLayout($scope.layoutData,$scope.selectedTemplateType);
		$modal.open({
            templateUrl: 'templates/common/saveSuccess.html',
            controller: 'SaveSuccessBoxCtrl',
            size: 'sm',
            resolve: {
              model: function () {
                return {
                  message: 'rightContent.layoutManagement.saveLayoutSuccess'
                };
              }
            }
          })
	}
	
	var getLayout = function(){
		layoutService.getLayout().then(function(data){
			if(!_.isUndefined(data.data)){
				$scope.layoutData = angular.fromJson(data.data);
			}
			if(!_.isUndefined(data.type)){
				$scope.selectedTemplateType = Number(data.type);
			}
	    },
	    function(errorMessage){
	    	responseService.errorResponse("Failed to get layout.");
	    });
	}
	
	getLayout();
}]);