linkerCloud.controller('ProductsController', ['$scope','$location','productsService',function($scope,$location,productsService) {
	$scope.$on('$stateChangeSuccess', function(event, toState, toParams, fromState, fromParams){ 
         productsService.selectNav($location.path()); 
    });	
	$scope.logged = (_.isUndefined(sessionStorage.username) || _.isEmpty(sessionStorage.username)) ? false : true;
	$scope.navigators = productsService.getNavs();
	$scope.defaultState = $scope.logged ? productsService.getDefaultState() : "products.mpp";
	if($scope.$state.current.name == "products"){
		$scope.$state.go($scope.defaultState);
	}

}]);