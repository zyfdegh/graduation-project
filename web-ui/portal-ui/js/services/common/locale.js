function langService(){
    
	return {
		getSupportedLangs : function(){
			var supportedLangs =  [{
          			"name" :  "en",
          			"display" : "English"        
		    },{
		          "name" :  "zh",
		          "display" : "中文" 
		    }];
			return supportedLangs;
		} 
	}
};
   
linkerCloud.factory('langService', [langService]);