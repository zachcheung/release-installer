# Release Installer

A simple tool to install packages from GitHub/GitLab releases. Tested with most of the [Prometheus exporters](https://raw.githubusercontent.com/prometheus/docs/c725947ff1f4aac33ea8664c9dc413ab93fd3ab4/content/docs/instrumenting/exporters.md).

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
release-installer <REPO>
```

It is recommended to test in a container before installing a package.

```shell
docker run --rm ghcr.io/zachcheung/release-installer -dir /tmp <REPO>
```

#### Example

* Public GitHub Repo

```console
/ # release-installer goreleaser/example
2024/07/15 01:08:46 Downloading example_1.3.0_linux_amd64.tar.gz from https://github.com/goreleaser/example/releases/download/v1.3.0/example_1.3.0_linux_amd64.tar.gz
2024/07/15 01:08:48 Downloaded example_1.3.0_linux_amd64.tar.gz
2024/07/15 01:08:48 Installed example to /usr/local/bin
```

* Private GitHub Repo

```shell
release-installer -token <TOKEN> <PRIVATE_REPO>
```

The token should have the `repo` scope if using `Personal access tokens (classic)`.

* Public GitLab Repo

```console
/ # release-installer -provider gitlab goreleaser/example
2024/07/12 08:26:25 Downloading example_2.0.7_linux_amd64.tar.gz from https://gitlab.com/goreleaser/example/-/releases/v2.0.7/downloads/example_2.0.7_linux_amd64.tar.gz
2024/07/12 08:26:25 Downloaded example_2.0.7_linux_amd64.tar.gz
2024/07/12 08:26:25 Installed example to /usr/local/bin
```

* Private GitLab Repo (in the Self-Repo CI Job)

```shell
release-installer -provider gitlab -url $CI_SERVER_URL -token $GITLAB_TOKEN $CI_PROJECT_ID
```

The `GITLAB_TOKEN` should be set and have at least the `read_api` scope.

If the value of `-url` contains `gitlab`, the `-provider` can be omitted.

#### Supported Providers

* GitHub
* GitLab

Token is required when repo is private.

#### Supported Compressed Package

* gzip

## License

[MIT](LICENSE)
