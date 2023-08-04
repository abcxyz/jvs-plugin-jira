# To use a new JVS release, update the base image to a new version.
FROM us-docker.pkg.dev/abcxyz-artifacts/docker-images/jvsctl:v0.1.0-alpha1

COPY jvs-plugin-jira /var/jvs/plugins/jvs-plugin-jira

# Normally we would set this to run as "nobody". But goreleaser builds the
# binary locally and sometimes it will mess up the permission and cause "exec
# user process caused: permission denied".
#
# USER nobody

# Run the CLI
ENTRYPOINT ["/bin/jvsctl"]
