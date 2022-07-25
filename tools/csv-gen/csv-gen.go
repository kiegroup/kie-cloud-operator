package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/RHsyseng/operator-utils/pkg/logs"
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
	csvversion "github.com/operator-framework/api/pkg/lib/version"
	csvv1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	sdkVersion "github.com/operator-framework/operator-sdk/version"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/rand"
)

var log = logs.GetLogger("csv.generator")

var (
	ibm             = "IBM"
	channel         = "8.x-stable"
	major, minor, _ = defaults.GetMajorMinorMicro(constants.CurrentVersion)
	csvs            = []csvSetting{
		{
			Name:         "bamoe-businessautomation",
			DisplayName:  "IBM Business Automation (DEV)",
			OperatorName: "bamoe-business-automation-operator",
			Registry:     "quay.io",
			Context:      "kiegroup",
			ImageName:    "kie-cloud-operator",
			Tag:          version.Version,
			Maturity:     "dev",
			Dev:          true,
		},
		{
			Name:         "bamoe-businessautomation",
			DisplayName:  "IBM Business Automation",
			OperatorName: "bamoe-business-automation-operator",
			Registry:     constants.ImageRegistryBrew,
			Context:      constants.ImageContextBrew,
			ImageName:    "bamoe-" + major + "-rhpam-rhel8-operator",
			Tag:          version.Version,
			Maturity:     "test",
		},
		{
			Name:         "bamoe-businessautomation",
			DisplayName:  "IBM Business Automation",
			OperatorName: "bamoe-business-automation-operator",
			Registry:     constants.ImageRegistryStage,
			Context:      constants.IBMBamoeImageContext,
			ImageName:    constants.IBMBamoeImagePrefix + "-rhel8-operator",
			Tag:          version.Version,
			Maturity:     channel,
		},
	}
)

var (
	ver = flag.String("version", version.CsvVersion, "set CSV version")
	// TODO uncomment for next 8.x release.
	// replaces = flag.String("replaces", version.CsvPriorVersion, "set CSV version to replace")
)

type csvSetting struct {
	Name         string `json:"name"`
	DisplayName  string `json:"displayName"`
	OperatorName string `json:"operatorName"`
	Registry     string `json:"repository"`
	Context      string `json:"context"`
	ImageName    string `json:"imageName"`
	Tag          string `json:"tag"`
	Maturity     string `json:"maturity"`
	Dev          bool   `json:"dev"`
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
	Permissions        []csvPermissions `json:"permissions"`
	ClusterPermissions []csvPermissions `json:"clusterPermissions"`
	Deployments        []csvDeployments `json:"deployments"`
}
type annotationsStruct struct {
	Annotations map[string]string `json:"annotations"`
}
type image struct {
	Name  string `json:"name"`
	Image string `json:"image"`
}

