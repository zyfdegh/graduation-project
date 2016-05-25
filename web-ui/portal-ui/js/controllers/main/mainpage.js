linkerCloud.controller('MainNavController', ['$scope','$location', 'logoutService','$window', 'responseService','productsService','langService',function($scope,$location,logoutService,$window,responseService,productsService,langService) {  
    $scope.supportedLangs = langService.getSupportedLangs();
     var storage = $scope.$translate.storage();
     var key = $scope.$translate.storageKey();
     var lang = storage.get(key)||'zh';
      $scope.selectedLang =  _.find($scope.supportedLangs,function(langObj){ 
           return langObj.name == lang;           
     });   
     $scope.$translate.use($scope.selectedLang.name);
     $scope.logged = (_.isUndefined(sessionStorage.username) || _.isEmpty(sessionStorage.username)) ? false : true;
     if($scope.logged){
       $scope.currentUser = {
          "name" : sessionStorage.username          
       };
     }
    // $scope.$translate("mainMenu.market").then(function(translate){
    //    var a=translate;
    //    console.log(a);
    // });
	  $scope.navigators = [        
    		{"name" : "mainMenu.market","href":"/home","ngclass": "active","ngshow":true},
    		{"name" : "mainMenu.workspace","href":"/products","ngclass": "","ngshow":$scope.logged},
    		{"name" : "mainMenu.solutions","href":"/solutions","ngclass": "","ngshow":true},
    		{"name" : "mainMenu.pricing","href":"/pricing","ngclass": "","ngshow":true},
    		{"name" : "mainMenu.partners","href":"/partners","ngclass": "","ngshow":true},
    		{"name" : "mainMenu.documents","href":"/documents","ngclass": "","ngshow":true}
  	];
    $scope.isCollapsed = true;
  	$scope.leftNavs = productsService.getNavs();

  	$scope.selectthis = function(name){
  		_.each($scope.navigators,function(item){
  			if(item.name != name){
  				item.ngclass = "";
  			}else{
  				item.ngclass = "active";
  			}
  		})

  	}
  	
  	function forRefresh(){
  		var path = $location.path();
  		_.each($scope.navigators,function(item){
  			if(path != ""){
  				if(item.href != path && path.indexOf("/products")<0){
	  				item.ngclass = "";
	  			}else if(path.indexOf("/products")>=0){
	  				if(item.href == "/products"){
	  					item.ngclass = "active";
	  				}else{
	  					item.ngclass = "";
	  				}
	  			}else{
	  				item.ngclass = "active";
	  			}
  			}
  		})
  	}
  	
  	forRefresh();
 
  	$scope.logout = function(){ 		 
          var key;         
          for (var i = sessionStorage.length - 1; i >= 0; i--) {
              key = sessionStorage.key(i);
              sessionStorage.removeItem(key);
          }	
          logoutService.doLogout().then(function(response){
            if(responseService.successResponse(response)){          
               window.location="/portal-ui/index.html"; 
            }
          },
          function(errorMessage){
            responseService.errorResponse(errorMessage);        
          });
          
  	};
    $scope.login = function(){
          window.location="/portal-ui/login.html"; 
    };
    $scope.signup = function(){
          window.location="/portal-ui/login.html#/signup"; 
    };
    $scope.languageSwitch = function(lang){
          $scope.selectedLang = lang;
          $scope.$translate.use(lang.name);  
          loadBundles(lang.name);
    };
	
	function loadBundles(lang) {
		$.i18n.properties({
			name:'Messages', 
			path:'js/non-angular/i18n/' + lang + '/', 
			mode:'map'
		});
	}
}]);