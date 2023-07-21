# The jvs image.
ARG JVS_IMAGE

FROM ${JVS_IMAGE}

# The folder should be consistent with JustificationConfig.PluginDir.
# https://github.com/abcxyz/jvs/blob/main/pkg/config/justification_config.go#L49
ARG PLUGIN_DIR

COPY jvs-plugin-jira ${PLUGIN_DIR}/jvs-plugin-jira

# Normally we would set this to run as "nobody". But goreleaser builds the
# binary locally and sometimes it will mess up the permission and cause "exec
# user process caused: permission denied".
#
# USER nobody

# Run the CLI
ENTRYPOINT ["/bin/jvsctl"]
