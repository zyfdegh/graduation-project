linkerCloud.controller('ConfirmCtrl', ['$scope', '$modalInstance','model',
    function ($scope, $modalInstance,model) {
      $scope.id = model.id;
      $scope.close = function (res) {
        $modalInstance.close(res);
      }            
    }
])
.controller('ActionSuccessBoxCtrl', ['$scope', '$modalInstance', 'model',
    function ($scope, $modalInstance, model) {
      $scope.id = model.id;
      $scope.close = function (res) {
        $modalInstance.close(res);
      };
    }
])
.controller('DeleteConfirmCtrl', ['$scope', '$modalInstance', 'model',
    function ($scope, $modalInstance, model) {
      $scope.title = model.title;
      $scope.message = model.message;
      $scope.buttons = model.buttons;
      $scope.close = function (res) {
        $modalInstance.close(res);
      };
      $scope.doaction = function (res,action) {
      	action();
        $modalInstance.close(res);
      };
    }
])
.controller('OrderPageCtrl', ['$scope', '$modalInstance', 'model',
    function ($scope, $modalInstance, model) {
      $scope.id = model.id;
      $scope.apps = model.apps;
      $scope.close = function (res) {
        $modalInstance.close(res);
      };
    }
])
.controller('SaveSuccessBoxCtrl', ['$scope', '$modalInstance', 'model',
    function ($scope, $modalInstance, model) {
      $scope.message = model.message;
      $scope.close = function (res) {
        $modalInstance.close(res);
      };
    }
])
.controller('billingPriceCtrl', ['$scope', '$modalInstance', 'model',
    function ($scope, $modalInstance, model) {
		$scope.price = model.price;
	    $scope.desc = model.desc;
	    $scope.close = function (res) {
	        $modalInstance.close(res);
	    };
    }
])