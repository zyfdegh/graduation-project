docker run -p 8080:8080 -it graduation-project/marathon bash
mkdir /etc/marathon/conf
config /etc/marathon/conf/zk
config /etc/marathon/conf/hostname
config /etc/hosts
nohup marathon --no-logger --event_subscriber http_callback > /var/log/marathon.log 2>&1 &
