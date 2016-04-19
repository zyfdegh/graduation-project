echo "begin to deploy cluster by ansible..."

ansible-playbook -i hosts deploy.yml

echo "deploy mesos cluster end"