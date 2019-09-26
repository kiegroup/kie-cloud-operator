package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/blang/semver"
	api "github.com/kiegroup/kie-cloud-operator/pkg/apis/app/v2"
	"github.com/kiegroup/kie-cloud-operator/pkg/components"
	"github.com/kiegroup/kie-cloud-operator/pkg/controller/kieapp/constants"
	"github.com/kiegroup/kie-cloud-operator/pkg/controller/kieapp/defaults"
	"github.com/kiegroup/kie-cloud-operator/tools/util"
	"github.com/kiegroup/kie-cloud-operator/version"
	oappsv1 "github.com/openshift/api/apps/v1"
	buildv1 "github.com/openshift/api/build/v1"
	oimagev1 "github.com/openshift/api/image/v1"
	routev1 "github.com/openshift/api/route/v1"
	csvv1 "github.com/operator-framework/operator-lifecycle-manager/pkg/api/apis/operators/v1alpha1"
	olmversion "github.com/operator-framework/operator-lifecycle-manager/pkg/lib/version"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	replacesCsvVersion = "1.2.0"
	rh                 = "Red Hat, Inc."
	maturity           = "stable"
	major, minor, _    = defaults.MajorMinorMicro(constants.CurrentVersion)
	csvs               = []csvSetting{
		{
			Name:         "kiecloud",
			DisplayName:  "Kie Cloud",
			OperatorName: "kie-cloud-operator",
			CsvDir:       "community",
			Registry:     "quay.io",
			Context:      "kiegroup",
			ImageName:    "kie-cloud-operator",
			Tag:          version.Version,
		},
		{
			Name:         "businessautomation",
			DisplayName:  "Business Automation",
			OperatorName: "business-automation-operator",
			CsvDir:       "redhat",
			Registry:     "registry.redhat.io",
			Context:      "rhpam-" + major,
			ImageName:    "rhpam" + major + minor + "-operator",
			Tag:          constants.VersionConstants[constants.CurrentVersion].ImageTag,
		},
	}
)

type csvSetting struct {
	Name         string `json:"name"`
	DisplayName  string `json:"displayName"`
	OperatorName string `json:"operatorName"`
	CsvDir       string `json:"csvDir"`
	Registry     string `json:"repository"`
	Context      string `json:"context"`
	ImageName    string `json:"imageName"`
	Tag          string `json:"tag"`
}
type csvPermissions struct {
	ServiceAccountName string              `json:"serviceAccountName"`
	Rules              []rbacv1.PolicyRule `json:"rules"`
}
type csvDeployments struct {
	Name string                `json:"name"`
	Spec appsv1.DeploymentSpec `json:"spec,omitempty"`
}
type csvStrategySpec struct {
	Permissions []csvPermissions `json:"permissions"`
	Deployments []csvDeployments `json:"deployments"`
}
type channel struct {
	Name       string `json:"name"`
	CurrentCSV string `json:"currentCSV"`
}
type packageStruct struct {
	PackageName string    `json:"packageName"`
	Channels    []channel `json:"channels"`
}

