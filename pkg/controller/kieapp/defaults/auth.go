package defaults

import (
	v1 "github.com/kiegroup/kie-cloud-operator/pkg/apis/app/v1"
	"github.com/kiegroup/kie-cloud-operator/pkg/controller/kieapp/constants"
	"github.com/kiegroup/kie-cloud-operator/pkg/controller/kieapp/shared"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
)

const ssoHostnameVar = "HOSTNAME_HTTPS"
const ssoClientVar = "SSO_CLIENT"

// ConfigureHostname sets the HOSTNAME_HTTPS environment variable with the provided hostname
// IF not yet set AND SSO auth is configured AND SSO_CLIENT exists
func ConfigureHostname(object *v1.CustomObject, cr *v1.KieApp, hostname string) {
	if cr.Spec.Auth.SSO == nil {
		return
	}
	for dcIdx := range object.DeploymentConfigs {
		dc := &object.DeploymentConfigs[dcIdx]
		for containerIdx := range dc.Spec.Template.Spec.Containers {
			container := &dc.Spec.Template.Spec.Containers[containerIdx]
			if pos := shared.GetEnvVar(ssoClientVar, container.Env); pos == -1 {
				continue
			}
			if pos := shared.GetEnvVar(ssoHostnameVar, container.Env); pos == -1 {
				container.Env = append(container.Env, corev1.EnvVar{
					Name:  ssoHostnameVar,
					Value: hostname,
				})
			} else if len(container.Env[pos].Value) == 0 {
				container.Env[pos].Value = hostname
			}
		}
	}
}

func configureAuth(spec v1.KieAppSpec, envTemplate *v1.EnvTemplate) (err error) {
	if spec.Auth.SSO == nil && spec.Auth.LDAP == nil && spec.Auth.RoleMapper == nil {
		return
	}
	if spec.Auth.SSO != nil && spec.Auth.LDAP != nil {
		err = errors.New("multiple authentication types not supported")
	} else if spec.Auth.SSO == nil && spec.Auth.LDAP == nil && spec.Auth.RoleMapper != nil {
		err = errors.New("roleMapper configuration must be declared together with SSO or LDAP")
	} else if spec.Auth.SSO != nil {
		err = configureSSO(spec.Auth.SSO, envTemplate)
	} else if spec.Auth.LDAP != nil {
		err = configureLDAP(spec.Auth.LDAP, envTemplate)
	}
	if spec.Auth.RoleMapper != nil {
		configureRoleMapper(spec.Auth.RoleMapper, envTemplate)
	}
	return
}

func configureSSO(config *v1.SSOAuthConfig, envTemplate *v1.EnvTemplate) error {
	if len(config.URL) == 0 || len(config.Realm) == 0 {
		return errors.New("neither url nor realm can be empty")
	}
	if len(config.Clients.Servers) != 0 && len(config.Clients.Servers) != len(envTemplate.Servers) {
		return errors.New("the number of Server SSO clients defined must match the number of KIE Servers")
	}
	// Set defaults
	if len(config.PrincipalAttribute) == 0 {
		config.PrincipalAttribute = constants.SSODefaultPrincipalAttribute
	}
	envTemplate.Auth.SSO = *config.DeepCopy()
	envTemplate.Console.SSOAuthClient = *config.Clients.Console.DeepCopy()
	if len(config.Clients.Servers) > 0 {
		for i := range envTemplate.Servers {
			envTemplate.Servers[i].SSOAuthClient = *config.Clients.Servers[i].DeepCopy()
		}
	}

	return nil
}

func configureLDAP(config *v1.LDAPAuthConfig, envTemplate *v1.EnvTemplate) error {
	if len(config.URL) == 0 {
		return errors.New("the url must not be empty")
	}
	envTemplate.Auth.LDAP = *config.DeepCopy()
	return nil
}

func configureRoleMapper(config *v1.RoleMapperAuthConfig, envTemplate *v1.EnvTemplate) {
	if config != nil {
		envTemplate.Auth.RoleMapper = *config.DeepCopy()
	}
}
