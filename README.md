# Release Installer

A simple tool to install package from GitLab release.

## Installation

* install script

```shell
curl -fsSL https://raw.githubusercontent.com/zachcheung/release-installer/main/install.sh | sh
```

* [releases](https://github.com/zachcheung/release-installer/releases)

* go install

```shell
go install github.com/zachcheung/release-installer@latest
```

## Usage

```shell
release-installer -url https://gitlab.example.com <REPO>
```

#### Example

* Public GitLab Repo

```shell
/ # release-installer -url https://gitlab.com goreleaser/example
2024/07/12 08:26:25 Downloading example_2.0.7_linux_amd64.tar.gz from https://gitlab.com/goreleaser/example/-/releases/v2.0.7/downloads/example_2.0.7_linux_amd64.tar.gz
2024/07/12 08:26:25 Downloaded example_2.0.7_linux_amd64.tar.gz
2024/07/12 08:26:25 Installed example to /usr/local/bin
```

* Private GitLab Repo in GitLab CI Job

```shell
release-installer -url ${CI_SERVER_URL} ${CI_PROJECT_ID}
```

#### Supported Providers

* GitLab (token is required when repo is private)

#### Supported Compressed Package

* gzip

## License

[MIT](LICENSE)
