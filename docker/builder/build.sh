#!/bin/bash

TILES_DIR="valhalla_tiles"
CONF_DIR="conf"

echo "Evironment:"
printenv

if [[ -z "${ROOT_DIR}" ]]; then
  echo "ROOT_DIR environemnt variable must be provided"
  exit 1
fi

cd $ROOT_DIR
mkdir $TILES_DIR $CONF_DIR

if [[ -z "${PBF_URL}" ]]; then
  echo "PBF_URL environemnt variable must be provided"
  exit 1
fi

PBF_FILE_NAME=$(basename $PBF_URL)

echo "Downloading PBF from $PBF_URL"
wget $PBF_URL

echo "Building configuration file..."
valhalla_build_config --mjolnir-tile-dir $ROOT_DIR/$TILES_DIR \
  --mjolnir-tile-extract $ROOT_DIR/valhalla_tiles.tar \
  --mjolnir-timezone $ROOT_DIR/$TILES_DIR/timezones.sqlite \
  --mjolnir-admin $ROOT_DIR/$TILES_DIR/admins.sqlite \
  --mjolnir-traffic-extract $ROOT_DIR/traffic.tar > $ROOT_DIR/$CONF_DIR/valhalla.json

echo "Building admins..."
valhalla_build_admins --config ./$CONF_DIR/valhalla.json $PBF_FILE_NAME

echo "Building timezones..."
valhalla_build_timezones > ./$TILES_DIR/timezones.sqlite

echo "Building tiles..."
valhalla_build_tiles --config ./$CONF_DIR/valhalla.json $PBF_FILE_NAME

echo "Packing files into tar file..."
find $TILES_DIR | sort -n | tar -cf "valhalla_tiles.tar" --no-recursion -T -
