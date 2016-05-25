linkerCloud.controller('ThemeController', ['$scope',function($scope) {
  	
    var stylesheets = [
    	{
    		"name" : "default",
  			"hrefs" : ["css/themes/default/common.css","css/themes/default/main.css","css/themes/default/product.css","css/themes/default/solutions.css"]
  		},
  		{
  			"name" : "dark",
  			"hrefs" : ["css/themes/dark/common.css","css/themes/dark/main.css","css/themes/dark/product.css"]
  		}
  	];
  	
  	var initTheme = function(){
  		sessionStorage.theme = "default";
  		var theme = sessionStorage.theme;
  		$scope.stylesheet = _.find(stylesheets,function(css){
  			return css.name == theme;
  		});
  	}
  	
  	initTheme();
}]);