package configs

import (
    "github.com/alecthomas/gobundle"
)

var ConfigBundle = gobundle.NewBuilder("config").Add(
    "configs/common-env.json", []byte("{\n  \"SSO_DISABLE_SSL_CERTIFICATE_VALIDATION\":\"FALSE\",\n  \"SSO_OPENIDCONNECT_DEPLOYMENTS\":\"ROOT.war\",\n  \"KIE_ADMIN_PWD\":\"RedHat\",\n  \"KIE_SERVER_CONTROLLER_PWD\":\"RedHat\",\n  \"KIE_SERVER_PWD\":\"RedHat\",\n  \"KIE_SERVER_USER\":\"executionUser\",\n  \"KIE_MBEANS\":\"enabled\",\n  \"KIE_SERVER_CONTROLLER_USER\":\"controllerUser\",\n  \"KIE_ADMIN_USER\":\"adminUser\"\n}"),
).Add(
    "configs/console-env.json", []byte("{\n  \"PROBE_DISABLE_BOOT_ERRORS_CHECK\":\"true'\",\n  \"PROBE_IMPL\":\"probe.eap.jolokia.EapProbe\",\n  \"KIE_MAVEN_USER\":\"mavenUser\",\n  \"KIE_MAVEN_PWD\":\"RedHat\"\n}"),
).Add(
    "configs/server-env.json", []byte("{\n  \"KIE_SERVER_CONTROLLER_PROTOCOL\":\"ws\",\n  \"MAVEN_REPOS\":\"RHPAMCENTR,EXTERNAL\",\n  \"RHPAMCENTR_MAVEN_REPO_PASSWORD\":\"RedHat\",\n  \"RHPAMCENTR_MAVEN_REPO_USERNAME\":\"mavenUser\",\n  \"RHPAMCENTR_MAVEN_REPO_PATH\":\"/maven2/\"\n}"),
).Build()
