linkerCloud.directive('layoutdesigner', function() {
    return {
    	restrict : 'E',
    	template: 	'<div class="layout-designer-block" ng-repeat="layout in layoutData" style="width:{{layout.width}}; opacity:{{layout.enable?1:0.3}}" data-type="{{layout.type}}">'+
			      		'<div style="width:50%;float: left;">'+
			      			'<span style="float: left;margin-left: 10px;" ng-if="!layout.edit">{{layout.title | translate}}</span>'+
			      			'<input style="float: left;margin-left: 10px;width:200px" ng-if="layout.edit" type="text" ng-model="layout.title"/>'+
			      		'</div>'+
			      		'<div style="width:50%;float:right;">'+
			      			'<span class="glyphicon glyphicon-remove-circle layout-designer-button" ng-if="layout.enable" ng-click="layout.enable = false" title="Hide Block"></span>'+
			      			'<span class="glyphicon glyphicon-ok-circle layout-designer-button" ng-if="!layout.enable" ng-click="layout.enable = true" title="Show Block"></span>'+
			      			'<span class="glyphicon glyphicon-arrow-down layout-designer-button" ng-if="isShowButton($index,1) == 1" ng-click="downPos($index)"></span>'+
			      			'<span class="glyphicon glyphicon-arrow-up layout-designer-button" ng-if="isShowButton($index,2) == 1" ng-click="upPos($index)"></span>'+
			      			'<span class="glyphicon glyphicon-arrow-left layout-designer-button" ng-if="isShowButton($index,3) == 1" ng-click="leftPos($index)"></span>'+
			      			'<span class="glyphicon glyphicon-arrow-right layout-designer-button" ng-if="isShowButton($index,4) == 1" ng-click="rightPos($index)"></span>'+
//			      			'<span class="glyphicon glyphicon-edit layout-designer-button" ng-if="!layout.edit && layout.type != \'advertise\'" ng-click="layout.edit = true" title="Edit Title"></span>'+
//			      			'<span class="glyphicon glyphicon-floppy-disk layout-designer-button" ng-if="layout.edit" ng-click="layout.edit = false" title="Save Title"></span>'+
			      		'</div>'+
			      	'</div>'      	
    }
});
