linkerCloud.directive('featuredservice', function() {
    return {
    	restrict : 'E',
    	template: 	'<div class="mpp-block" style="width:{{layout.featured.width}}" ng-if="layout.featured.display">'+
						'<div class="mpp-block-title">{{layout.featured.title | translate}}</div>'+ 
						'<div class="mpp-block-content">'+
							'<div class="service-item" ng-repeat="service in featuredServices">'+
			                    '<div class="service-img">'+
			                        '<img src="{{service.image}}" style="width:60px;height:60px;">'+
							 	'</div>'+
							 	'<div class="service-name" ng-cloak>{{service.name}}</div>'+
								'<div class="service-cart">'+
									'<span class="glyphicon glyphicon-shopping-cart"  aria-hidden="true"></span>&nbsp;{{\'mainPage.order\' | translate}}'+
								'</div>'+
							 '</div>'+ 	
						'</div>'+ 
					'</div>'
    }
});