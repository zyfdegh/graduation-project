var express = require('express');
var request = require('request');
var bodyParser = require('body-parser');
var path = require('path');

var app = express();

//designer ui
var staticPath=path.resolve(path.dirname(__dirname), 'dest');
app.use('/portal-ui', express.static(staticPath));
// app.use('/linker', express.static('linker_bk'));
//parse post payload
app.use(bodyParser.json()); // for parsing application/json
app.use(bodyParser.urlencoded({
	extended: true
})); // for parsing application/x-www-form-urlencoded

//cross domain
app.use(function(req, res, next) {
	res.header('Access-Control-Allow-Credentials', true);
	res.header('Access-Control-Allow-Origin', req.headers.origin);
	res.header('Access-Control-Allow-Methods', 'GET,PUT,POST,DELETE');
	res.header('Access-Control-Allow-Headers', 'X-Requested-With, X-HTTP-Method-Override, Content-Type, Accept');
	next();
});

//conf
var config = require('konphyg')(path.resolve(__dirname, 'conf'));
var serviceConf = config('service');

//api request redirect
//app apis
app.get('/apps', function(req, res) {
	var options = {
		url: serviceConf.serverUrl + serviceConf.app_service,
		method: 'GET'
	};
	var callback = function(error, response, body) {
		//console.log(response.statusCode)
		if (!error && response.statusCode == 200) {
			res.status(200).send(body);
		} else {
			res.status(response.statusCode).send(error);
		}
	};
	request(options, callback);
});

app.post('/apps', function(req, res) {
	var options = {
		url: serviceConf.serverUrl + serviceConf.app_service,
		method: 'POST',
		json: true,
		body: JSON.stringify(req.body)
	};
	var callback = function(error, response, body) {
		//console.log(response.statusCode)
		if (!error && response.statusCode == 201) {
			res.status(200).send(body);
		} else {
			res.status(response.statusCode).send(error);
		}
	};
	request(options, callback);
});

app.put('/apps/:appid', function(req, res) {
	var options = {
		url: serviceConf.serverUrl + serviceConf.app_service + req.params.appid,
		method: 'PUT',
		json: true,
		body: JSON.stringify(req.body)
	};
	var callback = function(error, response, body) {
		//console.log(response.statusCode)
		if (!error && response.statusCode == 200) {
			res.status(200).send(body);
		} else {
			res.status(response.statusCode).send(error);
		}
	};
	request(options, callback);
});

app.delete('/apps/:appid', function(req, res) {
	var options = {
		url: serviceConf.serverUrl + serviceConf.app_service + req.params.appid,
		method: 'DELETE'
	};
	var callback = function(error, response, body) {
		//console.log(response.statusCode)
		if (!error && response.statusCode == 200) {
			res.status(200).send(body);
		} else {
			res.status(response.statusCode).send(error);
		}
	};
	request(options, callback);
});
//app apis end

//servicegroup apis
app.get('/serviceGroups', function(req, res) {
	var options = {
		url: serviceConf.serverUrl + serviceConf.serviceGroup_service,
		method: 'GET'
	};
	var callback = function(error, response, body) {
		//console.log(response.statusCode)
		if (!error && response.statusCode == 200) {
			res.status(200).send(body);
		} else {
			res.status(response.statusCode).send(error);
		}
	};
	request(options, callback);
});

app.post('/serviceGroups', function(req, res) {
	var options = {
		url: serviceConf.serverUrl + serviceConf.serviceGroup_service,
		method: 'POST',
		json: true,
		body: JSON.stringify(req.body)
	};
	var callback = function(error, response, body) {
		//console.log(response.statusCode)
		if (!error && response.statusCode == 201) {
			res.status(200).send(body);
		} else {
			res.status(response.statusCode).send(error);
		}
	};
	request(options, callback);
});

