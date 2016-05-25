'use strict';
var request = require('request');
require('sugar');

var logger = global.obj.logger;
var urlCfg = global.obj.urlCfg;
var linkerConf = global.obj.linkerPortalCfg;

var providerUtil = require('../utils/providerUtil');
var ProviderUtil = new providerUtil("controllerProvider");


module.exports = function (app) {
      app.get('/docker/image', function(req, res) {    
          var url = "";
          var options = {
            url: linkerConf.linkerRepo.url,
            method: 'GET',
            json: true
          };
          var callback = function(error, response, body) {
            if (!error && response.statusCode == 200) {
              logger.info('Success ' + options.method +' request ' + options.url);
              // support v1 and v2 repositories
              if (options.url.indexOf("/v1/") != -1) {
                logger.info('get linker v1 repositories api.');
                res.status(200).send(body);
              }else {
                logger.info('get linker v2 repositories api.');
                var v2result = body.repositories.map(function(imageName){
                  return {
                    "name": imageName,
                    "description": ""
                  }
                });
                res.status(200).send({"results":v2result})
              }
            } else {
              if (error) {
                logger.error('Error ' + options.method +' request ' + options.url, error);
                res.status(500).send(error.errno);
              }else if(response.statusCode >= 400){
                logger.error('Error ' + response.statusCode + ' ' + options.method + ' request ' + options.url);
                res.status(response.statusCode).send(error);
              } 
            }
          };
          logger.trace("Start to get all image from linker repo " + options.url);
          request(options, callback);
    });
    app.get('/docker/imageTag', function(req, res) {    
          var imageName = req.query.imageName;
          var url = linkerConf.linkerRepo.tagUrl.replace(/{variable}/g, encodeURIComponent(imageName));
          var options = {
            url: url,
            method: 'GET',
            json: true
          };
          var callback = function(error, response, body) {
            if (!error && response.statusCode == 200) {
              // support v1 and v2 repositories
              if (options.url.indexOf("/v2/") != -1) {
                logger.info('get linker v2 repositories api.');
                res.status(200).send(body);
              }else {
                logger.info('get linker v1 repositories api.');
                var v1result = {
                  "name":imageName, 
                  "tags":Object.keys(body)
                }
                res.status(200).send(v1result)
            } 
          }else {
              if (error) {
                logger.error('Error ' + options.method +' request ' + options.url, error);
                res.status(500).send(error.errno);
              }else if(response.statusCode >= 400){
                logger.error('Error ' + response.statusCode + ' ' + options.method + ' request ' + options.url);
                res.status(response.statusCode).send(error);
              } 
            }
          };
          logger.trace("Start to get all image from linker repo " + options.url);
          request(options, callback);
    });

};