#!/bin/bash

echo "Stopping currenlty running docker-compose containers.."
docker-compose stop &2> /dev/null
num_of_backends=${1:-5}

# docker-compose use this as prefix for all running containers
project_name=$(basename $PWD)

for b in $(seq 1 $num_of_backends); do hosts="${hosts}${project_name}_api_server$b:8080,"; done

export BACKENDS=$(echo $hosts | sed 's/.$//')

echo "running $num_of_backends instances of api server"

docker-compose -p ${project_name} up --build --scale api_server=${num_of_backends} -e BACKENDS=${BACKENDS}
