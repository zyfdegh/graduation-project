var groupItemTemplate = '<div style="text-align: center;">'+
							'<div id="group_<%=node.id%>" class="group" ondrop="dropHere(event)" ondragover="allowDrop(event)" data-groupid="<%=node.id%>" data-grouptype="normalgroup">'+
							'<%_.each(node.data.apps,function(app){%>'+
								'<%var simappid = app.id;%>'+
								'<%var appid = node.id + "/" + simappid;%>'+
								'<%var instances = app.instances;%>'+
								'<%if(allow_update){%>'+
								'<div class="group-app-item" data-appid="<%=appid%>" data-toggle="context" oncontextmenu="generateAppContext(event)" draggable="true" ondragstart="startDeleteApp(event)">'+
								'<%}else{%>'+
								'<div class="group-app-item" data-appid="<%=appid%>">'+
								'<%}%>'+
									'<div class="a-app-image-container"><img src="<%=app.imageSrc%>" class="container-image" draggable="false"/></div>'+
									'<div class="a-app-item-id"><div class="a-app-item-id-label"><%=simappid%><span class="superscript"><%=instances%></span></div></div>'+
								'</div>'+
							'<%})%>'+
							'</div>'+
							'<div style="font-size:16px;margin-top:10px;cursor:auto">'+
								'<%if(allow_update){%>'+
								'<span data-toggle="context" data-groupid="<%=node.id%>" data-grouptype="normalgroup" oncontextmenu="generateContext(event)"><%=node.data.id%></span>'+
								'<%}else{%>'+
								'<span data-groupid="<%=node.id%>" data-grouptype="normalgroup"><%=node.data.id%></span>'+
								'<%}%>'+
							'</div>'+
						'</div>';

var serviceGroupTemplate = 	'<%if(allow_update){%>'+
							'<div class="service-group" data-toggle="context" data-groupid="<%=node.id%>" data-grouptype="template" ondrop="dropHere(event)" ondragover="allowDrop(event)" oncontextmenu="generateContext(event)">'+
							'<%}else{%>'+
							'<div class="service-group" data-groupid="<%=node.id%>" data-grouptype="template">'+
							'<%}%>'+	
								'<div class="service-group-image-container"><img src="<%=node.data.imageSrc%>" class="container-image" draggable="false"/></div>'+
								'<div class="a-app-item-id"><div class="a-app-item-id-label"><%=node.data.id%></div></div>'+
							'</div>';

var inputDepIDTemplate = '<div>'+
							'<input type="text" id="depid" style="height:25px;width:200px;font-size:14px;" data-pgid="<%=parentgroup.pgid%>" data-pgtype="<%=parentgroup.pgtype%>" onkeydown="addDependency(event)" onfocusout="closeSecondDialog()"/>'+
						'</div>';

var inputNewGroupIDTemplate = '<div>'+
								'<input type="text" id="newid" style="height:25px;width:200px;font-size:14px;" value="<%=s_gid%>" data-gid="<%=gid%>" data-gtype="<%=gtype%>" onkeydown="renameGroupID(event)" onfocusout="closeSecondDialog()"/>'+
							'</div>';

var inputNewTemplateIDTemplate = '<div>'+
									'<input type="text" id="newid" style="height:25px;width:200px;font-size:14px;" value="<%=idSuffix(s_gid)%>" data-gid="<%=gid%>" data-gtype="<%=gtype%>" onkeydown="renameGroupID(event)" onfocusout="closeSecondDialog()"/>'+
								'</div>';
						
var appDetailTemplate = '<form style="padding-top:20px" id="appdetailform" novalidate="novalidate">'+
							'<div style="font-size:28px;font-weight: 500;margin-bottom: 20px;"><%=$.i18n.prop("app_details")%></div>'+
							'<hr />'+
							'<div style="width:700px;height:50px;margin-top:20px">'+
								'<div style="float: left;height:30px;padding-top:10px;width:150px;text-align:center"><%=$.i18n.prop("app_id")%></div>'+
								'<div style="float:right;width:550px;padding-top:10px;">'+
									'<span style="height:30px">{{openedApp.id}}</span>'+
								'</div>'+
							'</div>'+
							'<div style="width:700px;height:60px">'+
								'<div style="float: left;height:30px;padding-top:10px;width:150px;text-align:center"><%=$.i18n.prop("cpus")%></div>'+
								'<div style="float:right;width:550px;">'+
									'<input type="number" id="cpus" min="0.1" step="0.1" v-model="openedApp.cpus" style="height:30px;width:550px" required/>'+
								'</div>'+
							'</div>'+
							'<div style="width:700px;height:60px">'+
								'<div style="float: left;height:30px;padding-top:10px;width:150px;text-align:center"><%=$.i18n.prop("memory")%></div>'+
								'<div style="float:right;width:550px;">'+
									'<input type="number" id="memory" min="1" step="1" v-model="openedApp.mem" style="height:30px;width:550px" required/>'+
								'</div>'+
							'</div>'+
							'<div style="width:700px;height:60px">'+
								'<div style="float: left;height:30px;padding-top:10px;width:150px;text-align:center"><%=$.i18n.prop("instances")%></div>'+
								'<div style="float:right;width:550px;">'+
									'<input type="number" id="instances" min="1" step="1" v-model="openedApp.instances" style="height:30px;width:550px" required/>'+
								'</div>'+
							'</div>'+
