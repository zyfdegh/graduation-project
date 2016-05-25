var navigators = [        			
		{"category" : "leftNav.category.usageInformation","child":[
			{"name" : "leftNav.usageInformation.dashboard","href":"/products/dashboard","ngclass": "active"},
			{"name" : "leftNav.usageInformation.serviceSubscriptions","href":"/products/services","ngclass": ""},			
		]},
		{"category" : "leftNav.category.onboard","child":[
			{"name" : "leftNav.linkOps.contentManagement","href":"/products/onboard/file","ngclass": ""}					
		]},
		{"category" : "leftNav.category.serviceDesigner","href":"/products/designer","ngclass": "","child":[
			{"name" : "leftNav.serviceDesigner.appModel","href":"/products/designer/app","ngclass": ""},
			{"name" : "leftNav.serviceDesigner.serviceModel","href":"/products/designer/service","ngclass": ""},			
		]},
		{"category" : "leftNav.category.resources","child":[
			{"name" : "leftNav.resources.platformAccount","href":"/products/resources","ngclass": ""}
		]},
		{"category" : "leftNav.category.identityManagement","href":"/products/identity","ngclass": "","child":[			
			{"name" : "leftNav.identityManagement.tenant","href":"/products/identity/tenant","ngclass": ""},
		]},	
		{"category" : "leftNav.category.payment","child":[			
			{"name" : "leftNav.payment.billing","href":"/products/mb/billing","ngclass": ""}			
		]},	
		{"category" : "leftNav.category.linkOps","child":[			
			{"name" : "leftNav.linkOps.project","href":"/products/linkops/project","ngclass": ""},
			{"name" : "leftNav.linkOps.contentManagement","href":"/products/linkops/content","ngclass": ""},
			// {"name" : "leftNav.linkOps.toolsInformation","href":"/products/linkops/tool","ngclass": ""},		
		]},			
		{"category" : "leftNav.category.customization","child":[			
			{"name" : "leftNav.customization.layoutManagement","href":"/products/layout","ngclass": ""}				
		]}
  	];

var selectedPath = "/products/dashboard";
var defaultState = "products.dashboard";

var selectNav = function(path){
	_.each(navigators,function(item){
  		if(item.href != path){
  			item.ngclass = "";
  		}else{
  			item.ngclass = "active";
  		}
  		if(!_.isUndefined(item.child)){
  			_.each(item.child,function(subitem){
  				if(subitem.href != path){
		  			subitem.ngclass = "";
		  		}else{
		  			subitem.ngclass = "active";
		  		}
  			})
  		}
  	});
  	selectedPath = path;
};

var getNavs = function(){
	return sessionStorage.navigators ? sessionStorage.navigators : navigators;
}

var getCurrentPath = function(){
	return selectedPath;
}
var getDefaultState = function(){
	return defaultState;
}

function productsService($location){
	return {
		"selectNav" : selectNav,
		"getNavs" : getNavs,
		"getCurrentPath" : getCurrentPath,
		"getDefaultState" : getDefaultState
	}
}

linkerCloud.factory('productsService', ['$location', productsService]);