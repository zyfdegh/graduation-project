'use strict';

require('sugar');
var logger = require('../utils/logger');

module.exports.ensureAuthenticated = function (req, res, next) {
    if(!req.session || !req.session.token){
    	var error={"name":"Please Sign In First!","code":"401"};
        res.status(401).send(error);
    }else{
    	return next();
    }
};
