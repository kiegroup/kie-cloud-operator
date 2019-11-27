export const BACKEND_URL = "/api";
export const SMART_ROUTER_NA_TITLE =
  "Smart Router is not applicable for RHDM environment.";
export const ENV_KEY = "env";
export const ENV_FIELD = "Environment";
export const INSTALLATION_STEP = "Installation";
export const SMART_ROUTER_STEP = "Smart Router";
export const RHDM_ENV_PREFIX = "rhdm";
export const CONSOLE_STEP = "Console";
export const SECURITY_STEP = "Security";
export const GITHOOKS_FIELD = "GitHooks";
export const GITHOOKS_FROM_FIELD = "$.spec.objects.console.gitHooks.from";
export const KIND_FIELD = "Kind";
export const NAME_FIELD = "Name";
export const GITHOOKS_KIND_KEY = "GitHooks_Kind";
export const ROLEMAPPER_KIND_KEY = "RoleMapper_Kind";
export const SECURITY_NAME_JSONPATH = "$.spec.auth.roleMapper.from.name";
export const CONSOLE_NAME_JSONPATH =
  "$.spec.objects.console.gitHooks.from.name";
export const GITHOOKS_ENVS = [
  "rhpam-trial",
  "rhdm-trial",
  "rhpam-authoring",
  "rhdm-authoring",
  "rhpam-authoring-ha",
  "rhdm-authoring-ha"
];
export const GITHOOKS_ERR_MSG =
  "Name is mandatory, if Kind is not empty for GitHooks.";
export const ROLEMAPPER_ERR_MSG =
  "Name is mandatory, if Kind is not empty for RoleMapper.";
