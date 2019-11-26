package defaults

import (
	"strconv"
	"testing"

	api "github.com/kiegroup/kie-cloud-operator/pkg/apis/app/v2"
	"github.com/kiegroup/kie-cloud-operator/pkg/controller/kieapp/constants"
	"github.com/kiegroup/kie-cloud-operator/pkg/controller/kieapp/test"
	appsv1 "github.com/openshift/api/apps/v1"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestConfigureHostnameNoAuth(t *testing.T) {
	object := &api.CustomObject{
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
	cr := &api.KieApp{}
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
	object := &api.CustomObject{
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
	cr := &api.KieApp{
		Spec: api.KieAppSpec{
			Auth: api.KieAppAuthObject{
				SSO: &api.SSOAuthConfig{
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
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: api.KieAppSpec{
			Environment: "rhpam-trial",
			Auth: api.KieAppAuthObject{
				SSO:  &api.SSOAuthConfig{},
				LDAP: &api.LDAPAuthConfig{},
			},
		},
	}
	_, err := GetEnvironment(cr, test.MockService())
	assert.EqualError(t, err, "multiple authentication types not supported")
}

func TestAuthOnlyRoleMapper(t *testing.T) {
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: api.KieAppSpec{
			Environment: "rhpam-trial",
			Auth: api.KieAppAuthObject{
				RoleMapper: &api.RoleMapperAuthConfig{},
			},
		},
	}
	_, err := GetEnvironment(cr, test.MockService())
	assert.EqualError(t, err, "roleMapper configuration must be declared together with SSO or LDAP")
}

func TestAuthNotConfigured(t *testing.T) {
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: api.KieAppSpec{
			Environment: "rhpam-trial",
		},
	}
	_, err := GetEnvironment(cr, test.MockService())
	assert.Nil(t, err, "Error getting trial environment")
}

func TestAuthSSOEmptyConfig(t *testing.T) {
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: api.KieAppSpec{
			Environment: "rhpam-trial",
			Auth: api.KieAppAuthObject{
				SSO: &api.SSOAuthConfig{},
			},
		},
	}
	_, err := GetEnvironment(cr, test.MockService())
	assert.EqualError(t, err, "neither url nor realm can be empty")
}

func TestExternalOnlyDefaultConfig(t *testing.T) {
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: api.KieAppSpec{
			Environment: "rhpam-trial",
			CommonConfig: api.CommonConfig{
				AdminUser:     "testUser",
				AdminPassword: "testPwd",
			},
			Auth: api.KieAppAuthObject{
				SSO: &api.SSOAuthConfig{
					URL:   "https://sso.example.com:8080",
					Realm: "rhpam-test",
				},
			},
			Version: "7.6.0",
		},
	}
	env, err := GetEnvironment(cr, test.MockService())
	assert.Nil(t, err, "Error getting trial environment")

	expectedEnvs := []corev1.EnvVar{
		{Name: "EXTERNAL_AUTH_ONLY", Value: "false"},
		{Name: "KIE_ADMIN_USER", Value: "testUser"},
		{Name: "KIE_ADMIN_PWD", Value: "testPwd"},
		{Name: "KIE_SERVER_USER", Value: "testUser"},
		{Name: "KIE_SERVER_PWD", Value: "testPwd"},
		{Name: "KIE_SERVER_CONTROLLER_USER", Value: "testUser"},
		{Name: "KIE_SERVER_CONTROLLER_PWD", Value: "testPwd"},
	}

	for _, expectedEnv := range expectedEnvs {
		assert.Contains(t, env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env, expectedEnv, "Console should contain env %v", expectedEnv)
		for i := range env.Servers {
			assert.Contains(t, env.Servers[i].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env, expectedEnv, "Server %v should contain env %v", i, expectedEnv)
		}
	}
}

func TestExternalOnlyConfig(t *testing.T) {
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: api.KieAppSpec{
			Environment: "rhpam-trial",
			CommonConfig: api.CommonConfig{
				AdminUser:     "testUser",
				AdminPassword: "testPwd",
			},
			Auth: api.KieAppAuthObject{
				ExternalOnly: true,
				SSO: &api.SSOAuthConfig{
					URL:   "https://sso.example.com:8080",
					Realm: "rhpam-test",
				},
			},
			Version: "7.6.0",
		},
	}
	env, err := GetEnvironment(cr, test.MockService())
	assert.Nil(t, err, "Error getting trial environment")

	expectedEnvs := []corev1.EnvVar{
		{Name: "EXTERNAL_AUTH_ONLY", Value: "true"},
		{Name: "KIE_ADMIN_USER", Value: "testUser"},
		{Name: "KIE_ADMIN_PWD", Value: "testPwd"},
		{Name: "KIE_SERVER_USER", Value: "testUser"},
		{Name: "KIE_SERVER_PWD", Value: "testPwd"},
		{Name: "KIE_SERVER_CONTROLLER_USER", Value: "testUser"},
		{Name: "KIE_SERVER_CONTROLLER_PWD", Value: "testPwd"},
	}

	for _, expectedEnv := range expectedEnvs {
		assert.Contains(t, env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env, expectedEnv, "Console should contain env %v", expectedEnv)
		for i := range env.Servers {
			assert.Contains(t, env.Servers[i].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env, expectedEnv, "Server %v should contain env %v", i, expectedEnv)
		}
	}
}

func TestExternalOnlyConfigOldVersions(t *testing.T) {
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: api.KieAppSpec{
			Environment: "rhpam-trial",
			CommonConfig: api.CommonConfig{
				AdminUser:     "testUser",
				AdminPassword: "testPwd",
			},
			Auth: api.KieAppAuthObject{
				SSO: &api.SSOAuthConfig{
					URL:   "https://sso.example.com:8080",
					Realm: "rhpam-test",
				},
			},
			Version: "7.5.1",
		},
	}
	env, err := GetEnvironment(cr, test.MockService())
	assert.Nil(t, err, "Error getting trial environment")

	expectedEnvs := []corev1.EnvVar{
		{Name: "KIE_ADMIN_USER", Value: "testUser"},
		{Name: "KIE_ADMIN_PWD", Value: "testPwd"},
		{Name: "KIE_SERVER_USER", Value: "executionUser"},
		{Name: "KIE_SERVER_PWD", Value: "RedHat"},
		{Name: "KIE_SERVER_CONTROLLER_USER", Value: "controllerUser"},
		{Name: "KIE_SERVER_CONTROLLER_PWD", Value: "RedHat"},
	}

	for _, expectedEnv := range expectedEnvs {
		assert.Contains(t, env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env, expectedEnv, "Console should contain env %v", expectedEnv)
		for i := range env.Servers {
			assert.Contains(t, env.Servers[i].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env, expectedEnv, "Server %v should contain env %v", i, expectedEnv)
		}
	}

	externalAuthEnv := corev1.EnvVar{Name: "EXTERNAL_AUTH_ONLY"}
	assert.NotContains(t, env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env, externalAuthEnv, "Console should not contain env %v", externalAuthEnv)
	for i := range env.Servers {
		assert.NotContains(t, env.Servers[i].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env, externalAuthEnv, "Server %v not should contain env %v", i, externalAuthEnv)
	}
}

