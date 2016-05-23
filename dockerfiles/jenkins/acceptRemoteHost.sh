#!/usr/bin/expect
# spawn git -c core.askpass=true ls-remote -h ssh://ywang2@192.168.3.61:29418/OnlyTest HEAD
spawn git -c core.askpass=true ls-remote -h ssh://linker@[lindex $argv 0]:29418/
expect {
"yes/no" {
 send "yes\n"
	}
}
expect eof
