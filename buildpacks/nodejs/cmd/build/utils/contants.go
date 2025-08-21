package utils

import (
	"os"
)

var CNB_LAYERS_DIR = os.Getenv("CNB_LAYERS_DIR")

var NODE_JS_LAYERS_DIR = CNB_LAYERS_DIR + "/nodejs"

var NODE_JS_LAUNCH_FILE = CNB_LAYERS_DIR + "/nodejs.toml"

var LAUNCHER_FILE = CNB_LAYERS_DIR + "/launch.toml"

var START_COMMAND = NODE_JS_LAYERS_DIR + "/bin/" + "npm run preview"

var NPM_INSTALL_COMMAND = NODE_JS_LAYERS_DIR + "/bin/node " + NODE_JS_LAYERS_DIR + "/bin/" + "npm install"

var BUILD_COMMAND = NODE_JS_LAYERS_DIR + "/bin/node " + NODE_JS_LAYERS_DIR + "/bin/" + "npm run build"