func TestAuthSSOConfig(t *testing.T) {
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: api.KieAppSpec{
			Environment: "rhpam-trial",
			Auth: api.KieAppAuthObject{
				SSO: &api.SSOAuthConfig{
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

func TestAuthSSOConfigWithClients(t *testing.T) {
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: api.KieAppSpec{
			Environment: "rhpam-trial",
			Auth: api.KieAppAuthObject{
				SSO: &api.SSOAuthConfig{
					URL:   "https://sso.example.com:8080",
					Realm: "rhpam-test",
				},
			},
			Objects: api.KieAppObjects{
				Console: api.ConsoleObject{
					SSOClient: &api.SSOAuthClient{
						Name:          "test-rhpamcentr-client",
						Secret:        "supersecret",
						HostnameHTTP:  "test-rhpamcentr.example.com",
						HostnameHTTPS: "secure-test-rhpamcentr.example.com",
					},
				},
				Servers: []api.KieServerSet{
					{
						Name:        "one",
						Deployments: Pint(2),
						SSOClient: &api.SSOAuthClient{
							Name:   "test-kieserver-a-client",
							Secret: "supersecret-a",
						},
					},
					{
						Deployments: Pint(3),
						SSOClient: &api.SSOAuthClient{
							Name:          "test-kieserver-b-client",
							Secret:        "supersecret-b",
							HostnameHTTPS: "test-kieserver-b.example.com",
						},
					},
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
		assert.Contains(t, env.Servers[1].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env, expectedEnv, "Server 1 does not contain env %v", expectedEnv)
	}
	expectedServerClientEnvs = []corev1.EnvVar{
		{Name: "SSO_SECRET", Value: "supersecret-b"},
		{Name: "SSO_CLIENT", Value: "test-kieserver-b-client"},
		{Name: "HOSTNAME_HTTPS", Value: "test-kieserver-b.example.com"},
	}
	for _, expectedEnv := range expectedServerClientEnvs {
		assert.Contains(t, env.Servers[2].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env, expectedEnv, "Server 2 does not contain env %v", expectedEnv)
		assert.Contains(t, env.Servers[3].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env, expectedEnv, "Server 3 does not contain env %v", expectedEnv)
		assert.Contains(t, env.Servers[4].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env, expectedEnv, "Server 4 does not contain env %v", expectedEnv)
	}
}

func TestAuthLDAPEmptyConfig(t *testing.T) {
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: api.KieAppSpec{
			Environment: "rhpam-trial",
			Auth: api.KieAppAuthObject{
				LDAP: &api.LDAPAuthConfig{},
			},
		},
	}
	_, err := GetEnvironment(cr, test.MockService())
	assert.EqualError(t, err, "the url must not be empty")
}

func TestAuthLDAPConfig(t *testing.T) {
	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: api.KieAppSpec{
			Environment: "rhpam-trial",
			Objects: api.KieAppObjects{
				Servers: []api.KieServerSet{
					{Deployments: Pint(2)},
				},
			},
			Auth: api.KieAppAuthObject{
				LDAP: &api.LDAPAuthConfig{
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
	var defaultMode int32 = 420
	tests := []struct {
		name                string
		roleMapper          *api.RoleMapperAuthConfig
		expectedVolumeMount *corev1.VolumeMount
		expectedVolume      *corev1.Volume
		expectedPath        string
	}{{
		name: "RoleMapper config is set with defaults",
		roleMapper: &api.RoleMapperAuthConfig{
			RolesProperties: "mapping.properties",
		},
		expectedVolumeMount: nil,
		expectedVolume:      nil,
		expectedPath:        constants.RoleMapperDefaultDir + "/mapping.properties",
	}, {
		name: "RoleMapper config has ReplaceRole",
		roleMapper: &api.RoleMapperAuthConfig{
			RolesProperties: "mapping.properties",
			ReplaceRole:     true,
		},
		expectedVolumeMount: nil,
		expectedVolume:      nil,
		expectedPath:        constants.RoleMapperDefaultDir + "/mapping.properties",
	}, {
		name: "RoleMapper config from a ConfigMap",
		roleMapper: &api.RoleMapperAuthConfig{
			RolesProperties: "mapping.properties",
			From: &corev1.ObjectReference{
				Name: "test-cm",
				Kind: "ConfigMap",
			},
		},
		expectedVolumeMount: &corev1.VolumeMount{
			Name:      constants.RoleMapperVolume,
			MountPath: constants.RoleMapperDefaultDir,
			ReadOnly:  true,
		},
		expectedVolume: &corev1.Volume{
			Name: constants.RoleMapperVolume,
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: "test-cm",
					},
					DefaultMode: &defaultMode,
				},
			},
		},
		expectedPath: constants.RoleMapperDefaultDir + "/mapping.properties",
	}, {
		name: "RoleMapper config from a Secret",
		roleMapper: &api.RoleMapperAuthConfig{
			RolesProperties: "mapping.properties",
			From: &corev1.ObjectReference{
				Name: "test-secret",
				Kind: "Secret",
			},
		},
		expectedVolumeMount: &corev1.VolumeMount{
			Name:      constants.RoleMapperVolume,
			MountPath: constants.RoleMapperDefaultDir,
			ReadOnly:  true,
		},
		expectedVolume: &corev1.Volume{
			Name: constants.RoleMapperVolume,
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: "test-secret",
				},
			},
		},
		expectedPath: constants.RoleMapperDefaultDir + "/mapping.properties",
	}, {
		name: "RoleMapper config from a PersistentVolumeClaim",
		roleMapper: &api.RoleMapperAuthConfig{
			RolesProperties: "mapping.properties",
			From: &corev1.ObjectReference{
				Name: "test-pvc",
				Kind: "PersistentVolumeClaim",
			},
		},
		expectedVolumeMount: &corev1.VolumeMount{
			Name:      constants.RoleMapperVolume,
			MountPath: constants.RoleMapperDefaultDir,
			ReadOnly:  true,
		},
		expectedVolume: &corev1.Volume{
			Name: constants.RoleMapperVolume,
			VolumeSource: corev1.VolumeSource{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
					ClaimName: "test-pvc",
				},
			},
		},
		expectedPath: constants.RoleMapperDefaultDir + "/mapping.properties",
	}, {
		name: "RoleMapper config is mounted on a different path",
		roleMapper: &api.RoleMapperAuthConfig{
			RolesProperties: "/other/path/mapping.properties",
			From: &corev1.ObjectReference{
				Name: "test-cm",
				Kind: "ConfigMap",
			},
		},
		expectedVolumeMount: &corev1.VolumeMount{
			Name:      constants.RoleMapperVolume,
			MountPath: "/other/path",
			ReadOnly:  true,
		},
		expectedVolume: &corev1.Volume{
			Name: constants.RoleMapperVolume,
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: "test-cm",
					},
					DefaultMode: &defaultMode,
				},
			},
		},
		expectedPath: "/other/path/mapping.properties",
	}}

	cr := &api.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: api.KieAppSpec{
			Environment: "rhpam-trial",
			Objects: api.KieAppObjects{
				Servers: []api.KieServerSet{
					{Deployments: Pint(2)},
				},
			},
			Auth: api.KieAppAuthObject{
				LDAP: &api.LDAPAuthConfig{
					URL:    "ldaps://ldap.example.com",
					BindDN: "cn=admin,dc=example,dc=com",
				},
			},
		},
	}

	for _, item := range tests {
		cr.Spec.Auth.RoleMapper = item.roleMapper
		env, err := GetEnvironment(cr, test.MockService())
		assert.Nil(t, err, "Error getting trial environment")

		expectedEnvs := []corev1.EnvVar{
			{Name: "AUTH_LDAP_URL", Value: "ldaps://ldap.example.com"},
			{Name: "AUTH_LDAP_BIND_DN", Value: "cn=admin,dc=example,dc=com"},
			{Name: "AUTH_LDAP_BIND_CREDENTIAL"},
			{Name: "AUTH_ROLE_MAPPER_ROLES_PROPERTIES", Value: item.expectedPath},
			{Name: "AUTH_ROLE_MAPPER_REPLACE_ROLE", Value: strconv.FormatBool(item.roleMapper.ReplaceRole)},
		}
		for _, expectedEnv := range expectedEnvs {
			assert.Containsf(t, env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env, expectedEnv, "Test %s - Console does not contain env %v", item.name, expectedEnv)
			if item.expectedVolume != nil {
				assert.Containsf(t, env.Console.DeploymentConfigs[0].Spec.Template.Spec.Volumes, *item.expectedVolume, "Test %s failed", item.name)
			}
			if item.expectedVolumeMount != nil {
				assert.Containsf(t, env.Console.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].VolumeMounts, *item.expectedVolumeMount, "Test %s failed", item.name)
			}

			for i := range env.Servers {
				assert.Containsf(t, env.Servers[i].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env, expectedEnv, "Test %s - Server %v does not contain env %v", item.name, i, expectedEnv)
				if item.expectedVolume != nil {
					assert.Containsf(t, env.Servers[i].DeploymentConfigs[0].Spec.Template.Spec.Volumes, *item.expectedVolume, "Test %s failed", item.name)
				}
				if item.expectedVolumeMount != nil {
					assert.Containsf(t, env.Servers[i].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].VolumeMounts, *item.expectedVolumeMount, "Test %s failed", item.name)
				}
			}
		}
	}
}
