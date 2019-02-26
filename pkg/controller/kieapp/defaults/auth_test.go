package defaults

import (
	"testing"

	v1 "github.com/kiegroup/kie-cloud-operator/pkg/apis/app/v1"
	"github.com/kiegroup/kie-cloud-operator/pkg/controller/kieapp/test"
	appsv1 "github.com/openshift/api/apps/v1"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestConfigureHostnameNoAuth(t *testing.T) {
	object := &v1.CustomObject{
		DeploymentConfigs: []appsv1.DeploymentConfig{
			{
				Spec: appsv1.DeploymentConfigSpec{
					Template: &corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{},
							},
						},
					},
				},
			},
		},
	}
	cr := &v1.KieApp{}
	hostname := "https://rhpam.example.com"

	ConfigureHostname(object, cr, hostname)

	httpsHostname := corev1.EnvVar{
		Name:  ssoHostnameVar,
		Value: hostname,
	}
	assert.NotContains(t, object.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env, httpsHostname)
}

func TestConfigureHostname(t *testing.T) {
	testHostname := "test-hostname.example.com"
	object := &v1.CustomObject{
		DeploymentConfigs: []appsv1.DeploymentConfig{
			{
				Spec: appsv1.DeploymentConfigSpec{
					Template: &corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Env: []corev1.EnvVar{},
								},
								{
									Env: []corev1.EnvVar{{
										Name:  ssoClientVar,
										Value: "test-client",
									}},
								},
								{
									Env: []corev1.EnvVar{{
										Name:  ssoClientVar,
										Value: "test-client",
									}, {
										Name:  ssoHostnameVar,
										Value: "",
									}},
								},
								{
									Env: []corev1.EnvVar{{
										Name:  ssoClientVar,
										Value: "test-client",
									}, {
										Name:  ssoHostnameVar,
										Value: testHostname,
									}},
								},
							},
						},
					},
				},
			},
		},
	}
	cr := &v1.KieApp{
		Spec: v1.KieAppSpec{
			Auth: v1.KieAppAuthObject{
				SSO: &v1.SSOAuthConfig{
					URL:   "https://sso.example.com",
					Realm: "therealm",
				},
			},
		},
	}
	hostname := "rhpam.example.com"

	ConfigureHostname(object, cr, hostname)

	httpsHostname := corev1.EnvVar{
		Name:  ssoHostnameVar,
		Value: hostname,
	}
	assert.NotContains(t, object.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env, httpsHostname)
	assert.Contains(t, object.DeploymentConfigs[0].Spec.Template.Spec.Containers[1].Env, httpsHostname)
	assert.Contains(t, object.DeploymentConfigs[0].Spec.Template.Spec.Containers[2].Env, httpsHostname)
	httpsHostname = corev1.EnvVar{
		Name:  ssoHostnameVar,
		Value: testHostname,
	}
	assert.Contains(t, object.DeploymentConfigs[0].Spec.Template.Spec.Containers[3].Env, httpsHostname)
}

func TestAuthMultipleType(t *testing.T) {
	cr := &v1.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: v1.KieAppSpec{
			Environment: "rhpam-trial",
			Auth: v1.KieAppAuthObject{
				SSO:  &v1.SSOAuthConfig{},
				LDAP: &v1.LDAPAuthConfig{},
			},
		},
	}
	_, err := GetEnvironment(cr, test.MockService())
	assert.EqualError(t, err, "multiple authentication types not supported")
}

func TestAuthOnlyRoleMapper(t *testing.T) {
	cr := &v1.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: v1.KieAppSpec{
			Environment: "rhpam-trial",
			Auth: v1.KieAppAuthObject{
				RoleMapper: &v1.RoleMapperAuthConfig{},
			},
		},
	}
	_, err := GetEnvironment(cr, test.MockService())
	assert.EqualError(t, err, "roleMapper configuration must be declared together with SSO or LDAP")
}

func TestAuthNotConfigured(t *testing.T) {
	cr := &v1.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: v1.KieAppSpec{
			Environment: "rhpam-trial",
		},
	}
	_, err := GetEnvironment(cr, test.MockService())
	assert.Nil(t, err, "Error getting trial environment")
}

func TestAuthSSOEmptyConfig(t *testing.T) {
	cr := &v1.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: v1.KieAppSpec{
			Environment: "rhpam-trial",
			Auth: v1.KieAppAuthObject{
				SSO: &v1.SSOAuthConfig{},
			},
		},
	}
	_, err := GetEnvironment(cr, test.MockService())
	assert.EqualError(t, err, "neither url nor realm can be empty")
}

