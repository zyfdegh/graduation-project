function showCPConfigs(){
	var self = this;
	
	var compiledTemplate = _.template(cpConfigsTemplate);
	$("#cp_content").empty();
	$("#cp_content").append(compiledTemplate);
	
	if(self.openedCP.configurations.length==0){
		$("#cpConfigList").css("width","100%");
		$("#cpConfigDetail").css("width","0%");
		var compiledTemplate = _.template('<div style="width:50%;margin:auto;margin-top:130px"><%=$.i18n.prop("no_cp_found")%></div>');
		$(".cp-configuration-list").empty();
		$(".cp-configuration-list").append(compiledTemplate);
	}
	
	new Vue({
	  el: '#cpConfig',
	  data: self.openedCP
	});
	
	if(self.openedCP.configurations.length>0){
		$("#cpConfigList").css("width","20%");
		$("#cpConfigDetail").css("width","78%");
		$(".cp-configuration-item-0").addClass("cp-configuration-item-active");
		var selectedConfig = self.openedCP.configurations[0];
		self.showConfigDetail(selectedConfig);
	}
	
	self.showCPHeader();
	self.showCPButtons();
}

function changeConfig(event){
	$(".cp-configuration-item-active").removeClass("cp-configuration-item-active");
	var index = $(event.currentTarget).data("index");
	$(".cp-configuration-item-"+index).addClass("cp-configuration-item-active");
	var selectedConfig = this.openedCP.configurations[index];
	this.showConfigDetail(selectedConfig);
}

function removeConfig(event){
	event.stopPropagation();
	var index = $(event.currentTarget).parent().data("index");
	this.openedCP.configurations.splice(index,1);
	this.showCPConfigs();
}

function newConfig(){
	var config = {
		"name" : "newConfig",
		"preconditions" : [],
		"steps" : []
	};
	this.openedCP.configurations.push(config);
	this.showCPConfigs();
}

//edit config name
function startEditName(event){
	var self = this;
	var editbutton = $(event.currentTarget);
	var configname_span = editbutton.parent().find(".configname");
	var configname_input = editbutton.parent().find(".confignameinput");
	var savebutton = editbutton.parent().find(".glyphicon-floppy-disk");

	configname_span.hide();
	configname_input.show();
	editbutton.hide();
	savebutton.show();
}

function finishEditName(event){
	var self = this;
	var savebutton = $(event.currentTarget);
	var configname_span = savebutton.parent().find(".configname");
	var configname_input = savebutton.parent().find(".confignameinput");
	var editbutton = savebutton.parent().find(".glyphicon-edit");

	configname_input.hide();
	configname_span.show();
	savebutton.hide();
	editbutton.show();
}
//edit config name

function showConfigDetail(selectedConfig){
	var compiledTemplate = _.template(cpConfigDetailTemplate);
	$("#cpConfigDetail").empty();
	$("#cpConfigDetail").append(compiledTemplate);
	
	if(selectedConfig.hasConfigIPStep == null){
		selectedConfig.hasConfigIPStep = false;
	}
	
	new Vue({
	  el: '#cpConfigDetail',
	  data: selectedConfig
	});

	//show conditions
	this.showConditionsDetail(selectedConfig);
	
	//show steps
	this.showStepsDetail(selectedConfig);
}

function showConditionsDetail(selectedConfig){
	var preconditions = selectedConfig.preconditions;
	if(selectedConfig.my_preconditions == null){
		selectedConfig.my_preconditions = [];
		_.each(preconditions,function(precondition){
			var targetapp = precondition.condition.split(" ")[0];
			var operator = precondition.condition.split(" ")[1];
			var value = precondition.condition.split(" ")[2];
			var data = {
				"targetapp" : targetapp,
				"operator" : operator,
				"value" : value
			}
			selectedConfig.my_preconditions.push(data);
		});
	}
		
	$("#precondition_list").empty();
	var configTemplate = _.template(cpConditionItemTemplate);
	$("#precondition_list").append(configTemplate);
		
	this.precondition_list_vm = new Vue({
	  el: '#precondition_list',
	  data: selectedConfig
	});
}

