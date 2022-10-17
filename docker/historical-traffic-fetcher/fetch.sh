#!/bin/bash

TRAFFIC_DIR = "traffic"

echo "Evironment:"
printenv

if [[ -z "${ROOT_DIR}" ]]; then
  echo "ROOT_DIR environemnt variable must be provided"
  exit 1
fi

cd $ROOT_DIR
if [ ! -d "${TRAFFIC_DIR}" ]; then
  mkdir $TRAFFIC_DIR
fi

cd $TRAFFIC_DIR

if [[ -z "${URL}" ]]; then
  echo "ROOT_DIR environemnt variable must be provided"
  exit 1
fi

echo "Downloading historical traffic data from $URL"
curl -O $URL historical_traffic_data

echi "Adding Historical traffic data..."
valhalla_add_predicted_traffic historical_traffic_data