'use strict';
var request = require('request');

global.obj.data = require("../test/projects.json");

var urlCfg = global.obj.urlCfg;
var linkerConf = global.obj.linkerPortalCfg;
var logger = require('../utils/logger');
var fs = require('fs');
var path = require('path');
var Authentication = require('../utils/authentication');

var providerUtil = require('../utils/providerUtil');
var ProviderUtil = new providerUtil("linkopsProvider");
var multipartMiddleware = global.obj.multipartMiddleware;
var handleError = function(error){
     return {"name" : error.errno};
}
module.exports = function (app) {
      //upload file
      app.post('/linkops/uploadFile', Authentication.ensureAuthenticated, multipartMiddleware, function(req, res) {
        var options = {
          url: linkerConf.linkopsProvider + urlCfg.linkops_api.content,
          method: 'POST',
          headers: {
           'Content-Type': "multipart/form-data",
           'X-Auth-Token' : req.session.token
          },
          json:true
        };
        var imageName = req.body.imagename;
        var dockerFile = req.body.dockerfile;
        var version = req.body.version;
        var email = req.body.email;
        var file = req.files;

        var callback = function(error, response, body) {
          if (!error && (response.statusCode == 200 ||response.statusCode == 201)) {
            res.status(200).send(body);
            logger.info('Success ' + options.method +' request ' + options.url);
          } else {
          	if (error) {
              logger.error('Error ' + options.method +' request ' + options.url, error);
              res.status(500).send(handleError(error));
            }else if(response.statusCode >= 400){
              logger.error('Error ' + response.statusCode + ' ' + options.method + ' request ' + options.url);
              res.status(response.statusCode).send(response.body.error);
            } 
          }
        };
        logger.trace("Start to get all projects by request " + options.url);
        var r = request(options, callback); 
        var form = r.form();
        
        form.append("imagename",imageName);
        form.append("dockerfile",dockerFile);
        form.append("zipfilename", file.file.name);
        form.append("version",version);
        form.append("email",email);
        form.append('file', fs.createReadStream(file.file.path));
     });
     //list file
     app.get('/linkops/content', Authentication.ensureAuthenticated, function(req, res) {
        var options = {
          url: linkerConf.linkopsProvider + urlCfg.linkops_api.content,
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
               res.status(500).send(handleError(error));
            }else if(response.statusCode >= 400){
              logger.error('Error ' + response.statusCode + ' ' + options.method + ' request ' + options.url);
              res.status(response.statusCode).send(response.body.error);
            } 
          }
        };      
        request(options, callback);
    });
     //delete file
    app.delete('/linkops/content/:id',  Authentication.ensureAuthenticated, function(req, res) {
        var options = {
            url: linkerConf.linkopsProvider + urlCfg.linkops_api.content + "/" + req.params.id,
            method: 'DELETE',
            headers: {
             'X-Auth-Token': req.session.token
            }
        };
        var callback = function(error, response, body) {
          if (!error && response.statusCode == 200) {
            res.status(200).send(body);
           
          } else {
            if (error) {
              res.status(500).send(handleError(error));
            }else if(response.statusCode >= 400){
              
              res.status(response.statusCode).send(response.body.error);
            }
          }
        };
       
        request(options, callback);
    });
    //update file
    app.put('/linkops/content/:id',  Authentication.ensureAuthenticated, function(req, res) {
        var options = {
            url: linkerConf.linkopsProvider + urlCfg.linkops_api.content + "/" + req.params.id,
            method: 'put',
            headers: {
             'X-Auth-Token': req.session.token
            }
        };
        var callback = function(error, response, body) {
          if (!error && response.statusCode == 200) {
            res.status(200).send(body);
           
          } else {
            if (error) {
              res.status(500).send(handleError(error));
            }else if(response.statusCode >= 400){
              
              res.status(response.statusCode).send(response.body.error);
            }
          }
        };
       
        request(options, callback);
    });
    //list env
    app.get('/linkops/env', Authentication.ensureAuthenticated,function(req, res) {
        var options = {
          url: linkerConf.linkopsProvider + urlCfg.linkops_api.env,
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
              res.status(500).send(handleError(error));
            }else if(response.statusCode >= 400){
              logger.error('Error ' + response.statusCode + ' ' + options.method + ' request ' + options.url);
              res.status(response.statusCode).send(response.body.error);
            } 
          }
        };      
        request(options, callback);
    });
    //refresh env information
    app.post('/linkops/env/:orderid', Authentication.ensureAuthenticated,function(req, res) {
        var options = {
          url: linkerConf.linkopsProvider + urlCfg.linkops_api.env + "/notify?orderid=" + req.params.orderid,
          method: 'POST',
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
              res.status(500).send(handleError(error));
            }else if(response.statusCode >= 400){
              logger.error('Error ' + response.statusCode + ' ' + options.method + ' request ' + options.url);
              res.status(response.statusCode).send(response.body.error);
            } 
          }
        };      
        request(options, callback);
    });
    //create env
    app.post('/linkops/env', Authentication.ensureAuthenticated, function(req, res) {
        var envName = req.body.name;
        var options = {
          url: linkerConf.linkopsProvider + urlCfg.linkops_api.env + "?envname=" + envName,
          method: 'POST',
          json:true,
          headers: {
             'X-Auth-Token': req.session.token
          }
        };
        var callback = function(error, response, body) {
          if (!error && (response.statusCode == 200 ||response.statusCode == 201)) {
            res.status(200).send(body);           
          } else {
            if (error) {
              logger.error('Error ' + options.method +' request ' + options.url, error);
              res.status(500).send(handleError(error));
            }else if(response.statusCode >= 400){
              logger.error('Error ' + response.statusCode + ' ' + options.method + ' request ' + options.url);
              res.status(response.statusCode).send(response.body.error);
            } 
          }
        };      
        request(options, callback);
    });
    //list project in env
    app.get('/linkops/project/:envid', Authentication.ensureAuthenticated,function(req, res) {
        var options = {
          url: linkerConf.linkopsProvider + urlCfg.linkops_api.project + "?opsenv_id=" + req.params.envid+"&status=running",
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
              res.status(500).send(handleError(error));
            }else if(response.statusCode >= 400){
              logger.error('Error ' + response.statusCode + ' ' + options.method + ' request ' + options.url);
              res.status(response.statusCode).send(response.body.error);
            } 
          }
        };      
        request(options, callback);
    });

    //delete env
    app.delete('/linkops/env/:id',  Authentication.ensureAuthenticated, function(req, res) {
        var options = {
            url: linkerConf.linkopsProvider + urlCfg.linkops_api.env + "/" + req.params.id,
            method: 'DELETE',
            headers: {
             'X-Auth-Token': req.session.token
            }
        };
        var callback = function(error, response, body) {
          if (!error && response.statusCode == 200) {
            res.status(200).send(body);
           
          } else {
            if (error) {
              res.status(500).send(handleError(error));
            }else if(response.statusCode >= 400){
              
              res.status(response.statusCode).send(response.body.error);
            }
          }
        };
       
        request(options, callback);
    });
    //create project in env
    app.post('/linkops/project',  Authentication.ensureAuthenticated, function(req, res) {
      var options = {
        url: linkerConf.linkopsProvider + urlCfg.linkops_api.project + "?opsenv_id=" + req.body.envId + "&name=" + req.body.name+"&sm_id="+req.body.selectedServiceGroup,
        method: 'POST',
        json: true,
        headers: {
             'X-Auth-Token': req.session.token
        }
      };
      var callback = function(error, response, body) {
        if (!error && (response.statusCode == 200 ||response.statusCode == 201)) {
          res.status(200).send(body);
          logger.info('Success ' + options.method +' request ' + options.url);
        } else {
          if (error) {
            logger.error('Error ' + options.method +' request ' + options.url, error);
            res.status(500).send(handleError(error));
          }else if(response.statusCode >= 400){
            logger.error('Error ' + response.statusCode + ' ' + options.method + ' request ' + options.url);
            res.status(response.statusCode).send(response.body.error);
          } 
        }
      };
      logger.trace("Start to create project by request " + options.url);
      request(options, callback);
    });
    //delete project
    app.delete('/linkops/project/:id',  Authentication.ensureAuthenticated, function(req, res) {
        var options = {
            url: linkerConf.linkopsProvider + urlCfg.linkops_api.project + "/" + req.params.id,
            method: 'DELETE',
            headers: {
             'X-Auth-Token': req.session.token
            }
        };
        var callback = function(error, response, body) {
          if (!error && response.statusCode == 200) {
            res.status(200).send(body);           
          } else {
            if (error) {
              res.status(500).send(handleError(error));
            }else if(response.statusCode >= 400){             
              res.status(response.statusCode).send(response.body.error);
            }
          }
        };
       
        request(options, callback);
    });
    
     //list job in project
    app.get('/linkops/job/:projectid', Authentication.ensureAuthenticated,function(req, res) {
        var options = {
          url: linkerConf.linkopsProvider + urlCfg.linkops_api.job + "?projectid=" + req.params.projectid,
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
              res.status(500).send(handleError(error));
            }else if(response.statusCode >= 400){
              logger.error('Error ' + response.statusCode + ' ' + options.method + ' request ' + options.url);
              res.status(response.statusCode).send(response.body.error);
            } 
          }
        };      
        request(options, callback);
    });

    //create job for project
    app.post('/linkops/job',  Authentication.ensureAuthenticated, function(req, res) {
      var options = {
        url: linkerConf.linkopsProvider + urlCfg.linkops_api.job + "?projectid=" + req.body.projectId + "&jobname=" + req.body.name+"&version="+req.body.version+"&branch="+req.body.branch+"&autodeploy="+req.body.autoDeploy,
        method: 'POST',
        json: true,
        headers: {
             'X-Auth-Token': req.session.token
        }
      };
      var callback = function(error, response, body) {
        if (!error && (response.statusCode == 200 ||response.statusCode == 201)) {
          res.status(200).send(body);
          logger.info('Success ' + options.method +' request ' + options.url);
        } else {
          if (error) {
            logger.error('Error ' + options.method +' request ' + options.url, error);
            res.status(500).send(handleError(error));
          }else if(response.statusCode >= 400){
            logger.error('Error ' + response.statusCode + ' ' + options.method + ' request ' + options.url);
            res.status(response.statusCode).send(response.body.error);
          } 
        }
      };
      logger.trace("Start to create project by request " + options.url);
      request(options, callback);
    });
    
    //delete job
    app.delete('/linkops/job/:name',  Authentication.ensureAuthenticated, function(req, res) {
        var options = {
            url: linkerConf.linkopsProvider + urlCfg.linkops_api.job + "/" + req.params.name + "?projectid=" + req.query.projectId,
            method: 'DELETE',
            headers: {
             'X-Auth-Token': req.session.token
            }
        };
        var callback = function(error, response, body) {
          if (!error && response.statusCode == 200) {
            res.status(200).send(body);           
          } else {
            if (error) {
              res.status(500).send(handleError(error));
            }else if(response.statusCode >= 400){             
              res.status(response.statusCode).send(response.body.error);
            }
          }
        };
       
        request(options, callback);
    });
    //terminate job
    app.delete('/linkops/job/terminate/:id',  Authentication.ensureAuthenticated, function(req, res) {
        var options = {
            url: linkerConf.linkopsProvider + urlCfg.linkops_api.job + "/" + req.params.id + "/jobenvs",
            method: 'DELETE',
            headers: {
             'X-Auth-Token': req.session.token
            }
        };
        var callback = function(error, response, body) {
          if (!error && response.statusCode == 200) {
            res.status(200).send(body);           
          } else {
            if (error) {
              res.status(500).send(handleError(error));
            }else if(response.statusCode >= 400){             
              res.status(response.statusCode).send(response.body.error);
            }
          }
        };
       
        request(options, callback);
    });
    //build job
    app.put('/linkops/job/:id',  Authentication.ensureAuthenticated, function(req, res) {
        var options = {
            url: linkerConf.linkopsProvider + urlCfg.linkops_api.job + "/" + req.params.id + "/build?projectid=" + req.query.projectId,
            method: 'PUT',
            headers: {
             'X-Auth-Token': req.session.token
            }
        };
        var callback = function(error, response, body) {
          if (!error && response.statusCode == 200) {
            res.status(200).send(body);           
          } else {
            if (error) {
              res.status(500).send(handleError(error));
            }else if(response.statusCode >= 400){             
              res.status(response.statusCode).send(response.body.error);
            }
          }
        };
       
        request(options, callback);
    });
    //deploy job
    app.post('/linkops/job/:id/jobenv',  Authentication.ensureAuthenticated, function(req, res) {
      var options = {
        url: linkerConf.linkopsProvider + urlCfg.linkops_api.job + "/" + req.params.id + "/jobenvs",
        method: 'POST',
        json: true,
        headers: {
             'X-Auth-Token': req.session.token
        }
      };
      var callback = function(error, response, body) {
        if (!error && (response.statusCode == 200 ||response.statusCode == 201)) {
          res.status(200).send(body);
          logger.info('Success ' + options.method +' request ' + options.url);
        } else {
          if (error) {
            logger.error('Error ' + options.method +' request ' + options.url, error);
            res.status(500).send(handleError(error));
          }else if(response.statusCode >= 400){
            logger.error('Error ' + response.statusCode + ' ' + options.method + ' request ' + options.url);
            res.status(response.statusCode).send(response.body.error);
          } 
        }
      };
      logger.trace("Start to create project by request " + options.url);
      request(options, callback);
    });
     //list artifact in project
    app.get('/linkops/artifact/:projectid', Authentication.ensureAuthenticated,function(req, res) {
        var options = {
          url: linkerConf.linkopsProvider + urlCfg.linkops_api.project + "/" + req.params.projectid + "/artifacts",
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
              res.status(500).send(handleError(error));
            }else if(response.statusCode >= 400){
              logger.error('Error ' + response.statusCode + ' ' + options.method + ' request ' + options.url);
              res.status(response.statusCode).send(response.body.error);
            } 
          }
        };      
        request(options, callback);
    });
     //create artifact in project
    app.post('/linkops/artifact', Authentication.ensureAuthenticated,function(req, res) {
        var content = {"name":req.body.name,"type":req.body.type,"group_id":req.body.groupId,"df_ids":req.body.selectedDockerfiles};
        var options = {
          url: linkerConf.linkopsProvider + urlCfg.linkops_api.project + "/" + req.body.projectId + "/artifacts",
          method: 'POST',
          json:true,
          headers: {
             'X-Auth-Token': req.session.token
          },
          body: JSON.stringify(content)
      
        };
        var callback = function(error, response, body) {
          if (!error && (response.statusCode == 200 ||response.statusCode == 201)) {
            res.status(200).send(body);           
          } else {
            if (error) {
              logger.error('Error ' + options.method +' request ' + options.url, error);
              res.status(500).send(handleError(error));
            }else if(response.statusCode >= 400){
              logger.error('Error ' + response.statusCode + ' ' + options.method + ' request ' + options.url);
              res.status(response.statusCode).send(response.body.error);
            } 
          }
        };      
        request(options, callback);
    });
   //delete artifact
    app.delete('/linkops/artifact/:id',  Authentication.ensureAuthenticated, function(req, res) {
        var options = {
            url: linkerConf.linkopsProvider + urlCfg.linkops_api.project + "/" + req.query.projectId + "/artifacts/" + req.params.id,
            method: 'DELETE',
            headers: {
             'X-Auth-Token': req.session.token
            }
        };
        var callback = function(error, response, body) {
          if (!error && response.statusCode == 200) {
            res.status(200).send(body);           
          } else {
            if (error) {
              res.status(500).send(handleError(error));
            }else if(response.statusCode >= 400){             
              res.status(response.statusCode).send(response.body.error);
            }
          }
        };
       
        request(options, callback);
    });
     //UPDATE artifact
    app.put('/linkops/artifact/:id',  Authentication.ensureAuthenticated, function(req, res) {
        var content = {"name":req.body.name,"type":req.body.type,"group_id":req.body.groupId,"df_ids":req.body.selectedDockerfiles};
        console.log("aaaaa");
        var options = {
            url: linkerConf.linkopsProvider + urlCfg.linkops_api.project + "/" + req.body.projectId + "/artifacts/" + req.body.id,
            method: 'PUT',
            json:true,
            headers: {
             'X-Auth-Token': req.session.token
            },
            body:JSON.stringify(content)
        };
        var callback = function(error, response, body) {
          if (!error && response.statusCode == 200) {
            res.status(200).send(body);           
          } else {
            if (error) {
              res.status(500).send(handleError(error));
            }else if(response.statusCode >= 400){             
              res.status(response.statusCode).send(response.body.error);
            }
          }
        };
       
        request(options, callback);
    });
      //list project envs deployed by job
    app.get('/linkops/projectenvs/:jobid', Authentication.ensureAuthenticated,function(req, res) {
        var queryParam = JSON.stringify({"job_id" : req.params.jobid});
        var options = {
          url: linkerConf.linkopsProvider + urlCfg.linkops_api.projectenvs + "?query=" + queryParam,
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
              res.status(500).send(handleError(error));
            }else if(response.statusCode >= 400){
              logger.error('Error ' + response.statusCode + ' ' + options.method + ' request ' + options.url);
              res.status(response.statusCode).send(response.body.error);
            } 
          }
        };      
        request(options, callback);
    });
    //delete project env of job
    app.delete('/linkops/projectenvs/:id',  Authentication.ensureAuthenticated, function(req, res) {
        var options = {
            url: linkerConf.linkopsProvider + urlCfg.linkops_api.projectenvs + "/" + req.params.id,
            method: 'DELETE',
            headers: {
             'X-Auth-Token': req.session.token
            }
        };
        var callback = function(error, response, body) {
          if (!error && response.statusCode == 200) {
            res.status(200).send(body);           
          } else {
            if (error) {
              res.status(500).send(handleError(error));
            }else if(response.statusCode >= 400){             
              res.status(response.statusCode).send(response.body.error);
            }
          }
        };
       
        request(options, callback);
    });



};