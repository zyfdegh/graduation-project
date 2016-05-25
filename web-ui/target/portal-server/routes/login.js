'use strict';
var request = require('request');

var urlCfg = global.obj.urlCfg;
var linkerConf = global.obj.linkerPortalCfg;
var logger = require('../utils/logger');

var providerUtil = require('../utils/providerUtil');
var ProviderUtil = new providerUtil("controllerProvider");

module.exports = function (app) {     
    app.post('/user/login', ProviderUtil.parseControllerUrl, function(req, res) {
      var options = {
        url: ProviderUtil.rebuildUrl(global.obj.controller_url + urlCfg.user_api.login),
        method: 'POST',
        json: true,
        body: JSON.stringify(req.body)
      };
      var callback = function(error, response, body) {
        if (!error && response.statusCode == 200) {         
          req.session.token = body.data.id;        
          res.status(200).send(body);
          logger.info('Success ' + options.method +' request ' + options.url);
        } else {
          if (error) {
            logger.error('Error ' + options.method +' request ' + options.url, error);
            res.status(500).send(error.errno); //"ECONNREFUSED"
          }else if(response.statusCode >= 400){
            logger.error('Error ' + response.statusCode + ' ' + options.method + ' request ' + options.url);
            res.status(response.statusCode).send(response.body.error.name);
          } 
        }
      };
      request(options, callback);
    });
    app.post('/user/registry', ProviderUtil.parseControllerUrl, function(req, res) {
      var options = {
        url: ProviderUtil.rebuildUrl(global.obj.controller_url + urlCfg.user_api.registry),
        method: 'POST',
        json: true,
        body: JSON.stringify(req.body)
      };
      var callback = function(error, response, body) {
        if (!error && response.statusCode == 201) {
          res.status(200).send(body);
          logger.info('Success ' + options.method +' request ' + options.url);
        } else {
          if (error) {
            logger.error('Error ' + options.method +' request ' + options.url, error);
            res.status(500).send(error.errno);
          }else if(response.statusCode >= 400){
            logger.error('Error ' + response.statusCode + ' ' + options.method + ' request ' + options.url);
            res.status(response.statusCode).send(response.body.error.name);
          } 
        }
      };
      request(options, callback);
    });
    app.get('/logout', function(req, res){
       req.session.destroy(function(err) {
          if(err){           
            res.status(500).send("Log out failed!");
          }else{
            res.status(200).send();
          }
       })
     
    });
    app.get('/user/active',ProviderUtil.parseControllerUrl, function(req,res){
        var activeCode = req.query.activeCode;
        var uid = req.query.uid;
        var options = {
          url: ProviderUtil.rebuildUrl(global.obj.controller_url + urlCfg.user_api.active + "?uid=" + uid + "&activeCode=" + activeCode),
          method: 'GET',
          json: true
        };
        var callback = function(error, response, body) {
          if (!error && response.statusCode == 200) {
            logger.info('Success ' + options.method +' request ' + options.url);
            res.redirect("portal-ui/login.html#/activeSuccess");
          } else {
            if (error) {
              logger.error('Error ' + options.method +' request ' + options.url, error);
              res.status(500).send(error.errno);
            }else if(response.statusCode >= 400){
              logger.error('Error ' + response.statusCode + ' ' + options.method + ' request ' + options.url);
              var encodedReason = encodeURIComponent(response.body.error.name);
              res.redirect("portal-ui/login.html#/activeFailed?uid="+uid+"&reason=" + encodedReason);
           
            } 
          }
        };
      request(options, callback);

    });
    app.get('/user/reactive',ProviderUtil.parseControllerUrl, function(req,res){       
        var uid = req.query.uid;
        var options = {
          url: ProviderUtil.rebuildUrl(global.obj.controller_url + urlCfg.user_api.reactive + "/" + uid),
          method: 'GET',
          json: true
        };
        var callback = function(error, response, body) {
          if (!error && (response.statusCode == 201 || response.statusCode == 200)) {
            // res.redirect("portal-ui/login.html");
            res.status(200).send(body);
          } else {
            if (error) {
              logger.error('Error ' + options.method +' request ' + options.url, error);
              res.status(500).send(error.errno);
            }else if(response.statusCode >= 400){
              logger.error('Error ' + response.statusCode + ' ' + options.method + ' request ' + options.url);
              var encodedReason = encodeURIComponent(response.body.error.name);
              res.redirect("portal-ui/login.html#/activeFailed?uid="+uid+"&reason=" + encodedReason);
            
            } 
          }
        };
      request(options, callback);

    });
  
};