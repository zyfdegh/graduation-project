'use strict';
require('sugar');
var zookeeper = require('node-zookeeper-client');

var linkerPortalCfg = global.obj.linkerPortalCfg;

var rootPath = "/controller";

var ZkUtil = function(zookeeper_url) {
      this.url = zookeeper_url;
      this.client = zookeeper.createClient(this.url);
      this.controllerEndpoints = [];  
};

ZkUtil.prototype.getClient = function() {
  if (this.client != null) {
    return this.client;
  } else {
    this.client = zookeeper.createClient(this.url);
    return this.client;
  }
};

ZkUtil.prototype.connect = function() {
  var self = this;
  if (this.client != null) {
    this.client.once('connected', function() {
      self.watchController();
    }).connect();
  }
};

ZkUtil.prototype.closeConnection = function() {
  if (this.client != null) {
    this.client.close();
  }
};

ZkUtil.prototype.setControllerEndpoints = function(children) {
  var self = this;
  global.obj.controller_urls = [];
  var childrenLen = children.length;
  if (childrenLen > 0) {
    children.forEach(function(child) {
      self.client.getData(rootPath + '/' + child, function(error, data, stat) {
        if (error) {
          console.log(error.stack);
          return;
        }
        global.obj.controller_urls.push(data.toString('utf8'));
      });
    });
  } else {
    console.log('Can not connect to controller from zookeeper');
    return;
  }
};

ZkUtil.prototype.getControllerUrl = function(req, res, next) {
      var childrenLen = global.obj.controller_urls.length;
      global.obj.controller_url = global.obj.controller_urls[Math.floor(Math.random() * childrenLen)];   
  // console.log('Got data: %s', global.app.controller_url);
  // next();
};

ZkUtil.prototype.watchController = function() {
  var self = this;
  this.client.getChildren(rootPath, function(event) {
    console.log('Got event: %s.', event);
    self.watchController()
  }, function(error, children, stats) {
    if (error) {
      console.log(error.stack);
      return;
    }
    console.log('children are: %j.', children);
    self.setControllerEndpoints(children);
  });
};

module.exports = ZkUtil