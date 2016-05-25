linkerCloud.controller('DashboardController', ['$scope' ,'$location','responseService','dashboardService', function($scope,$location,responseService,dashboardService) {  	
    $scope.cpuChartData = [];
    $scope.memoryChartData = [];
    $scope.diskChartData = [];
    $scope.getServiceInstances = function(){
        dashboardService.getServiceInstances().then(function(response){
            if(responseService.successResponse(response)){          
                var tempArray = [];
                var tempData = _.pairs(response.data.status_num || {});
                _.each(tempData,function(data){
                    var tempObj = {"key":"","y":""};
                    tempObj.key = data[0];
                    tempObj.y = data[1];
                    tempArray.push(tempObj); 
                });
                 $scope.pieChartData = tempArray;
              }
        }, function(error){
            responseService.errorResp(error);  
        })
    };
    $scope.getResources = function(){
         dashboardService.getResources().then(function(response){
            if(responseService.successResponse(response)){          
                var tempProviders = response.data.providers || {};
                var providerTypes = _.keys(tempProviders);

                var cpuArray = [];
                var memoryArray = [];
                var diskArray = [];

                    var cpuObj = {};
                    var memoryObj = {};
                    var diskObj = {};

                    var totalInfo = tempProviders["AliCloud"].total || {};                                       
                    var cpuValueAli = totalInfo.cpus * 0.8;
                    var memoryValueAli = totalInfo.mems * 0.8;
                    var diskValueAli = totalInfo.disks * 0.8;
                    var cpuValueIDC = totalInfo.cpus * 0.2;
                    var memoryValueIDC = totalInfo.mems * 0.2;
                    var diskValueIDC = totalInfo.disks * 0.2;

                    var valueObj = {"label":"AliCloud","value" : cpuValueAli};
                    cpuArray.push(valueObj);
                    var valueObj = {"label":"Linker","value" : cpuValueIDC};
                    cpuArray.push(valueObj);
                    cpuObj = {"key" : "CPU", "values":cpuArray};
                    $scope.cpuChartData.push(cpuObj);
                    var valueObj = {"label":"AliCloud","value" : memoryValueAli};
                    memoryArray.push(valueObj);
                    var valueObj = {"label":"Linker","value" : memoryValueIDC};
                    memoryArray.push(valueObj);
                    memoryObj = {"key" : "Memory", "values":memoryArray};
                    $scope.memoryChartData.push(memoryObj);
                    var valueObj = {"label":"AliCloud","value" : diskValueAli};
                    diskArray.push(valueObj);
                    var valueObj = {"label":"Linker","value" : diskValueIDC};
                    diskArray.push(valueObj);
                    diskObj = {"key" : "Disk", "values":diskArray};
                    $scope.diskChartData.push(diskObj);


              }
        }, function(error){
            responseService.errorResp(error);  
        })
    };
	$scope.pieChartOptions = {
	     chart: {
                type: 'pieChart',
                width: $(window).width() * .4 * 0.75,
                height: $(window).height() * 0.3,
                x: function(d){return d.key;},
                y: function(d){return d.y;},
                showLabels: true,
                growOnHover:false,
                labelSunbeamLayout:true,
                transitionDuration: 500,
                labelThreshold: 0.01,
                valueFormat: d3.format(',f'),
                legend: {
                    margin: {
                        top: 5,
                        right: 35,
                        bottom: 5,
                        left: 0
                    }
                }
                
            }
	};
    $scope.cpuChartOptions = {
        chart: {
            type: 'multiBarHorizontalChart',
            width: $(window).width() * .5 * 0.75,
            height: $(window).height() * 0.3,
            x: function(d){return d.label;},
            y: function(d){return d.value;},       
            showControls: true,
            showValues: true,
            transitionDuration: 500,
            valueFormat: d3.format(',.2f'),
            xAxis: {
                showMaxMin: false
            },
            yAxis: {
                axisLabel: 'Values',
                tickFormat: function(d){
                    return d3.format(',.2f')(d);
                }
            }
        }
    };
    $scope.memeoryChartOptions = {
        chart: {
            type: 'multiBarHorizontalChart',
            width: $(window).width() * .5 * 0.75,
            height: $(window).height() * 0.3,
            x: function(d){return d.label;},
            y: function(d){return d.value;},
            //yErr: function(d){ return [-Math.abs(d.value * Math.random() * 0.3), Math.abs(d.value * Math.random() * 0.3)] },
            showControls: true,
            showValues: true,
            transitionDuration: 500,
            valueFormat: d3.format(',f'),
            xAxis: {
                showMaxMin: false
            },
            yAxis: {
                axisLabel: 'Values',
                tickFormat: function(d){
                    return d3.format(',f')(d);
                }
            }
        }
    };
     $scope.diskChartOptions = {
        chart: {
            type: 'multiBarHorizontalChart',
            width: $(window).width() * .5 * 0.75,
            height: $(window).height() * 0.3,
            x: function(d){return d.label;},
            y: function(d){return d.value;},
            //yErr: function(d){ return [-Math.abs(d.value * Math.random() * 0.3), Math.abs(d.value * Math.random() * 0.3)] },
            showControls: true,
            showValues: true,
            transitionDuration: 500,
            valueFormat: d3.format(',f'),
            xAxis: {
                showMaxMin: false
            },
            yAxis: {
                axisLabel: 'Values',
                tickFormat: function(d){
                    return d3.format(',f')(d);
                }
            }
        }
    };
    $scope.getServiceInstances();
	$scope.getResources();

     // [
 //            {
 //                key: "Helion",
 //                y: 5
 //            },
 //            {
 //                key: "Ali",
 //                y: 2
 //            },
 //            {
 //                key: "Linker",
 //                y: 9
 //            },
 //            {
 //                key: "AWS",
 //                y: 7
 //            }
 //        ];

     // $scope.cpuChartData = [
     //        {
     //            "key": "Used",
     //            "values": [
     //                {
     //                    "label" : "Helion" ,
     //                    "value" : 10
     //                } ,
     //                {
     //                    "label" : "Ali" ,
     //                    "value" : 3
     //                } ,
     //                {
     //                    "label" : "Linker" ,
     //                    "value" : 23
     //                } ,
     //                {
     //                    "label" : "AWS" ,
     //                    "value" : 4
     //                } 
     //            ]
     //        },
     //        {
     //            "key": "Available",
               
     //            "values": [
     //                {
     //                    "label" : "Helion" ,
     //                    "value" : 30
     //                } ,
     //                {
     //                    "label" : "Ali" ,
     //                    "value" : 33
     //                } ,
     //                {
     //                    "label" : "Linker" ,
     //                    "value" : 53
     //                } ,
     //                {
     //                    "label" : "AWS" ,
     //                    "value" : 23
     //                } 
     //            ]
     //        }
     //  ];
    
   



	
 
}]);