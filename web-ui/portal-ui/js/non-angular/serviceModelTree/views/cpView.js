var cpDialogTemplate = '<div class="modal-dialog" style="width:1200px" onclick="hideAppTree()" id="cp_dialog">'+
							'<div class="modal-content">'+
								'<div class="modal-header" id="cp_header">'+
								'</div>'+
								'<div class="modal-body" style="height:450px;overflow: auto;" id="cp_content">'+
								'</div>'+
								'<div class="modal-footer" id="cp_buttons">'+ 
								'</div>'+
							'</div>'+
						'</div>';

var cpHeaderTemplate = '<h4><%=$.i18n.prop("configuration_package")%> / <%=stepname%></h4>'+
						'<span class="glyphicon glyphicon-remove cp-close-button" data-dismiss="modal" title="<%=$.i18n.prop("close")%>"></span>';
						
var cpButtonsTemplate = '<%if(currentstep == 1){%>'+
							'<button class="btn btn-primary" onclick="wizardControl(1,1)"><%=$.i18n.prop("next")%></button>'+
						'<%}else if(currentstep == 2){%>'+
							'<button class="btn btn-primary" style="float:left" onclick="wizardControl(2,-1)"><%=$.i18n.prop("previous")%></button>'+
							'<button class="btn btn-primary" onclick="wizardControl(2,1)"><%=$.i18n.prop("next")%></button>'+
						'<%}else if(currentstep == 3){%>'+
							'<button class="btn btn-primary" style="float:left" onclick="wizardControl(3,-1)"><%=$.i18n.prop("previous")%></button>'+
							'<button class="btn btn-primary" data-dismiss="modal" onclick="saveCP();"><%=$.i18n.prop("save")%></button>'+
						'<%}%>';
						
									
var cpBasicTemplate = 	'<form class="cp-basic-form" id="cpBasic">'+
							'<div class="cp-basic-form-row">'+
								'<div class="cp-basic-form-label"><%=$.i18n.prop("service_group_id")%></div>'+
								'<div class="cp-basic-form-input">'+
									'<span class="cp-basic-form-content">{{service_group_id}}</span>'+
								'</div>'+
							'</div>'+
							'<div style="width:100%;height:60px;">'+
								'<div class="cp-basic-form-label"><%=$.i18n.prop("app_id")%></div>'+
								'<div class="cp-basic-form-input">'+
									'<span class="cp-basic-form-content">{{app_container_id}}</span>'+
								'</div>'+
							'</div>'+
//							'<div style="width:100%;height:60px;">'+
//								'<div class="cp-basic-form-label"><%=$.i18n.prop("image")%></div>'+
//								'<div class="cp-basic-form-input" style="padding-top:2px;">'+
//									'<input type="text" id="cp_image" v-model="image" placeholder="<%=$.i18n.prop("placeholder_image")%>" class="cp-basic-form-content form-control" required/>'+
//								'</div>'+
//							'</div>'+
						'</form>';
						
var cpConfigsTemplate = '<div style="width:97%;margin:auto;" id="cpConfig">'+
							'<div style="width:20%;float:left" id="cpConfigList">'+
								'<div class="a-apps-title" style="width:100%;margin-top:5px"><%=$.i18n.prop("configurations")%>'+
									'<span class="glyphicon glyphicon-plus" style="float:right;margin-right:10px;font-size:14px;cursor:pointer" title="<%=$.i18n.prop("add_configuration")%>" onclick="newConfig()">'+
								'</span></div>'+
								'<div class="cp-configuration-list">'+
									'<span v-repeat="configurations" class="cp-configuration-item cp-configuration-item-{{$index}}" data-index="{{$index}}" onclick="changeConfig(event)">'+
										'<span class="configname" title="{{name}}">{{name}}</span>'+
			      						'<input class="confignameinput" type="text" v-model="name"/>'+
										'<span class="glyphicon glyphicon-remove" style="float:right;margin-right:10px;font-size:14px;cursor:pointer;" title="<%=$.i18n.prop("remove_config")%>" onclick="removeConfig(event)"></span>'+
										'<span class="glyphicon glyphicon-edit" title="<%=$.i18n.prop("edit_name")%>" style="float:right;margin-right:10px;font-size:14px;cursor:pointer;" onclick="startEditName(event)"></span>'+
				      					'<span class="glyphicon glyphicon-floppy-disk" title="<%=$.i18n.prop("save_name")%>" style="float:right;margin-right:10px;font-size:14px;cursor:pointer;display:none" onclick="finishEditName(event)"></span>'+
			      					'</span>'+
								'</div>'+
							'</div>'+
							'<div style="width:78%;float:right" id="cpConfigDetail">'+	
							'</div>'+
						'</div>';

