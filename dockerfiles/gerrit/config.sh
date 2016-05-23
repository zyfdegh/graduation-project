#!/bin/bash

local_ip=`ifconfig eth1|sed -n 2p|awk  '{ print $2 }'|awk -F : '{ print $2 }'`;
mysql_ip=$2;
ldap_ip=$4;
# 29418,8080
containerports=$6;
# 32000,32001
hostports=$8;

portcontainers=(${containerports//,/ })
porthosts=(${hostports//,/ })
#finalport = ""

containerindex=1
for containerport in ${portcontainers[@]}
do
    if [ $containerport = '8080' ]
        then
           finalindex=`expr $containerindex`
    fi
    containerindex=`expr $containerindex + 1`
done

hostindex=1
for hostport in ${porthosts[@]}
do
    if [ $hostindex -eq $finalindex ]
        then
           finalport=$hostport
    fi
    hostindex=`expr $hostindex + 1`
done

sed -i "s/\(canonicalWebUrl *= *http\:\/\/\).*/\1$HOST\:${finalport}\//" $GERRIT_HOME/gerrit/etc/gerrit.config
sed -i "s/\(hostname *= *\).*/\1${mysql_ip}/" $GERRIT_HOME/gerrit/etc/gerrit.config
sed -i "s/\(server *= *ldap\:\/\/*\).*/\1${ldap_ip}/" $GERRIT_HOME/gerrit/etc/gerrit.config