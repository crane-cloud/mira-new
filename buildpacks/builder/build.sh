#!/usr/bin/env bash
set -e

PLATFORM=${PLATFORM:-linux/amd64}

RUN_IMAGE=cranecloudplatform/buildpacks-run:latest
BUILD_IMAGE=cranecloudplatform/buildpacks-build:latest

# Build the base images
echo "Building base images for Crane Cloud Platform Builder..."
echo "Using platform: ${PLATFORM}"
echo "BUILDING ${BUILD_IMAGE}..."
docker build --platform=${PLATFORM} \
  -t "${BUILD_IMAGE}" \
  ./build

echo "BUILDING ${RUN_IMAGE}..."
docker build --platform=${PLATFORM} \
  -t "${RUN_IMAGE}" \
  /run

echo
echo "BASE IMAGES BUILT!"
echo
echo "Images:"
for IMAGE in "${BUILD_IMAGE}" "${RUN_IMAGE}"; do
  echo "    ${IMAGE}"
done