var cpConfigDetailTemplate = '<div class="a-apps-title" style="width:100%;margin-top:5px" onclick="openConditions()"><%=$.i18n.prop("preconditions")%>'+
									'<span class="glyphicon glyphicon-plus" style="float:right;margin-right:40px;font-size:14px;cursor:pointer" title="<%=$.i18n.prop("add_precondition")%>" onclick="newCondition()">'+
									'</span></div>'+
								'<div class="cp-configuration-conditions" id="precondition_list" style="overflow: auto;">'+
								'</div>'+
								'<div class="a-apps-title" style="width:100%;margin-top:5px;padding-left: 100px" onclick="openSteps()"><%=$.i18n.prop("steps")%>'+
									'<span class="glyphicon glyphicon-plus" style="float:right;margin-right:40px;font-size:14px;cursor:pointer;margin-top:-2px;" title="<%=$.i18n.prop("add_step")%>" onclick="newStep()"></span>'+
									'<span style="float:right;margin-right:40px;font-size:12px;"><%=$.i18n.prop("config_ip")%></span>'+
									'<span style="float:right;margin-right:6px;"><input type="checkbox" style="margin:0px" id="isConfigIP" v-model="hasConfigIPStep"></span>'+
								'</div>'+
								'<div class="cp-configuration-conditions" id="step_list" style="overflow: auto;display:none">'+
								'</div>';
								
var cpConditionItemTemplate = 	'<div style="width:50%;margin:auto;margin-top:80px;" v-if="my_preconditions.length == 0"><%=$.i18n.prop("no_precondition_found")%></div>'+
								'<div style="width:100%;height:45px;border-bottom:1px solid #d5d9dc;" v-repeat="my_preconditions" data-index="{{$index}}">'+
							      		'<div style="width:40%;height:45px;float: left;margin-left:20px;">'+
							      			'<input class="condition_app form-control" type="text" v-model="targetapp" title="{{targetapp}}" style="float:left;width:80%;margin-top:5px">'+
							      			'<span class="glyphicon glyphicon-collapse-down" style="float:left;margin-left:5px;margin-top:13px;font-size:16px;cursor:pointer" title="<%=$.i18n.prop("change_app")%>" onclick="openAppTree(event,\'condition\')"></span>'+
							      		'</div>'+
							      		'<div style="width:30%;height:45px;float: left;margin-left:10px;">'+
							      			'<select class="condition_operator form-control" style="margin-top:5px;width:80%" v-model="operator">'+
							      				'<option value="-eq"><%=$.i18n.prop("eq")%></option>'+
							      				'<option value="-ne"><%=$.i18n.prop("neq")%></option>'+
							      				'<option value="-gt"><%=$.i18n.prop("gt")%></option>'+
							      				'<option value="-lt"><%=$.i18n.prop("lt")%></option>'+
							      				'<option value="-ge"><%=$.i18n.prop("gtoeq")%></option>'+
							      				'<option value="-le"><%=$.i18n.prop("ltoeq")%></option>'+
							      			'</select>'+
							      		'</div>'+
							      		'<div style="width:15%;height:45px;float: left;margin-left:10px;">'+
							      			'<input class="condition_value form-control" type="number" min="1" v-model="value" step="1" style="width:60%;margin-top:5px">'+
							      		'</div>'+
							      		'<div style="width:10%;height:45px;float:left;text-align:center;">'+
							      			'<span class="glyphicon glyphicon-remove" style="margin-top:13px;font-size:16px;cursor:pointer" title="<%=$.i18n.prop("remove_precondition")%>" onclick="removeCondition(event)"></span>'+
							      		'</div>'+
							      	'</div>';

