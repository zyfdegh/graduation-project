var serviceTreeItemTemplate = '<div style="text-align: center;">' +
	'<div id="group_<%=node.id%>" class="group" data-groupid="<%=node.id%>" style="height:<%=node.data.$height%>px;">' +
	'<%_.each(node.data.apps,function(app){%>' +
	'<%var simappid = app.id;%>'+
	'<%var appid = node.id + "/" + simappid;%>' +
	'<%_.each(app.instance_ids,function(instance_id){%>' +
	'<%readIPAddress(instance_id);%>' +
	'<div class="group-app-item" style="cursor:move" data-appid="<%=appid%>" id="<%=instance_id%>" oncontextmenu="generateServiceContext(event)" >'+
		'<div class="a-app-image-container"><img src="<%=app.imageSrc%>" class="container-image" draggable="false"/></div>'+
		'<div class="a-app-item-id"><div class="a-app-item-id-label"><%=simappid%></div></div>'+
	'</div>'+
	'<%})%>' +
	'<%})%>' +
	'</div>' +
	'<div style="font-size:16px;margin-top:10px"><span data-groupid="<%=node.id%>"><%=node.data.id%></span></div>' +
	'</div>';

var serviceTreeRootTemplate = '<div class="service-group" data-groupid="<%=node.id%>">'+
								'<div class="service-group-image-container"><img src="<%=node.data.imageSrc%>" class="container-image" draggable="false"/></div>'+
								'<div class="a-app-item-id"><div class="a-app-item-id-label"><%=node.data.id%></div></div>'+
							'</div>';

var serviceScaleTemplate = '<form name="scaleform">'+
							'<div>' +
								'<input type="number" id="scaleText" name="scaleText" style="height:25px;width:200px;font-size:14px;" min="1" step="1" data-operate="<%=operate%>" data-gid="<%=gid%>" placeholder="<%=placeholder%>" onkeydown="doScale(event)" onfocusout="closeSecondDialog()" required/>' +
							'</div></form>';
	
var cadvisorTemplate = '<div class="modal-dialog" style="width:1200px">'+
							'<div class="modal-content">'+
								'<div class="modal-header">'+
									'<h4><%=$.i18n.prop("metering_for")%> <%=appid%></h4>'+
									'<span class="glyphicon glyphicon-remove cp-close-button" data-dismiss="modal" title="Close"></span>'+
								'</div>'+
								'<div class="modal-body" id="embed_cadvisor">'+
									 '<iframe width="1170" height="450" frameborder="0" allowfullscreen=""></iframe>'+
								'</div>'+
							'</div>'+
						'</div>';
