'use strict';
var request = require('request');

var logger = global.obj.logger;
var urlCfg = global.obj.urlCfg;
var linkerConf = global.obj.linkerPortalCfg;

var providerUtil = require('../utils/providerUtil');
var ProviderUtil = new providerUtil("controllerProvider");


module.exports = function (app) {
      app.post('/dockerHub', function(req, res) {    
      var url = "";
      if(req.body.url == ""){
        url = linkerConf.dockerHub.url + "query="+req.body.imageName + "&page=1";
      }else{
        url = req.body.url;
      }
      var options = {
        url: url,
        method: 'GET'
        
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
      logger.trace("Start to get all image from dockerHub " + options.url);
      request(options, callback);
    });

};