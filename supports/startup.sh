#!/usr/bin/env bash

export LANG=en_US.UTF-8
export LC_ALL=en_US.UTF-8

PROGRAM_NAME=fireye
PROGRAM_CONF=config.yaml
PROGRAM_DIR=/var/apps/app/
PROGRAM_BIN=${PROGRAM_DIR}/${PROGRAM_NAME}

CMD="${PROGRAM_BIN} -c ${PROGRAM_DIR}/${PROGRAM_CONF}"

echo "Starting ..."
echo "${CMD}"
$CMD
