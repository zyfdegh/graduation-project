var login = angular.module('login',['ngRoute','ui.bootstrap','pascalprecht.translate','ngCookies']);

login.config(function($routeProvider, $locationProvider) {
	$routeProvider
		.when('/', {
    		templateUrl: 'templates/login/signin.html',
    		controller: 'LoginController'
  		})
  		.when('/signup', {
    		templateUrl: 'templates/login/signup.html',
    		controller: 'SignUpController'
  		})
      .when('/activeSuccess', {
        templateUrl: 'templates/login/active_success.html',
        controller: 'ActiveSuccessController'
      })
      .when('/activeFailed', {
        templateUrl: 'templates/login/active_failed.html',
        controller: 'ActiveFailedController'
      })
  		.otherwise({  
            redirectTo: '/'
       });
})
.config(['$translateProvider',function($translateProvider) {
     var lang="zh";
     $translateProvider.useStaticFilesLoader({
        prefix: '/portal-ui/locales/',
        suffix: '.json'
     });
     $translateProvider.useLocalStorage();
     $translateProvider.useSanitizeValueStrategy('escape');
     $translateProvider.preferredLanguage(lang);    
}])
.run(['$rootScope', '$translate',
    function ($rootScope, $translate) {
       $rootScope.$translate = $translate;      
    }
  ])
.run(['$rootScope',
    function ($rootScope) {
      // Wait for window load
      $(window).load(function () {
        $rootScope.loaded = true;
      });
    }
  ]);