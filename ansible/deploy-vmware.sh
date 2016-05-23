echo "begin to deploy cluster by ansible..."

ansible-playbook -i hosts-vmware --user=root --private-key=centos7_key deploy.yml

echo "deploy mesos cluster end"