package defaults

import (
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestConsoleEnvironmentDefaults(t *testing.T) {
	defaults := ConsoleEnvironmentDefaults()
	logrus.Debugf("Loaded common defaults as %v", defaults)
	assert.Equal(t, defaults["SSO_OPENIDCONNECT_DEPLOYMENTS"], "ROOT.war", "Expected ROOT.war as the value for SSO_OPENIDCONNECT_DEPLOYMENTS")
	assert.Equal(t, defaults["KIE_MAVEN_USER"], "mavenUser", "Expected mavenUser as the value for KIE_MAVEN_USER")
}

func TestServerEnvironmentDefaults(t *testing.T) {
	defaults := ServerEnvironmentDefaults()
	logrus.Debugf("Loaded server defaults as %v", defaults)
	assert.Equal(t, defaults["SSO_OPENIDCONNECT_DEPLOYMENTS"], "ROOT.war", "Expected ROOT.war as the value for SSO_OPENIDCONNECT_DEPLOYMENTS")
	assert.Equal(t, defaults["RHPAMCENTR_MAVEN_REPO_USERNAME"], "mavenUser", "Expected mavenUser as the value for RHPAMCENTR_MAVEN_REPO_USERNAME")
}