func TestAuthSSOConfig(t *testing.T) {
	cr := &v1.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: v1.KieAppSpec{
			Environment: "rhpam-trial",
			Auth: v1.KieAppAuthObject{
				SSO: &v1.SSOAuthConfig{
					URL:   "https://sso.example.com:8080",
					Realm: "rhpam-test",
				},
			},
		},
	}
	env, err := GetEnvironment(cr, test.MockService())
	assert.Nil(t, err, "Error getting trial environment")

	expectedEnvs := []corev1.EnvVar{
		{Name: "SSO_URL", Value: "https://sso.example.com:8080"},
		{Name: "SSO_REALM", Value: "rhpam-test"},
		{Name: "SSO_OPENIDCONNECT_DEPLOYMENTS", Value: "ROOT.war"},
		{Name: "SSO_PRINCIPAL_ATTRIBUTE", Value: "preferred_username"},
		{Name: "SSO_DISABLE_SSL_CERTIFICATE_VALIDATION", Value: "false"},
		{Name: "SSO_USERNAME"},
		{Name: "SSO_PASSWORD"},
		{Name: "SSO_PASSWORD"},
	}
	for _, expectedEnv := range expectedEnvs {
		assert.Contains(t, env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env, expectedEnv, "Console should contain env %v", expectedEnv)
		for i := range env.Servers {
			assert.Contains(t, env.Servers[i].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env, expectedEnv, "Server %v should contain env %v", i, expectedEnv)
		}
	}

	expectedClientEnvs := []corev1.EnvVar{
		{Name: "SSO_SECRET"},
		{Name: "SSO_CLIENT"},
		{Name: "HOSTNAME_HTTP"},
		{Name: "HOSTNAME_HTTPS"},
	}
	for _, expectedEnv := range expectedClientEnvs {
		assert.Contains(t, env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env, expectedEnv, "Console should contain env %v", expectedEnv)
		for i := range env.Servers {
			assert.Contains(t, env.Servers[i].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env, expectedEnv, "Server %v should contain env %v", i, expectedEnv)
		}
	}
}

func TestAuthSSOConfigWithMismatchedClients(t *testing.T) {
	cr := &v1.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: v1.KieAppSpec{
			Environment: "rhpam-trial",
			Auth: v1.KieAppAuthObject{
				SSO: &v1.SSOAuthConfig{
					URL:   "https://sso.example.com:8080",
					Realm: "rhpam-test",
					Clients: v1.SSOAuthClients{
						Console: v1.SSOAuthClient{
							Name:          "test-rhpamcentr-client",
							Secret:        "supersecret",
							HostnameHTTP:  "test-rhpamcentr.example.com",
							HostnameHTTPS: "secure-test-rhpamcentr.example.com",
						},
						Servers: []v1.SSOAuthClient{
							{
								Name:   "test-kieserver-a-client",
								Secret: "supersecret-a",
							},
							{
								Name:          "test-kieserver-b-client",
								Secret:        "supersecret-b",
								HostnameHTTPS: "test-kieserver-b.example.com",
							},
						},
					},
				},
			},
			Objects: v1.KieAppObjects{
				Server: &v1.CommonKieServerSet{
					Deployments: 3,
				},
			},
		},
	}
	_, err := GetEnvironment(cr, test.MockService())
	assert.Error(t, err, "the number of Server SSO clients defined must match the number of KIE Servers")

}

func TestAuthSSOConfigWithClients(t *testing.T) {
	cr := &v1.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: v1.KieAppSpec{
			Environment: "rhpam-trial",
			Auth: v1.KieAppAuthObject{
				SSO: &v1.SSOAuthConfig{
					URL:   "https://sso.example.com:8080",
					Realm: "rhpam-test",
					Clients: v1.SSOAuthClients{
						Console: v1.SSOAuthClient{
							Name:          "test-rhpamcentr-client",
							Secret:        "supersecret",
							HostnameHTTP:  "test-rhpamcentr.example.com",
							HostnameHTTPS: "secure-test-rhpamcentr.example.com",
						},
						Servers: []v1.SSOAuthClient{
							{
								Name:   "test-kieserver-a-client",
								Secret: "supersecret-a",
							},
							{
								Name:          "test-kieserver-b-client",
								Secret:        "supersecret-b",
								HostnameHTTPS: "test-kieserver-b.example.com",
							},
						},
					},
				},
			},
			Objects: v1.KieAppObjects{
				Server: &v1.CommonKieServerSet{
					Deployments: 2,
				},
			},
		},
	}
	env, err := GetEnvironment(cr, test.MockService())
	assert.Nil(t, err, "Error getting trial environment")

	expectedEnvs := []corev1.EnvVar{
		{Name: "SSO_URL", Value: "https://sso.example.com:8080"},
		{Name: "SSO_REALM", Value: "rhpam-test"},
		{Name: "SSO_OPENIDCONNECT_DEPLOYMENTS", Value: "ROOT.war"},
		{Name: "SSO_PRINCIPAL_ATTRIBUTE", Value: "preferred_username"},
		{Name: "SSO_DISABLE_SSL_CERTIFICATE_VALIDATION", Value: "false"},
		{Name: "SSO_USERNAME"},
		{Name: "SSO_PASSWORD"},
	}
	for _, expectedEnv := range expectedEnvs {
		assert.Contains(t, env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env, expectedEnv, "Console does not contain env %v", expectedEnv)
		for i := range env.Servers {
			assert.Contains(t, env.Servers[i].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env, expectedEnv, "Server %v does not contain env %v", i, expectedEnv)
		}
	}

	expectedConsoleClientEnvs := []corev1.EnvVar{
		{Name: "SSO_SECRET", Value: "supersecret"},
		{Name: "SSO_CLIENT", Value: "test-rhpamcentr-client"},
		{Name: "HOSTNAME_HTTP", Value: "test-rhpamcentr.example.com"},
		{Name: "HOSTNAME_HTTPS", Value: "secure-test-rhpamcentr.example.com"},
	}
	for _, expectedEnv := range expectedConsoleClientEnvs {
		assert.Contains(t, env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env, expectedEnv, "Console does not contain env %v", expectedEnv)
	}

	expectedServerClientEnvs := []corev1.EnvVar{
		{Name: "SSO_SECRET", Value: "supersecret-a"},
		{Name: "SSO_CLIENT", Value: "test-kieserver-a-client"},
	}
	for _, expectedEnv := range expectedServerClientEnvs {
		assert.Contains(t, env.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env, expectedEnv, "Server 0 does not contain env %v", expectedEnv)
	}
	expectedServerClientEnvs = []corev1.EnvVar{
		{Name: "SSO_SECRET", Value: "supersecret-b"},
		{Name: "SSO_CLIENT", Value: "test-kieserver-b-client"},
		{Name: "HOSTNAME_HTTPS", Value: "test-kieserver-b.example.com"},
	}
	for _, expectedEnv := range expectedServerClientEnvs {
		assert.Contains(t, env.Servers[1].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env, expectedEnv, "Server 1 does not contain env %v", expectedEnv)
	}
}

