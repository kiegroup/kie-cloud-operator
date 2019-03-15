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

func configureAuth(cr *v1.KieApp, envTemplate *v1.EnvTemplate) (err error) {
	if cr.Spec.Auth.SSO == nil && cr.Spec.Auth.LDAP == nil && cr.Spec.Auth.RoleMapper == nil {
		return
	}
	if cr.Spec.Auth.SSO != nil && cr.Spec.Auth.LDAP != nil {
		err = errors.New("multiple authentication types not supported")
	} else if cr.Spec.Auth.SSO == nil && cr.Spec.Auth.LDAP == nil && cr.Spec.Auth.RoleMapper != nil {
		err = errors.New("roleMapper configuration must be declared together with SSO or LDAP")
	} else if cr.Spec.Auth.SSO != nil {
		err = configureSSO(cr, envTemplate)
	} else if cr.Spec.Auth.LDAP != nil {
		err = configureLDAP(cr.Spec.Auth.LDAP, envTemplate)
	}
	if cr.Spec.Auth.RoleMapper != nil {
		configureRoleMapper(cr.Spec.Auth.RoleMapper, envTemplate)
	}
	return
}

func configureSSO(cr *v1.KieApp, envTemplate *v1.EnvTemplate) error {
	if len(cr.Spec.Auth.SSO.URL) == 0 || len(cr.Spec.Auth.SSO.Realm) == 0 {
		return errors.New("neither url nor realm can be empty")
	}
	// Set defaults
	if len(cr.Spec.Auth.SSO.PrincipalAttribute) == 0 {
		cr.Spec.Auth.SSO.PrincipalAttribute = constants.SSODefaultPrincipalAttribute
	}
	envTemplate.Auth.SSO = *cr.Spec.Auth.SSO.DeepCopy()
	if cr.Spec.Objects.Console.SSOClient != nil {
		envTemplate.Console.SSOAuthClient = *cr.Spec.Objects.Console.SSOClient.DeepCopy()
	}
	if cr.Spec.Auth.SSO != nil {
		for index := range envTemplate.Servers {
			serverSet, _ := GetServerSet(cr, index)
			if serverSet.SSOClient != nil {
				envTemplate.Servers[index].SSOAuthClient = *serverSet.SSOClient.DeepCopy()
			}
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
