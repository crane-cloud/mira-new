# Define the base image
FROM ubuntu:noble

# Install packages that we want to make available at run time
RUN apt-get update && \
  apt-get install -y xz-utils ca-certificates && \
  rm -rf /var/lib/apt/lists/*

# Create user and group
ARG cnb_uid=1000
ARG cnb_gid=1000
RUN groupadd cnb --gid ${cnb_gid} && \
  useradd --uid ${cnb_uid} --gid ${cnb_gid} -m -s /bin/bash cnb

# Set user and group
USER ${cnb_uid}:${cnb_gid}

# Set required CNB target information
LABEL io.buildpacks.base.distro.name="ubuntu"
LABEL io.buildpacks.base.distro.version="24.04"