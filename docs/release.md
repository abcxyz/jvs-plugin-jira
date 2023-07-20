# Release

**JVS-PLUGIN-JIRA is not an official Google product.**

We leverage [goreleaser](https://goreleaser.com/) for SCM (GitHub) release.


## New Release

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

```sh
# goreleaser expects a "clean" repo to release so commit any local changes if
# needed.
git add . && git commit -m "local changes"

# goreleaser expects a tag.
# The tag must be a semantic version https://semver.org/
# DON'T push the tag if you're not releasing.
git tag -f -a v0.0.0-$(git rev-parse --short HEAD)

# Use goreleaser to build the images.
# All the images will be tagged with the git tag given earlier.
goreleaser release --clean --skip-publish
```
