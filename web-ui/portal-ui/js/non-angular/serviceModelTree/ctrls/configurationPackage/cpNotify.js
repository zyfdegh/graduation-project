function showCPNotify(){
	var self = this;
	
	var compiledTemplate = _.template(cpNotifyTemplate);
	$("#cp_content").empty();
	$("#cp_content").append(compiledTemplate);
	
	this.notify_list_vm = new Vue({
	  el: '#cpNotify',
	  data: self.openedCP
	});
	
	if(self.openedCP.notifies.length>0){
		$("#notification_list").sortable({
	        tolerance: 'pointer',
	        revert: 'invalid',
	        forceHelperSize: true,
	        stop : function(event,ui){
	        	var totalNum = $(event.target).children().length;
	        	var new_my_notifies = [];
	        	for(var i=0;i<totalNum;i++){
	        		var index = $(event.target).children().eq(i).data("index");
	        		new_my_notifies.push(self.notify_list_vm.notifies[index]); 
	        	}
	        	self.notify_list_vm.notifies = new_my_notifies;
				self.showCPNotify(self.notify_list_vm.$data);
	        }
	    });
	}
 
	self.showCPHeader();
	self.showCPButtons();
}

function newNotify(){
	var notify = {
		"notify_path" : "",
		"scope" : ""
	};
	this.openedCP.notifies.push(notify);
	this.showCPNotify();
}

function removeNotify(event){
	var index = $(event.currentTarget).parent().parent().data("index");
	this.openedCP.notifies.splice(index,1);
	this.showCPNotify();
}