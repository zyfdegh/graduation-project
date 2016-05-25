global.obj = {};

var express = require('express');
var cookieParser = require('cookie-parser');
var session = require('express-session');
var bodyParser = require('body-parser');
var path = require('path');
var fs = require('fs');
// var multer  = require('multer');
// var upload = multer();
var multipart = require('connect-multiparty');
var multipartMiddleware = global.obj.multipartMiddleware = multipart();


var portalConfig = global.obj.config = require('konphyg')(path.resolve(path.dirname(__dirname), 'portal-server/conf'));
var linkerPortalCfg = global.obj.linkerPortalCfg= portalConfig('linker-portal');
var logger = global.obj.logger = require('./utils/logger');
global.obj.urlCfg = portalConfig('url');


if(linkerPortalCfg.controllerProvider.ha.enabled){    
	var ZkUtil = require('./utils/zkUtil');
	global.obj.zkUtil_controller = new ZkUtil(linkerPortalCfg.controllerProvider.ha.zookeeper_url);
};
// if(linkerPortalCfg.controllerProvider.ha.enabled){
// 	global.obj.zkUtil_cicd = new ZkUtil(linkerPortalCfg.cicdProvider.ha.zookeeper_url);
// }


var sessionStore;
var app = express();
app.use(cookieParser());

var staticPath=path.resolve(path.dirname(__dirname), 'portal-ui');
app.use('/portal-ui', express.static(staticPath));
//parse post payload
app.use(bodyParser.json()); // for parsing application/json
app.use(bodyParser.urlencoded({
	extended: true
})); // for parsing application/x-www-form-urlencoded
// app.use(bodyParser()); 
//cross domain
app.use(function(req, res, next) {
	res.header('Access-Control-Allow-Credentials', true);
	res.header('Access-Control-Allow-Origin', req.headers.origin);
	res.header('Access-Control-Allow-Methods', 'GET,PUT,POST,DELETE');
	res.header('Access-Control-Allow-Headers', 'X-Requested-With, X-HTTP-Method-Override, Content-Type, Accept');
	next();
});

if (linkerPortalCfg.ha.enabled) {
  /* use one redis as session storage */
  console.log('HA is enabled, setting up Redis for session store');
      var RedisStore = require('connect-redis')(session);
      sessionStore = new RedisStore(linkerPortalCfg.ha.redis.options);
 
  /* TODO for using sentinel
  var SentinelUtil = require('./utils/redis-sentinel-util');
  var sentinelUtil = new SentinelUtil(linkerPortalCfg.ha.options);
  logger.info('HA is enabled, setting up Redis for session store');
  var RedisStore = require('connect-redis')(session);

   sentinelUtil.getMaster(linkerPortalCfg.ha.options.masterName,function(master){	
	    console.log('Sentinel master store configuration', {
			        "host": master[0],
			        "port": master[1]
			      }); 
		sessionStore = new RedisStore({
			        "host": master[0],
			        "port": master[1]
			      });
		app.use(session({
			  key: 'express.sid',
			  store: sessionStore,
			  secret: "aaa",
			  resave: false,
			  saveUninitialized: true,
			  cookie: { secure: false }
		}));
		app.listen(linkerPortalCfg.port, function() {
			logger.info("linker portal started, listen port " + linkerPortalCfg.port + ".")
	    	console.log("linker portal started, listen port " + linkerPortalCfg.port + ".")
		});  
   });
   */
}else{
	logger.info('Without HA, using MemoryStore');
	var MemoryStore = session.MemoryStore;
	  sessionStore = new MemoryStore();
}
app.use(session({
	  key: 'express.sid',
	  store: sessionStore,
	  secret: "aaa",
	  resave: false,
	  saveUninitialized: false
	  // cookie: { secure: true }
})); 
// dynamically add all the API routes
fs.readdirSync(path.join(__dirname, 'routes')).forEach(function (file) {
  if (file[0] === '.') {
    return;
  }

  require(path.join(__dirname,  'routes', file))(app);

});
 app.listen(linkerPortalCfg.port, function() {		
    	if(linkerPortalCfg.controllerProvider.ha.enabled){
    		global.obj.zkUtil_controller.connect();
    	}else{
    		logger.info("linker portal started, listen port " + linkerPortalCfg.port + ".")
    	    console.log("linker portal started, listen port " + linkerPortalCfg.port + ".")
    	}
});






  
