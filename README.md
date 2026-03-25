# Logdy - terminal logs in web browser

<p align="center">
<img src="https://github.com/logdyhq/logdy-core/assets/1653294/9ec8cb3f-0b8f-4523-b600-377444734b9d" height=100/>
</p>

<p align="center">
<strong> <a href="https://logdy.dev">Webpage</a> | 
<a href="https://demo.logdy.dev">Demo</a> | 
<a href="https://logdy.dev/docs/quick-start">Docs</a> | 
<a href="https://github.com/logdyhq/logdy-core/releases">Download</a> | 
<a href="https://logdy.dev/blog">Blog</a> | </strong> 
<a href="https://github.com/logdyhq/logdy-core/actions/workflows/test.yml">
  <img src="https://github.com/logdyhq/logdy-core/actions/workflows/test.yml/badge.svg"/>
</a>
</p>

### Latest version: 0.17.0 (1 June 2025) - [Read announcement](https://logdy.dev/blog/post/logdy-new-version-announcement-v017)

Logdy is a lightweight, single-binary log viewer that works just like `grep`, `awk`, `sed`, or `jq`. Simply add it to your PATH—no installation, deployment, or compilation required. It runs entirely locally, ensuring security and privacy. [Read more](https://logdy.dev/docs/what-is-logdy).

### Key features
* Zero-dependency single binary
* Embedded Web UI
* Real-time log viewing and filtering
* Secure local operation
* Multiple input modes (files, stdin, sockets, REST API)
* Custom parsers and columns with TypeScript support (code editor with types support)
* Go library integration
* Cross-build support for Raspberry Pi (ARMv6/ARMv7/ARM64)

### Standalone use
```bash
# Use with any shell command
$ tail -f file.log | logdy
WebUI started, visit http://localhost:8080

# Read log files
$ logdy follow app-out.log --full-read
WebUI started, visit http://localhost:8080
```
More use [modes in the docs.](https://logdy.dev/docs/explanation/command-modes)

### Use as a Go library
```go
package main

import "github.com/logdyhq/logdy-core/logdy"

func main(){
  logdyLogger := logdy.InitializeLogdy(logdy.Config{
    ServerIp:       "127.0.0.1",
    ServerPort:     "8080",
  }, nil)

  // app code...

  logdyLogger.LogString("This is a message")
  logdyLogger.Log(logdy.Fields{"msg": "supports structured logs too", "url": "some url here"})
}
```
Check [docs](https://logdy.dev/docs/golang-logs-viewer) or [example app](https://github.com/logdyhq/logdy-core/blob/main/example-app/main.go).

## Demo of the UI
Visit [demo.logdy.dev](https://demo.logdy.dev)


![autogenerate](https://github.com/logdyhq/logdy-core/assets/1653294/bfe09fa8-bbba-46fa-b54d-503f796c7b57)

Visit [logdy.dev](http://logdy.dev) for more info and detailed documentation.

##### Project status: New features added actively.

Logdy is in active development, with new features being added regularly. Feedback is welcome from early adopters. Feel free to post [Issues](https://github.com/logdyhq/logdy-core/issues), [Pull Requests](https://github.com/logdyhq/logdy-core/pulls) and contribute in the [Discussions](https://github.com/logdyhq/logdy-core/discussions). Stay tuned for updates, visit [Logdy Blog](https://logdy.dev/blog).

## Install using script
The command below will download the latest release and add the executable to your system's PATH. You can also use it to update Logdy.

```bash
curl https://logdy.dev/install.sh | sh
```

## Install with Homebrew (MacOS)
On MacOS you can use homebrew to install Logdy.

```bash
brew install logdy
```

## Download precompiled binary

Navigate to [releases](https://github.com/logdyhq/logdy-core/releases) Github page and download the latest release for your architecture.

```bash
wget https://github.com/logdyhq/logdy-core/releases/download/vX.Y.Z/logdy_linux_arm64;
mv logdy_linux_arm64 logdy;
chmod +x logdy;
```
Additionally, you can [add the binary to your PATH](https://logdy.dev/docs/how-tos#how-to-add-logdy-to-path) for easier access.

## Raspberry Pi

`logdy` runs well on Raspberry Pi with 32-bit and 64-bit OS images.

### Build Raspberry Pi binaries
Use the dedicated script to generate binaries for:
- `linux/armv6` (Pi Zero / Pi 1)
- `linux/armv7` (Pi 2 / Pi 3 with 32-bit OS)
- `linux/arm64` (Pi 3/4/5 with 64-bit OS)

```bash
./scripts/build_rpi.sh X.Y.Z
```

Generated files:
- `bin/logdy_linux_armv6`
- `bin/logdy_linux_armv7`
- `bin/logdy_linux_arm64`

### Install as a service on Raspberry Pi (systemd)
This repository includes a hardened service unit at `packaging/systemd/logdy.service`.

Example installation flow on Raspberry Pi:
```bash
sudo install -m 0755 ./bin/logdy_linux_arm64 /usr/local/bin/logdy
sudo ./scripts/install_rpi_service.sh /usr/local/bin/logdy
```

By default, service mode starts:
- `logdy --ui-ip 0.0.0.0 --port 8080 stdin`

Then access the UI from your network:
- `http://<raspberry-pi-ip>:8080`
## Quick start
Whatever the below command will produce to the output, will be forwarded to a Web UI.
```bash
node index.js | logdy
```
The following should appear
```
INFO[2024-02...] WebUI started, visit http://localhost:8080    port=8080
```
Open the URL Address and start building parsers, columns and filters.

There are multiple other ways you can run Logdy, check the [docs](https://logdy.dev/docs/explanation/command-modes).

## Install Go library
```bash
go get -u github.com/logdyhq/logdy-core/logdy
```
[Read more](https://logdy.dev/docs/golang-logs-viewer) about how to use Logdy embedded into your Go app.

## Documentation

For product documentation navigate to the [official docs](https://logdy.dev/docs/quick-start).

## CLI Usage

```bash
Usage:
  logdy [command] [flags]
  logdy [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  demo        Starts a demo mode, random logs will be produced, the [number] defines a number of messages produced per second
  follow      Follows lines added to files. Example `logdy follow foo.log /var/log/bar.log`
  forward     Forwards the STDIN to a specified port, example `tail -f file.log | logdy forward 8123`
  help        Help about any command
  socket      Sets up a port to listen on for incoming log messages. Example `logdy socket 8233`. You can setup multiple ports `logdy socket 8123 8124 8125`
  stdin       Listens to STDOUT/STDERR of a provided command. Example `logdy stdin "npm run dev"`
  utils       A set of utility commands that help working with large files

Flags:
      --api-key string                API key (send as a header Authorization)
      --append-to-file string         Path to a file where message logs will be appended, the file will be created if it doesn't exist
      --append-to-file-raw            When 'append-to-file' is set, raw lines without metadata will be saved to a file
      --bulk-window int               A time window during which log messages are gathered and send in a bulk to a client. Decreasing this window will improve the 'real-time' feeling of messages presented on the screen but could decrease UI performance (default 100)
      --config string                 Path to a file where a config (json) for the UI is located
      --disable-ansi-code-stripping   Use this flag to disable Logdy from stripping ANSI sequence codes
  -t, --fallthrough                   Will fallthrough all of the stdin received to the terminal as is (will display incoming messages)
  -h, --help                          help for logdy
      --max-message-count int         Max number of messages that will be stored in a buffer for further retrieval. On buffer overflow, oldest messages will be removed. (default 100000)
  -n, --no-analytics                  Opt-out from sending anonymous analytical data that helps improve Logdy
  -u, --no-updates                    Opt-out from checking updates on program startup
  -p, --port string                   Port on which the Web UI will be served (default "8080")
      --rotate-file-size string       If set, how big the file can grow before being rotated, used K/M/G to describe the size, example: 15M -> 15 megabytes (env: LOGDY_ROTATE_FILE_SIZE)
      --ui-ip string                  Bind Web UI server to a specific IP address (default "127.0.0.1")
      --ui-pass string                Password that will be used to authenticate in the UI
  -v, --verbose                       Verbose logs
      --version                       version for logdy
```

**It is possible to also provide some of the options as ENV variables, read more in the docs**

## Development
For development, we recommend running `demo` mode
```bash
go run . demo 1
```

The above command will start Logdy in `demo` mode with 1 log message produced per second.
You can read more about [demo mode](https://logdy.dev/docs/demo-mode).

If you would like to develop with UI, check [readme for logdy-ui](https://github.com/logdyhq/logdy-ui) for instructions how to run both together during development.

## Building

This repository uses static asset embedding during compilation. This way, the UI is served from a single binary. Before you build make sure you copy a compiled [UI](https://github.com/logdyhq/logdy-ui) (follow the instructions about building) in `http/assets` directory. The UI is already commited to this repository, so you don't have to do anymore actions.

Look at `http/embed.go` for more details on how UI is embedded into the binary.

For a local architecture build:
```bash
go build
```

## Releasing
For cross-architecture builds use the repository scripts below. This generates binaries in `bin/` and includes Raspberry Pi targets.

```bash
./build_across x.x.x
./scripts/build_rpi.sh x.x.x
```

Once it's ready, publish the binaries in a new Github release. Again, don't forget to update the version.

```bash
ghr vx.x.x bin/
```