func main() {
	flag.Parse()
	imageShaMap := map[string]string{}
	for _, csv := range csvs {
		operatorName := csv.Name + "-operator"
		templateStruct := &csvv1.ClusterServiceVersion{}
		templateStruct.SetGroupVersionKind(csvv1.SchemeGroupVersion.WithKind("ClusterServiceVersion"))
		templateStrategySpec := csvv1.StrategyDetailsDeployment{}
		deployment := components.GetDeployment(csv.OperatorName, csv.Registry, csv.Context, csv.ImageName, csv.Tag, "Always", csv.Dev)
		templateStrategySpec.DeploymentSpecs = append(templateStrategySpec.DeploymentSpecs, []csvv1.StrategyDeploymentSpec{{Name: csv.OperatorName, Spec: deployment.Spec}}...)
		role := components.GetRole(csv.OperatorName)
		templateStrategySpec.Permissions = append(templateStrategySpec.Permissions, []csvv1.StrategyDeploymentPermissions{{ServiceAccountName: deployment.Spec.Template.Spec.ServiceAccountName, Rules: role.Rules}}...)
		clusterRole := components.GetClusterRole(csv.OperatorName)
		templateStrategySpec.ClusterPermissions = append(templateStrategySpec.ClusterPermissions, []csvv1.StrategyDeploymentPermissions{{ServiceAccountName: deployment.Spec.Template.Spec.ServiceAccountName, Rules: clusterRole.Rules}}...)
		templateStruct.Spec.InstallStrategy.StrategySpec = templateStrategySpec
		templateStruct.Spec.InstallStrategy.StrategyName = "deployment"

		csvVersionedName := operatorName + "." + *ver
		random := rand.String(10)
		csvVersion := csvversion.OperatorVersion{}
		csvVersion.Version = semver.MustParse(*ver)
		if csv.Maturity != channel {
			csvVersionedName = csvVersionedName + "-dev-" + random
			csvVersion.Version.Build = []string{random}
		}
		templateStruct.Name = csvVersionedName
		templateStruct.Spec.Version = csvVersion
		templateStruct.Namespace = "placeholder"
		descrip := "Deploys and manages IBM Business Automation Manager Open Editions environment."
		repository := "https://github.com/kiegroup/kie-cloud-operator"
		examples := []string{"{\x22apiVersion\x22:\x22app.kiegroup.org/v2\x22,\x22kind\x22:\x22KieApp\x22,\x22metadata\x22:{\x22name\x22:\x22rhpam-trial\x22},\x22spec\x22:{\x22environment\x22:\x22rhpam-trial\x22}}"}
		templateStruct.SetAnnotations(
			map[string]string{
				"createdAt":      time.Now().Format("2006-01-02 15:04:05"),
				"containerImage": deployment.Spec.Template.Spec.Containers[0].Image,
				"description":    descrip,
				"categories":     "Integration & Delivery",
				"certified":      "true",
				// TODO after next 8.0.0 release replace Basic Install with Seamless Upgrades
				"capabilities":        "Basic Install",
				"repository":          repository,
				"support":             ibm,
				"tectonic-visibility": "ocs",
				"alm-examples":        "[" + strings.Join(examples, ",") + "]",
				"operators.openshift.io/infrastructure-features": "[\"Disconnected\"]",
			},
		)
		templateStruct.SetLabels(
			map[string]string{
				"operator-" + csv.Name:            "true",
				"operatorframework.io/os.linux":   "supported",
				"operatorframework.io/arch.amd64": "supported",
			},
		)
		templateStruct.Spec.Keywords = []string{"kieapp", "pam", "decision", "kie", "cloud", "bpm", "process", "automation", "operator"}
		// TODO Uncomment after IBM BAMOE 8.0.0 release
		// templateStruct.Spec.Replaces = operatorName + "." + *replaces
		templateStruct.Spec.Description = descrip + "\n\n* **IBM Process Automation Manager** is a platform for developing containerized microservices and applications that automate business decisions and processes. It includes business process management (BPM), business rules management (BRM), and business resource optimization and complex event processing (CEP) technologies. It also includes a user experience platform to create engaging user interfaces for process and decision services with minimal coding.\n\n * **Red Hat Decision Manager** is a platform for developing containerized microservices and applications that automate business decisions. It includes business rules management, complex event processing, and resource optimization technologies. Organizations can incorporate sophisticated decision logic into line-of-business applications and quickly update underlying business rules as market conditions change."
		templateStruct.Spec.DisplayName = csv.DisplayName
		templateStruct.Spec.Maturity = csv.Maturity
		templateStruct.Spec.Provider = csvv1.AppLink{Name: ibm}
		templateStruct.Spec.Links = []csvv1.AppLink{
			{Name: "Product Page", URL: "https://access.redhat.com/products/red-hat-process-automation-manager"},
			{Name: "Documentation", URL: "https://access.redhat.com/documentation/en-us/red_hat_process_automation_manager/" + major + "." + minor + "/#category-deploying-red-hat-process-automation-manager-on-openshift"},
		}
		templateStruct.Spec.Icon = []csvv1.Icon{
			{
				Data:      "PHN2ZyBpZD0iTGF5ZXJfMSIgZGF0YS1uYW1lPSJMYXllciAxIiB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHdpZHRoPSI1MS44IiBoZWlnaHQ9IjE5LjMzIiB2aWV3Qm94PSIwIDAgNTEuOCAxOS4zMyI+PGRlZnM+PHN0eWxlPi5jbHMtMXtmaWxsOiM0MjZhYjM7fTwvc3R5bGU+PC9kZWZzPjxwYXRoIGNsYXNzPSJjbHMtMSIgZD0iTTM4Ljc2LDkuMiwzOC4zLDcuODZIMzAuNjFWOS4yWm0uOSwyLjU3LS40Ny0xLjM1SDMwLjYxdjEuMzVabTUuNzEsMTUuMzdoNi43MVYyNS44SDQ1LjM3djEuMzRabTAtMi41Nmg2LjcxVjIzLjIzSDQ1LjM3djEuMzVabTAtMi41N2g0VjIwLjY3aC00VjIyWm00LTMuOWgtNHYxLjM0aDRWMTguMTFabS00LTEuMjJoNFYxNS41NUg0MS43M2wtLjM4LDEuMDhMNDEsMTUuNTVIMzMuM3YxLjM0aDRWMTUuNjZsLjQ0LDEuMjNoNy4xOGwuNDMtMS4yM3YxLjIzWm00LTMuOUg0Mi42MmwtLjQ3LDEuMzRINDkuNFYxM1pNMzMuMywxOS40NWg0VjE4LjExaC00djEuMzRabTAsMi41Nmg0VjIwLjY3aC00VjIyWm0tMi42OSwyLjU3aDYuNzFWMjMuMjNIMzAuNjF2MS4zNVptMCwyLjU2aDYuNzFWMjUuOEgzMC42MXYxLjM0Wk00NC40LDcuODYsNDMuOTMsOS4yaDguMTVWNy44NlpNNDMsMTEuNzdoOVYxMC40Mkg0My41MUw0MywxMS43N1pNMzMuMywxNC4zM2g3LjI1TDQwLjA4LDEzSDMzLjN2MS4zNFptNS4zNSw1LjEySDQ0bC40Ny0xLjM0SDM4LjE4bC40NywxLjM0Wm0uOSwyLjU2aDMuNTlsLjQ3LTEuMzRIMzkuMDhMMzkuNTUsMjJabS45LDIuNTdoMS43OWwuNDctMS4zNUg0MGwuNDcsMS4zNVptLjksMi41Ni40Ni0xLjM0aC0uOTNsLjQ3LDEuMzRabS0yNi44NCwwaDkuODhhNS4xMSw1LjExLDAsMCwwLDMuNDYtMS4zNEgxNC41MXYxLjM0Wk0yNSwyMC42N1YyMmg0LjUxYTUuNDEsNS40MSwwLDAsMC0uMTctMS4zNFpNMTcuMTksMjJoNFYyMC42N2gtNFYyMlpNMjUsMTQuMzNoNC4zNEE1LjQxLDUuNDEsMCwwLDAsMjkuNTEsMTNIMjV2MS4zNFptLTcuODEsMGg0VjEzaC00djEuMzRabTcuMi02LjQ3SDE0LjUxVjkuMkgyNy44NWE1LjEzLDUuMTMsMCwwLDAtMy40Ni0xLjM0Wm00LjQ0LDIuNTZIMTQuNTF2MS4zNUgyOS4zN2E1LjMsNS4zLDAsMCwwLS41NC0xLjM1Wk0xNy4xOSwxNS41NXYxLjM0SDI3LjcxYTUuMzYsNS4zNiwwLDAsMCwxLjEyLTEuMzRabTEwLjUyLDIuNTZIMTcuMTl2MS4zNEgyOC44M2E1LjM2LDUuMzYsMCwwLDAtMS4xMi0xLjM0Wm0tMTMuMiw2LjQ3SDI4LjgzYTUuMyw1LjMsMCwwLDAsLjU0LTEuMzVIMTQuNTF2MS4zNVpNMy43Nyw5LjJoOS40VjcuODZIMy43N1Y5LjJabTAsMi41N2g5LjRWMTAuNDJIMy43N3YxLjM1Wk0xMC40OCwxM2gtNHYxLjM0aDRWMTNabS00LDMuOWg0VjE1LjU1aC00djEuMzRabTAsMi41Nmg0VjE4LjExaC00djEuMzRabTAsMi41Nmg0VjIwLjY3aC00VjIyWk0zLjc3LDI0LjU4aDkuNFYyMy4yM0gzLjc3djEuMzVabTAsMS4yMmg5LjR2MS4zNEgzLjc3VjI1LjhabTUwLjgyLjMyYy4wOSwwLC4xMywwLC4xMy0uMTJ2LS4wN2MwLS4wOCwwLS4xMi0uMTMtLjEySDU0LjR2LjMxWm0tLjE5LjU2aC0uMjZWMjUuNjFoLjQ5QS4zMi4zMiwwLDAsMSw1NSwyNmEuMzIuMzIsMCwwLDEtLjIuMzJsLjI1LjQxaC0uMjlsLS4yLS4zN0g1NC40di4zN1ptLjk0LS40OHYtLjEzYS43OS43OSwwLDAsMC0xLjU3LDB2LjEzYS43OS43OSwwLDAsMCwxLjU3LDBabS0xLjgxLS4wNmExLDEsMCwxLDEsMSwxLjA1LDEsMSwwLDAsMS0xLTEuMDVaIiB0cmFuc2Zvcm09InRyYW5zbGF0ZSgtMy43NyAtNy44NikiLz48L3N2Zz4=",
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
				Description: "A project prescription running an IBM BAMOE environment.",
				Name:        "kieapps." + api.SchemeGroupVersion.Group,
				Resources: []csvv1.APIResourceReference{
					{
						Kind:    "DeploymentConfig",
						Version: oappsv1.GroupVersion.String(),
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
						Version: routev1.GroupVersion.String(),
					},
					{
						Kind:    "BuildConfig",
						Version: buildv1.GroupVersion.String(),
					},
					{
						Kind:    "ImageStream",
						Version: oimagev1.GroupVersion.String(),
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
						Path:         "upgrades.enabled",
						XDescriptors: []string{"urn:alm:descriptor:com.tectonic.ui:booleanSwitch"},
					},
					{
						Description:  "Set true to enable automatic minor product version upgrades, it is disabled by default. Requires spec.upgrades.enabled to be true.",
						DisplayName:  "Include minor version upgrades",
						Path:         "upgrades.minor",
						XDescriptors: []string{"urn:alm:descriptor:com.tectonic.ui:booleanSwitch"},
					},
					{
						Description:  "Set true to enable image tags, disabled by default. This will leverage image tags instead of the image digests.",
						DisplayName:  "Use Image Tags",
						Path:         "useImageTags",
						XDescriptors: []string{"urn:alm:descriptor:com.tectonic.ui:booleanSwitch"},
					},
					{
						Description:  "Environment deployed.",
						DisplayName:  "Environment",
						Path:         "environment",
						XDescriptors: []string{"urn:alm:descriptor:com.tectonic.ui:label"},
					},
				},
				StatusDescriptors: []csvv1.StatusDescriptor{
					{
						Description:  "Product version installed.",
						DisplayName:  "Version",
						Path:         "version",
						XDescriptors: []string{"urn:alm:descriptor:com.tectonic.ui:label"},
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
					{
						Description:  "Deployments for the KieApp environment.",
						DisplayName:  "Deployments",
						Path:         "deployments",
						XDescriptors: []string{"urn:alm:descriptor:com.tectonic.ui:podStatuses"},
					},
				},
			},
		}

		bundleDir := "deploy/olm-catalog/prod/" + *ver + "/"
		if csv.Maturity == "dev" {
			bundleDir = "deploy/olm-catalog/dev/" + *ver + "/"
			templateStruct.Annotations["certified"] = "false"
			deployFile := "deploy/operator.yaml"
			createFile(deployFile, deployment)
			roleFile := "deploy/role.yaml"
			createFile(roleFile, role)
		}
		if csv.Maturity == "test" {
			bundleDir = "deploy/olm-catalog/test/" + *ver + "/"
		}
		csvFile := bundleDir + "manifests/" + operatorName + ".clusterserviceversion.yaml"

		var templateInterface interface{}
		templateInterface = templateStruct

		// find and replace images with SHAs where necessary
		templateByte, err := json.Marshal(templateInterface)
		if err != nil {
			log.Error(err)
		}
		for from, to := range imageShaMap {
			if to != "" {
				templateByte = bytes.ReplaceAll(templateByte, []byte(from), []byte(to))
			}
		}
		if err = json.Unmarshal(templateByte, &templateInterface); err != nil {
			log.Error(err)
		}
		createFile(csvFile, &templateInterface)

		annotationsdata := annotationsStruct{
			Annotations: map[string]string{
				"operators.operatorframework.io.bundle.channel.default.v1": channel,
				"operators.operatorframework.io.bundle.channels.v1":        channel,
				"operators.operatorframework.io.bundle.manifests.v1":       "manifests/",
				"operators.operatorframework.io.bundle.mediatype.v1":       "registry+v1",
				"operators.operatorframework.io.bundle.metadata.v1":        "metadata/",
				"operators.operatorframework.io.bundle.package.v1":         operatorName,
				"operators.operatorframework.io.metrics.builder":           "operator-sdk-" + sdkVersion.Version,
				"operators.operatorframework.io.metrics.mediatype.v1":      "metrics+v1",
				"operators.operatorframework.io.metrics.project_layout":    "go",
			},
		}
		annotationsFile := bundleDir + "metadata/annotations.yaml"
		createFile(annotationsFile, &annotationsdata)

		// TODO uncomment for next 8.x release.
		// create prior-version snippet yaml sample
		//versionSnippet := &api.KieApp{}
		//versionSnippet.Name = "prior-version"
		//versionSnippet.SetAnnotations(map[string]string{
		//	"consoleName":    "snippet-" + versionSnippet.Name,
		//	"consoleTitle":   "Prior Product Version",
		//	"consoleDesc":    "Use this snippet to deploy a prior product version",
		//	"consoleSnippet": "true",
		//})
		//versionSnippet.SetGroupVersionKind(api.SchemeGroupVersion.WithKind("KieApp"))
		//jsonByte, err := json.Marshal(versionSnippet)
		//if err != nil {
		//	log.Error(err)
		//}
		//if jsonByte, err = sjson.DeleteBytes(jsonByte, "metadata.creationTimestamp"); err != nil {
		//	log.Error(err)
		//}
		//if jsonByte, err = sjson.DeleteBytes(jsonByte, "status"); err != nil {
		//	log.Error(err)
		//}
		//if jsonByte, err = sjson.DeleteBytes(jsonByte, "spec"); err != nil {
		//	log.Error(err)
		//}
		//if jsonByte, err = sjson.SetBytes(jsonByte, "spec.version", []byte(constants.PriorVersion)); err != nil {
		//	log.Error(err)
		//}
		//var snippetObj interface{}
		//if err = json.Unmarshal(jsonByte, &snippetObj); err != nil {
		//	log.Error(err)
		//}
		//createFile("deploy/crs/"+api.SchemeGroupVersion.Version+"/snippets/prior_version.yaml", snippetObj)
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

func createFile(path string, obj interface{}) {
	os.MkdirAll(filepath.Dir(path), os.ModePerm)
	f, err := os.Create(path)
	defer f.Close()
	if err != nil {
		fmt.Println(err)
		return
	}
	writer := bufio.NewWriter(f)
	util.MarshallObject(obj, writer)
	writer.Flush()
}

// CopyFile copies a file from src to dst. If src and dst files exist, and are
// the same, then return success. Otherise, attempt to create a hard link
// between the two files. If that fail, copy the file contents from src to dst.
func CopyFile(src, dst string) (err error) {
	sfi, err := os.Stat(src)
	if err != nil {
		return
	}
	if !sfi.Mode().IsRegular() {
		// cannot copy non-regular files (e.g., directories,
		// symlinks, devices, etc.)
		return fmt.Errorf("CopyFile: non-regular source file %s (%q)", sfi.Name(), sfi.Mode().String())
	}
	dfi, err := os.Stat(dst)
	if err != nil {
		if !os.IsNotExist(err) {
			return
		}
	} else {
		if !(dfi.Mode().IsRegular()) {
			return fmt.Errorf("CopyFile: non-regular destination file %s (%q)", dfi.Name(), dfi.Mode().String())
		}
		if os.SameFile(sfi, dfi) {
			return
		}
	}
	if err = os.Link(src, dst); err == nil {
		return
	}
	err = copyFileContents(src, dst)
	return
}

// copyFileContents copies the contents of the file named src to the file named
// by dst. The file will be created if it does not already exist. If the
// destination file exists, all it's contents will be replaced by the contents
// of the source file.
func copyFileContents(src, dst string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return
	}
	defer func() {
		cerr := out.Close()
		if err == nil {
			err = cerr
		}
	}()
	if _, err = io.Copy(out, in); err != nil {
		return
	}
	err = out.Sync()
	return
}
