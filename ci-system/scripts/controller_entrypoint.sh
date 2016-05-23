#!/bin/bash
SCRIPT=$(readlink -f "$0")
SCRIPTPATH=$(dirname "$SCRIPT")

echo ZK_SERVERS=${ZK_SERVERS}
echo USERMGMT_ENDPOINT=${USERMGMT_ENDPOINT}
echo CONTROLLER_HOSTNAME=${CONTROLLER_HOSTNAME}

# do configuration
# config zk from env
if [[ -z ${ZK_SERVERS} ]]; then
	#statements
	# sed -i "s/\(zookeeper.url *= *\).*/\1${ZK_SERVERS}/" ${SCRIPTPATH}/controller.properties
# else
	echo "ZK_SERVERS environment is required."
	exit 1
fi

if [[ -z ${USERMGMT_ENDPOINT} ]]; then
	#statements
	# sed -i "s/\(usermgmt.endpoint *= *\).*/\1${USERMGMT_ENDPOINT}/" ${SCRIPTPATH}/controller.properties
# else
	echo "USERMGMT_ENDPOINT environment is required."
	exit 1
fi

# config usermgmt to disable ha
sed -i "s/\(usermgmt.ha *= *\).*/\1false/" ${SCRIPTPATH}/controller.properties

# config mongo using cluster alias
sed -i "s/\(db.alias *= *\).*/\1cluster/" ${SCRIPTPATH}/controller.properties

# config swagger ui 
# skippath=echo ${SCRIPTPATH//\//\\/}
sed -i "s/\(swagger.file.path *= *\).*/\1${SCRIPTPATH//\//\\/}\/swagger-ui\/dist/" ${SCRIPTPATH}/controller.properties

cmd="${SCRIPTPATH}/controller -config ${SCRIPTPATH}/controller.properties"

if [[ ! -z ${CONTROLLER_HOSTNAME} ]]; then
	#statements
	cmd="${cmd} -hostname ${CONTROLLER_HOSTNAME}"
	# ${SCRIPTPATH}/controller --config ${SCRIPTPATH}/controller.properties -hostname ${CONTROLLER_HOSTNAME}
# else
	# ${SCRIPTPATH}/controller --config ${SCRIPTPATH}/controller.properties
fi

exec ${cmd}