//							'<div style="width:700px;height:100px">'+
//								'<div style="float: left;height:30px;padding-top:25px;width:150px;text-align:center"><%=$.i18n.prop("command")%></div>'+
//								'<div style="float:right;width:550px;">'+
//									'<textarea placeholder="<%=$.i18n.prop("placeholder_command")%>" id="cmd" style="width:550px;height:80px;resize: none;" v-model="openedApp.cmd"></textarea>'+
//								'</div>'+
//							'</div>'+


							'<div style="width:700px;height:120px">'+
								'<div style="float: left;height:30px;padding-top:10px;width:150px;text-align:center"><%=$.i18n.prop("docker_image")%></div>'+
                                 
                                 '<label class="radio-inline">'+
		                            '<input type="radio" name="picked" v-model="radio.repoType" value="linker"/><%=$.i18n.prop("linker_images")%>' +                       
		                          '</label>' +
		                          '<label class="radio-inline">'+		                            
		                            '<input type="radio" name="picked" v-model="radio.repoType" value="dockerhub" /><%=$.i18n.prop("dockerhub_images")%>'+	       
		                          '</label>' +
		                          '<div class="form-group" style="margin-top:10px;float:left;margin-left:150px;" v-show="radio.repoType==\'linker\'">' + 
			                      	 '<label style="font-weight:bold;"><%=$.i18n.prop("select_an_image")%> </label>'+
				                 	 '<select id="linkerImageList" class="form-control" v-model="dockerImage.fromLinker" style="width:500px;" options="dockerImages" v-on="change:getImageTags"></select>'+
		                             '<label style="font-weight:bold;margin-top:10px;"><%=$.i18n.prop("select_an_image_tag")%> </label>'+
				                 	 '<select id="tagList" class="form-control" v-model="imageTag.tag" style="width:500px;" options="imageTags" ></select>'+
		                          	 '<p style="margin-top:7px;font-style:italic;color:red;"><%=$.i18n.prop("note_linker_image")%></p>'+
		                          '</div>'+
		                          '<div class="form-group" style="margin-top:10px;float:left;margin-left:150px;" v-show="radio.repoType==\'dockerhub\'">' +
			                           '<label style="font-weight:bold;"><%=$.i18n.prop("input_an_image")%></label>'+		         
				                       '<input type="text" id="docker_image" v-model="dockerImage.fromDockerhub" class="form-control" required placeholder="<%=$.i18n.prop("placeholder_dockerhub_image")%>" style="width:500px;"/>' +
				                       '<p style="margin-top:7px;font-style:italic;color:red;"><%=$.i18n.prop("note_dockerhub_image_1")%><a href="https://hub.docker.com" target="_blank"><%=$.i18n.prop("note_dockerhub_image_2")%></a><%=$.i18n.prop("note_dockerhub_image_3")%></p>' + 
		                          '</div>'+


								// '<div style="float:right;width:550px;">'+
								// 	'<input type="text" id="docker_image" v-model="container.docker.image" style="height:30px;width:550px" required/>'+
								// '</div>'+
							'</div>'+
							'<div style="width:700px;height:120px">'+
								'<div style="float:left;height:120px;padding-top:35px;width:150px;text-align:center"><%=$.i18n.prop("docker_parameters")%>'+
									'<span class="glyphicon glyphicon-plus" style="margin-left: 10px;cursor:pointer" title="<%=$.i18n.prop("add_parameter")%>" onclick="newContainerParameter()"/>'+
								'</div>'+
								'<div style="height:100px;overflow: auto;width:550px;float:right;" v-if="openedApp.container.docker.parameters.length>0">'+
									'<div style="margin-left:10px;width:520px;height: 30px;" v-repeat="openedApp.container.docker.parameters" data-index="{{$index}}">'+
										'<input type="checkbox" v-model="editable" title="<%=$.i18n.prop("make_parameter_editable")%>" style="float:left;margin-top:3px"/>'+
										'<div style="width:140px;float:left;margin-left:15px;height:30px">'+
											'<input type="text" style="width:100%" id="docker_par_name_{{$index}}" v-model="vKey" placeholder="<%=$.i18n.prop("parameter_name")%>" required/>'+
										'</div>'+
										'<div style="width:140px;float:left;margin-left:15px;height:30px">'+
											'<input type="text" style="width:100%" id="docker_par_value_{{$index}}" v-model="vValue" placeholder="<%=$.i18n.prop("parameter_value")%>" required/>'+
										'</div>'+
										'<div style="width:100px;float:left;margin-left:15px;height:30px">'+
											'<input type="text" style="width:100%" id="docker_par_desc_{{$index}}" v-model="description" placeholder="<%=$.i18n.prop("parameter_desc")%>"/>'+
										'</div>'+
										'<span class="glyphicon glyphicon-remove" style="float:left;cursor:pointer;margin-left:20px;margin-top:3px" title="<%=$.i18n.prop("remove_parameter")%>" onclick="removeContainerParameter(event)"/>'+
									'</div>'+
								'</div>'+
								'<div style="width:60%;margin:auto;text-align: center;font-size:16px;height:30px;padding-top:50px;" v-if="openedApp.container.docker.parameters.length==0">'+
									'<%=$.i18n.prop("no_parameters")%>'+
								'</div>'+
							'</div>'+
							'<div style="width:700px;height:120px;float:left">'+
								'<div style="float:left;height:120px;padding-top:35px;width:150px;text-align:center"><%=$.i18n.prop("container_volume")%>'+
									'<span class="glyphicon glyphicon-plus" style="margin-left: 10px;cursor:pointer" title="<%=$.i18n.prop("add_volume")%>" onclick="newContainerVolume()"/>'+
								'</div>'+
								'<div style="height:100px;overflow: auto;width:550px;float:right;" v-if="openedApp.container.volumes.length>0">'+
									'<div style="margin-left:10px;width:520px;height: 30px;" v-repeat="openedApp.container.volumes" data-index="{{$index}}">'+
										'<div style="width:140px;float:left;margin-left:15px;height:30px">'+
											'<input type="text" style="width:100%" id="volume_containerpath_{{$index}}" v-model="containerPath" placeholder="<%=$.i18n.prop("containerPath")%>" required/>'+
										'</div>'+
										'<div style="width:140px;float:left;margin-left:15px;height:30px">'+
											'<input type="text" style="width:100%" id="volume_hostpath_{{$index}}" v-model="hostPath" placeholder="<%=$.i18n.prop("hostPath")%>" required/>'+
										'</div>'+
										'<div style="width:100px;float:left;margin-left:15px;height:30px">'+
											'<select v-model="mode" style="width:100%;">'+  
												'<option value="RO"><%=$.i18n.prop("ro")%></option>'+
												'<option value="RW"><%=$.i18n.prop("rw")%></option>'+
							              	'</select>'+
										'</div>'+
										'<span class="glyphicon glyphicon-remove" style="float:left;cursor:pointer;margin-left:20px;margin-top:3px" title="<%=$.i18n.prop("remove_volume")%>" onclick="removeContainerVolume(event)"/>'+
									'</div>'+
								'</div>'+
								'<div style="width:60%;margin:auto;text-align: center;font-size:16px;height:30px;padding-top:35px;" v-if="openedApp.container.volumes.length==0">'+
									'<%=$.i18n.prop("no_volumes")%>'+
								'</div>'+
							'</div>'+
							'<div style="width:700px;height:120px;float:left">'+
								'<div style="float: left;height:30px;padding-top:35px;width:150px;text-align:center"><%=$.i18n.prop("env")%></div>'+
								'<div style="float:right;width:550px;">'+
									'<textarea placeholder="<%=$.i18n.prop("placeholder_env")%>" id="env" style="width:550px;height:100px;resize: none;" v-model="openedApp.env"></textarea>'+
									'<label id="env-error" class="error" for="env"></label>'+
								'</div>'+
							'</div>'+
							'<div style="width:700px;height:60px;float:left">'+
								'<div style="float: left;height:30px;padding-top:10px;width:150px;text-align:center"><%=$.i18n.prop("expose_ports")%></div>'+
								'<div style="float:right;width:550px;margin-top:8px">'+
									'<input type="radio" name="exposePorts" v-model="openedApp.exposePorts" value="yes"> <%=$.i18n.prop("expose_yes")%>'+
									'<input type="radio" name="exposePorts" v-model="openedApp.exposePorts" value="no" style="margin-left:30px"> <%=$.i18n.prop("expose_no")%>'+
								'</div>'+
							'</div>'+
							'<div style="width:700px;height:120px;float:left">'+
								'<div style="float: left;height:30px;padding-top:35px;width:150px;text-align:center"><%=$.i18n.prop("constraints")%></div>'+
								'<div style="float:right;width:550px;">'+
									'<textarea placeholder="<%=$.i18n.prop("placeholder_constraints")%>" id="constraints" style="width:550px;height:100px;resize: none;" v-model="openedApp.constraints"></textarea>'+
									'<label id="constraints-error" class="error" for="env"></label>'+
								'</div>'+
							'</div>'+
						'</form>';
