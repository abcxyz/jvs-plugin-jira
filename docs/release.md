# Release

**JVS-PLUGIN-JIRA is not an official Google product.**

We leverage [goreleaser](https://goreleaser.com/) for SCM (GitHub) release.


## New Release

1.   If a new JVS base image is needed, update [Dockerfile](https://github.com/abcxyz/jvs-plugin-jira/blob/main/Dockerfile)
to point to a new JVS base image, and merge the PR.

2.   Run the commands below to cut the release.

```sh
# The version you're going to release.
REL_VER=v0.0.x

# Tag
git tag -f -a $REL_VER -m $REL_VER

# Push tag. This will trigger the release workflow.
git push origin $REL_VER
```

## Manually Release Images

Or if you want to build/push images for local development.

**1.  Prepare JVS image**

Pre-built JVS release images can be found from `us-docker.pkg.dev/abcxyz-artifacts/docker-images`. Please ensure to select the
image compatible with your JIRA plugin.

If you want to build a JVS image for local development, follow [this guidance](https://github.com/abcxyz/jvs/blob/main/docs/release.md#manually-release-images).

**2.  Overlay JIRA plugin image**

Update [Dockerfile](https://github.com/abcxyz/jvs-plugin-jira/blob/main/Dockerfile)
accordingly to point to the JVS image, then run the commands below to manually
build the image.

```sh
# Set the container registry for the images, for example:
export CONTAINER_REGISTRY=us-docker.pkg.dev/my-project/images

# goreleaser expects a "clean" repo to release so commit any local changes if
# needed.
git add . && git commit -m "local changes"

# goreleaser expects a tag.
# The tag must be a semantic version https://semver.org/
# DON'T push the tag if you're not releasing.
git tag -f -a v0.0.0-$(git rev-parse --short HEAD)

# goreleaser will tag the image with the git tag, optionally, override it by:
export DOCKER_TAG=mytag

# Use goreleaser to build the images.
# All the images will be tagged with the git tag given earlier.
# Use --skip-validate flag if you are trying to release from a older commit.
# To use a new JVS release, update the base image to a new version in 
# .goreleaser.docker.yaml.
goreleaser release -f .goreleaser.docker.yaml --clean
```
