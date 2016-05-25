linkerCloud.directive('layout1', function() {
    return {
    	restrict : 'E',
    	template: 	'<div class="{{selectedTemplateType == 1?\'layout-template-active\':\'layout-template\'}}" ng-click="selectTemplate($event,1)">'+
			      		'<div style="width:100%;height:20%;border-bottom: 1px solid #d5d9dc;"></div>'+
			      		'<div style="width:100%;height:20%;border-bottom: 1px solid #d5d9dc;"></div>'+
			      		'<div style="width:100%;height:20%;border-bottom: 1px solid #d5d9dc;"></div>'+
			      		'<div style="width:100%;height:20%;border-bottom: 1px solid #d5d9dc;"></div>'+
			      		'<div style="width:100%;height:20%;"></div>'+
			      	'</div>'
    }
});

linkerCloud.directive('layout2', function() {
    return {
    	restrict : 'E',
    	template: 	'<div class="{{selectedTemplateType == 2?\'layout-template-active\':\'layout-template\'}}" ng-click="selectTemplate($event,2)">'+
			      		'<div style="width:100%;height:33.3%;border-bottom: 1px solid #d5d9dc;"></div>'+
			      		'<div style="width:50%;height:33.3%;border-bottom: 1px solid #d5d9dc;border-right:1px solid #d5d9dc;float:left"></div>'+
			      		'<div style="width:50%;height:33.3%;border-bottom: 1px solid #d5d9dc;float:left"></div>'+
			      		'<div style="width:50%;height:33.3%;border-right:1px solid #d5d9dc;float:left"></div>'+
			      		'<div style="width:50%;height:33.3%;float:left"></div>'+
			      	'</div>'
    }
});

linkerCloud.directive('layout3', function() {
    return {
    	restrict : 'E',
    	template: 	'<div class="{{selectedTemplateType == 3?\'layout-template-active\':\'layout-template\'}}" ng-click="selectTemplate($event,3)">'+
			      		'<div style="width:50%;height:33.3%;border-bottom: 1px solid #d5d9dc;border-right:1px solid #d5d9dc;float:left"></div>'+
			      		'<div style="width:50%;height:33.3%;border-bottom: 1px solid #d5d9dc;float:left"></div>'+
			      		'<div style="width:100%;height:33.3%;border-bottom: 1px solid #d5d9dc;float:left"></div>'+
			      		'<div style="width:50%;height:33.3%;border-right:1px solid #d5d9dc;float:left"></div>'+
			      		'<div style="width:50%;height:33.3%;float:left"></div>'+
			      	'</div>'
    }
});

linkerCloud.directive('layout4', function() {
    return {
    	restrict : 'E',
    	template: 	'<div class="{{selectedTemplateType == 4?\'layout-template-active\':\'layout-template\'}}" ng-click="selectTemplate($event,4)">'+
			      		'<div style="width:50%;height:33.3%;border-bottom: 1px solid #d5d9dc;border-right:1px solid #d5d9dc;float:left"></div>'+
			      		'<div style="width:50%;height:33.3%;border-bottom: 1px solid #d5d9dc;float:left"></div>'+
			      		'<div style="width:50%;height:33.3%;border-bottom: 1px solid #d5d9dc;border-right:1px solid #d5d9dc;float:left"></div>'+
			      		'<div style="width:50%;height:33.3%;border-bottom: 1px solid #d5d9dc;float:left"></div>'+
			      		'<div style="width:100%;height:33.3%;float:left"></div>'+
			      	'</div>'
    }
});