#!/bin/bash

# get ldap env then restart
# LDAP_SERVER="ldap:\/\/127.0.0.1"
# LDAP_ROOT_DN="dc=linkernetworks,dc=com"
# LDAP_MANAGER_DN="cn=admin,dc=linkernetworks,dc=com"

LDAP_SERVER=$2;
GERRIT_IP=$4;

/acceptRemoteHost.sh ${GERRIT_IP}

echo "\${LDAP_SERVER}=ldap:\/\/${LDAP_SERVER}"
echo "\${LDAP_ROOT_DN}=${LDAP_ROOT_DN}"
echo "\${LDAP_MANAGER_DN}=${LDAP_MANAGER_DN}"

config_xml_path=${JENKINS_HOME}/config.xml
echo "config.xml path is ${config_xml_path}"

sed -i "s/%ldap_server%/ldap\:\/\/${LDAP_SERVER}/" ${config_xml_path}
sed -i "s/%ldap_root_dn%/${LDAP_ROOT_DN}/" ${config_xml_path}
sed -i "s/%ldap_manager_dn%/${LDAP_MANAGER_DN}/" ${config_xml_path}

/restart.sh

gen-ssh-key() {
  expect << END
  spawn ssh-keygen -t rsa
  expect "Enter file in which to save the key (/root/.ssh/id_rsa):"
  send "\r"
  expect "Enter passphrase (empty for no passphrase):"
  send "\r"
  expect "Enter same passphrase again:"
  send "\r"

expect eof
END
}

gen-ssh-key

curl --digest --user linker:password -X POST --header "Content-Type: text/plain" http://$GERRIT_IP:8080/a/accounts/self/sshkeys -d@/root/.ssh/id_rsa.pub

echo "config.sh success "
