var winston = require('winston'),
  	path = require('path');
require('sugar');
var LogConf = (global.obj) ? global.obj.linkerPortalCfg : {
  	logging: {
    	console: {
      		enabled: false
    	},
   	 	file: {
      		enabled: false
    	}
  	}
};

var LogLevels = {
  	levels: {
    	error: 4,
    	warn: 3,
    	info: 2,
    	debug: 1,
   	 	trace: 0
  	},
  	colors: {
    	error: 'red',
    	warn: 'orange',
    	info: 'green',
    	debug: 'blue',
    	trace: 'yellow'
  	}
};

var transports = [];
if (LogConf.logging.console.enabled) {
  	transports.push(new(winston.transports.Console)({
    	level: LogConf.logging.console.level ? LogConf.logging.console.level.toLowerCase() : 'info',
    	json: true,
    	handleExceptions: true,
    	colorize: true,
    	timestamp: function () {
      		return Date.create().format('{dd} {Mon} {yyyy} {hh}:{mm}:{ss},{fff}');
    	}
  	}));
}
if (LogConf.logging.file.enabled) {
  	var logFileName = path.resolve('logs', 'linker-server.log');
  	transports.push(new(winston.transports.File)({
    	level: LogConf.logging.file.level ? LogConf.logging.file.level.toLowerCase() : 'info',
   	 	filename: logFileName,
    	handleExceptions: true,
    	maxsize: LogConf.logging.file.maxSizeMB * 1024 * 1024 || 10485760, //10M
    	maxFile: LogConf.logging.file.maxFile || 10,
    	timestamp: function () {
      		return Date.create().format('{dd} {Mon} {yyyy} {hh}:{mm}:{ss},{fff}');
    	},
    	json: true
  	}));
}
var logger = new(winston.Logger)({
  levels: LogLevels.levels,
  transports: transports,
  exitOnError: false
});
winston.addColors(LogLevels.colors);
module.exports = logger;
