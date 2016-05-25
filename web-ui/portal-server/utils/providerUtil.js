'use strict';

require('sugar');
var controllerHA = global.obj.linkerPortalCfg.controllerProvider.ha;

var ProviderUtil = function (providerType) {
  this.providerType = providerType ? providerType : "controllerProvider";
};

ProviderUtil.prototype.rebuildUrl = function(path){
	return "http://" + path;
};
ProviderUtil.prototype.parseControllerUrl = function(req,res,next){    
    if(controllerHA.enabled){
         global.obj.zkUtil_controller.getControllerUrl();        
    }else{
         global.obj.controller_url = controllerHA.controller_url;
    }
    return next();
};

module.exports = ProviderUtil;
