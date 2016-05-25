linkerCloud.directive('dragapptoservice', function() {
    return {
    	restrict : 'E',
    	template: '<div ng-repeat="app in availableApps" class="a-app-item" style="cursor:move" data-appid="{{app.id}}" draggable="true">'+
						'<div class="a-app-image-container"><img src="{{app.imageSrc}}" class="container-image" draggable="false"/></div>'+
						'<div class="a-app-item-id"><div class="a-app-item-id-label" ng-cloak>{{app.id}}</div></div>'+
					'</div>',
   		link : function(scope, element) {
   			var el = element[0];
            el.addEventListener(
                'dragstart',
                function(e) {
                    scope.dragApp(e);
                },
                false
            );
        }
    }
});

linkerCloud.directive('dragtemplatetoservice', function() {
    return {
    	restrict : 'E',
    	template	: '<div ng-repeat="template in serviceGroups">' +
    			'<div class="a-template-item" style="cursor:move" data-groupid="{{template.id}}" draggable="true">'+
					'<div class="service-group-image-container"><img src="{{template.imageSrc}}" class="container-image" draggable="false"/></div>'+
					'<div class="a-app-item-id"><div class="a-app-item-id-label">{{template.displayName}}</div></div>'+
			'</div>'+
			'<div style="text-align:center;margin-bottom:10px">$ {{template.price}}</div>'+
			'</div>',
   		link : function(scope, element) {
   			var el = element[0];
            el.addEventListener(
                'dragstart',
                function(e) {
                    scope.dragTemplate(e);
                },
                false
            );
        }
    }
});