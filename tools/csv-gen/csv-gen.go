package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/RHsyseng/operator-utils/pkg/logs"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/blang/semver"
	"github.com/heroku/docker-registry-client/registry"
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
	"github.com/tidwall/sjson"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var log = logs.GetLogger("csv.generator")

var (
	rh              = "Red Hat"
	maturity        = "stable"
	major, minor, _ = defaults.MajorMinorMicro(constants.CurrentVersion)
	csvs            = []csvSetting{
		{
			Name:         "kiecloud",
			DisplayName:  "Business Automation",
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
			Registry:     constants.ImageRegistry,
			Context:      "rhpam-" + major,
			ImageName:    "rhpam-rhel8-operator",
			Tag:          constants.CurrentVersion,
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
	Permissions        []csvPermissions `json:"permissions"`
	ClusterPermissions []csvPermissions `json:"clusterPermissions"`
	Deployments        []csvDeployments `json:"deployments"`
}
type channel struct {
	Name       string `json:"name"`
	CurrentCSV string `json:"currentCSV"`
}
type packageStruct struct {
	PackageName string    `json:"packageName"`
	Channels    []channel `json:"channels"`
}
type image struct {
	Name  string `json:"name"`
	Image string `json:"image"`
}

func main() {
	imageShaMap := map[string]string{}
	for _, csv := range csvs {
		operatorName := csv.Name + "-operator"
		templateStruct := &csvv1.ClusterServiceVersion{}
		templateStruct.SetGroupVersionKind(csvv1.SchemeGroupVersion.WithKind("ClusterServiceVersion"))
		csvStruct := &csvv1.ClusterServiceVersion{}
		strategySpec := &csvStrategySpec{}
		json.Unmarshal(csvStruct.Spec.InstallStrategy.StrategySpecRaw, strategySpec)

		templateStrategySpec := &csvStrategySpec{}
		deployment := components.GetDeployment(csv.OperatorName, csv.Registry, csv.Context, csv.ImageName, csv.Tag, "Always")
		templateStrategySpec.Deployments = append(templateStrategySpec.Deployments, []csvDeployments{{Name: csv.OperatorName, Spec: deployment.Spec}}...)
		role := components.GetRole(csv.OperatorName)
		templateStrategySpec.Permissions = append(templateStrategySpec.Permissions, []csvPermissions{{ServiceAccountName: deployment.Spec.Template.Spec.ServiceAccountName, Rules: role.Rules}}...)
		clusterRole := components.GetClusterRole(csv.OperatorName)
		templateStrategySpec.ClusterPermissions = append(templateStrategySpec.ClusterPermissions, []csvPermissions{{ServiceAccountName: deployment.Spec.Template.Spec.ServiceAccountName, Rules: clusterRole.Rules}}...)
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
		templateStruct.Spec.Replaces = operatorName + "." + version.PriorVersion
		templateStruct.Spec.Description = descrip + "\n\n* **Red Hat Process Automation Manager** is a platform for developing containerized microservices and applications that automate business decisions and processes. It includes business process management (BPM), business rules management (BRM), and business resource optimization and complex event processing (CEP) technologies. It also includes a user experience platform to create engaging user interfaces for process and decision services with minimal coding.\n\n * **Red Hat Decision Manager** is a platform for developing containerized microservices and applications that automate business decisions. It includes business rules management, complex event processing, and resource optimization technologies. Organizations can incorporate sophisticated decision logic into line-of-business applications and quickly update underlying business rules as market conditions change.\n\n[See more](https://www.redhat.com/en/products/process-automation)."
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

		opMajor, opMinor, _ := defaults.MajorMinorMicro(version.Version)
		csvFile := "deploy/catalog_resources/" + csv.CsvDir + "/" + opMajor + "." + opMinor + "/" + csvVersionedName + ".clusterserviceversion.yaml"

		imageName, _, _ := defaults.GetImage(deployment.Spec.Template.Spec.Containers[0].Image)
		relatedImages := []image{}

		if csv.OperatorName == "kie-cloud-operator" {
			templateStruct.Annotations["certified"] = "false"
			deployFile := "deploy/operator.yaml"
			createFile(deployFile, deployment)
			roleFile := "deploy/role.yaml"
			createFile(roleFile, role)
		}
		relatedImages = append(relatedImages, image{Name: imageName, Image: deployment.Spec.Template.Spec.Containers[0].Image})

		// create image-references file for automated ART digest find/replace
		imageRef := constants.ImageRef{
			TypeMeta: metav1.TypeMeta{
				APIVersion: oimagev1.GroupVersion.String(),
				Kind:       "ImageStream",
			},
			Spec: constants.ImageRefSpec{
				Tags: []constants.ImageRefTag{
					{
						// Needs to match the component name for upstream and downstream.
						Name: "rhpam-7-rhel8-operator-container",
						From: &corev1.ObjectReference{
							// Needs to match the image that is in your CSV that you want to replace.
							Name: deployment.Spec.Template.Spec.Containers[0].Image,
							Kind: "DockerImage",
						},
					},
				},
			},
		}
		sort.Sort(sort.Reverse(sort.StringSlice(constants.SupportedVersions)))
		for _, imageVersion := range constants.SupportedVersions {
			for _, i := range constants.Images {
				imageURL := i.Registry + ":" + imageVersion
				imageRef.Spec.Tags = append(imageRef.Spec.Tags, constants.ImageRefTag{
					Name: i.Component,
					From: &corev1.ObjectReference{
						Name: imageURL,
						Kind: "DockerImage",
					},
				})
				if imageVersion >= "7.7.0" {
					relatedImages = append(relatedImages, getRelatedImage(imageURL))
				}
			}
		}

		// add ancillary images to image-references file
		imageRef.Spec.Tags = append(imageRef.Spec.Tags, constants.ImageRefTag{
			Name: constants.OauthComponent,
			From: &corev1.ObjectReference{
				Name: constants.OauthImageURL,
				Kind: "DockerImage",
			},
		})
		imageRef.Spec.Tags = append(imageRef.Spec.Tags, constants.ImageRefTag{
			Name: constants.OseCli311Component,
			From: &corev1.ObjectReference{
				Name: constants.OseCli311ImageURL,
				Kind: "DockerImage",
			},
		})
		imageRef.Spec.Tags = append(imageRef.Spec.Tags, constants.ImageRefTag{
			Name: constants.MySQL57Component,
			From: &corev1.ObjectReference{
				Name: constants.MySQL57ImageURL,
				Kind: "DockerImage",
			},
		})
		imageRef.Spec.Tags = append(imageRef.Spec.Tags, constants.ImageRefTag{
			Name: constants.PostgreSQL10Component,
			From: &corev1.ObjectReference{
				Name: constants.PostgreSQL10ImageURL,
				Kind: "DockerImage",
			},
		})
		imageRef.Spec.Tags = append(imageRef.Spec.Tags, constants.ImageRefTag{
			Name: constants.Datagrid73Component,
			From: &corev1.ObjectReference{
				Name: constants.Datagrid73ImageURL,
				Kind: "DockerImage",
			},
		})
		imageRef.Spec.Tags = append(imageRef.Spec.Tags, constants.ImageRefTag{
			Name: constants.Broker75Component,
			From: &corev1.ObjectReference{
				Name: constants.Broker75ImageURL,
				Kind: "DockerImage",
			},
		})

		// add ancillary images to relatedImages
		relatedImages = append(relatedImages, getRelatedImage(constants.OauthImageURL))
		relatedImages = append(relatedImages, getRelatedImage(constants.OseCli311ImageURL))
		relatedImages = append(relatedImages, getRelatedImage(constants.MySQL57ImageURL))
		relatedImages = append(relatedImages, getRelatedImage(constants.PostgreSQL10ImageURL))
		relatedImages = append(relatedImages, getRelatedImage(constants.Datagrid73ImageURL))
		relatedImages = append(relatedImages, getRelatedImage(constants.Broker75ImageURL))

		if logs.GetBoolEnv("DIGESTS") {
			url := "https://" + constants.ImageRegistry
			if val, ok := os.LookupEnv("REGISTRY"); ok {
				url = val
			}
			username := "" // anonymous
			password := "" // anonymous
			if userToken := strings.Split(os.Getenv("USER_TOKEN"), ":"); len(userToken) > 1 {
				username = userToken[0]
				password = userToken[1]
			}
			hub, err := registry.New(url, username, password)
			if err != nil {
				log.Error(err)
			} else {
				defaultCheckRedirect := hub.Client.CheckRedirect
				for _, tagRef := range imageRef.Spec.Tags {
					hub.Client.CheckRedirect = defaultCheckRedirect
					if _, ok := imageShaMap[tagRef.From.Name]; !ok {
						imageShaMap[tagRef.From.Name] = ""
						imageName, imageTag, imageContext := defaults.GetImage(tagRef.From.Name)
						repo := imageContext + "/" + imageName
						tags, err := hub.Tags(repo)
						if err != nil {
							log.Error(err)
						}
						// do not follow redirects - this is critical so we can get the registry digest from Location in redirect response
						hub.Client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
							return http.ErrUseLastResponse
						}
						if _, exists := find(tags, imageTag); exists {
							req, err := http.NewRequest("GET", url+"/v2/"+repo+"/manifests/"+imageTag, nil)
							if err != nil {
								log.Error(err)
							}
							req.Header.Add("Accept", "application/vnd.docker.distribution.manifest.v2+json")
							resp, err := hub.Client.Do(req)
							if err != nil {
								log.Error(err)
							}
							if resp != nil {
								defer resp.Body.Close()
							}
							if resp.StatusCode == 302 || resp.StatusCode == 301 {
								digestURL, err := resp.Location()
								if err != nil {
									log.Error(err)
								}
								if digestURL != nil {
									if url := strings.Split(digestURL.EscapedPath(), "/"); len(url) > 1 {
										imageShaMap[tagRef.From.Name] = strings.ReplaceAll(tagRef.From.Name, ":"+imageTag, "@"+url[len(url)-1])
									}
								}
							}
						}
					}
				}
			}
		}
		imageFile := "deploy/catalog_resources/" + csv.CsvDir + "/" + opMajor + "." + opMinor + "/" + "image-references"
		createFile(imageFile, imageRef)

		var templateInterface interface{}
		if len(relatedImages) > 0 {
			templateJSON, err := json.Marshal(templateStruct)
			if err != nil {
				log.Error(err)
			}
			result, err := sjson.SetBytes(templateJSON, "spec.relatedImages", relatedImages)
			if err != nil {
				log.Error(err)
			}
			if err = json.Unmarshal(result, &templateInterface); err != nil {
				log.Error(err)
			}
		} else {
			templateInterface = templateStruct
		}

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

		packageFile := "deploy/catalog_resources/" + csv.CsvDir + "/" + csv.Name + ".package.yaml"
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

func getRelatedImage(imageURL string) image {
	imageName, _, _ := defaults.GetImage(imageURL)
	return image{
		Name:  imageName,
		Image: imageURL,
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

func find(slice []string, val string) (int, bool) {
	for i, item := range slice {
		if item == val {
			return i, true
		}
	}
	return -1, false
}