linkerCloud.controller('BillingController', ['$scope','$location','responseService', 'billingService', function($scope,$location,responseService,billingService) {
	$scope.initDateComponent = function(param) {
			laydate({
				elem: '#' + param,
				isclear: true,
				istoday: false,
				choose: function(date) {
					$scope[param] = date;
				}
			});
			$scope.$translate(['rightContent.billing.sun', 'rightContent.billing.mon', 'rightContent.billing.tue','rightContent.billing.wed',
			'rightContent.billing.thu','rightContent.billing.fri','rightContent.billing.sat','rightContent.billing.clear','rightContent.billing.ok']).then(function (translations) {
				$("#laydate_table thead tr th:eq(0)").text(translations['rightContent.billing.sun']);
				$("#laydate_table thead tr th:eq(1)").text(translations['rightContent.billing.mon']);
				$("#laydate_table thead tr th:eq(2)").text(translations['rightContent.billing.tue']);
				$("#laydate_table thead tr th:eq(3)").text(translations['rightContent.billing.wed']);
				$("#laydate_table thead tr th:eq(4)").text(translations['rightContent.billing.thu']);
				$("#laydate_table thead tr th:eq(5)").text(translations['rightContent.billing.fri']);
				$("#laydate_table thead tr th:eq(6)").text(translations['rightContent.billing.sat']);
				$("#laydate_clear").text(translations['rightContent.billing.clear']);
				$("#laydate_ok").text(translations['rightContent.billing.ok']);
			 });
	};
		
	$scope.currentPage = 1;
	$scope.totalPage = 1;
	$scope.recordPerPage = 10;
	$scope.totalrecords = 0;
	
	$scope.getBillRecords = function(search){
		var query = {};
		if($scope.filter_date_from !="" || $scope.filter_date_to !="" || $scope.filter_transaction_type != ""){
			
			var from = $scope.filter_date_from !="" ? new Date($scope.filter_date_from).getTime()/1000 : "";
			var to = $scope.filter_date_to !="" ? new Date($scope.filter_date_to).getTime()/1000 : "";
			var filterdate = {};
			if(from != ""){
				filterdate.$gt = from;
				query.date = filterdate;
			}
			if(from != ""){
				filterdate.$lt = to;
				query.date = filterdate;
			}	
			if($scope.filter_transaction_type != ""){
				query.transaction_type = $scope.filter_transaction_type;
			}
		}
		query = JSON.stringify(query);
		if(search){
			$scope.currentPage = 1;
		}
		var skip = ($scope.currentPage - 1) * $scope.recordPerPage;
		var limit = $scope.recordPerPage;
		
		billingService.getBillRecords(query,skip,limit).then(function(data){
		    	if(responseService.successResponse(data)){
		    		$scope.totalrecords = data.count;
		    		$scope.totalPage = Math.ceil($scope.totalrecords/$scope.recordPerPage);
		    		_.each(data.data,function(record){
		    			record.servicegroup = idToSimple(record.sg_id);
		    			record.price = record.price.toFixed(2);
		    		})
		    		$scope.reports = data.data;
		    	}
	    },
	    function(errorMessage){
	    		responseService.errorResp(errorMessage);
	    });
	};
	
	var initialize = function() {
	    $scope.filter_date_from = "";
	    $scope.filter_date_to = "";
	    $scope.filter_transaction_type = "";
	    $scope.$watch('currentPage', function() {
			$scope.getBillRecords(false);
		});
	};
	initialize();
}]);