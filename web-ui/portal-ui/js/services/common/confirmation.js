function confirmService($window,$modal){
	var deleteConfirm = function($scope){
		$modal.open({
		    templateUrl: 'templates/common/confirm.html',
		    controller: 'DeleteConfirmCtrl',
		    size: 'sm',
		    resolve: {
		        model: function () {
		            return $scope.confirm;
		        }
		    }
	   });
	};
	
	return {
		"deleteConfirm" : deleteConfirm
	}
}
   
linkerCloud.factory('confirmService', ['$window', '$modal',confirmService]);