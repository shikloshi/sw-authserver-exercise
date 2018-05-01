#!/bin/bash

set -e

num_of_backends=${1:-5}

server_name=api_server
api_server_port=8080

# docker-compose use this as prefix for all running containers
project_name=$(basename $PWD)

for b in $(seq 1 $num_of_backends); do hosts="${hosts}${project_name}_${server_name}_$b:${api_server_port},"; done

echo $hosts

export BACKENDS=$(echo $hosts | sed 's/.$//')

echo "Creating haproxy backend servers to..."

echo "Add the following to haproxy/haproxy.cfg"
echo "========================================"

echo "Add this this to get_api_backend"
for i in $(seq 1 $num_of_backends); 
do
    echo "server api$i ${project_name}_${server_name}_$i:${api_server_port}"
done
echo "========================================"

echo "Add this this to post_api_backend"
echo "server api-replicator ${project_name}_replicate_service_1:8080"
echo "========================================"

echo "Press any key when ready..."
read

echo "running $num_of_backends instances of api server"

echo "Stopping currenlty running docker-compose containers.."
docker-compose stop &2> /dev/null

docker-compose -p ${project_name} up --build --scale api_server=${num_of_backends}
