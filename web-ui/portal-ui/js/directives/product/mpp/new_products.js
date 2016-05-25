linkerCloud.directive('newservice', function() {
    return {
    	restrict : 'E',
    	template: 	'<div class="mpp-block" style="width:{{layout.new.width}}" ng-if="layout.new.display && serviceGroups.length>0">'+
						'<div class="mpp-block-title">{{layout.new.title | translate}}</div>'+
						'<div class="mpp-block-content">'+
							'<div ng-repeat="serviceGroup in serviceGroups" class="service-item" ng-click="openOrderPage(serviceGroup);">'+
			                    '<div class="service-img">'+
			                        '<img src="{{serviceGroup.imageSrc}}" style="width:60px;height:60px;">'+
							 	'</div>'+
							 	'<div class="service-name" ng-cloak>{{serviceGroup.displayName}}</div>'+
								'<div class="service-cart">'+
									'<span class="glyphicon glyphicon-shopping-cart"  aria-hidden="true"></span>&nbsp;{{\'mainPage.order\' | translate}}'+
								'</div>'+
							 '</div>'+ 	
						'</div>'+ 
					'</div>'
    }
});