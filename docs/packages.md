# LinuxKit packages

LinuxKit packages a container images which are pull using the `moby`
tool and assembled into bootable Linux images. LinuxKit comes with a
number of [packages](../pkg) which are core part of LinuxKit, but
users can add their own packages to the YAML files.

All LinuxKit packages are:
- Signed with Docker Content Trust.
- Multi-arch manifests to work on multiple architectures.
- Derived from well-known (and signed) sources for repeatable builds.
- Build with multi-stage builds to minimise their size.

## Package source

A package source consists of a directory containing at least two files:

- `build.yml`: contains metadata associated with the package
- `Dockerfile`: contains the steps to build the package.

`build.yml` contains the following fields:

- `image` _(string)_: *(mandatory)* The name of the image to build
- `org` _(string)_: The hub/registry organisation to which this package belongs
- `arches` _(list of string)_: The architectures which this package should be built for (valid entries are `GOARCH` names)
- `gitrepo` _(string)_: The git repository where the package source is kept.
- `network` _(bool)_: Allow network access during the package build (default: no)
- `disable-content-trust` _(bool)_: Disable Docker content trust for this package (default: no)
- `disable-cache` _(bool)_: Disable build cache for this package (default: no)

## Building packages

### Prerequisites

Before you can build packages you need:
- Docker version 17.06 or newer. If you are on a Mac you also need
  `docker-credential-osxkeychain.bin`, which comes with Docker for Mac.
- `make`, `notary`, `base64`, `jq`, and `expect`
- A *recent* version of `manifest-tool` which you can build with `make
  bin/manifest-tool`, or `go get github.com:estesp/manifest-tool`, or
  via the LinuxKit homebrew tap with `brew install --HEAD
  manifest-tool`. `manifest-tool` must be in your path.
- The LinuxKit tool `linuxkit` which must be in your path.

Further, when building packages you need to be logged into hub with
`docker login` as some of the tooling extracts your hub credentials
during the build.

### Build packages as a maintainer

If you have write access to the `linuxkit` organisation on hub, you
should also be set up with signing keys for packages and your signing
key should have a passphrase, which we call `<passphrase>` throughout.

All official LinuxKit packages are multi-arch manifests and most of
them are available for amd64 and aarm64. Official images *must* be
build on both architectures and they must be build *in sequence*, i.e.,
they can't be build in parallel.

To build a package on an architecture:

```
DOCKER_CONTENT_TRUST_REPOSITORY_PASSPHRASE="<passphrase>" linuxkit pkg push «path-to-package»
```

`«path-to-package»` is the path to the package's source directory
(containing at least `build.yml` and `Dockerfile`). It can be `.` if
the package is in the current directory.

**Note:** You *must* be logged into hub (`docker login`) and the
passphrase for the key *must* be supplied as an environment
variable. The build process has to resort to using `expect` to drive
`notary` so none of the credentials can be entered interactively.

This will:
- Build a local images as `linuxkit/<image>:<hash>-<arch>`
- Push it to hub
- Sign it with your key
- Create a manifest called `linuxkit/<image>:<hash>` (note no `-<arch>`)
- Push the manifest to hub
- Sign the manifest

If you repeat the same on another architecture, a new manifest will be
pushed and signed containing the previous and the new
architecture. The YAML files should consume the package as:
`linuxkit/<image>:<hash>`.


Since it is not very good to have your passphrase in the clear (or
even stashed in your shell history), we recommend using a password
manager with a CLI interface, such as LastPass or `pass`. You can then
invoke the build like this (for LastPass):

```
DOCKER_CONTENT_TRUST_REPOSITORY_PASSPHRASE=$(lpass show <key> --password) linuxkit pkg push «path-to-package»
```

### Build packages as a developer

If you want to develop packages or test them locally, it is best to
override the hub organisation used. You may also want to disable
signing while developing. A typical example would be:

```
linuxkit pkg build -org=wombat -disable-content-trust «path-to-package»
```

This will create a local image: `wombat/<image>:<hash>-<arch>` which
you can use in your local YAML files for testing. If you need to test
on other systems you can push the image to your hub account and pull
from a different system by issuing:

```
linuxkit pkg build -org=wombat -disable-content-trust push
```

This will push both `wombat/<image>:<hash>-<arch>` and
`wombat/<image>:<hash>` to hub.

Finally, if you are tired of the long hashes you can override the hash
with:

```
linuxkit pkg build -org=wombat -disable-content-trust -hash=foo push
```

and this will create `wombat/<image>:foo-<arch>` and
`wombat/<image>:foo` for use in your YAML files.

