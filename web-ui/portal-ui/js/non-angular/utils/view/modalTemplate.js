var nonAngularAlertTemplate = '<div class="modal-dialog">'+
								'<div class="modal-content">'+
								  '<div class="modal-header">'+    
								    '<h4><%=alert.title%></h4>'+
								  '</div>'+
								  '<div class="modal-body">'+
								    '<div class="modal-icon-area">'+
								      '<span class="<%=alert.type%>-icon glyphicon glyphicon-<%=alert.sign%>-sign" aria-hidden="true" ></span>'+ 
								    '</div>'+
								    '<div class="modal-font-area">'+
								      '<span><%=alert.message%></span>'+ 
								    '</div>'+ 
								  '</div>'+
								  '<div class="modal-footer">'+  
								  	  '<%if(alert.modaltype == "notify"){%>'+
								      '<button class="btn btn-primary" data-dismiss="modal" onclick="clearModalMask();"><%=$.i18n.prop("ok")%></button>'+
								      '<%}else if(alert.modaltype == "dconfirm"){%>'+
								      '<button class="btn btn-primary" data-dismiss="modal" onclick="clearModalMask();<%=alert.actions[0]%>"><%=$.i18n.prop("delete")%></button>'+
								      '<button class="btn btn-primary" data-dismiss="modal" onclick="clearModalMask();"><%=$.i18n.prop("cancel")%></button>'+
								      '<%}%>'+
								  '</div>'+
								  '</div>'+
								'</div>';