#!/usr/bin/env bash
set -e

PLATFORM=${PLATFORM:-linux/amd64}

RUN_IMAGE=cranecloud/buildpacks-run:latest
BUILD_IMAGE=cranecloud/buildpacks-build:latest

# Build the base images
echo "Building base images for Crane Cloud Platform Builder..."
echo "Using platform: ${PLATFORM}"
echo "BUILDING ${BUILD_IMAGE}..."
docker buildx build --platform=${PLATFORM} \
  -t "${BUILD_IMAGE}" \
  ./build --push

echo "BUILDING ${RUN_IMAGE}..."
docker buildx build --platform=${PLATFORM} \
  -t "${RUN_IMAGE}" \
  ./run --push

echo
echo "BASE IMAGES BUILT!"
echo
echo "Images:"
for IMAGE in "${BUILD_IMAGE}" "${RUN_IMAGE}"; do
  echo "    ${IMAGE}"
done