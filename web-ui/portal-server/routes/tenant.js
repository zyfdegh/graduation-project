'use strict';
var request = require('request');

var urlCfg = global.obj.urlCfg;
var linkerConf = global.obj.linkerPortalCfg;

var logger = require('../utils/logger');

var Authentication = require('../utils/authentication');

var providerUtil = require('../utils/providerUtil');
var ProviderUtil = new providerUtil("controllerProvider");

module.exports = function (app) {
      app.get('/tenant', Authentication.ensureAuthenticated,ProviderUtil.parseControllerUrl, function(req, res) {
        var options = {
          url: ProviderUtil.rebuildUrl(global.obj.controller_url + urlCfg.user_api.tenant),
          method: 'GET',
          json:true,
          headers: {
             'X-Auth-Token': req.session.token
          }
        };
        var callback = function(error, response, body) {
          if (!error && response.statusCode == 200) {
            res.status(200).send(body);           
          } else {
          	if (error) {
              logger.error('Error ' + options.method +' request ' + options.url, error);
              res.status(500).send(error.errno); //"ECONNREFUSED"
            }else if(response.statusCode >= 400){
              logger.error('Error ' + response.statusCode + ' ' + options.method + ' request ' + options.url);
              res.status(response.statusCode).send(response.body.error);
            } 
          }
        };      
        request(options, callback);
    });
    
    
  
};