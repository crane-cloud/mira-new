package utils

import (
	"os"
)

var APP_TYPE = os.Getenv("APP_TYPE")

var OUTPUT_DIR = os.Getenv("OUTPUT_DIR")

var APP_START_COMMAND = os.Getenv("START_COMMAND")

var APP_BUILD_COMMAND = os.Getenv("BUILD_COMMAND")

var CNB_LAYERS_DIR = os.Getenv("CNB_LAYERS_DIR")

var NODE_JS_LAYERS_DIR = CNB_LAYERS_DIR + "/nodejs"

var NODE_JS_LAUNCH_FILE = CNB_LAYERS_DIR + "/nodejs.toml"

var LAUNCHER_FILE = CNB_LAYERS_DIR + "/launch.toml"

var SSR_START_COMMAND = NODE_JS_LAYERS_DIR + "/bin/" + APP_START_COMMAND

var NPM_INSTALL_COMMAND = NODE_JS_LAYERS_DIR + "/bin/node " + NODE_JS_LAYERS_DIR + "/bin/" + "npm install"

var BUILD_COMMAND = NODE_JS_LAYERS_DIR + "/bin/node " + NODE_JS_LAYERS_DIR + "/bin/" + APP_BUILD_COMMAND

var NODE_JS_BUILD_TOML = NODE_JS_LAYERS_DIR + "/build.toml"
var NODE_JS_LAUNCH_TOML = NODE_JS_LAYERS_DIR + "/launch.toml"

var CADDY_LAYERS_DIR = NODE_JS_LAYERS_DIR + "/caddy"

var CADDY_FILE_PATH = CADDY_LAYERS_DIR + "/Caddyfile"

var CADDY_START_COMMAND = CADDY_LAYERS_DIR + "/" + "caddy run --config " + CADDY_FILE_PATH + " --adapter caddyfile"
