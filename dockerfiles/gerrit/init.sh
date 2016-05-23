#!/bin/bash
echo "start init gerrit review db"
mysql -h ${2} -P3306 -uroot -p123456 < /create_gerrit_user.sql

echo "start backup config file"
cp $GERRIT_HOME/gerrit/etc/gerrit.config $GERRIT_HOME/gerrit/etc/gerrit.config.bak

echo "start init gerrit"
java -jar $GERRIT_WAR init --batch -d ${GERRIT_HOME}/gerrit

$GERRIT_HOME/gerrit/bin/gerrit.sh stop

echo "replace config"
cp $GERRIT_HOME/gerrit/etc/gerrit.config.bak $GERRIT_HOME/gerrit/etc/gerrit.config

echo "start gerrit again"
$GERRIT_HOME/gerrit/bin/gerrit.sh start

echo "add  linker user into db"
mysql -h ${2} -P3306 -uroot -p123456 reviewdb < /init.sql

echo "done"