app.put('/serviceGroups/:sgid', function(req, res) {
	var options = {
		url: serviceConf.serverUrl + serviceConf.serviceGroup_service + req.params.sgid,
		method: 'PUT',
		json: true,
		body: JSON.stringify(req.body)
	};
	var callback = function(error, response, body) {
		//console.log(response.statusCode)
		if (!error && response.statusCode == 200) {
			res.status(200).send(body);
		} else {
			res.status(response.statusCode).send(error);
		}
	};
	request(options, callback);
});

app.delete('/serviceGroups/:sgid', function(req, res) {
	var options = {
		url: serviceConf.serverUrl + serviceConf.serviceGroup_service + req.params.sgid,
		method: 'DELETE'
	};
	var callback = function(error, response, body) {
		//console.log(response.statusCode)
		if (!error && response.statusCode == 200) {
			res.status(200).send(body);
		} else {
			res.status(response.statusCode).send(error);
		}
	};
	request(options, callback);
});
//servicegroup apis end

//order service apis
app.post('/serviceGroupOrders', function(req, res) {
	var options = {
		url: serviceConf.serverUrl + serviceConf.serviceGroupOrders_service + "?query=" + req.query.query,
		method: 'POST'
	};
	var callback = function(error, response, body) {
		//console.log(response.statusCode)
		if (!error && (response.statusCode == 201 || response.statusCode == 200)) {
			res.status(200).send(body);
		} else {
			res.status(response.statusCode).send(error);
		}
	};
	request(options, callback);
});
//order service apis end

//list instance apis
app.get('/groupInstances', function(req, res) {
	var options = {
		url: serviceConf.serverUrl + serviceConf.instance_service,
		method: 'GET'
	};
	var callback = function(error, response, body) {
		//console.log(response.statusCode)
		if (!error && response.statusCode == 200) {
			res.status(200).send(body);
		} else {
			res.status(response.statusCode).send(error);
		}
	};
	request(options, callback);
});
//list instance apis end

app.get('/groupInstances/:sgid', function(req, res) {
	var options = {
		url: serviceConf.serverUrl + serviceConf.instance_service + req.params.sgid,
		method: 'GET'
	};
	var callback = function(error, response, body) {
		// console.log(body)
		if (!error && response.statusCode == 200) {
			res.status(200).send(body);
		} else {
			res.status(response.statusCode).send(error);
		}
	};
	request(options, callback);
});

app.put('/groupInstances/scaleApp/:sgid', function(req, res) {
	var options = {
		url: serviceConf.serverUrl + serviceConf.instance_service + "scaleApp/" + req.params.sgid + "?appId=" + req.query.appId + "&num=" + req.query.num,
		// params: {
		// 	"appId": req.query.appId,
		// 	"num": req.query.num
		// },
		method: 'PUT'
	};
	console.log(options.url);
	console.log(options.params);
	var callback = function(error, response, body) {
		//console.log(response.statusCode)
		if (!error && response.statusCode == 200) {
			res.status(200).send(body);
		} else {
			res.status(response.statusCode).send(error);
		}
	};
	request(options, callback);
});

app.delete('/groupInstances/:sgid', function(req, res) {
	var options = {
		url: serviceConf.serverUrl + serviceConf.serviceGroupOrders_service + req.params.sgid,
		method: 'DELETE'
	};
	var callback = function(error, response, body) {
		//console.log(response.statusCode)
		if (!error && response.statusCode == 200) {
			res.status(200).send(body);
		} else {
			res.status(response.statusCode).send(error);
		}
	};
	request(options, callback);
});

//list app instance  apis
app.get('/appInstances/:instanceid', function(req, res) {
	var options = {
		url: serviceConf.serverUrl + serviceConf.app_instance_service + req.params.instanceid,
		method: 'GET'
	};
	var callback = function(error, response, body) {
		//console.log(response.statusCode)
		if (!error && response.statusCode == 200) {
			res.status(200).send(body);
		} else {
			res.status(response.statusCode).send(error);
		}
	};
	request(options, callback);
});
//list instance apis end

app.listen(3000, function() {
	//console.log("linker model designer started, listen port 3000.")
});