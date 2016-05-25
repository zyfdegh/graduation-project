'use strict';
var request = require('request');

var logger = global.obj.logger;
var urlCfg = global.obj.urlCfg;
var linkerConf = global.obj.linkerPortalCfg;

var Authentication = require('../utils/authentication');

var providerUtil = require('../utils/providerUtil');
var ProviderUtil = new providerUtil("controllerProvider");


module.exports = function (app) {
      app.get('/apps', Authentication.ensureAuthenticated, ProviderUtil.parseControllerUrl, function(req, res) {
      var options = {
        url: ProviderUtil.rebuildUrl(global.obj.controller_url + urlCfg.model_api.app_service),
        method: 'GET',
        json:true,
        headers: {
             'X-Auth-Token': req.session.token
        }
      };
      var callback = function(error, response, body) {
        if (!error && response.statusCode == 200) {
          res.status(200).send(body);
          logger.info('Success ' + options.method +' request ' + options.url);
        } else {
        	if (error) {
            logger.error('Error ' + options.method +' request ' + options.url, error);
            res.status().send(error);
          }else if(response.statusCode >= 400){
            logger.error('Error ' + response.statusCode + ' ' + options.method + ' request ' + options.url);
            res.status(response.statusCode).send(response.body.error);
          } 
        }
      };
      logger.trace("Start to get all apps by request " + options.url);
      request(options, callback);
    });
	
	app.get('/apps/operations/:appid', Authentication.ensureAuthenticated, ProviderUtil.parseControllerUrl, function(req, res) {
      var options = {
        url: ProviderUtil.rebuildUrl(global.obj.controller_url + urlCfg.model_api.app_operations + req.params.appid),
        method: 'GET',
        json:true,
        headers: {
             'X-Auth-Token': req.session.token
        }
      };
      var callback = function(error, response, body) {
        if (!error && response.statusCode == 200) {
          res.status(200).send(body);
          logger.info('Success ' + options.method +' request ' + options.url);
        } else {
        	if (error) {
            logger.error('Error ' + options.method +' request ' + options.url, error);
            res.status().send(error);
          }else if(response.statusCode >= 400){
            logger.error('Error ' + response.statusCode + ' ' + options.method + ' request ' + options.url);
            res.status(response.statusCode).send(response.body.error);
          } 
        }
      };
      logger.trace("Start to get operations of apps by request " + options.url);
      request(options, callback);
    });
    
    app.post('/apps', Authentication.ensureAuthenticated, ProviderUtil.parseControllerUrl, function(req, res) {
      var options = {
        url : ProviderUtil.rebuildUrl(global.obj.controller_url + urlCfg.model_api.app_service), 
        method: 'POST',
        json: true,
        body: JSON.stringify(req.body),
        headers: {
             'X-Auth-Token': req.session.token
        }
      };
      var callback = function(error, response, body) {
        if (!error && response.statusCode == 201) {
          res.status(200).send(body);
          logger.info('Success ' + options.method +' request ' + options.url);
        } else {
          if (error) {
            logger.error('Error ' + options.method +' request ' + options.url, error);
            res.status().send(error);
          }else if(response.statusCode >= 400){
            logger.error('Error ' + response.statusCode + ' ' + options.method + ' request ' + options.url);
            res.status(response.statusCode).send(error);
          } 
        }
      };
      logger.trace("Start to create app by request " + options.url);
      request(options, callback);
    });

    app.put('/apps/:appid', Authentication.ensureAuthenticated, ProviderUtil.parseControllerUrl, function(req, res) {
      var options = {
        url: ProviderUtil.rebuildUrl(global.obj.controller_url + urlCfg.model_api.app_service + req.params.appid),
        method: 'PUT',
        json: true,
        body: JSON.stringify(req.body),
        headers: {
             'X-Auth-Token': req.session.token
        }
      };
      var callback = function(error, response, body) {
        //console.log(response.statusCode)
        if (!error && response.statusCode == 200) {
          res.status(200).send(body);
          logger.info('Success ' + options.method +' request ' + options.url);
        } else {
          if (error) {
            logger.error('Error ' + options.method +' request ' + options.url, error);
            res.status().send(error);
          }else if(response.statusCode >= 400){
            logger.error('Error ' + response.statusCode + ' ' + options.method + ' request ' + options.url);
            res.status(response.statusCode).send(error);
          } 
        }
      };
      logger.trace("Start to update app by request " + options.url);
      request(options, callback);
    });

    app.delete('/apps/:appid', Authentication.ensureAuthenticated, ProviderUtil.parseControllerUrl, function(req, res) {
      var options = {
        url: ProviderUtil.rebuildUrl(global.obj.controller_url + urlCfg.model_api.app_service + req.params.appid),
        method: 'DELETE',
        headers: {
             'X-Auth-Token': req.session.token
        }
      };
      var callback = function(error, response, body) {
        //console.log(response.statusCode)
        if (!error && response.statusCode == 200) {
          res.status(200).send(body);
          logger.info('Success ' + options.method +' request ' + options.url);
        } else {
          if (error) {
            logger.error('Error ' + options.method +' request ' + options.url, error);
            res.status().send(error);
          }else if(response.statusCode >= 400){
            logger.error('Error ' + response.statusCode + ' ' + options.method + ' request ' + options.url);
            res.status(response.statusCode).send(error);
          }
        }
      };
      logger.trace("Start to delete app by request " + options.url);
      request(options, callback);
    });
  //app apis end

  //servicegroup apis
    app.get('/serviceGroups/published', ProviderUtil.parseControllerUrl, function(req, res) {
      var options = {
        url: ProviderUtil.rebuildUrl(global.obj.controller_url + urlCfg.model_api.serviceGroup_service_published + "?query=" + req.query.query),
        method: 'GET',
        json:true   
      };
      var callback = function(error, response, body) {
        //console.log(response.statusCode)
        if (!error && response.statusCode == 200) {
          res.status(200).send(body);
          logger.info('Success ' + options.method +' request ' + options.url);
        } else {
          if (error) {
            logger.error('Error ' + options.method +' request ' + options.url, error);
            res.status().send(error);
          }else if(response.statusCode >= 400){
            logger.error('Error ' + response.statusCode + ' ' + options.method + ' request ' + options.url);
            res.status(response.statusCode).send(response.body.error);
          }
        }
      };
      logger.trace("Start to get service groups by request " + options.url);
      request(options, callback);
    });
    app.get('/serviceGroups', Authentication.ensureAuthenticated, ProviderUtil.parseControllerUrl, function(req, res) {
      var options = {
        url: ProviderUtil.rebuildUrl(global.obj.controller_url + urlCfg.model_api.serviceGroup_service),
        method: 'GET',
        json:true,
        headers: {
             'X-Auth-Token': req.session.token
        }
      };
      var callback = function(error, response, body) {
        //console.log(response.statusCode)
        if (!error && response.statusCode == 200) {
          res.status(200).send(body);
          logger.info('Success ' + options.method +' request ' + options.url);
        } else {
          if (error) {
            logger.error('Error ' + options.method +' request ' + options.url, error);
            res.status().send(error);
          }else if(response.statusCode >= 400){
            logger.error('Error ' + response.statusCode + ' ' + options.method + ' request ' + options.url);
            res.status(response.statusCode).send(response.body.error);
          }
        }
      };
      logger.trace("Start to get service groups by request " + options.url);
      request(options, callback);
    });
	
	app.get('/serviceGroups/operations/:sgid', Authentication.ensureAuthenticated, ProviderUtil.parseControllerUrl, function(req, res) {
      var options = {
        url: ProviderUtil.rebuildUrl(global.obj.controller_url + urlCfg.model_api.serviceGroup_operations + req.params.sgid),
        method: 'GET',
        json:true,
        headers: {
             'X-Auth-Token': req.session.token
        }
      };
      var callback = function(error, response, body) {
        if (!error && response.statusCode == 200) {
          res.status(200).send(body);
          logger.info('Success ' + options.method +' request ' + options.url);
        } else {
        	if (error) {
            logger.error('Error ' + options.method +' request ' + options.url, error);
            res.status().send(error);
          }else if(response.statusCode >= 400){
            logger.error('Error ' + response.statusCode + ' ' + options.method + ' request ' + options.url);
            res.status(response.statusCode).send(response.body.error);
          } 
        }
      };
      logger.trace("Start to get operations of service groups by request " + options.url);
      request(options, callback);
    });
    
    app.post('/serviceGroups', Authentication.ensureAuthenticated, ProviderUtil.parseControllerUrl, function(req, res) {
     var options = {
        url: ProviderUtil.rebuildUrl(global.obj.controller_url + urlCfg.model_api.serviceGroup_service),
        method: 'POST',
        json: true,
        body: JSON.stringify(req.body),
        headers: {
             'X-Auth-Token': req.session.token
        }
      };
      var callback = function(error, response, body) {
        //console.log(response.statusCode)
        if (!error && response.statusCode == 201) {
          res.status(200).send(body);
          logger.info('Success ' + options.method +' request ' + options.url);
        } else {
          if (error) {
            logger.error('Error ' + options.method +' request ' + options.url, error);
            res.status().send(error);
          }else if(response.statusCode >= 400){
            logger.error('Error ' + response.statusCode + ' ' + options.method + ' request ' + options.url);
            res.status(response.statusCode).send(response.body.error);
          }
        }
      };
      logger.trace("Start to add service groups by request " + options.url);
      request(options, callback);
    });

    app.put('/serviceGroups/:sgid', Authentication.ensureAuthenticated, ProviderUtil.parseControllerUrl, function(req, res) {
      var options = {
        url: ProviderUtil.rebuildUrl(global.obj.controller_url + urlCfg.model_api.serviceGroup_service + req.params.sgid),
        method: 'PUT',
        json: true,
        body: JSON.stringify(req.body),
        headers: {
             'X-Auth-Token': req.session.token
        }
      };
      var callback = function(error, response, body) {
        //console.log(response.statusCode)
        if (!error && response.statusCode == 200) {
          res.status(200).send(body);
          logger.info('Success ' + options.method +' request ' + options.url);
        } else {
          if (error) {
            logger.error('Error ' + options.method +' request ' + options.url, error);
            res.status().send(error);
          }else if(response.statusCode >= 400){
            logger.error('Error ' + response.statusCode + ' ' + options.method + ' request ' + options.url);
            res.status(response.statusCode).send(error);
          }
        }
      };
      logger.trace("Start to update service groups by request " + options.url);
      request(options, callback);
    });
	
	app.put('/serviceGroups/publish/:sgid', Authentication.ensureAuthenticated, ProviderUtil.parseControllerUrl, function(req, res) {
      var options = {
        url: ProviderUtil.rebuildUrl(global.obj.controller_url + urlCfg.model_api.publish_serviceGroup_service + req.params.sgid),
        method: 'PUT',
        headers: {
             'X-Auth-Token': req.session.token
        }
      };
      var callback = function(error, response, body) {
        if (!error && response.statusCode == 200) {
          res.status(200).send(body);
          logger.info('Success ' + options.method +' request ' + options.url);
        } else {
          if (error) {
            logger.error('Error ' + options.method +' request ' + options.url, error);
            res.status().send(error);
          }else if(response.statusCode >= 400){
            logger.error('Error ' + response.statusCode + ' ' + options.method + ' request ' + options.url);
            res.status(response.statusCode).send(error);
          }
        }
      };
      logger.trace("Start to publish service groups by request " + options.url);
      request(options, callback);
    });
    
     app.put('/serviceGroups/unpublish/:sgid', Authentication.ensureAuthenticated, ProviderUtil.parseControllerUrl, function(req, res) {
      var options = {
        url: ProviderUtil.rebuildUrl(global.obj.controller_url + urlCfg.model_api.unpublish_serviceGroup_service + req.params.sgid),
        method: 'PUT',
        headers: {
             'X-Auth-Token': req.session.token
        }
      };
      var callback = function(error, response, body) {
        //console.log(response.statusCode)
        if (!error && response.statusCode == 200) {
          res.status(200).send(body);
          logger.info('Success ' + options.method +' request ' + options.url);
        } else {
          if (error) {
            logger.error('Error ' + options.method +' request ' + options.url, error);
            res.status().send(error);
          }else if(response.statusCode >= 400){
            logger.error('Error ' + response.statusCode + ' ' + options.method + ' request ' + options.url);
            res.status(response.statusCode).send(error);
          }
        }
      };
      logger.trace("Start to unpublish service groups by request " + options.url);
      request(options, callback);
    });
    
    app.put('/serviceGroups/submit/:sgid', Authentication.ensureAuthenticated, ProviderUtil.parseControllerUrl, function(req, res) {
      var options = {
        url: ProviderUtil.rebuildUrl(global.obj.controller_url + urlCfg.model_api.submit_serviceGroup_service + req.params.sgid),
        method: 'PUT',
        headers: {
             'X-Auth-Token': req.session.token
        }
      };
      var callback = function(error, response, body) {
        if (!error && response.statusCode == 200) {
          res.status(200).send(body);
          logger.info('Success ' + options.method +' request ' + options.url);
        } else {
          if (error) {
            logger.error('Error ' + options.method +' request ' + options.url, error);
            res.status().send(error);
          }else if(response.statusCode >= 400){
            logger.error('Error ' + response.statusCode + ' ' + options.method + ' request ' + options.url);
            res.status(response.statusCode).send(error);
          }
        }
      };
      logger.trace("Start to submit service groups by request " + options.url);
      request(options, callback);
    });
    
    app.delete('/serviceGroups/:sgid', Authentication.ensureAuthenticated, ProviderUtil.parseControllerUrl, function(req, res) {
      var options = {
        url: ProviderUtil.rebuildUrl(global.obj.controller_url + urlCfg.model_api.serviceGroup_service + req.params.sgid),
        method: 'DELETE',
        headers: {
             'X-Auth-Token': req.session.token
        }
      };
      var callback = function(error, response, body) {
        //console.log(response.statusCode)
        if (!error && response.statusCode == 200) {
          res.status(200).send(body);
          logger.info('Success ' + options.method +' request ' + options.url);
        } else {
          if (error) {
            logger.error('Error ' + options.method +' request ' + options.url, error);
            res.status().send(error);
          }else if(response.statusCode >= 400){
            logger.error('Error ' + response.statusCode + ' ' + options.method + ' request ' + options.url);
            res.status(response.statusCode).send(error);
          }
        }
      };
      logger.trace("Start to update service groups by request " + options.url);
      request(options, callback);
    });
  //servicegroup apis end

  //order service apis
  	app.get('/serviceGroupOrders', Authentication.ensureAuthenticated, ProviderUtil.parseControllerUrl, function(req, res) {
      var options = {
        url: ProviderUtil.rebuildUrl(global.obj.controller_url + urlCfg.model_api.serviceGroupOrders_service  + "?query=" + req.query.query),
        method: 'GET',
        json:true,
        headers: {
             'X-Auth-Token': req.session.token
        }
      };
      var callback = function(error, response, body) {
        if (!error && (response.statusCode == 201 || response.statusCode == 200)) {
          res.status(200).send(body);
          logger.info('Success ' + options.method +' request ' + options.url);
        } else {
          if (error) {
            logger.error('Error ' + options.method +' request ' + options.url, error);
            res.status().send(error);
          }else if(response.statusCode >= 400){
            logger.error('Error ' + response.statusCode + ' ' + options.method + ' request ' + options.url);
            res.status(response.statusCode).send(response.body.error);
          }
        }
      };
      logger.trace("Start to order service by request " + options.url);
      request(options, callback);
    });
    
    app.get('/serviceGroupOrders/operations/:sgoid', Authentication.ensureAuthenticated, ProviderUtil.parseControllerUrl, function(req, res) {
      var options = {
        url: ProviderUtil.rebuildUrl(global.obj.controller_url + urlCfg.model_api.serviceGroupOrders_operations + req.params.sgoid),
        method: 'GET',
        json:true,
        headers: {
             'X-Auth-Token': req.session.token
        }
      };
      var callback = function(error, response, body) {
        if (!error && response.statusCode == 200) {
          res.status(200).send(body);
          logger.info('Success ' + options.method +' request ' + options.url);
        } else {
        	if (error) {
            logger.error('Error ' + options.method +' request ' + options.url, error);
            res.status().send(error);
          }else if(response.statusCode >= 400){
            logger.error('Error ' + response.statusCode + ' ' + options.method + ' request ' + options.url);
            res.status(response.statusCode).send(response.body.error);
          } 
        }
      };
      logger.trace("Start to get operations of sgo by request " + options.url);
      request(options, callback);
    });
    
    app.post('/serviceGroupOrders', Authentication.ensureAuthenticated, ProviderUtil.parseControllerUrl, function(req, res) {
      var options = {
        url: ProviderUtil.rebuildUrl(global.obj.controller_url + urlCfg.model_api.serviceGroupOrders_service),      
        method: 'POST',
        body: JSON.stringify(req.body),
        json:true,
        headers: {
             'X-Auth-Token': req.session.token
        }
      };
      var callback = function(error, response, body) {
        //console.log(response.statusCode)
        if (!error && (response.statusCode == 201 || response.statusCode == 200)) {
          res.status(200).send(body);
          logger.info('Success ' + options.method +' request ' + options.url);
        } else {
          if (error) {
            logger.error('Error ' + options.method +' request ' + options.url, error);
            res.status().send(error);
          }else if(response.statusCode >= 400){
            logger.error('Error ' + response.statusCode + ' ' + options.method + ' request ' + options.url);
            res.status(response.statusCode).send(response.body.error);
          }
        }
      };
      logger.trace("Start to order service by request " + options.url);
      request(options, callback);
    });
    
    //scale to
    app.put('/serviceGroupOrders/:orderid/scaleApp', Authentication.ensureAuthenticated, ProviderUtil.parseControllerUrl, function(req, res) {
      var options = {
        url: ProviderUtil.rebuildUrl(global.obj.controller_url + urlCfg.model_api.serviceGroupOrders_service  + req.params.orderid + "/scaleApp" + "?appId=" + req.query.appId + "&num=" + req.query.num),      
        method: 'PUT',
        headers: {
             'X-Auth-Token': req.session.token
        }
      };
      var callback = function(error, response, body) {
        //console.log(response.statusCode)
        if (!error && response.statusCode == 200) {
          res.status(200).send(body);
          logger.info('Success ' + options.method +' request ' + options.url);
        } else {
          if (error) {
            logger.error('Error ' + options.method +' request ' + options.url, error);
            res.status().send(error);
          }else if(response.statusCode >= 400){
            logger.error('Error ' + response.statusCode + ' ' + options.method + ' request ' + options.url);
            res.status(response.statusCode).send(error);
          }
        }
      };
      logger.trace("Start to scale app by request " + options.url);
      request(options, callback);
    });
		//scale to end
		
		//terminate order
		app.delete('/serviceGroupOrders/:orderid', Authentication.ensureAuthenticated, ProviderUtil.parseControllerUrl, function(req, res) {
      var options = {
        url: ProviderUtil.rebuildUrl(global.obj.controller_url + urlCfg.model_api.serviceGroupOrders_service + req.params.orderid),
        method: 'DELETE',
        headers: {
             'X-Auth-Token': req.session.token
        }
      };
      var callback = function(error, response, body) {
        if (!error && response.statusCode == 200) {
          res.status(200).send(body);
          logger.info('Success ' + options.method +' request ' + options.url);
        } else {
          if (error) {
            logger.error('Error ' + options.method +' request ' + options.url, error);
            res.status().send(error);
          }else if(response.statusCode >= 400){
            logger.error('Error ' + response.statusCode + ' ' + options.method + ' request ' + options.url);
            res.status(response.statusCode).send(error);
          }
        }
      };
      logger.trace("Start to terminate service instance by request " + options.url);
      request(options, callback);
    });
		//terminate order end
		
  //order service apis end

  //list instance apis
    app.get('/groupInstances', Authentication.ensureAuthenticated, ProviderUtil.parseControllerUrl, function(req, res) {
      var options = {
        url: ProviderUtil.rebuildUrl(global.obj.controller_url + urlCfg.model_api.instance_service),
        method: 'GET',
        headers: {
             'X-Auth-Token': req.session.token
        }
      };
      var callback = function(error, response, body) {
        //console.log(response.statusCode)
        if (!error && response.statusCode == 200) {
          res.status(200).send(body);
          logger.info('Success ' + options.method +' request ' + options.url);
        } else {
          if (error) {
            logger.error('Error ' + options.method +' request ' + options.url, error);
            res.status().send(error);
          }else if(response.statusCode >= 400){
            logger.error('Error ' + response.statusCode + ' ' + options.method + ' request ' + options.url);
            res.status(response.statusCode).send(response.body.error);
          }
        }
      };
      logger.trace("Start to get service instances by request " + options.url);
      request(options, callback);
    });
  //list instance apis end

    app.get('/groupInstances/:sgid', Authentication.ensureAuthenticated, ProviderUtil.parseControllerUrl, function(req, res) {
      var options = {
        url: ProviderUtil.rebuildUrl(global.obj.controller_url + urlCfg.model_api.instance_service + req.params.sgid),
        method: 'GET',
        headers: {
             'X-Auth-Token': req.session.token
        }
      };
      var callback = function(error, response, body) {
        // console.log(body)
        if (!error && response.statusCode == 200) {
          res.status(200).send(body);
          logger.info('Success ' + options.method +' request ' + options.url);
        } else {
          if (error) {
            logger.error('Error ' + options.method +' request ' + options.url, error);
            res.status().send(error);
          }else if(response.statusCode >= 400){
            logger.error('Error ' + response.statusCode + ' ' + options.method + ' request ' + options.url);
            res.status(response.statusCode).send(error);
          }
        }
      };
      logger.trace("Start to get service instance detail by request " + options.url);
      request(options, callback);
    });
    
  //list app instance  apis
    app.get('/appInstances/:instanceid', Authentication.ensureAuthenticated, ProviderUtil.parseControllerUrl, function(req, res) {
      var options = {
        url: ProviderUtil.rebuildUrl(global.obj.controller_url + urlCfg.model_api.app_instance_service + req.params.instanceid),
        method: 'GET',
        headers: {
             'X-Auth-Token': req.session.token
        }
      };
      var callback = function(error, response, body) {
        if (!error && response.statusCode == 200) {
          res.status(200).send(body);
          logger.info('Success ' + options.method +' request ' + options.url);
        } else {
          if (error) {
            logger.error('Error ' + options.method +' request ' + options.url, error);
            res.status().send(error);
          }else if(response.statusCode >= 400){
            logger.error('Error ' + response.statusCode + ' ' + options.method + ' request ' + options.url);
            res.status(response.statusCode).send(error);
          }
        }
      };
      logger.trace("Start to get service instance app detail by request " + options.url);
      request(options, callback);
    });
  //list instance apis end
  
  //app config
		app.get('/appConfigs', Authentication.ensureAuthenticated, ProviderUtil.parseControllerUrl, function(req, res) {
      var options = {
        url: ProviderUtil.rebuildUrl(global.obj.controller_url + urlCfg.model_api.app_config_service + "?query=" + req.query.query),
        method: 'GET',
        json:true,
        headers: {
             'X-Auth-Token': req.session.token
        }
      };
      var callback = function(error, response, body) {
        if (!error && response.statusCode == 200) {
          res.status(200).send(body);
          logger.info('Success ' + options.method +' request ' + options.url);
        } else {
          if (error) {
            logger.error('Error ' + options.method +' request ' + options.url, error);
            res.status().send(error);
          }else if(response.statusCode >= 400){
            logger.error('Error ' + response.statusCode + ' ' + options.method + ' request ' + options.url);
            res.status(response.statusCode).send(error);
          }
        }
      };
      logger.trace("Start to get configuration package by request " + options.url);
      request(options, callback);
    });
    
    app.post('/appConfigs', Authentication.ensureAuthenticated, ProviderUtil.parseControllerUrl, function(req, res) {
		 if(req.body.configurations == undefined){
     		req.body.configurations = [];
     	}else{
     		for(var i=0;i<req.body.configurations.length;i++){
     			if(req.body.configurations[i].preconditions == undefined){
     				req.body.configurations[i].preconditions = [];
     			}
     			if(req.body.configurations[i].steps == undefined){
     				req.body.configurations[i].steps = [];
     			}
     		}
     	}
     	if(req.body.notifies == undefined){
     		req.body.notifies = [];
     	}
      var options = {
        url : ProviderUtil.rebuildUrl(global.obj.controller_url + urlCfg.model_api.app_config_service), 
        method: 'POST',
        json: true,
        body: JSON.stringify(req.body),
        headers: {
             'X-Auth-Token': req.session.token
        }
      };
      var callback = function(error, response, body) {
        if (!error && response.statusCode == 201) {
          res.status(200).send(body);
          logger.info('Success ' + options.method +' request ' + options.url);
        } else {
          if (error) {
            logger.error('Error ' + options.method +' request ' + options.url, error);
            res.status().send(error);
          }else if(response.statusCode >= 400){
            logger.error('Error ' + response.statusCode + ' ' + options.method + ' request ' + options.url);
            res.status(response.statusCode).send(error);
          } 
        }
      };
      logger.trace("Start to create configuration package by request " + options.url);
      request(options, callback);
    });

    app.put('/appConfigs/:cpid', Authentication.ensureAuthenticated, ProviderUtil.parseControllerUrl, function(req, res) {
		if(req.body.configurations == undefined){
     		req.body.configurations = [];
     	}else{
     		for(var i=0;i<req.body.configurations.length;i++){
     			if(req.body.configurations[i].preconditions == undefined){
     				req.body.configurations[i].preconditions = [];
     			}
     			if(req.body.configurations[i].steps == undefined){
     				req.body.configurations[i].steps = [];
     			}
     		}
     	}
     	if(req.body.notifies == undefined){
     		req.body.notifies = [];
     	}
      var options = {
        url: ProviderUtil.rebuildUrl(global.obj.controller_url + urlCfg.model_api.app_config_service + req.params.cpid),
        method: 'PUT',
        json: true,
        body: JSON.stringify(req.body),
        headers: {
             'X-Auth-Token': req.session.token
        }
      };
      var callback = function(error, response, body) {
        //console.log(response.statusCode)
        if (!error && response.statusCode == 200) {
          res.status(200).send(body);
          logger.info('Success ' + options.method +' request ' + options.url);
        } else {
          if (error) {
            logger.error('Error ' + options.method +' request ' + options.url, error);
            res.status().send(error);
          }else if(response.statusCode >= 400){
            logger.error('Error ' + response.statusCode + ' ' + options.method + ' request ' + options.url);
            res.status(response.statusCode).send(error);
          } 
        }
      };
      logger.trace("Start to update configuration package by request " + options.url);
      request(options, callback);
    });
    
    app.delete('/appConfigs', Authentication.ensureAuthenticated, ProviderUtil.parseControllerUrl, function(req, res) {
      var options = {
        url: ProviderUtil.rebuildUrl(global.obj.controller_url + urlCfg.model_api.app_config_service + "?query=" + req.query.query),
        method: 'DELETE',
        headers: {
             'X-Auth-Token': req.session.token
        }
      };
      var callback = function(error, response, body) {
        if (!error && response.statusCode == 200) {
          res.status(200).send(body);
          logger.info('Success ' + options.method +' request ' + options.url);
        } else {
          if (error) {
            logger.error('Error ' + options.method +' request ' + options.url, error);
            res.status().send(error);
          }else if(response.statusCode >= 400){
            logger.error('Error ' + response.statusCode + ' ' + options.method + ' request ' + options.url);
            res.status(response.statusCode).send(error);
          }
        }
      };
      logger.trace("Start to delete configuration package by request " + options.url);
      request(options, callback);
    });
  //app config end
  
  //billing
  app.get('/billing', Authentication.ensureAuthenticated, ProviderUtil.parseControllerUrl, function(req, res) {
      var options = {
        url: ProviderUtil.rebuildUrl(global.obj.controller_url + urlCfg.payment_api.billing),
        method: 'GET',
        json:true,
        headers: {
             'X-Auth-Token': req.session.token
        }
      };
      var callback = function(error, response, body) {
        if (!error && response.statusCode == 200) {
          res.status(200).send(body);
          logger.info('Success ' + options.method +' request ' + options.url);
        } else {
        	if (error) {
            logger.error('Error ' + options.method +' request ' + options.url, error);
            res.status().send(error);
          }else if(response.statusCode >= 400){
            logger.error('Error ' + response.statusCode + ' ' + options.method + ' request ' + options.url);
            res.status(response.statusCode).send(response.body.error);
          } 
        }
      };
      logger.trace("Start to get billing models by request " + options.url);
      request(options, callback);
    });
    
    app.get('/billing/all', ProviderUtil.parseControllerUrl, function(req, res) {
      var options = {
        url: ProviderUtil.rebuildUrl(global.obj.controller_url + urlCfg.payment_api.billing_all),
        method: 'GET',
        json:true
      };
      var callback = function(error, response, body) {
        if (!error && response.statusCode == 200) {
          res.status(200).send(body);
          logger.info('Success ' + options.method +' request ' + options.url);
        } else {
        	if (error) {
            logger.error('Error ' + options.method +' request ' + options.url, error);
            res.status().send(error);
          }else if(response.statusCode >= 400){
            logger.error('Error ' + response.statusCode + ' ' + options.method + ' request ' + options.url);
            res.status(response.statusCode).send(response.body.error);
          } 
        }
      };
      logger.trace("Start to get billing models by request " + options.url);
      request(options, callback);
    });
    
    app.post('/billing', Authentication.ensureAuthenticated, ProviderUtil.parseControllerUrl, function(req, res) {
      var options = {
        url: ProviderUtil.rebuildUrl(global.obj.controller_url + urlCfg.payment_api.billing),      
        method: 'POST',
        body: JSON.stringify(req.body),
        headers: {
             'X-Auth-Token': req.session.token
        }
      };
      var callback = function(error, response, body) {
        //console.log(response.statusCode)
        if (!error && (response.statusCode == 201 || response.statusCode == 200)) {
          res.status(200).send(body);
          logger.info('Success ' + options.method +' request ' + options.url);
        } else {
          if (error) {
            logger.error('Error ' + options.method +' request ' + options.url, error);
            res.status().send(error);
          }else if(response.statusCode >= 400){
            logger.error('Error ' + response.statusCode + ' ' + options.method + ' request ' + options.url);
            res.status(response.statusCode).send(error);
          }
        }
      };
      logger.trace("Start to create billing model by request " + options.url);
      request(options, callback);
    });
    
    app.put('/billing/:billingid', Authentication.ensureAuthenticated, ProviderUtil.parseControllerUrl, function(req, res) {
      var options = {
        url: ProviderUtil.rebuildUrl(global.obj.controller_url + urlCfg.payment_api.billing  + req.params.billingid),      
        method: 'PUT',
        body: JSON.stringify(req.body),
        headers: {
             'X-Auth-Token': req.session.token
        }
      };
      var callback = function(error, response, body) {
        //console.log(response.statusCode)
        if (!error && response.statusCode == 200) {
          res.status(200).send(body);
          logger.info('Success ' + options.method +' request ' + options.url);
        } else {
          if (error) {
            logger.error('Error ' + options.method +' request ' + options.url, error);
            res.status().send(error);
          }else if(response.statusCode >= 400){
            logger.error('Error ' + response.statusCode + ' ' + options.method + ' request ' + options.url);
            res.status(response.statusCode).send(error);
          }
        }
      };
      logger.trace("Start to update billing model by request " + options.url);
      request(options, callback);
    });
    
    app.get('/userAccounts', Authentication.ensureAuthenticated, ProviderUtil.parseControllerUrl, function(req, res) {
      var options = {
        url: ProviderUtil.rebuildUrl(global.obj.controller_url + urlCfg.payment_api.userAccounts + "?count=true&query=" + req.query.query + "&skip=" + req.query.skip + "&limit=" + req.query.limit),
        method: 'GET',
        json:true,
        headers: {
             'X-Auth-Token': req.session.token
        }
      };
      var callback = function(error, response, body) {
        if (!error && response.statusCode == 200) {
          res.status(200).send(body);
          logger.info('Success ' + options.method +' request ' + options.url);
        } else {
        	if (error) {
            logger.error('Error ' + options.method +' request ' + options.url, error);
            res.status().send(error);
          }else if(response.statusCode >= 400){
            logger.error('Error ' + response.statusCode + ' ' + options.method + ' request ' + options.url);
            res.status(response.statusCode).send(response.body.error);
          } 
        }
      };
      logger.trace("Start to get billing records by request " + options.url);
      request(options, callback);
    });
  //billing end
 }