'use strict';
var client = require('redis-sentinel-client');

var SentinelUtil = function(options) {
  this.sentinelClient = client.createClient(options);
};

SentinelUtil.prototype.getMaster = function(masterName,callback) {
  if (this.sentinelClient != null) {
    this.sentinelClient.getSentinel().send_command("SENTINEL", ["get-master-addr-by-name",masterName],function(error,master){
         callback(master);
    });
  } else {
	    this.client = client.createClient(this.endpoints);
	    this.sentinelClient.getSentinel().send_command("SENTINEL", ["get-master-addr-by-name",masterName],function(error,master){
	         callback(master);
	    });
  }
};
module.exports = SentinelUtil;
