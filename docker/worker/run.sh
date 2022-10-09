#!/bin/bash

CONF_DIR="conf"
THREADS=${THREADS_PER_POD:=2}

if [[ -z "${ROOT_DIR}" ]]; then
  echo "ROOT_DIR environemnt variable must be provided"
  exit 1
fi

cd $ROOT_DIR

echo "Starting Valhalla server with $THREADS threads..."
valhalla_service ./$CONF_DIR/valhalla.json $THREADS