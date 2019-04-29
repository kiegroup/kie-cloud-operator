export const MockupData_JSON = {
  pages: [
    {
      fields: [
        {
          label: "Application Name",
          default: "rhpam-trial",
          required: true,
          jsonPath: "$.metadata.name",
          type: "text"
        },
        {
          label: "Environment",
          default: "env2 default",
          required: true,
          description: "The name of the environment used as a baseline",
          jsonPath: "$.spec.environment",

          type: "dropDown"
        },
        {
          label: "Image Registry",
          default: "rhpam-trial",
          required: false,
          jsonPath: "$.spec.imageRegistry.registry",
          type: "text"
        },
        {
          label: "Insecure",
          default: false,
          required: false,
          jsonPath: "$.spec.imageRegistry.insecure",
          type: "checkbox"
        },
        {
          label: "User Name",
          default: "rhpam-trial",
          required: false,
          jsonPath: "$.spec.commonConfig.adminUser",
          type: "text"
        },
        {
          label: "Password",
          default: "env2 default",
          required: false,
          jsonPath: "$.spec.commonConfig.adminPassword",
          type: "password"
        }
      ],
      buttons: [
        {
          label: "Edit Yaml",
          action: "submit"
        },
        {
          label: "Cancel",
          action: "close"
        },
        {
          label: "Deploy",
          action: "submit"
        },

        {
          label: "Customize",
          action: "next"
        }
      ]
    },
    {
      fields: [
        {
          label: "SSO",
          type: "checkbox",
          jsonPath: "$.spec.auth.sso",
          default: "false"
        },
        {
          label: "url",
          type: "text",
          jsonPath: "$.spec.auth.sso.url",
          default: "",
          description: "RH-SSO URL"
        },
        {
          label: "realm",
          type: "text",
          jsonPath: "$.spec.auth.sso.realm",
          default: "",
          description: "RH-SSO Realm name"
        },
        {
          label: "Admin User",
          type: "text",
          jsonPath: "$.spec.auth.sso.adminuser",
          default: "",
          description:
            "RH-SSO Realm Admin Username used to create the Client if it doesn't exist"
        },
        {
          label: "Admin Password",
          type: "password",
          jsonPath: "$.spec.auth.sso.adminPassword",
          default: "",
          description: "RH-SSO Realm Admin Password used to create the Client"
        },
        {
          label: "Disable SSL Cert Validation",
          type: "checkbox",
          jsonPath: "$.spec.auth.sso.disableSSLCertValidation",
          default: "false",
          description: "RH-SSO Disable SSL Certificate Validation"
        },
        {
          label: "Principal Attribute",
          type: "text",
          jsonPath: "$.spec.auth.sso.principalAttribute",
          default: "",
          description: "RH-SSO Principal Attribute to use as username"
        },
        {
          label: "LDAP",
          type: "checkbox",
          jsonPath: "$.spec.auth.ldap",
          default: "false"
        },
        {
          label: "url",
          type: "text",
          jsonPath: "$.spec.auth.ldap.url",
          default: "",
          description: " LDAP Endpoint to connect for authentication"
        },
        {
          label: "bindDN",
          type: "text",
          jsonPath: "$.spec.auth.ldap.bindDN",
          default: "",
          description: "Bind DN used for authentication"
        },
        {
          label: "bind Credential",
          type: "password",
          jsonPath: "$.spec.auth.ldap.bindCredential",
          default: "",
          description: "LDAP Credentials used for authentication"
        },
        {
          label: "jaasSecurityDomain",
          type: "password",
          jsonPath: "$.spec.auth.ldap.jaasSecurityDomain",
          default: "",
          description:
            "The JMX ObjectName of the JaasSecurityDomain used to decrypt the password."
        },
        {
          label: "baseCtxDN",
          type: "checkbox",
          jsonPath: "$.spec.auth.ldap.baseCtxDN",
          default: "false",
          description:
            "LDAP Base DN of the top-level context to begin the user search."
        },
        {
          label: "baseFilter",
          type: "text",
          jsonPath: "$.spec.auth.ldap.baseFilter",
          default: "",
          description:
            "LDAP search filter used to locate the context of the user to authenticate. The input username or userDN obtained from the login module callback is substituted into the filter anywhere a {0} expression is used. A common example for the search filter is (uid={0})."
        },
        {
          label: "searchScope",
          type: "dropDown",
          jsonPath: "$.spec.auth.ldap.searchScope",
          default: "",
          description: "The search scope to use."
        },
        {
          label: "searchTimeLimit",
          type: "text",
          jsonPath: "$.spec.auth.ldap.searchTimeLimit",
          default: "",
          description: "The timeout in milliseconds for user or role searches."
        },
        {
          label: "distinguishedNameAttribute",
          type: "text",
          jsonPath: "$.spec.auth.ldap.distinguishedNameAttribute",
          default: "",
          description:
            "The name of the attribute in the user entry that contains the DN of the user. This may be necessary if the DN of the user itself contains special characters, backslash for example, that prevent correct user mapping. If the attribute does not exist, the entry’s DN is used."
        },
        {
          label: "parseUsername",
          type: "checkbox",
          jsonPath: "$.spec.auth.ldap.parseUsername",
          default: "false",
          description:
            "A flag indicating if the DN is to be parsed for the username. If set to true, the DN is parsed for the username. If set to false the DN is not parsed for the username. This option is used together with usernameBeginString and usernameEndString."
        },
        {
          label: "usernameBeginString",
          type: "text",
          jsonPath: "$.spec.auth.ldap.usernameBeginString",
          default: "",
          description:
            "Defines the String which is to be removed from the start of the DN to reveal the username. This option is used together with usernameEndString and only taken into account if parseUsername is set to true."
        },
        {
          label: "usernameBeginString",
          type: "text",
          jsonPath: "$.spec.auth.ldap.usernameBeginString",
          default: "",
          description:
            "Defines the String which is to be removed from the end of the DN to reveal the username. This option is used together with usernameBeginString and only taken into account if parseUsername is set to true."
        },
        {
          label: "roleAttributeID",
          type: "text",
          jsonPath: "$.spec.auth.ldap.roleAttributeID",
          default: "",
          description: "Name of the attribute containing the user roles"
        },
        {
          label: "rolesCtxDN",
          type: "password",
          jsonPath: "$.spec.auth.ldap.rolesCtxDN",
          default: "",
          description:
            "The fixed DN of the context to search for user roles. This is not the DN where the actual roles are, but the DN where the objects containing the user roles are. For example, in a Microsoft Active Directory server, this is the DN where the user account is."
        },
        {
          label: "roleFilter",
          type: "password",
          jsonPath: "$.spec.auth.ldap.roleFilter",
          default: "",
          description:
            "A search filter used to locate the roles associated with the authenticated user. The input username or userDN obtained from the login module callback is substituted into the filter anywhere a {0} expression is used. The authenticated userDN is substituted into the filter anywhere a {1} is used. An example search filter that matches on the input username is (member={0}). An alternative that matches on the authenticated userDN is (member={1})."
        },
        {
          label: "roleRecursion",
          type: "number",
          jsonPath: "$.spec.auth.ldap.baseCtxDN",
          default: "false",
          description:
            "The number of levels of recursion the role search will go below a matching context. Disable recursion by setting this to 0."
        },
        {
          label: "defaultRole",
          type: "text",
          jsonPath: "$.spec.auth.ldap.defaultRole",
          default: "",
          description: "A role included for all authenticated users"
        },
        {
          label: "roleNameAttributeID",
          type: "text",
          jsonPath: "$.spec.auth.ldap.roleNameAttributeID",
          default: "",
          description:
            "Name of the attribute within the roleCtxDN context which contains the role name. If the roleAttributeIsDN property is set to true, this property is used to find the role object’s name attribute."
        },
        {
          label: "parseRoleNameFromDN",
          type: "checkbox",
          jsonPath: "$.spec.auth.ldap.parseRoleNameFromDN",
          default: "false",
          description:
            "A flag indicating if the DN returned by a query contains the roleNameAttributeID. If set to true, the DN is checked for the roleNameAttributeID. If set to false, the DN is not checked for the roleNameAttributeID. This flag can improve the performance of LDAP queries."
        },
        {
          label: "roleAttributeIsDN",
          type: "checkbox",
          jsonPath: "$.spec.auth.ldap.roleAttributeIsDN",
          default: "false",
          description:
            "Whether or not the roleAttributeID contains the fully-qualified DN of a role object. If false, the role name is taken from the value of the roleNameAttributeId attribute of the context name. Certain directory schemas, such as Microsoft Active Directory, require this attribute to be set to true."
        },
        {
          label: "referralUserAttributeIDToCheck",
          type: "text",
          jsonPath: "$.spec.auth.ldap.referralUserAttributeIDToCheck",
          default: "",
          description:
            " If you are not using referrals, you can ignore this option. When using referrals, this option denotes the attribute name which contains users defined for a certain role, for example member, if the role object is inside the referral. Users are checked against the content of this attribute name. If this option is not set, the check will always fail, so role objects cannot be stored in a referral tree"
        },
        {
          label: "Roles Properties",
          type: "text",
          jsonPath: "$.spec.auth.roleMapper.rolesProperties",
          default: "",
          description:
            " When present, the RoleMapping Login Module will be configured to use the provided file. This property defines the fully-qualified file path and name of a properties file or resource which maps roles to replacement roles. The format is original_role=role1,role2,role3"
        },
        {
          label: "Replace Role",
          type: "checkbox",
          jsonPath: "$.spec.auth.roleMapper.replaceRole",
          default: "false",
          description:
            " Whether to add to the current roles, or replace the current roles with the mapped ones. Replaces if set to true."
        },

        {
          label: "version",
          type: "text",
          jsonPath: "$.spec.commonConfig.version",
          default: "",
          description: "The version of the application deployment"
        },
        {
          label: "ImageTag",
          type: "text",
          jsonPath: "$.spec.commonConfig.imageTag",
          default: "",
          description: "The tag to use for the application images."
        },
        {
          label: "keyStorePassword",
          type: "password",
          jsonPath: "$.spec.commonConfig.keyStorePassword",
          default: "",
          description: "The password to use for keystore generation."
        },
        {
          label: "DB Password",
          type: "password",
          jsonPath: "$.spec.commonConfig.dbPassword",
          default: "",
          description: "The password to use for databases."
        },
        {
          label: "amqPassword",
          type: "password",
          jsonPath: "$.spec.commonConfig.amqPassword",
          default: "",
          description: "The password to use for amq user"
        },
        {
          label: "amqClusterPassword",
          type: "password",
          jsonPath: "$.spec.commonConfig.amqClusterPassword",
          default: "",
          description: "RH-SSO Realm Admin Password used to create the Client"
        },
        {
          label: "controllerPassword",
          type: "password",
          jsonPath: "$.spec.commonConfig.controllerPassword",
          default: "",
          description: "The password to use for the controllerUser."
        },
        {
          label: "serverPassword",
          type: "password",
          jsonPath: "$.spec.commonConfig.serverPassword",
          default: "",
          description: "The password to use for the executionUser."
        },
        {
          label: "mavenPassword",
          type: "password",
          jsonPath: "$.spec.commonConfig.mavenPassword",
          default: "",
          description: "The password to use for the mavenUser."
        }
      ],
      buttons: [
        {
          label: "Deploy",
          action: "submit"
        },
        {
          label: "Edit Yaml",
          action: "submit"
        },
        {
          label: "Back",
          action: "back"
        },
        {
          label: "Next",
          action: "next"
        }
      ]
    }
  ]
};