func TestAuthLDAPEmptyConfig(t *testing.T) {
	cr := &v1.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: v1.KieAppSpec{
			Environment: "rhpam-trial",
			Auth: v1.KieAppAuthObject{
				LDAP: &v1.LDAPAuthConfig{},
			},
		},
	}
	_, err := GetEnvironment(cr, test.MockService())
	assert.EqualError(t, err, "the url must not be empty")
}

func TestAuthLDAPConfig(t *testing.T) {
	cr := &v1.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: v1.KieAppSpec{
			Environment: "rhpam-trial",
			Objects: v1.KieAppObjects{
				Server: &v1.CommonKieServerSet{
					Deployments: 2,
				},
			},
			Auth: v1.KieAppAuthObject{
				LDAP: &v1.LDAPAuthConfig{
					URL:    "ldaps://ldap.example.com",
					BindDN: "cn=admin,dc=example,dc=com",
				},
			},
		},
	}
	env, err := GetEnvironment(cr, test.MockService())
	assert.Nil(t, err, "Error getting trial environment")

	expectedEnvs := []corev1.EnvVar{
		{Name: "AUTH_LDAP_URL", Value: "ldaps://ldap.example.com"},
		{Name: "AUTH_LDAP_BIND_DN", Value: "cn=admin,dc=example,dc=com"},
		{Name: "AUTH_LDAP_BIND_CREDENTIAL"},
	}
	for _, expectedEnv := range expectedEnvs {
		assert.Contains(t, env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env, expectedEnv, "Console does not contain env %v", expectedEnv)
		for i := range env.Servers {
			assert.Contains(t, env.Servers[i].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env, expectedEnv, "Server %v does not contain env %v", i, expectedEnv)
		}
	}
}

func TestAuthRoleMapperConfig(t *testing.T) {
	cr := &v1.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: v1.KieAppSpec{
			Environment: "rhpam-trial",
			Objects: v1.KieAppObjects{
				Server: &v1.CommonKieServerSet{
					Deployments: 2,
				},
			},
			Auth: v1.KieAppAuthObject{
				LDAP: &v1.LDAPAuthConfig{
					URL:    "ldaps://ldap.example.com",
					BindDN: "cn=admin,dc=example,dc=com",
				},
				RoleMapper: &v1.RoleMapperAuthConfig{
					RolesProperties: "mapping.properties",
					ReplaceRole:     true,
				},
			},
		},
	}
	env, err := GetEnvironment(cr, test.MockService())
	assert.Nil(t, err, "Error getting trial environment")

	expectedEnvs := []corev1.EnvVar{
		{Name: "AUTH_LDAP_URL", Value: "ldaps://ldap.example.com"},
		{Name: "AUTH_LDAP_BIND_DN", Value: "cn=admin,dc=example,dc=com"},
		{Name: "AUTH_LDAP_BIND_CREDENTIAL"},
		{Name: "AUTH_ROLE_MAPPER_ROLES_PROPERTIES", Value: "mapping.properties"},
		{Name: "AUTH_ROLE_MAPPER_REPLACE_ROLE", Value: "true"},
	}
	for _, expectedEnv := range expectedEnvs {
		assert.Contains(t, env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env, expectedEnv, "Console does not contain env %v", expectedEnv)
		for i := range env.Servers {
			assert.Contains(t, env.Servers[i].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env, expectedEnv, "Server %v does not contain env %v", i, expectedEnv)
		}
	}
}
