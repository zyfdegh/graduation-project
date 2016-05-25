linkerCloud.directive('mppadvertise', function() {
    return {
    	restrict : 'E',
    	template: 	'<div style="width:{{layout.advertise.width}}" ng-if="layout.advertise.display">'+
					    '<div id="slides_control" class="mpp-advertise" ng-switch="isLoading">'+
					    		'<div style="width:60%;margin:auto;margin-top:50px;text-align:center" ng-switch-when="true">Images loading... {{ percentLoaded }}%</div>'+
					    		'<div ng-switch-when="false" ng-switch="isSuccessful">'+
					    			'<div ng-switch-when="true">'+
								'<carousel interval="myInterval">'+
								    	'<slide ng-repeat="slide in slides" active="slide.active">'+
								        	'<img ng-src="{{slide.image}}">'+
								        	'<div class="carousel-caption"></div>'+
								    '</slide>'+
								'</carousel>'+
								'</div>'+
								'<div style="width:60%;margin:auto;margin-top:50px;text-align:center" ng-switch-when="false">'+		 
									'Images failed to load'+
								'</div>'+	 
							'</div>'+
					    '</div>'+
					    '<div class="mpp-recommend">'+
			                  '<div class="mpp-recommend-service-title">{{recommend.title}}</div><br/>'+
			                  '<button type="button" class="btn btn-success btn-xs" ng-repeat="t in [1,2,3,4]" style="margin-left: 5px;">'+
									'<span class="glyphicon glyphicon-star" aria-hidden="true"></span>'+
							  '</button><br/><br/>{{recommend.desc}}<br/><br/>'+
					    '</div>'+
					'</div>'
    }
});