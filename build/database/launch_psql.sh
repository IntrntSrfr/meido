#!/bin/bash

IMAGE_NAME="meido_psql"
DOCKERFILE_LOCATION="$(pwd)/Dockerfile"

# Check if the image exists
if [[ "$(docker images -q $IMAGE_NAME 2> /dev/null)" == "" ]]; then
  # Build the image if it does not exist
  echo "Image not found. Building image..."
  docker build -t $IMAGE_NAME -f $DOCKERFILE_LOCATION .
fi

# Check if the container is already running
if [[ "$(docker ps -q -f name=$IMAGE_NAME 2> /dev/null)" == "" ]]; then
  # Run the image if the container is not already running
  echo "Container not running. Starting container..."
  #docker run --name $IMAGE_NAME -d $IMAGE_NAME
  docker run --name $IMAGE_NAME -e POSTGRES_DB=meido -e POSTGRES_PASSWORD=password -d -p 5432:5432 $IMAGE_NAME
fi

