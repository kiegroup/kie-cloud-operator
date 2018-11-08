package defaults

type EnvTemplate struct {
	Template    `json:",inline"`
	ServerCount []Template `json:"serverCount,omitempty"`
}

type Template struct {
	ApplicationName    string `json:"applicationName,omitempty"`
	Version            string `json:"version,omitempty"`
	ImageTag           string `json:"imageTag,omitempty"`
	KeyStorePassword   string `json:"keyStorePassword,omitempty"`
	AdminPassword      string `json:"adminPassword,omitempty"`
	ControllerPassword string `json:"controllerPassword,omitempty"`
	ServerPassword     string `json:"serverPassword,omitempty"`
	MavenPassword      string `json:"mavenPassword,omitempty"`
}
