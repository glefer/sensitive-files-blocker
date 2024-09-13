# Traefik Sensitive file blocker plugin

![Banner](./.assets/icon.webp)

This plugin allows Traefik users to block access to sensitive files and directories that should not be publicly accessible. Common examples of sensitive files are .env files, .git/ directories, and other configuration or version control files that could expose private data or system vulnerabilities if accessed externally.


[![Build Status](https://github.com/glefer/sensitive-files-blocker/actions/workflows/main.yml/badge.svg?branch=main)](https://github.com/glefer/sensitive-files-blocker/actions)
[![codecov](https://codecov.io/github/glefer/sensitive-files-blocker/graph/badge.svg?token=MX6K3NPPAO)](https://codecov.io/github/glefer/sensitive-files-blocker)
[![Go Report Card](https://goreportcard.com/badge/github.com/glefer/sensitive-files-blocker)](https://goreportcard.com/report/github.com/glefer/sensitive-files-blocker)
![Go Version](https://img.shields.io/github/go-mod/go-version/glefer/sensitive-files-blocker?style=flat-square)
![Latest Release](https://img.shields.io/github/v/release/glefer/sensitive-files-blocker?style=flat-square&sort=semver)

## Installation
It is possible to install the [plugin locally](https://traefik.io/blog/using-private-plugins-in-traefik-proxy-2-5/) or to install it through [Traefik Pilot](https://pilot.traefik.io/plugins).

## Features
* Blocks predefined sensitive files.
* Customizable list of files or directories to block.
*  Lightweight and easy to install as a Traefik middleware plugin.
* Flexible installation methods: via Traefik Pilot, local mode, or Docker.

### Standard

This procedure will install the plugin via the [Traefik Plugin registry](https://plugins.traefik.io/install).

Add this configuration in your `traefik.yml` file.

```yaml
experimental:
  plugins:
    sensitive-files-blocker:
      moduleName: "github.com/glefer/sensitive-files-blocker"
      version: "v0.0.1"
```
### Local Mode

Download the latest release of the plugin and save it to a location the Traefik container can reach. 

```bash
git clone git@github.com:glefer/sensitive-files-blocker.git
```


The source code of the plugin should be organized as follows:

```
./plugins-local/
    └── src
        └── github.com
            └── glefer
                └── sensitive-files-blocker
                    ├── .assets/
                    ├── .github/
                    ├── .gitignore
                    ├── .traefik.yml
                    ├── go.mod
                    ├── LICENSE
                    ├── main.go
                    ├── main_test.go
                    ├── Makefile
                    └── readme.md
```

```yaml
# Static configuration
experimental:
  localPlugins:
    sensitive-files-blocker:
      moduleName: github.com/glefer/sensitive-files-blocker
```

(In the above example, the `sensitive-files-blocker` plugin will be loaded from the path `$GOPATH/plugins-local/src/github.com/glefer/sensitive-files-blocker`.)


> If `$GOPATH` is not set, plugins-local must be placed at the root of the server.

### Docker

To run this plugin in a Docker environment, add the following configuration to your compose.yml file to mount `plugins-local` directory in the right container path.

```yaml

services:
  traefik:
    image: "traefik:v3.1"
    container_name: "traefik"
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - "/var/run/docker.sock:/var/run/docker.sock:ro"
      - "/path/to/plugins-local/:/plugins-local/"

```


## Configuration

Here is an example of a file provider dynamic configuration (given here in YAML), where the interesting part is the `http.middlewares` section:

```yaml
# Dynamic configuration
http:
  middlewares:
    sensitive-files-blocker:
      plugin:
        sensitive-files-blocker:
          files:
            - ^.env$
            - ^.git
```

The files field accepts a list of regex patterns representing files or directories to block. You can customize it based on the files that should be restricted from access.

### Enable middleware for all routers

In order to enable this middleware for all routers, you can add this configuration in the `traefik.yml` file.

```yaml
entryPoints:
  http:
    address: ":80"
    http:
      middlewares:
        - sensitive-files-blocker@file
```

This configuration ensures that any requests going through all routers will have the sensitive file blocker middleware applied.


### Enable Middleware for a Specific Router (YAML Format)
```yaml
http:
  routers:
    Router-1:
      rule: "Host(`example.com`)"
      service: "service-1"
      middlewares:
        - sensitive-files-blocker@file
```

### Enable Middleware for a Specific Router (docker compose format)
```yaml
services:
  my-container:
    # ...
    labels:
      - "traefik.http.routers.my-container.rule=Host(`example.com`)"
      - "traefik.http.routers.my-app.entrypoints=websecure"
      - "traefik.http.routers.my-app.tls.certresolver=myresolver"
      - "traefik.http.middlewares.my-app-blocker.plugin.sensitive-files-blocker.files=^.env$,^.git"

```

## Contributing

Contributions are welcome! Feel free to open issues, submit pull requests, or suggest new features.
Testing

To run tests:
```bash
make test
```

This will run all tests defined in main_test.go.

## License

This project is licensed under the MIT License. See the LICENSE file for more details.