func main() {
	for _, csv := range csvs {
		operatorName := csv.Name + "-operator"
		templateStruct := &csvv1.ClusterServiceVersion{}
		templateStruct.SetGroupVersionKind(csvv1.SchemeGroupVersion.WithKind("ClusterServiceVersion"))
		csvStruct := &csvv1.ClusterServiceVersion{}
		strategySpec := &csvStrategySpec{}
		json.Unmarshal(csvStruct.Spec.InstallStrategy.StrategySpecRaw, strategySpec)
		permissions := strategySpec.Permissions

		templateStrategySpec := &csvStrategySpec{}
		deployment := components.GetDeployment(csv.OperatorName, csv.Registry, csv.Context, csv.ImageName, csv.Tag, "Always")
		templateStrategySpec.Deployments = append(templateStrategySpec.Deployments, []csvDeployments{{Name: csv.OperatorName, Spec: deployment.Spec}}...)
		role := components.GetRole(csv.OperatorName)
		templateStrategySpec.Permissions = append(templateStrategySpec.Permissions, []csvPermissions{{ServiceAccountName: deployment.Spec.Template.Spec.ServiceAccountName, Rules: role.Rules}}...)
		templateStrategySpec.Permissions = append(templateStrategySpec.Permissions, permissions...)
		// Re-serialize deployments and permissions into csv strategy.
		updatedStrat, err := json.Marshal(templateStrategySpec)
		if err != nil {
			panic(err)
		}
		templateStruct.Spec.InstallStrategy.StrategySpecRaw = updatedStrat
		templateStruct.Spec.InstallStrategy.StrategyName = "deployment"
		csvVersionedName := operatorName + "." + version.Version
		templateStruct.Name = csvVersionedName
		templateStruct.Namespace = "placeholder"
		descrip := csv.DisplayName + " Operator for deployment and management of RHPAM/RHDM environments."
		repository := "https://github.com/kiegroup/kie-cloud-operator"
		examples := []string{"{\x22apiVersion\x22:\x22app.kiegroup.org/v2\x22,\x22kind\x22:\x22KieApp\x22,\x22metadata\x22:{\x22name\x22:\x22rhpam-trial\x22},\x22spec\x22:{\x22environment\x22:\x22rhpam-trial\x22}}"}
		templateStruct.SetAnnotations(
			map[string]string{
				"createdAt":           time.Now().Format("2006-01-02 15:04:05"),
				"containerImage":      deployment.Spec.Template.Spec.Containers[0].Image,
				"description":         descrip,
				"categories":          "Integration & Delivery",
				"certified":           "true",
				"capabilities":        "Seamless Upgrades",
				"repository":          repository,
				"support":             rh,
				"tectonic-visibility": "ocs",
				"alm-examples":        "[" + strings.Join(examples, ",") + "]",
			},
		)
		templateStruct.SetLabels(
			map[string]string{
				"operator-" + csv.Name: "true",
			},
		)
		templateStruct.Spec.Keywords = []string{"kieapp", "pam", "decision", "kie", "cloud", "bpm", "process", "automation", "operator"}
		var opVersion olmversion.OperatorVersion
		opVersion.Version = semver.MustParse(version.Version)
		templateStruct.Spec.Version = opVersion
		templateStruct.Spec.Replaces = operatorName + "." + replacesCsvVersion
		templateStruct.Spec.Description = descrip
		templateStruct.Spec.DisplayName = csv.DisplayName
		templateStruct.Spec.Maturity = maturity
		templateStruct.Spec.Maintainers = []csvv1.Maintainer{{Name: rh, Email: "bsig-cloud@redhat.com"}}
		templateStruct.Spec.Provider = csvv1.AppLink{Name: rh}
		templateStruct.Spec.Links = []csvv1.AppLink{
			{Name: "Product Page", URL: "https://access.redhat.com/products/red-hat-process-automation-manager"},
			{Name: "Documentation", URL: "https://access.redhat.com/documentation/en-us/red_hat_process_automation_manager/" + major + "." + minor + "/#category-deploying-red-hat-process-automation-manager-on-openshift"},
		}
		templateStruct.Spec.Icon = []csvv1.Icon{
			{
				Data:      "PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHZpZXdCb3g9IjAgMCAxMDAgMTAwIj48ZGVmcz48c3R5bGU+LmNscy0xe2ZpbGw6I2Q3MWUwMH0uY2xzLTJ7ZmlsbDojYzIxYTAwfS5jbHMtM3tmaWxsOiNmZmZ9LmNscy00e2ZpbGw6I2VhZWFlYX0uY2xzLTV7ZmlsbDojYjdiN2I3fS5jbHMtNntmaWxsOiNjZGNkY2R9PC9zdHlsZT48L2RlZnM+PHRpdGxlPkxvZ288L3RpdGxlPjxnIGlkPSJMYXllcl8xIiBkYXRhLW5hbWU9IkxheWVyIDEiPjxjaXJjbGUgY2xhc3M9ImNscy0xIiBjeD0iNTAiIGN5PSI1MCIgcj0iNTAiIHRyYW5zZm9ybT0icm90YXRlKC00NSA1MCA1MCkiLz48cGF0aCBjbGFzcz0iY2xzLTIiIGQ9Ik04NS4zNiAxNC42NGE1MCA1MCAwIDAgMS03MC43MiA3MC43MnoiLz48cGF0aCBjbGFzcz0iY2xzLTMiIGQ9Ik02NS43NiAzNC4yOEwxNS42IDQzLjE1djEuMTNhLjM0LjM0IDAgMCAwIC4zLjM0YzEuNDcuMTcgNy45MyAyLjExIDggMjMuNDlhLjQ2LjQ2IDAgMCAwIC4zNS40NGwyLjU5LjU3cy0xLjIxLTI1LjU0IDguNzctMjcuMDYgMTEuMiAyNy4yNyAxMS4zMyAzMS4xYS41NC41NCAwIDAgMCAuNDMuNTFsMy41MS43OHMuMDYtMzQuNTQgMTQuOTItMzYuODJ2LTMuMzV6Ii8+PHBhdGggY2xhc3M9ImNscy00IiBkPSJNNjUuMzUgMjcuNTZMMTYuMTggMzguNDJhLjc1Ljc1IDAgMCAwLS41OS43M3Y0bDUwLjE3LTguODd2LTYuNzZhMS42OCAxLjY4IDAgMCAwLS40MS4wNHoiLz48cGF0aCBjbGFzcz0iY2xzLTUiIGQ9Ik0zNS42MSA0Mi4wNWMtNC42MS43LTYuODMgNi41NC03Ljg5IDEyLjYxbDEzLjY1LTEuMzNjMC0uMTcuMDktLjM0LjEzLS41MXMuMTQtLjUzLjIxLS44bC4yLS42OHEuMTItLjQuMjUtLjhsLjItLjYyYy4xMi0uMzYuMjUtLjcxLjM5LTEuMDZsLjEyLS4zMmMtMS42NC00LjE3LTMuOTgtNi45OS03LjI2LTYuNDl6TTgyLjIzIDMxLjE5bC0xNi0zLjYyYTEuOSAxLjkgMCAwIDAtLjQyIDB2Ni43NmwxNy4wNiAyLjgzdi01LjIzYS43Ni43NiAwIDAgMC0uNjQtLjc0ek01My40MyA1My42MmwxOC40MS0xLjEzYzIuMS02LjA1IDUuNTEtMTEuNzUgMTEtMTIuOGwtMTctMi4wOGMtNi42OCAxLjEyLTEwLjM2IDguMjktMTIuNDEgMTYuMDF6Ii8+PHBhdGggY2xhc3M9ImNscy02IiBkPSJNNDEuNzEgNTJsLjEzLS40NS0uMTMuNDZ6TTQxLjkxIDUxLjM0bC0uMDYuMjIuMDctLjIzek0yNy43MiA1NC42NmE2OC4yNiA2OC4yNiAwIDAgMC0uOTMgMTJ2Mi40MkwzOSA2Ni4xYTEuMDYgMS4wNiAwIDAgMCAuODEtMSA1OC43MiA1OC43MiAwIDAgMSAxLjY5LTEyLjI2YzAgLjE2LS4wOS4zMy0uMTMuNDl6TTY1Ljc4IDM0LjI4bC4wMSAzLjM0IDE3LjAzIDIuMDd2LTIuNThsLTE3LjA0LTIuODN6TTUwLjg3IDc0LjQ0TDY4IDY4LjY4YS45Mi45MiAwIDAgMCAuNjMtLjc5IDcyLjQ2IDcyLjQ2IDAgMCAxIDMuMTgtMTUuNGwtMTguMzggMS4xM2E5MC45MSA5MC45MSAwIDAgMC0yLjU2IDIwLjgyek01My40MyA1My42MnoiLz48L2c+PC9zdmc+",
				MediaType: "image/svg+xml",
			},
		}
		tLabels := map[string]string{
			"alm-owner-" + csv.Name: operatorName,
			"operated-by":           csvVersionedName,
		}
		templateStruct.Spec.Labels = tLabels
		templateStruct.Spec.Selector = &metav1.LabelSelector{MatchLabels: tLabels}
		templateStruct.Spec.InstallModes = []csvv1.InstallMode{
			{Type: csvv1.InstallModeTypeOwnNamespace, Supported: true},
			{Type: csvv1.InstallModeTypeSingleNamespace, Supported: true},
			{Type: csvv1.InstallModeTypeMultiNamespace, Supported: false},
			{Type: csvv1.InstallModeTypeAllNamespaces, Supported: false},
		}
		templateStruct.Spec.CustomResourceDefinitions.Owned = []csvv1.CRDDescription{
			{
				Version:     api.SchemeGroupVersion.Version,
				Kind:        "KieApp",
				DisplayName: "KieApp",
				Description: "A project prescription running an RHPAM/RHDM environment.",
				Name:        "kieapps." + api.SchemeGroupVersion.Group,
				Resources: []csvv1.APIResourceReference{
					{
						Kind:    "DeploymentConfig",
						Version: oappsv1.SchemeGroupVersion.String(),
					},
					{
						Kind:    "StatefulSet",
						Version: appsv1.SchemeGroupVersion.String(),
					},
					{
						Kind:    "Role",
						Version: rbacv1.SchemeGroupVersion.String(),
					},
					{
						Kind:    "RoleBinding",
						Version: rbacv1.SchemeGroupVersion.String(),
					},
					{
						Kind:    "Route",
						Version: routev1.SchemeGroupVersion.String(),
					},
					{
						Kind:    "BuildConfig",
						Version: buildv1.SchemeGroupVersion.String(),
					},
					{
						Kind:    "ImageStream",
						Version: oimagev1.SchemeGroupVersion.String(),
					},
					{
						Kind:    "Secret",
						Version: corev1.SchemeGroupVersion.String(),
					},
					{
						Kind:    "PersistentVolumeClaim",
						Version: corev1.SchemeGroupVersion.String(),
					},
					{
						Kind:    "ServiceAccount",
						Version: corev1.SchemeGroupVersion.String(),
					},
					{
						Kind:    "Service",
						Version: corev1.SchemeGroupVersion.String(),
					},
				},
				SpecDescriptors: []csvv1.SpecDescriptor{
					{
						Description:  "Set true to enable automatic micro version product upgrades, it is disabled by default.",
						DisplayName:  "Enable Upgrades",
						Value:        util.RawMessagePointer("false"),
						Path:         "upgrades.enabled",
						XDescriptors: []string{"urn:alm:descriptor:com.tectonic.ui:booleanSwitch"},
					},
					{
						Description:  "Set true to enable automatic minor product version upgrades, it is disabled by default. Requires spec.upgrades.enabled to be true.",
						DisplayName:  "Include minor version upgrades",
						Value:        util.RawMessagePointer("false"),
						Path:         "upgrades.minor",
						XDescriptors: []string{"urn:alm:descriptor:com.tectonic.ui:booleanSwitch"},
					},
					{
						Description:  "Environment deployed.",
						DisplayName:  "Environment",
						Path:         "environment",
						XDescriptors: []string{"urn:alm:descriptor:com.tectonic.ui:label"},
					},
					{
						Description:  "Product version installed.",
						DisplayName:  "Version",
						Path:         "version",
						XDescriptors: []string{"urn:alm:descriptor:com.tectonic.ui:label"},
					},
				},
				StatusDescriptors: []csvv1.StatusDescriptor{
					{
						Description:  "Deployments for the KieApp environment.",
						DisplayName:  "Deployments",
						Path:         "deployments",
						XDescriptors: []string{"urn:alm:descriptor:com.tectonic.ui:podStatuses"},
					},
					{
						Description:  "Current phase.",
						DisplayName:  "Status",
						Path:         "phase",
						XDescriptors: []string{"urn:alm:descriptor:com.tectonic.ui:label"},
					},
					{
						Description:  "The address for accessing Business Central, if it is deployed.",
						DisplayName:  "Business/Decision Central URL",
						Path:         "consoleHost",
						XDescriptors: []string{"urn:alm:descriptor:org.w3:link"},
					},
				},
			},
		}

		if csv.OperatorName == "kie-cloud-operator" {
			templateStruct.Annotations["certified"] = "false"
			deployFile := "deploy/operator.yaml"
			createFile(deployFile, deployment)
			roleFile := "deploy/role.yaml"
			createFile(roleFile, role)
		}

		csvFile := "deploy/catalog_resources/" + csv.CsvDir + "/" + version.Version + "/" + csvVersionedName + ".clusterserviceversion.yaml"
		/*
			copyTemplateStruct := templateStruct.DeepCopy()
			copyTemplateStruct.Annotations["createdAt"] = ""
			data := &csvv1.ClusterServiceVersion{}
			if fileExists(csvFile) {
				yamlFile, err := ioutil.ReadFile(csvFile)
				if err != nil {
					log.Printf("yamlFile.Get err   #%v ", err)
				}
				err = yaml.Unmarshal(yamlFile, data)
				if err != nil {
					log.Fatalf("Unmarshal: %v", err)
				}
				data.Annotations["createdAt"] = ""
			}
			if !reflect.DeepEqual(copyTemplateStruct.Spec, data.Spec) ||
				!reflect.DeepEqual(copyTemplateStruct.Annotations, data.Annotations) ||
				!reflect.DeepEqual(copyTemplateStruct.Labels, data.Labels) {

				createFile(csvFile, templateStruct)
			}
		*/
		createFile(csvFile, templateStruct)

		packageFile := "deploy/catalog_resources/" + csv.CsvDir + "/" + version.Version + "/" + csv.Name + "." + version.Version + ".package.yaml"
		p, err := os.Create(packageFile)
		defer p.Close()
		if err != nil {
			fmt.Println(err)
			return
		}
		pwr := bufio.NewWriter(p)
		pwr.WriteString("#! package-manifest: " + csvFile + "\n")
		packagedata := packageStruct{
			PackageName: operatorName,
			Channels: []channel{
				{
					Name:       maturity,
					CurrentCSV: csvVersionedName,
				},
			},
		}
		util.MarshallObject(packagedata, pwr)
		pwr.Flush()
	}
}

// fileExists checks if a file exists and is not a directory before we
// try using it to prevent further errors.
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func createFile(filepath string, obj interface{}) {
	f, err := os.Create(filepath)
	defer f.Close()
	if err != nil {
		fmt.Println(err)
		return
	}
	writer := bufio.NewWriter(f)
	util.MarshallObject(obj, writer)
	writer.Flush()
}
