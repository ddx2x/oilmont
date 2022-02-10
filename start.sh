#!/bin/bash
set -o errexit
set -o nounset
set -o pipefail

REPO=k8sharbor.ddx2x.nip/ddx2x
COMMIT_ID=$(git rev-parse --verify HEAD)
VERSION="${VERSION:-"${COMMIT_ID:0:8}"}"

function build() {
    for DOCKERFILE in $(ls  -l images | grep -v ^d |awk '{print $9}')
    do
    APP=$(echo ${DOCKERFILE} | awk '{split($0,a,"."); print(a[2])}')
    docker build -t ${REPO}/${APP}:${VERSION} -f images/${DOCKERFILE} .
    docker push ${REPO}/${APP}:${VERSION}
    done
};

function run(){
  BASE_ARGS="--registry etcd --registry_address=127.0.0.1:2379"
  GATEWAY_ARGS=${BASE_ARGS}" ""api --handler=http --address 0.0.0.0:8080"
  DATABASE="mongodb://127.0.0.1:27017/admin"

  for DOCKERFILE in $(ls  -l images | grep -v ^d |awk '{print $9}')
    do
    APP=$(echo ${DOCKERFILE} | awk '{split($0,a,"."); print(a[2])}')
    ARGS=${BASE_ARGS}
    docker rm -f ${APP}

    if [[ ${APP} == "gateway" ]]; then
      ARGS=${GATEWAY_ARGS}
      docker run -tid --name=${APP} \
          -e STORAGE_URI=${DATABASE} \
          --net=host \
          ${REPO}/${APP}:${VERSION} \
          ${ARGS}
    else
      ARGS=${BASE_ARGS}
      docker run -tid --name=${APP} \
            -e STORAGE_URI=${DATABASE} \
            ${REPO}/${APP}:${VERSION} \
            ${ARGS}
    fi
    done
}


while true
do
  case "$1" in
  build)
      build
      shift
      ;;
  run)
      run
      shift
      ;;
  -h|--help)
      usage
      ;;
  esac
shift
done