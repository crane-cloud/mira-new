# MIRA

The new mira backend.

## Running

To run Mira in development follow the following steps.

**Clone the Mira Repository**

```sh
git clone https://github.com/crane-cloud/mira-new.git

cd mira-new
```

Then start up the Docker dependency services

```sh
docker compose -f scripts/compose.yml up -d
```

Mira is composed of two micro-service components. The API Server and the Image Builder. They can be started by running

```sh
# Starting the API Server
go run main.go api-server

# Starting the Image Builder
go run main.go image-builder
```

The API Server will listen at port `3000`. Incase you have the that port designated to another server, you can overide it by running the command with the `--port` flag.

```sh
go run main.go api-server --port <your-desired-port>
```

## Usage

To containerize source code into an image. You will have to send a POST request to `/images/containerize` path. The Content type is `multipart/form-data` with the following fields.

| Field         | Type                  | Description                                                    |
| ------------- | --------------------- | -------------------------------------------------------------- |
| `name`        | `string`              | Name of the source                                             |
| `type`        | `git` \| `file`       | Source type: git or file                                       |
| `file`        | `blob`                | A ZIP file containing the source code you want to containerize |
| `branch`      | `string`              | Git branch to use                                              |
| `repo`        | `string`              | Repository URL or file path                                    |
| `gitusername` | `string` *(optional)* | Git username (if required)                                     |
| `gitpassword` | `string` *(optional)* | Git password or token (if required)                            |

This will return a `JSON` response in this format.

```json
{
  "data": {
    "name": "mira-test",
    "runid": "1681d792-e193-4b0b-91f0-79f9279cc5b9",
    "wspath": "localhost:8080/drivers/streams/logs/mira/1681d792-e193-4b0b-91f0-79f9279cc5b9"
  },
  "message": "Image generation started"
}
```

### Logs

The response contains a `data.wspath` field that contains a URL. You can open a Websocket connection to this path and stream build logs. This log stream contains the entire buildpack lifecycle logs. You can filter out the logs for the lifecycle step you want, which is usually the build step.

> Note: If the buildpack produces logs that contain [ANSI escape codes](https://en.wikipedia.org/wiki/ANSI_escape_code) used for terminal color formatting. The log stream will also contain these, so when displaying on the frontend you can use a package like [ansi-to-html](https://www.npmjs.com/package/ansi-to-html) for formating, or you can just use a regex to strip them out.

You can also check out this [HTML template file](https://github.com/crane-cloud/mira-new/blob/main/public/logs.html) demonstrating this Entire process.