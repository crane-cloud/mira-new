# MIRA

The new mira backend.

## Installation

To install `mira` on your system. Run

```sh
curl -sL https://github.com/crane-cloud/mira-new/releases/download/v0.0.2/install.sh | bash

# OR

wget -qO- https://github.com/crane-cloud/mira-new/releases/download/v0.0.2/install.sh | bash
```

## Running

Mira is composed of two micro-service components. The API Server and the Image Builder. Then can be started by running

```sh
# Starting the API Server
mira api-server

# Starting the Image Builder
mira image-builder
```

The API Server will listen at port `3000`. Incase you have the that port designated to another server, you can overide it by running the command with the `--port` flag.

```sh
mira api-server --port <your-desired-port>
```