insert into accounts(registered_on, full_name, preferred_email, contact_filed_on, maximum_page_size, show_site_header,use_flash_clipboard,download_url, download_command, copy_self_on_email, date_format, time_format,reverse_patch_set_order,show_user_in_review, relative_date_in_change_table, comment_visibility_strategy,diff_view,  change_screen, size_bar_in_change_table, inactive,account_id) 
	values(now(),  'linker', 'linker@linkernetworks.com', null, 25, 'Y','Y','SSH','CHECKOUT', 'N', null, null, 'N', 'Y', 'N',null,null, null, 'N', 'N', '1');

insert into account_external_ids(account_id, email_address, password, external_id) values('1', 'linker@linkernetworks.com', 'password', 'gerrit:linker@linkernetworks.com');
insert into account_external_ids(account_id, email_address, password, external_id) values('1', null, 'password', 'username:linker');

insert into account_group_members(account_id, group_id) values(1, 1);
insert into account_group_members_audit (added_by, removed_by, removed_on, account_id, group_id, added_on) values(1, null, null, 1, 1, now());