var linkerCloud = angular.module('LinkerCloud',['ngRoute','ui.router','ui.bootstrap','nvd3','pascalprecht.translate','ngCookies']);

linkerCloud.run(['$rootScope', '$state', '$stateParams',
    function ($rootScope, $state, $stateParams) {
       $rootScope.$state = $state;
       $rootScope.$stateParams = $stateParams;
    }
  ]
).config(['$stateProvider', '$urlRouterProvider', function ($stateProvider, $urlRouterProvider) {
      $urlRouterProvider.otherwise("/home");
      $stateProvider
        .state("home", {
          url: "/home",
          templateUrl: "templates/main/home.html",
          controller: 'HomeController'
        })      
        .state("solutions", {
          url: "/solutions",
          templateUrl: 'templates/main/solutions.html',
          controller: 'SolutionsController'
        })
        .state("pricing", {
          url: "/pricing",
          templateUrl: 'templates/main/pricing.html',
          controller: 'PricingController'
        })
        .state("partners", {
          url: "/partners",
          templateUrl: 'templates/main/partners.html',
          controller: 'PartnersController'
        })
        .state("documents", {
          url: "/documents",
          templateUrl: 'templates/main/documents.html',
          controller: 'DocumentsController'
        })
}])
.config(['$stateProvider', '$urlRouterProvider', function ($stateProvider, $urlRouterProvider) {
      $stateProvider      
         .state("products", {
            url: "/products",
            templateUrl: 'templates/main/products.html',
            controller: 'ProductsController'
         })
         .state('products.dashboard', {
            url: '/dashboard',
            views:{
              'productsMainpage' : {
                   templateUrl: 'templates/product/dashboard.html',
                   controller: 'DashboardController'
              }
            }
           
        })
         
         .state('products.services', {
            url: '/services',
            views:{
              'productsMainpage' : {
                 templateUrl: 'templates/product/services.html',
                 controller: 'ServicesController'
              }
            }
            
        })         
             
         .state('products.layout', {
            url: '/layout',
            views:{
              'productsMainpage' : {
                   templateUrl: 'templates/product/layout.html',
                   controller: 'LayoutController'
              }
            }
           
        })       
         
}])
.config(['$stateProvider', '$urlRouterProvider', function ($stateProvider, $urlRouterProvider) {
      $stateProvider      
         .state("products.designer", {
            url: "/designer",
            views:{
              'productsMainpage' : {
                  templateUrl: 'templates/product/designer/main.html',
                  controller: 'DesignerController'
              }
            }
            
         })
         .state('products.designer.app', {
            url: '/app',
            views:{
              'productsMainpage@products' : {
                  templateUrl: 'templates/product/designer/app.html',
                  controller: 'AppDesignerController'
              }
            }
           
        })
         .state('products.designer.service', {
            url: '/service',
            views:{
              'productsMainpage@products' : {
                  templateUrl: 'templates/product/designer/service.html',
                  controller: 'ServiceDesignerController'
              }
            }
            
        })
         .state('products.designer.package', {
            url: '/package',
            views:{
              'productsMainpage@products' : {
                   templateUrl: 'templates/product/designer/package.html',
                   controller: 'PackageDesignerController'
              }
            }
           
        })
      
}])
.config(['$stateProvider', '$urlRouterProvider', function ($stateProvider, $urlRouterProvider) {
      $stateProvider      
         .state("products.resources", {
            url: "/resources",
            views:{
              'productsMainpage' : {
                    templateUrl: 'templates/product/resources/main.html',
                    controller: 'ResourcesController'
              }
            }            
         })
         .state('products.resources.account', {
            url: '/account',
            views:{
              'productsMainpage@products' : {
                     templateUrl: 'templates/product/resources/compute.html',
                     controller: 'ComputeResourcesController'
              }
            }
           
        })
         
}])
.config(['$stateProvider', '$urlRouterProvider', function ($stateProvider, $urlRouterProvider) {
      $stateProvider      
         .state("products.identity", {
            url: "/identity",
             views:{
              'productsMainpage' : {
                      templateUrl: 'templates/product/identity/main.html',
                      controller: 'IdentityController'
              }
            }
           
         })
        
         .state('products.identity.tenant', {
            url: '/tenant',
            views:{
              'productsMainpage@products' : {
                      templateUrl: 'templates/product/identity/tenant.html',
                      controller: 'TenantIdentityController'
              }
            }
           
        })
        .state('products.identity.role', {
            url: '/role',
            views:{
              'productsMainpage@products' : {
                       templateUrl: 'templates/product/identity/role.html',
                       controller: 'RoleIdentityController'
              }
            }
           
        })
        .state('products.identity.user', {
            url: '/user',
            views:{
              'productsMainpage@products' : {
                       templateUrl: 'templates/product/identity/user.html',
                       controller: 'UserIdentityController'
              }
            }
           
        })
}])
.config(['$stateProvider', '$urlRouterProvider', function ($stateProvider, $urlRouterProvider) {
      $stateProvider      
        .state('products.mb', {
            url: '/mb',
            views:{
              'productsMainpage' : {
                 templateUrl: 'templates/product/mb.html',
                 controller: 'MBController'
              }
            }
            
         })        
         .state('products.mb.metering', {
            url: '/metering',
            views:{
              'productsMainpage@products' : {
                      templateUrl: 'templates/product/mb/metering.html',
                      controller: 'MeteringController'
              }
            }
           
        })
        .state('products.mb.billing', {
            url: '/billing',
            views:{
              'productsMainpage@products' : {
                       templateUrl: 'templates/product/payment/billing.html',
                       controller: 'BillingController'
              }
            }
           
        })
        
}])
.config(['$stateProvider', '$urlRouterProvider', function ($stateProvider, $urlRouterProvider) {
      $stateProvider      
        .state('products.linkops', {
            url: '/linkops'           
        })     
         .state('products.linkops.project', {
            url: '/project',
            views:{
              'productsMainpage@products' : {
                      templateUrl: 'templates/product/linkops/project/main.html',
                      controller: 'ProjectController'
              }
            }
           
        })
        .state('products.linkops.content', {
            url: '/content',
            views:{
              'productsMainpage@products' : {
                       templateUrl: 'templates/product/linkops/content/main.html',
                       controller: 'ContentController'
              }
            }
           
        })
        .state('products.linkops.tool', {
            url: '/tool',
            views:{
              'productsMainpage@products' : {
                       templateUrl: 'templates/product/linkops/tool/main.html',
                       controller: 'ToolController'
              }
            }
           
        })
        
}])
.config(['$stateProvider', '$urlRouterProvider', function ($stateProvider, $urlRouterProvider) {
      $stateProvider      
        .state('products.onboard', {
            url: '/onboard'           
        })     
         .state('products.onboard.file', {
            url: '/file',
            views:{
              'productsMainpage@products' : {
                      templateUrl: 'templates/product/linkops/content/main.html',
                      controller: 'ContentController'
              }
            }
           
        })
     
        
}])
.config(['$translateProvider',function($translateProvider) {
     var lang="zh";
     $translateProvider.useStaticFilesLoader({
        prefix: '/portal-ui/locales/',
        suffix: '.json'
     });
     $translateProvider.useLocalStorage();
     $translateProvider.useSanitizeValueStrategy('escape');
     // $translateProvider.fallbackLanguage(lang);
     $translateProvider.preferredLanguage(lang);
     // $translateProvider.useMessageFormatInterpolation();
     
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