var cpStepItemTemplate = 	'<div style="width:50%;margin:auto;margin-top:80px;" v-if="my_steps.length == 0"><%=$.i18n.prop("no_step_found")%></div>'+
							'<div style="width:100%;height:200px;border-bottom:1px solid #d5d9dc;" v-repeat="my_steps" data-index="{{$index}}">'+
								'<div style="float:left;width:85%">'+
									'<div style="height:45px;width:100%">'+
								      	'<div style="width:47%;float:left">'+
								      		'<div style="float:left;margin-left:20px;margin-top:10px;width:80px"><%=$.i18n.prop("command")%></div>'+
								      		'<select class="form-control" style="margin-top:5px;width:200px;margin-left:20px;float:left" v-model="execute.command">'+
								      			'<option value="/init.sh"><%=$.i18n.prop("init_sh")%></option>'+
								      			'<option value="/config.sh"><%=$.i18n.prop("config_sh")%></option>'+
								      			'<option value="/start.sh"><%=$.i18n.prop("start_sh")%></option>'+
								      			'<option value="/restart.sh"><%=$.i18n.prop("restart_sh")%></option>'+
								      		'</select>'+
								      	'</div>'+
								      	'<div style="width:47%;height:45px;float:right">'+
								      		'<div style="float:left;margin-left:20px;margin-top:10px;width:80px"><%=$.i18n.prop("scope")%></div>'+
								      		'<select class="form-control" style="margin-top:5px;width:200px;margin-left:20px;float:left" v-model="scope">'+
								      			'<option value="ONLYME">ONLYME</option>'+
								      			'<option value="WITHOUTME">WITHOUTME</option>'+
								      			'<option value="ALL">ALL</option>'+
								      		'</select>'+
								      	'</div>'+
							      	'</div>'+
							      	'<div style="width:100%;">'+
							      		'<div style="float:left;margin-left:20px;margin-top:15px;width:80px">'+
							      			'<span><%=$.i18n.prop("parameters")%></span>'+
							      			'<span class="glyphicon glyphicon-plus" style="font-size:14px;cursor:pointer;margin-top:20px;margin-left:30px" title="<%=$.i18n.prop("add_parameter")%>" onclick="newParam(event)"></span>'+
							      		'</div>'+
							      		'<div style="margin-top:5px;width:70%;margin-left:50px;float:left;height:120px;overflow:auto">'+
							      			'<div v-repeat="execute.params" data-index="{{$index}}" style="margin-top:5px;height:45px">'+
							      				'<input class="form-control" type="text" v-model="name" style="width:30%;float:left" placeholder="<%=$.i18n.prop("parameter_name")%>">'+
							      				'<input class="form-control" type="text" v-model="value" style="width:50%;float:left;margin-left:10px" placeholder="<%=$.i18n.prop("parameter_value")%>">'+
							      				'<span class="glyphicon glyphicon-collapse-down" style="margin-top:10px;font-size:14px;cursor:pointer;float:left;margin-left:5px;" title="<%=$.i18n.prop("get_app_properties")%>" onclick="openAppTree(event,\'step\')"></span>'+
							      				'<span class="glyphicon glyphicon-remove" style="margin-top:10px;font-size:14px;cursor:pointer;float:left;margin-left:50px;" title="<%=$.i18n.prop("remove_parameter")%>" onclick="removeParam(event)"></span>'+
							      			'</div>'+
							      			'<div style="width:50%;" v-if="execute.params.length == 0"><%=$.i18n.prop("no_parameters_found")%></div>'+
							      		'</div>'+
							      	'</div>'+
							    '</div>'+
							    '<div style="width:15%;height:100%;float:right;text-align:center">'+
							      	'<span class="glyphicon glyphicon-remove" style="font-size:20px;cursor:pointer;margin-top:80px;" title="<%=$.i18n.prop("remove_step")%>" onclick="removeStep(event)"></span>'+
							  	'</div>'+
							'</div>';
							      	
var cp_app_tree_template = 	'<div style="z-index:1060;top:<%=top%>px;left:<%=left%>px;position: absolute;" id="apptreearea" onclick="event.stopPropagation()">'+
								'<ul id="apptree" class="ztree ui-menu ui-widget ui-widget-content ui-corner-bottom" style="width:<%=width%>px;height:200px;overflow: auto;"></ul>'+
							'</div>';
							
var cpNotifyTemplate =  '<div style="width:85%;margin:auto;margin-top:20px" id="cpNotify">'+
							'<div class="a-apps-title" style="width:100%;"><%=$.i18n.prop("notification_list")%>'+
								'<span class="glyphicon glyphicon-plus" style="float:right;margin-right:40px;font-size:14px;cursor:pointer" title="<%=$.i18n.prop("add_notification")%>" onclick="newNotify()"></span>'+
							'</div>'+
							'<div class="cp-configuration-conditions" id="notification_list" style="overflow: auto;">'+
								'<div style="width:50%;margin:auto;margin-top:100px" v-if="notifies.length == 0"><%=$.i18n.prop("no_notify_found")%></div>'+
								'<div style="width:100%;height:45px;border-bottom:1px solid #d5d9dc;" v-repeat="notifies" data-index="{{$index}}">'+
									'<div style="float:left;width:15%;text-align:center;margin-top:12px"><%=$.i18n.prop("notify_path")%></div>'+
							      	'<div style="width:30%;height:45px;float: left;">'+
							      		'<span class="glyphicon glyphicon-collapse-down" style="float:right;margin-top:13px;font-size:16px;cursor:pointer" title="<%=$.i18n.prop("change_notify")%>" onclick="openAppTree(event,\'notify\')"></span>'+
							      		'<input class="form-control" type="text" v-model="notify_path" title="{{notify_path}}" style="float:right;width:90%;margin-top:5px;margin-right:5px;">'+
							      	'</div>'+
							      	'<div style="float:left;width:10%;text-align:center;margin-top:12px;margin-left:20px"><%=$.i18n.prop("scope")%></div>'+
							      	'<select class="form-control" style="margin-top:5px;width:20%;float:left;" v-model="scope">'+
								      	'<option value="ONLYME">ONLYME</option>'+
								      	'<option value="WITHOUTME">WITHOUTME</option>'+
								      	'<option value="ALL">ALL</option>'+
								    '</select>'+
							      	'<div style="width:15%;height:45px;float:left;">'+
							      		'<span class="glyphicon glyphicon-remove" style="margin-top:13px;font-size:16px;cursor:pointer;margin-right:40px;float:right" title="<%=$.i18n.prop("remove_notify")%>" onclick="removeNotify(event)"></span>'+
							      	'</div>'+
							    '</div>'
							'</div>'+
						'</div>';
