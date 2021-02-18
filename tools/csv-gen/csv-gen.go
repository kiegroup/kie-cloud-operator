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
	"github.com/tidwall/sjson"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/rand"
)

var log = logs.GetLogger("csv.generator")

var (
	rh              = "Red Hat"
	channel         = "stable"
	major, minor, _ = defaults.GetMajorMinorMicro(constants.CurrentVersion)
	csvs            = []csvSetting{
		{
			Name:         "businessautomation",
			DisplayName:  "Business Automation (DEV)",
			OperatorName: "business-automation-operator",
			Registry:     "quay.io",
			Context:      "kiegroup",
			ImageName:    "kie-cloud-operator",
			Tag:          version.Version,
			Maturity:     "dev",
			Dev:          true,
		},
		{
			Name:         "businessautomation",
			DisplayName:  "Business Automation",
			OperatorName: "business-automation-operator",
			Registry:     constants.ImageRegistryBrew,
			Context:      constants.ImageContextBrew,
			ImageName:    "rhpam-" + major + "-rhpam-rhel8-operator",
			Tag:          version.Version,
			Maturity:     "test",
		},
		{
			Name:         "businessautomation",
			DisplayName:  "Business Automation",
			OperatorName: "business-automation-operator",
			Registry:     constants.ImageRegistryStage,
			Context:      "rhpam-" + major,
			ImageName:    "rhpam-rhel8-operator",
			Tag:          version.Version,
			Maturity:     channel,
		},
	}
)

var (
	ver      = flag.String("version", version.CsvVersion, "set CSV version")
	replaces = flag.String("replaces", version.CsvPriorVersion, "set CSV version to replace")
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
		descrip := "Deploys and manages Red Hat Process Automation Manager and Red Hat Decision Manager environments."
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
		templateStruct.Spec.Replaces = operatorName + "." + *replaces
		templateStruct.Spec.Description = descrip + "\n\n* **Red Hat Process Automation Manager** is a platform for developing containerized microservices and applications that automate business decisions and processes. It includes business process management (BPM), business rules management (BRM), and business resource optimization and complex event processing (CEP) technologies. It also includes a user experience platform to create engaging user interfaces for process and decision services with minimal coding.\n\n * **Red Hat Decision Manager** is a platform for developing containerized microservices and applications that automate business decisions. It includes business rules management, complex event processing, and resource optimization technologies. Organizations can incorporate sophisticated decision logic into line-of-business applications and quickly update underlying business rules as market conditions change.\n\n[See more](https://www.redhat.com/en/products/process-automation)."
		templateStruct.Spec.DisplayName = csv.DisplayName
		templateStruct.Spec.Maturity = csv.Maturity
		templateStruct.Spec.Maintainers = []csvv1.Maintainer{{Name: rh, Email: "bsig-cloud@redhat.com"}}
		templateStruct.Spec.Provider = csvv1.AppLink{Name: rh}
		templateStruct.Spec.Links = []csvv1.AppLink{
			{Name: "Product Page", URL: "https://access.redhat.com/products/red-hat-process-automation-manager"},
			{Name: "Documentation", URL: "https://access.redhat.com/documentation/en-us/red_hat_process_automation_manager/" + major + "." + minor + "/#category-deploying-red-hat-process-automation-manager-on-openshift"},
		}
		templateStruct.Spec.Icon = []csvv1.Icon{
			{
				Data:      "PHN2ZyBpZD0iTGF5ZXJfMSIgZGF0YS1uYW1lPSJMYXllciAxIiB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHZpZXdCb3g9IjAgMCA3MjEuMTUgNzIxLjE1Ij48ZGVmcz48c3R5bGU+LmNscy0xe2ZpbGw6I2RkMzkyNjt9LmNscy0ye2ZpbGw6I2NjMzQyNzt9LmNscy0ze2ZpbGw6I2ZmZjt9LmNscy00e2ZpbGw6I2U1ZTVlNDt9PC9zdHlsZT48L2RlZnM+PHRpdGxlPlByb2R1Y3RfSWNvbi1SZWRfSGF0LUF1dG9tYXRpb24tUkdCPC90aXRsZT48Y2lyY2xlIGNsYXNzPSJjbHMtMSIgY3g9IjM2MC41NyIgY3k9IjM2MC41NyIgcj0iMzU4LjU4Ii8+PHBhdGggY2xhc3M9ImNscy0yIiBkPSJNNjEzLjc4LDEwNy4wOSwxMDYuNzIsNjE0LjE2YzE0MC4xNCwxMzguNjIsMzY2LjExLDEzOC4xNiw1MDUuNjctMS40Uzc1Mi40LDI0Ny4yNCw2MTMuNzgsMTA3LjA5WiIvPjxwb2x5Z29uIGNsYXNzPSJjbHMtMyIgcG9pbnRzPSIzNzguOTcgMzI3LjQ4IDQ2MS43NyAxNTkuNTcgMjU5LjY3IDE1OS40OSAyNTkuNjcgNDEzLjEgMzA2Ljk3IDQxMy43OCAzOTMuMjcgMzI3LjQ3IDM3OC45NyAzMjcuNDgiLz48cG9seWdvbiBjbGFzcz0iY2xzLTQiIHBvaW50cz0iMzU5LjYgNTc4LjA2IDQ4Mi41NSAzMjcuNDUgMzkzLjI3IDMyNy40NyAzMDYuOTcgNDEzLjc4IDM1OS42IDQxNC41MiAzNTkuNiA1NzguMDYiLz48L3N2Zz4=",
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
		csvFile := bundleDir + "manifests/" + operatorName + "." + *ver + ".clusterserviceversion.yaml"

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

		// copy crd to manifests dir
		crdYaml := "kieapp.crd.yaml"
		crdPath := "deploy/crds/"
		crdFile := crdPath + crdYaml
		err = CopyFile(crdFile, bundleDir+"manifests/"+crdYaml)
		if err != nil {
			log.Error(err)
		}
		//crdSymLink := bundleDir + "manifests/" + crdYaml
		//os.Symlink("../../../../crds/"+crdYaml, crdSymLink)

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

		// create prior-version snippet yaml sample
		versionSnippet := &api.KieApp{}
		versionSnippet.Name = "prior-version"
		versionSnippet.SetAnnotations(map[string]string{
			"consoleName":    "snippet-" + versionSnippet.Name,
			"consoleTitle":   "Prior Product Version",
			"consoleDesc":    "Use this snippet to deploy a prior product version",
			"consoleSnippet": "true",
		})
		versionSnippet.SetGroupVersionKind(api.SchemeGroupVersion.WithKind("KieApp"))
		jsonByte, err := json.Marshal(versionSnippet)
		if err != nil {
			log.Error(err)
		}
		if jsonByte, err = sjson.DeleteBytes(jsonByte, "metadata.creationTimestamp"); err != nil {
			log.Error(err)
		}
		if jsonByte, err = sjson.DeleteBytes(jsonByte, "status"); err != nil {
			log.Error(err)
		}
		if jsonByte, err = sjson.DeleteBytes(jsonByte, "spec"); err != nil {
			log.Error(err)
		}
		if jsonByte, err = sjson.SetBytes(jsonByte, "spec.version", []byte(constants.PriorVersion1)); err != nil {
			log.Error(err)
		}
		var snippetObj interface{}
		if err = json.Unmarshal(jsonByte, &snippetObj); err != nil {
			log.Error(err)
		}
		createFile("deploy/crs/"+api.SchemeGroupVersion.Version+"/snippets/prior_version.yaml", snippetObj)
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
