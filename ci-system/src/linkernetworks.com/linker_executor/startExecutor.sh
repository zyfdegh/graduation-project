#!/bin/bash

echo "Running executor..."
SCRIPT=$(readlink -f "$0")
SCRIPTPATH=$(dirname "$SCRIPT")

if [[ ! -z ${MESOS_MASTER} ]]; then
	#statements
	ZK_SERVERS=`echo ${MESOS_MASTER} | cut -d'/' -f 3`
	echo "ZK_SERVERS=${ZK_SERVERS}"
	$SCRIPTPATH/linkerexecutor -config $SCRIPTPATH/executor.properties -zk ${ZK_SERVERS}
else
	$SCRIPTPATH/linkerexecutor -config $SCRIPTPATH/executor.properties
fi