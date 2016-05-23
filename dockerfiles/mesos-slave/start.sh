config /etc/hosts

config /etc/mesos/zk
#zk://172.17.0.1:2181,172.17.0.2:2181/mesos

mesos-slave --master=zk://172.17.0.1:2181,172.17.0.2:2181/mesos --no-systemd_enable_support --containerizers=docker
