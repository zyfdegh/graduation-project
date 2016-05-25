linkerCloud.directive('popularservice', function() {
    return {
    		restrict : 'E',
//  	template: 	'<div class="mpp-block" style="width:{{layout.popular.width}}" ng-if="layout.popular.display">'+
//						'<div class="mpp-block-title">{{layout.popular.title | translate}}</div>'+ 
//						'<div class="mpp-block-content">'+
//							'<div class="service-item" ng-repeat="service in popularServices">'+
//			                    '<div class="service-img">'+
//			                        '<img src="{{service.image}}" style="width:60px;height:60px;">'+
//							 	'</div>'+
//							 	'<div class="service-name" ng-cloak>{{service.name}}</div>'+
//								'<div class="service-cart">'+
//									'<span class="glyphicon glyphicon-shopping-cart"  aria-hidden="true"></span>&nbsp;{{\'mainPage.order\' | translate}}'+
//								'</div>'+
//							 '</div>'+ 	
//						'</div>'+ 
//					'</div>'
	 	templateUrl: "templates/product/mpp/popular_products.html"
    }
});