export const MockupData_JSON_SCHEMA = {
  required: ["spec"],
  properties: {
    spec: {
      type: "object",
      required: ["environment"],
      properties: {
        auth: {
          description: "Authentication integration configuration",
          type: "object",
          properties: {
            ldap: {
              description: "LDAP integration configuration",
              type: "object",
              required: ["url"],
              properties: {
                baseCtxDN: {
                  description:
                    "LDAP Base DN of the top-level context to begin the user search.",
                  type: "string"
                },
                baseFilter: {
                  description:
                    "DAP search filter used to locate the context of the user to authenticate. The input username or userDN obtained from the login module callback is substituted into the filter anywhere a {0} expression is used. A common example for the search filter is (uid={0}).",
                  type: "string"
                },
                bindCredential: {
                  description: "LDAP Credentials used for authentication",
                  type: "string",
                  format: "password"
                },
                bindDN: {
                  description: "Bind DN used for authentication",
                  type: "string"
                },
                defaultRole: {
                  description: "A role included for all authenticated users",
                  type: "string"
                },
                distinguishedNameAttribute: {
                  description:
                    "The name of the attribute in the user entry that contains the DN of the user. This may be necessary if the DN of the user itself contains special characters, backslash for example, that prevent correct user mapping. If the attribute does not exist, the entry’s DN is used.",
                  type: "string"
                },
                jaasSecurityDomain: {
                  description:
                    "The JMX ObjectName of the JaasSecurityDomain used to decrypt the password.",
                  type: "string"
                },
                parseRoleNameFromDN: {
                  description:
                    "A flag indicating if the DN returned by a query contains the roleNameAttributeID. If set to true, the DN is checked for the roleNameAttributeID. If set to false, the DN is not checked for the roleNameAttributeID. This flag can improve the performance of LDAP queries.",
                  type: "boolean"
                },
                parseUsername: {
                  description:
                    "A flag indicating if the DN is to be parsed for the username. If set to true, the DN is parsed for the username. If set to false the DN is not parsed for the username. This option is used together with usernameBeginString and usernameEndString.",
                  type: "boolean"
                },
                referralUserAttributeIDToCheck: {
                  description:
                    "If you are not using referrals, you can ignore this option. When using referrals, this option denotes the attribute name which contains users defined for a certain role, for example member, if the role object is inside the referral. Users are checked against the content of this attribute name. If this option is not set, the check will always fail, so role objects cannot be stored in a referral tree.",
                  type: "string"
                },
                roleAttributeID: {
                  description:
                    "Name of the attribute containing the user roles.",
                  type: "string"
                },
                roleAttributeIsDN: {
                  description:
                    "Whether or not the roleAttributeID contains the fully-qualified DN of a role object. If false, the role name is taken from the value of the roleNameAttributeId attribute of the context name. Certain directory schemas, such as Microsoft Active Directory, require this attribute to be set to true.",
                  type: "boolean"
                },
                roleFilter: {
                  description:
                    "A search filter used to locate the roles associated with the authenticated user. The input username or userDN obtained from the login module callback is substituted into the filter anywhere a {0} expression is used. The authenticated userDN is substituted into the filter anywhere a {1} is used. An example search filter that matches on the input username is (member={0}). An alternative that matches on the authenticated userDN is (member={1}).",
                  type: "string"
                },
                roleNameAttributeID: {
                  description:
                    "Name of the attribute within the roleCtxDN context which contains the role name. If the roleAttributeIsDN property is set to true, this property is used to find the role object’s name attribute.",
                  type: "string"
                },
                roleRecursion: {
                  description:
                    "The number of levels of recursion the role search will go below a matching context. Disable recursion by setting this to 0.",
                  type: "integer",
                  format: "int16"
                },
                rolesCtxDN: {
                  description:
                    "The fixed DN of the context to search for user roles. This is not the DN where the actual roles are, but the DN where the objects containing the user roles are. For example, in a Microsoft Active Directory server, this is the DN where the user account is.",
                  type: "string"
                },
                searchScope: {
                  description: "The search scope to use.",
                  type: "string",
                  enum: ["SUBTREE_SCOPE", "OBJECT_SCOPE", "ONELEVEL_SCOPE"]
                },
                searchTimeLimit: {
                  description:
                    "The timeout in milliseconds for user or role searches.",
                  type: "integer",
                  format: "int32"
                },
                url: {
                  description: "LDAP Endpoint to connect for authentication",
                  type: "string"
                },
                usernameBeginString: {
                  description:
                    "Defines the String which is to be removed from the start of the DN to reveal the username. This option is used together with usernameEndString and only taken into account if parseUsername is set to true.",
                  type: "string"
                },
                usernameEndString: {
                  description:
                    "Defines the String which is to be removed from the end of the DN to reveal the username. This option is used together with usernameBeginString and only taken into account if parseUsername is set to true.",
                  type: "string"
                }
              }
            },
            roleMapper: {
              description: "RoleMapper configuration",
              type: "object",
              required: ["rolesProperties"],
              properties: {
                replaceRole: {
                  description:
                    "Whether to add to the current roles, or replace the current roles with the mapped ones. Replaces if set to true.",
                  type: "boolean"
                },
                rolesProperties: {
                  description:
                    "When present, the RoleMapping Login Module will be configured to use the provided file. This property defines the fully-qualified file path and name of a properties file or resource which maps roles to replacement roles. The format is original_role=role1,role2,role3",
                  type: "string"
                }
              }
            },
            sso: {
              description: "RH-SSO integration configuration",
              type: "object",
              required: ["url", "realm"],
              properties: {
                adminPassword: {
                  description:
                    "RH-SSO Realm Admin Password used to create the Client",
                  type: "string",
                  format: "password"
                },
                adminUser: {
                  description:
                    "RH-SSO Realm Admin Username used to create the Client if it doesn't exist",
                  type: "string"
                },
                disableSSLCertValidation: {
                  description: "RH-SSO Disable SSL Certificate Validation",
                  type: "boolean"
                },
                principalAttribute: {
                  description: "RH-SSO Principal Attribute to use as username",
                  type: "string"
                },
                realm: {
                  description: "RH-SSO Realm name",
                  type: "string"
                },
                url: {
                  description: "RH-SSO URL",
                  type: "string"
                }
              }
            }
          }
        },
        environment: {
          description: "The name of the environment used as a baseline",
          type: "string",
          enum: [
            "rhdm-authoring-ha",
            "rhdm-authoring",
            "rhdm-optaweb-trial",
            "rhdm-production-immutable",
            "rhdm-trial",
            "rhpam-authoring-ha",
            "rhpam-authoring",
            "rhpam-production-immutable",
            "rhpam-production",
            "rhpam-trial"
          ]
        },
        imageRegistry: {
          description:
            "If required imagestreams are missing in both the 'openshift' and local namespaces, the operator will create said imagestreams locally using the registry specified here.",
          type: "object",
          properties: {
            insecure: {
              description:
                "A flag used to indicate the specified registry is insecure. Defaults to 'false'.",
              type: "boolean"
            },
            registry: {
              description:
                "Image registry's base 'url:port'. e.g. registry.example.com:5000. Defaults to 'registry.redhat.io'.",
              type: "string"
            }
          }
        },
        objects: {
          description: "Configuration of the RHPAM components",
          type: "object",
          properties: {
            console: {
              description: "Configuration of the RHPAM workbench",
              type: "object",
              properties: {
                env: {
                  type: "array",
                  items: {
                    type: "object",
                    required: ["name"],
                    oneOf: [
                      {
                        required: ["value"]
                      },
                      {
                        required: ["valueFrom"]
                      }
                    ],
                    properties: {
                      name: {
                        description: "Name of an environment variable",
                        type: "string"
                      },
                      value: {
                        description: "Value for that environment variable",
                        type: "string"
                      },
                      valueFrom: {
                        description:
                          "Source for the environment variable's value",
                        type: "object"
                      }
                    }
                  }
                },
                keystoreSecret: {
                  description: "Keystore secret name",
                  type: "string"
                },
                resources: {
                  type: "object",
                  properties: {
                    limits: {
                      type: "object"
                    },
                    requests: {
                      type: "object"
                    }
                  }
                },
                ssoClient: {
                  description:
                    "Client definitions used for creating the RH-SSO clients in the specified Realm",
                  type: "object",
                  properties: {
                    hostnameHTTP: {
                      description: "Hostname to set as redirect URL",
                      type: "string"
                    },
                    hostnameHTTPS: {
                      description: "Secure hostname to set as redirect URL",
                      type: "string"
                    },
                    name: {
                      description: "Client name",
                      type: "string"
                    },
                    secret: {
                      description: "Client secret",
                      type: "string",
                      format: "password"
                    }
                  }
                }
              }
            },
            servers: {
              description: "Configuration of the each individual KIE server",
              type: "array",
              minItems: 1,
              items: {
                description: "KIE Server configuration",
                type: "object",
                properties: {
                  build: {
                    description:
                      "Configuration of build configs for immutable KIE servers",
                    type: "object",
                    required: ["kieServerContainerDeployment", "gitSource"],
                    properties: {
                      artifactDir: {
                        description:
                          "List of directories from which archives will be copied into the deployment folder. If unspecified, all archives in /target will be copied.",
                        type: "string"
                      },
                      from: {
                        description:
                          "Image definition to use for all the servers",
                        type: "object",
                        required: ["kind", "name"],
                        properties: {
                          kind: {
                            description: "Object kind. e.g. ImageStreamTag",
                            type: "string"
                          },
                          name: {
                            description: "Object name",
                            type: "string"
                          },
                          namespace: {
                            description:
                              "Namespace where the object is located",
                            type: "string"
                          }
                        }
                      },
                      gitSource: {
                        type: "object",
                        required: ["uri", "reference"],
                        properties: {
                          contextDir: {
                            description:
                              "Context/subdirectory where the code is located, relatively to repo root",
                            type: "string"
                          },
                          reference: {
                            description: "Branch to use in the git repository",
                            type: "string"
                          },
                          uri: {
                            description: "Git URI for the s2i source",
                            type: "string"
                          }
                        }
                      },
                      kieServerContainerDeployment: {
                        description:
                          "The Maven GAV to deploy, e.g., rhpam-kieserver-library=org.openshift.quickstarts:rhpam-kieserver-library:1.4.0-SNAPSHOT",
                        type: "string"
                      },
                      mavenMirrorURL: {
                        description: "Maven mirror to use for S2I builds",
                        type: "string"
                      },
                      webhooks: {
                        type: "array",
                        minItems: 1,
                        items: {
                          description: "WebHook secretes for build configs",
                          type: "object",
                          required: ["type", "secret"],
                          properties: {
                            secret: {
                              description: "Secret value for webhook",
                              type: "string"
                            },
                            type: {
                              description:
                                "WebHook type, either GitHub or Generic",
                              type: "string",
                              enum: ["GitHub", "Generic"]
                            }
                          }
                        }
                      }
                    }
                  },
                  deployments: {
                    description: "Number of Server sets that will be deployed",
                    type: "integer",
                    format: "int"
                  },
                  env: {
                    type: "array",
                    items: {
                      type: "object",
                      required: ["name"],
                      oneOf: [
                        {
                          required: ["value"]
                        },
                        {
                          required: ["valueFrom"]
                        }
                      ],
                      properties: {
                        name: {
                          description: "Name of an environment variable",
                          type: "string"
                        },
                        value: {
                          description: "Value for that environment variable",
                          type: "string"
                        },
                        valueFrom: {
                          description:
                            "Source for the environment variable's value",
                          type: "object"
                        }
                      }
                    }
                  },
                  from: {
                    description: "Image definition to use for all the servers",
                    type: "object",
                    required: ["kind", "name"],
                    properties: {
                      kind: {
                        description: "Object kind",
                        type: "string",
                        enum: ["ImageStreamTag", "DockerImage"]
                      },
                      name: {
                        description: "Object name",
                        type: "string"
                      },
                      namespace: {
                        description: "Namespace where the object is located",
                        type: "string"
                      }
                    }
                  },
                  keystoreSecret: {
                    description: "Keystore secret name",
                    type: "string"
                  },
                  name: {
                    description: "Server name",
                    type: "string"
                  },
                  resources: {
                    type: "object",
                    properties: {
                      limits: {
                        type: "object"
                      },
                      requests: {
                        type: "object"
                      }
                    }
                  },
                  ssoClient: {
                    description:
                      "Client definitions used for creating the RH-SSO clients in the specified Realm",
                    type: "object",
                    properties: {
                      hostnameHTTP: {
                        description: "Hostname to set as redirect URL",
                        type: "string"
                      },
                      hostnameHTTPS: {
                        description: "Secure hostname to set as redirect URL",
                        type: "string"
                      },
                      name: {
                        description: "Client name",
                        type: "string"
                      },
                      secret: {
                        description: "Client secret",
                        type: "string",
                        format: "password"
                      }
                    }
                  }
                }
              }
            },
            smartRouter: {
              description: "Configuration of the RHPAM smart router",
              type: "object",
              properties: {
                env: {
                  type: "array",
                  items: {
                    type: "object",
                    required: ["name"],
                    oneOf: [
                      {
                        required: ["value"]
                      },
                      {
                        required: ["valueFrom"]
                      }
                    ],
                    properties: {
                      name: {
                        description: "Name of an environment variable",
                        type: "string"
                      },
                      value: {
                        description: "Value for that environment variable",
                        type: "string"
                      },
                      valueFrom: {
                        description:
                          "Source for the environment variable's value",
                        type: "object"
                      }
                    }
                  }
                },
                keystoreSecret: {
                  description: "Keystore secret name",
                  type: "string"
                },
                resources: {
                  type: "object",
                  properties: {
                    limits: {
                      type: "object"
                    },
                    requests: {
                      type: "object"
                    }
                  }
                }
              }
            }
          }
        }
      }
    }
  }
};
