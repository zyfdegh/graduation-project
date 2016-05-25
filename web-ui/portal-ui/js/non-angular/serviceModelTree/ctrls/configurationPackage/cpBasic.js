function openCPDialog(){
	var self = this;
	var compiledTemplate = _.template(cpDialogTemplate);
	$("#linker-alert").empty();
	$("#linker-alert").append(compiledTemplate);
	$("#linker-alert").modal("show");
	
	self.cp_current_step = 1;
	self.showCPSteps();
}

function showCPSteps(){
	var self = this;
	
	switch(self.cp_current_step){
		case 1 : self.showCPBasic(); break;
		case 2 : self.showCPConfigs(); break;
		case 3 : self.showCPNotify(); break;
	}
}

function showCPBasic(){
	var self = this;
	
	var compiledTemplate = _.template(cpBasicTemplate);
	$("#cp_content").empty();
	$("#cp_content").append(compiledTemplate);
	
	new Vue({
	  el: '#cpBasic',
	  data: self.openedCP
	});
	
	self.showCPHeader();
	self.showCPButtons();
}