function showStepsDetail(selectedConfig){
	var self = this;
	var steps = selectedConfig.steps;
	if(selectedConfig.my_steps == null){
		selectedConfig.my_steps = [];
		_.each(steps,function(step){
			var config_type = step.config_type;
			var execute = step.execute;
			var scope = step.scope;
			
			if(config_type == "command"){
				selectedConfig.hasConfigIPStep = true;
				$("#isConfigIP").prop("checked",true);
				$("#isConfigIP").trigger("change");
			}else{
				var executes = execute.split(" ");
				var command = executes[0];
				var params = [];
				for(var i=1;i<executes.length;i=i+2){
					var param = {
						"name" : executes[i].substring(2),
						"value" : executes[i+1]
					};
					params.push(param);
				}
				
				var my_step = {
					"config_type" : "docker",
					"execute" : {
						"command" : command,
						"params" : params
					},
					"scope" : scope
				}
				
				selectedConfig.my_steps.push(my_step);
			}		
		});
	}
		
	$("#step_list").empty();
	var stepTemplate = _.template(cpStepItemTemplate);
	$("#step_list").append(stepTemplate);
		
	this.step_list_vm = new Vue({
	  el: '#step_list',
	  data: selectedConfig
	});
	
	if(selectedConfig.my_steps.length>0){
		$("#step_list").sortable({
	        tolerance: 'pointer',
	        revert: 'invalid',
	        forceHelperSize: true,
	        stop : function(event,ui){
	        	var totalStepNum = $(event.target).children().length;
	        	var new_my_steps = [];
	        	for(var i=0;i<totalStepNum;i++){
	        		var index = $(event.target).children().eq(i).data("index");
	        		new_my_steps.push(self.step_list_vm.my_steps[index]); 
	        	}
	        	self.step_list_vm.my_steps = new_my_steps;
				self.showStepsDetail(self.step_list_vm.$data);
	        }
	    });
	}
}

//control panel open
function openConditions(){
	$("#precondition_list").show();
	$("#step_list").hide();
}

function openSteps(){
	$("#precondition_list").hide();
	$("#step_list").show();
}
//control end

//precondition 
function newCondition(){
	var data = {
		"targetapp" : "",
		"operator" : "",
		"value" : ""
	};
	this.precondition_list_vm.my_preconditions.push(data);
	this.showConditionsDetail(this.precondition_list_vm.$data);
}

function removeCondition(event){
	var index = $(event.currentTarget).parent().parent().data("index");
	this.precondition_list_vm.my_preconditions.splice(index,1);
	this.showConditionsDetail(this.precondition_list_vm.$data);
}
//precondition end

//steps
function newParam(event){
	var param = {
		"name" : "",
		"value" : ""
	};
	var index = $(event.currentTarget).parent().parent().parent().parent().data("index");
	var params = this.step_list_vm.my_steps[index].execute.params;
	params.push(param);
}

function removeParam(event){
	var stepindex = $(event.currentTarget).parent().parent().parent().parent().parent().data("index");
	var paramindex = $(event.currentTarget).parent().data("index");
	var params = this.step_list_vm.my_steps[stepindex].execute.params;
	params.splice(paramindex,1);
}

function newStep(){
	var my_step = {
			"config_type" : "docker",
			"execute" : {
				"command" : "",
				"params" : []
			},
			"scope" : ""
		}
	this.step_list_vm.my_steps.push(my_step);
	this.showStepsDetail(this.step_list_vm.$data);
}

function removeStep(event){
	var index = $(event.currentTarget).parent().parent().data("index");
	this.step_list_vm.my_steps.splice(index,1);
	this.showStepsDetail(this.step_list_vm.$data);
}
//steps end