package clusterresource

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"runtime/debug"
	"strings"
	"time"

	"github.com/flyteorg/flyteadmin/pkg/executioncluster"
	"github.com/flyteorg/flyteadmin/pkg/runtime"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/jsonmergepatch"
	"k8s.io/apimachinery/pkg/util/mergepatch"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"k8s.io/client-go/kubernetes/scheme"

	"k8s.io/apimachinery/pkg/api/meta"
	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/restmapper"

	"github.com/flyteorg/flyteadmin/pkg/repositories/models"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	v1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/flyteorg/flyteadmin/pkg/manager/impl/resources"
	managerinterfaces "github.com/flyteorg/flyteadmin/pkg/manager/interfaces"

	"github.com/flyteorg/flyteadmin/pkg/executioncluster/interfaces"

	"github.com/flyteorg/flyteadmin/pkg/common"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"

	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/flyteorg/flytestdlib/logger"
	"google.golang.org/grpc/codes"

	"github.com/flyteorg/flyteadmin/pkg/errors"
	"github.com/flyteorg/flyteadmin/pkg/repositories"
	repositoriesInterfaces "github.com/flyteorg/flyteadmin/pkg/repositories/interfaces"
	runtimeInterfaces "github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"
	"github.com/flyteorg/flytestdlib/promutils"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
)

const namespaceVariable = "namespace"
const projectVariable = "project"
const domainVariable = "domain"
const templateVariableFormat = "{{ %s }}"
const replaceAllInstancesOfString = -1

// The clusterresource Controller manages applying desired templatized kubernetes resource files as resources
// in the execution kubernetes cluster.
type Controller interface {
	Sync(ctx context.Context) error
	Run()
}

type controllerMetrics struct {
	Scope                           promutils.Scope
	SyncStarted                     prometheus.Counter
	KubernetesResourcesCreated      prometheus.Counter
	KubernetesResourcesCreateErrors prometheus.Counter
	ResourcesAdded                  prometheus.Counter
	ResourceAddErrors               prometheus.Counter
	TemplateReadErrors              prometheus.Counter
	TemplateDecodeErrors            prometheus.Counter
	AppliedTemplateExists           prometheus.Counter
	TemplateUpdateErrors            prometheus.Counter
	Panics                          prometheus.Counter
}

type FileName = string
type NamespaceName = string
type LastModTimeCache = map[FileName]time.Time
type NamespaceCache = map[NamespaceName]LastModTimeCache

type templateValuesType = map[string]string

type controller struct {
	db                     repositories.RepositoryInterface
	config                 runtimeInterfaces.Configuration
	executionCluster       interfaces.ClusterInterface
	resourceManager        managerinterfaces.ResourceInterface
	poller                 chan struct{}
	metrics                controllerMetrics
	lastAppliedTemplateDir string
	// Map of [namespace -> [templateFileName -> last modified time]]
	appliedTemplates NamespaceCache
}

var descCreatedAtSortParam, _ = common.NewSortParameter(admin.Sort{
	Direction: admin.Sort_DESCENDING,
	Key:       "created_at",
})

// Use a strategic-merge-patch to mimic `kubectl apply` behavior for serviceaccounts.
// Kubectl defaults to using the StrategicMergePatch strategy.
// However the controller-runtime only has an implementation for MergePatch which we were formerly
// using but failed to actually always merge resources in the Patch call.
// INTERESTINGLY Patch doesn't actually appear to update the majority of resources. We default to using Update but
// whitelist the specific set of resources that require a Patch to work instead.
// If you use update with a ServiceAccount - *every* call to Update results in a new corresponding secret being created
// which has the (not so) fun side-effect of overwhelming API server when this Sync script is run as a cron.
var strategicPatchTypes = map[string]bool{
	v1.ServiceAccountKind: true,
}

func (c *controller) templateAlreadyApplied(namespace NamespaceName, templateFile os.FileInfo) bool {
	namespacedAppliedTemplates, ok := c.appliedTemplates[namespace]
	if !ok {
		// There is no record of this namespace altogether.
		return false
	}
	timestamp, ok := namespacedAppliedTemplates[templateFile.Name()]
	if !ok {
		// There is no record of this file having ever been applied.
		return false
	}
	// The applied template file could have been modified, in which case we will need to apply it once more.
	return timestamp.Equal(templateFile.ModTime())
}

// Given a map of templatized variable names -> data source, this function produces an output that maps the same
// variable names to their fully resolved values (from the specified data source).
func populateTemplateValues(data map[string]runtimeInterfaces.DataSource) (templateValuesType, error) {
	templateValues := make(templateValuesType, len(data))
	collectedErrs := make([]error, 0)
	for templateVar, dataSource := range data {
		if templateVar == namespaceVariable || templateVar == projectVariable || templateVar == domainVariable {
			// The namespace variable is specifically reserved for system use only.
			collectedErrs = append(collectedErrs, errors.NewFlyteAdminErrorf(codes.InvalidArgument,
				"Cannot assign namespace template value in user data"))
			continue
		}
		var dataValue string
		if len(dataSource.Value) > 0 {
			dataValue = dataSource.Value
		} else if len(dataSource.ValueFrom.EnvVar) > 0 {
			dataValue = os.Getenv(dataSource.ValueFrom.EnvVar)
		} else if len(dataSource.ValueFrom.FilePath) > 0 {
			templateFile, err := ioutil.ReadFile(dataSource.ValueFrom.FilePath)
			if err != nil {
				collectedErrs = append(collectedErrs, errors.NewFlyteAdminErrorf(codes.InvalidArgument,
					"failed to substitute parameterized value for %s: unable to read value from: [%+v] with err: %v",
					templateVar, dataSource.ValueFrom.FilePath, err))
				continue
			}
			dataValue = string(templateFile)
		} else {
			collectedErrs = append(collectedErrs, errors.NewFlyteAdminErrorf(codes.InvalidArgument,
				"failed to substitute parameterized value for %s: unset or unrecognized ValueFrom: [%+v]", templateVar, dataSource.ValueFrom))
			continue
		}
		if len(dataValue) == 0 {
			collectedErrs = append(collectedErrs, errors.NewFlyteAdminErrorf(codes.InvalidArgument,
				"failed to substitute parameterized value for %s: unset. ValueFrom: [%+v]", templateVar, dataSource.ValueFrom))
			continue
		}
		templateValues[fmt.Sprintf(templateVariableFormat, templateVar)] = dataValue
	}
	if len(collectedErrs) > 0 {
		return nil, errors.NewCollectedFlyteAdminError(codes.InvalidArgument, collectedErrs)
	}
	return templateValues, nil
}

// Produces a map of template variable names and their fully resolved values based on configured defaults for each
// system-domain in the application config file.
func populateDefaultTemplateValues(defaultData map[runtimeInterfaces.DomainName]runtimeInterfaces.TemplateData) (
	map[string]templateValuesType, error) {
	defaultTemplateValues := make(map[string]templateValuesType)
	collectedErrs := make([]error, 0)
	for domainName, templateData := range defaultData {
		domainSpecificTemplateValues, err := populateTemplateValues(templateData)
		if err != nil {
			collectedErrs = append(collectedErrs, err)
			continue
		}
		defaultTemplateValues[domainName] = domainSpecificTemplateValues
	}
	if len(collectedErrs) > 0 {
		return nil, errors.NewCollectedFlyteAdminError(codes.InvalidArgument, collectedErrs)
	}
	return defaultTemplateValues, nil
}

// Fetches user-specified overrides from the admin database for template variables and their desired value
// substitutions based on the input project and domain. These database values are overlaid on top of the configured
// variable defaults for the specific domain as defined in the admin application config file.
func (c *controller) getCustomTemplateValues(
	ctx context.Context, project, domain string, domainTemplateValues templateValuesType) (templateValuesType, error) {
	if len(domainTemplateValues) == 0 {
		domainTemplateValues = make(templateValuesType)
	}
	customTemplateValues := make(templateValuesType)
	for key, value := range domainTemplateValues {
		customTemplateValues[key] = value
	}
	collectedErrs := make([]error, 0)
	// All override values saved in the database take precedence over the domain-specific defaults.
	resource, err := c.resourceManager.GetResource(ctx, managerinterfaces.ResourceRequest{
		Project:      project,
		Domain:       domain,
		ResourceType: admin.MatchableResource_CLUSTER_RESOURCE,
	})
	if err != nil {
		if _, ok := err.(errors.FlyteAdminError); !ok || err.(errors.FlyteAdminError).Code() != codes.NotFound {
			collectedErrs = append(collectedErrs, err)
		}
	}
	if resource != nil && resource.Attributes != nil && resource.Attributes.GetClusterResourceAttributes() != nil {
		for templateKey, templateValue := range resource.Attributes.GetClusterResourceAttributes().Attributes {
			customTemplateValues[fmt.Sprintf(templateVariableFormat, templateKey)] = templateValue
		}
	}
	if len(collectedErrs) > 0 {
		return nil, errors.NewCollectedFlyteAdminError(codes.InvalidArgument, collectedErrs)
	}
	return customTemplateValues, nil
}

// Obtains the REST interface for a GroupVersionResource
func getDynamicResourceInterface(mapping *meta.RESTMapping, dynamicClient dynamic.Interface, namespace NamespaceName) dynamic.ResourceInterface {
	if mapping.Scope.Name() == meta.RESTScopeNameNamespace {
		// namespaced resources should specify the namespace
		return dynamicClient.Resource(mapping.Resource).Namespace(namespace)
	}
	// for cluster-wide resources (e.g. namespaces)
	return dynamicClient.Resource(mapping.Resource)
}

// This represents the minimum closure of objects generated from a template file that
// allows for dynamically creating (or updating) the resource using server side apply.
type dynamicResource struct {
	obj     *unstructured.Unstructured
	mapping *meta.RESTMapping
}

// This function borrows heavily from the excellent example code here:
// https://ymmt2005.hatenablog.com/entry/2020/04/14/An_example_of_using_dynamic_client_of_k8s.io/client-go#Background-Server-Side-Apply
// to dynamically discover the GroupVersionResource for the templatized k8s object from the cluster resource config files
// which a dynamic client can use to create or mutate the resource.
func prepareDynamicCreate(target executioncluster.ExecutionTarget, config string) (dynamicResource, error) {
	dc, err := discovery.NewDiscoveryClientForConfig(&target.Config)
	if err != nil {
		return dynamicResource{}, err
	}
	mapper := restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(dc))

	decUnstructured := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
	obj := &unstructured.Unstructured{}
	_, gvk, err := decUnstructured.Decode([]byte(config), nil, obj)
	if err != nil {
		return dynamicResource{}, err
	}

	// Find GVR
	mapping, err := mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return dynamicResource{}, err
	}

	return dynamicResource{
		obj:     obj,
		mapping: mapping,
	}, nil
}

// This function loops through the kubernetes resource template files in the configured template directory.
// For each unapplied template file (wrt the namespace) this func attempts to
//   1) create k8s object resource from template by performing:
//      a) read template file
//      b) substitute templatized variables with their resolved values
//   2) create the resource on the kubernetes cluster and cache successful outcomes
func (c *controller) syncNamespace(ctx context.Context, project models.Project, domain runtimeInterfaces.Domain, namespace NamespaceName,
	templateValues, customTemplateValues templateValuesType) error {
	templateDir := c.config.ClusterResourceConfiguration().GetTemplatePath()
	if c.lastAppliedTemplateDir != templateDir {
		// Invalidate all caches
		c.lastAppliedTemplateDir = templateDir
		c.appliedTemplates = make(NamespaceCache)
	}
	templateFiles, err := ioutil.ReadDir(templateDir)
	if err != nil {
		return errors.NewFlyteAdminErrorf(codes.Internal,
			"Failed to read config template dir [%s] for namespace [%s] with err: %v",
			namespace, templateDir, err)
	}

	collectedErrs := make([]error, 0)
	for _, templateFile := range templateFiles {
		templateFileName := templateFile.Name()
		if filepath.Ext(templateFileName) != ".yaml" {
			// nothing to do.
			logger.Debugf(ctx, "syncing namespace [%s]: ignoring unrecognized filetype [%s]",
				namespace, templateFile.Name())
			continue
		}

		if c.templateAlreadyApplied(namespace, templateFile) {
			// nothing to do.
			logger.Debugf(ctx, "syncing namespace [%s]: templateFile [%s] already applied, nothing to do.", namespace, templateFile.Name())
			continue
		}

		// 1) create resource from template:
		k8sManifest, err := c.createResourceFromTemplate(ctx, templateDir, templateFileName, project, domain, namespace, templateValues, customTemplateValues)
		if err != nil {
			collectedErrs = append(collectedErrs, err)
			continue
		}

		// 2) create the resource on the kubernetes cluster and cache successful outcomes
		if _, ok := c.appliedTemplates[namespace]; !ok {
			c.appliedTemplates[namespace] = make(LastModTimeCache)
		}
		for _, target := range c.executionCluster.GetAllValidTargets() {
			dynamicObj, err := prepareDynamicCreate(target, k8sManifest)
			if err != nil {
				logger.Warningf(ctx, "Failed to transform kubernetes manifest for namespace [%s] "+
					"into a dynamic unstructured mapping with err: %v, manifest: %v", namespace, err, k8sManifest)
				collectedErrs = append(collectedErrs, err)
				c.metrics.KubernetesResourcesCreateErrors.Inc()
				continue
			}

			logger.Debugf(ctx, "Attempting to create resource [%+v] in cluster [%v] for namespace [%s]",
				dynamicObj.obj.GetKind(), target.ID, namespace)

			dr := getDynamicResourceInterface(dynamicObj.mapping, target.DynamicClient, namespace)
			_, err = dr.Create(ctx, dynamicObj.obj, metav1.CreateOptions{})

			if err != nil {
				if k8serrors.IsAlreadyExists(err) {
					logger.Debugf(ctx, "Type [%+v] in namespace [%s] already exists - attempting update instead",
						dynamicObj.obj.GetKind(), namespace)
					c.metrics.AppliedTemplateExists.Inc()

					currentObj, err := dr.Get(ctx, dynamicObj.obj.GetName(), metav1.GetOptions{})
					if err != nil {
						c.metrics.TemplateUpdateErrors.Inc()
						logger.Warningf(ctx, "Failed to get current resource from server [%+v] in namespace [%s] with err: %v",
							dynamicObj.obj.GetKind(), namespace, err)
						collectedErrs = append(collectedErrs, err)
						continue
					}

					modified, err := json.Marshal(dynamicObj.obj)
					if err != nil {
						c.metrics.TemplateUpdateErrors.Inc()
						logger.Warningf(ctx, "Failed to marshal resource [%+v] in namespace [%s] to json with err: %v",
							dynamicObj.obj.GetKind(), namespace, err)
						collectedErrs = append(collectedErrs, err)
						continue
					}

					patch, patchType, err := c.createPatch(dynamicObj.mapping, currentObj, modified, namespace)
					if err != nil {
						c.metrics.TemplateUpdateErrors.Inc()
						logger.Warningf(ctx, "Failed to create patch for resource [%+v] in namespace [%s] err: %v",
							dynamicObj.obj.GetKind(), namespace, err)
						collectedErrs = append(collectedErrs, err)
						continue
					}

					if string(patch) == "{}" {
						logger.Infof(ctx, "Resource [%+v] in namespace [%s] is not modified",
							dynamicObj.obj.GetKind(), namespace)
						continue
					}

					_, err = dr.Patch(ctx, dynamicObj.obj.GetName(),
						patchType, patch, metav1.PatchOptions{})
					if err != nil {
						c.metrics.TemplateUpdateErrors.Inc()
						logger.Warningf(ctx, "Failed to patch resource [%+v] in namespace [%s] with err: %v",
							dynamicObj.obj.GetKind(), namespace, err)
						collectedErrs = append(collectedErrs, err)
						continue
					}

					logger.Debugf(ctx, "Successfully updated resource [%+v] in namespace [%s]",
						dynamicObj.obj.GetKind(), namespace)
					c.appliedTemplates[namespace][templateFile.Name()] = templateFile.ModTime()
				} else {
					// Some error other than AlreadyExists was raised when we tried to Create the k8s object.
					c.metrics.KubernetesResourcesCreateErrors.Inc()
					logger.Warningf(ctx, "Failed to create kubernetes object from config template [%s] for namespace [%s] with err: %v",
						templateFileName, namespace, err)
					err := errors.NewFlyteAdminErrorf(codes.Internal,
						"Failed to create kubernetes object from config template [%s] for namespace [%s] with err: %v",
						templateFileName, namespace, err)
					collectedErrs = append(collectedErrs, err)
				}
			} else {
				logger.Debugf(ctx, "Created resource [%+v] for namespace [%s] in kubernetes",
					dynamicObj.obj.GetKind(), namespace)
				c.metrics.KubernetesResourcesCreated.Inc()
				c.appliedTemplates[namespace][templateFile.Name()] = templateFile.ModTime()
			}
		}
	}
	if len(collectedErrs) > 0 {
		return errors.NewCollectedFlyteAdminError(codes.Internal, collectedErrs)
	}
	return nil
}

var metadataAccessor = meta.NewAccessor()

// getLastApplied get last applied manifest from object's annotation
func getLastApplied(obj k8sruntime.Object) ([]byte, error) {
	annots, err := metadataAccessor.Annotations(obj)
	if err != nil {
		return nil, err
	}

	if annots == nil {
		return nil, nil
	}

	lastApplied, ok := annots[corev1.LastAppliedConfigAnnotation]
	if !ok {
		return nil, nil
	}

	return []byte(lastApplied), nil
}

func addResourceVersion(patch []byte, rv string) ([]byte, error) {
	var patchMap map[string]interface{}
	err := json.Unmarshal(patch, &patchMap)
	if err != nil {
		return nil, err
	}
	u := unstructured.Unstructured{Object: patchMap}
	a, err := meta.Accessor(&u)
	if err != nil {
		return nil, err
	}
	a.SetResourceVersion(rv)

	return json.Marshal(patchMap)
}

// createResourceFromTemplate this method perform following processes:
//      1) read template file pointed by templateDir and templateFileName
//      2) substitute templatized variables with their resolved values
// the method will return the kubernetes raw manifest
func (c *controller) createResourceFromTemplate(ctx context.Context, templateDir string,
	templateFileName string, project models.Project, domain runtimeInterfaces.Domain, namespace NamespaceName,
	templateValues, customTemplateValues templateValuesType) (string, error) {
	// 1) read the template file
	template, err := ioutil.ReadFile(path.Join(templateDir, templateFileName))
	if err != nil {
		logger.Warningf(ctx,
			"failed to read config template from path [%s] for namespace [%s] with err: %v",
			templateFileName, namespace, err)
		err := errors.NewFlyteAdminErrorf(
			codes.Internal, "failed to read config template from path [%s] for namespace [%s] with err: %v",
			templateFileName, namespace, err)
		c.metrics.TemplateReadErrors.Inc()
		return "", err
	}
	logger.Debugf(ctx, "successfully read template config file [%s]", templateFileName)

	// 2) substitute templatized variables with their resolved values
	// First, add the special case namespace template which is always substituted by the system
	// rather than fetched via a user-specified source.
	templateValues[fmt.Sprintf(templateVariableFormat, namespaceVariable)] = namespace
	templateValues[fmt.Sprintf(templateVariableFormat, projectVariable)] = project.Identifier
	templateValues[fmt.Sprintf(templateVariableFormat, domainVariable)] = domain.ID

	var k8sManifest = string(template)
	for templateKey, templateValue := range templateValues {
		k8sManifest = strings.Replace(k8sManifest, templateKey, templateValue, replaceAllInstancesOfString)
	}
	// Replace remaining template variables from domain specific defaults.
	for templateKey, templateValue := range customTemplateValues {
		k8sManifest = strings.Replace(k8sManifest, templateKey, templateValue, replaceAllInstancesOfString)
	}

	return k8sManifest, nil
}

// createPatch create 3-way merge patch of current object, original object (retrieved from last applied annotation), and the modification
// for native k8s resource, strategic merge patch is used
// for custom resource, json merge patch is used
// heavily inspired by kubectl's patcher
func (c *controller) createPatch(mapping *meta.RESTMapping, currentObj *unstructured.Unstructured, modified []byte, namespace string) ([]byte, types.PatchType, error) {
	current, err := k8sruntime.Encode(unstructured.UnstructuredJSONScheme, currentObj)
	if err != nil {
		return nil, "", fmt.Errorf("failed to encode [%+v] in namespace [%s] to json with err: %v",
			currentObj.GetKind(), namespace, err)
	}

	original, err := getLastApplied(currentObj)

	var patch []byte
	patchType := types.StrategicMergePatchType
	obj, err := scheme.Scheme.New(mapping.GroupVersionKind)
	switch {
	case err == nil:
		patchType = types.StrategicMergePatchType
		if patch == nil {
			lookupPatchMeta, err := strategicpatch.NewPatchMetaFromStruct(obj)
			if err != nil {
				return nil, "", fmt.Errorf("failed to create lookup patch meta for [%+v] in namespace [%s] with err: %v",
					currentObj.GetKind(), namespace, err)
			}

			patch, err = strategicpatch.CreateThreeWayMergePatch(original, modified, current, lookupPatchMeta, true)
			if err != nil {
				return nil, "", fmt.Errorf("failed to create 3 way merge patch for resource [%+v] in namespace [%s] with err: %v\noriginal:\n%s\nmodified:\n%s\ncurrent:\n%s",
					currentObj.GetKind(), namespace, err, original, modified, current)
			}
		}
	case k8sruntime.IsNotRegisteredError(err):
		patchType = types.MergePatchType
		preconditions := []mergepatch.PreconditionFunc{mergepatch.RequireKeyUnchanged("apiVersion"),
			mergepatch.RequireKeyUnchanged("kind"), mergepatch.RequireMetadataKeyUnchanged("name")}
		patch, err = jsonmergepatch.CreateThreeWayJSONMergePatch(original, modified, current, preconditions...)
		if err != nil {
			return nil, "", fmt.Errorf("failed to create 3 way merge patch for resource [%+v] in namespace [%s] with err: %v",
				currentObj.GetKind(), namespace, err)
		}
	case err != nil:
		return nil, "", fmt.Errorf("failed to create get instance of versioned object [%+v] in namespace [%s] with err: %v",
			currentObj.GetKind(), namespace, err)

	}

	if string(patch) == "{}" {
		// not modified
		return patch, patchType, nil
	}

	if currentObj.GetResourceVersion() != "" {
		patch, err = addResourceVersion(patch, currentObj.GetResourceVersion())
		if err != nil {
			return nil, "", fmt.Errorf("failed adding resource version for object [%+v] in namespace [%s] with err: %v",
				currentObj.GetKind(), namespace, err)
		}
	}

	return patch, patchType, nil
}

func (c *controller) Sync(ctx context.Context) error {
	defer func() {
		if err := recover(); err != nil {
			c.metrics.Panics.Inc()
			logger.Warningf(ctx, fmt.Sprintf("caught panic: %v [%+v]", err, string(debug.Stack())))
		}
	}()
	c.metrics.SyncStarted.Inc()
	logger.Debugf(ctx, "Running an invocation of ClusterResource Sync")

	// Prefer to sync projects most newly created to ensure their resources get created first when other resources exist.
	filter, err := common.NewSingleValueFilter(common.Project, common.NotEqual, "state", int32(admin.Project_ARCHIVED))
	if err != nil {
		return err
	}
	projects, err := c.db.ProjectRepo().List(ctx, repositoriesInterfaces.ListResourceInput{
		SortParameter: descCreatedAtSortParam,
		InlineFilters: []common.InlineFilter{filter},
	})
	if err != nil {
		return err
	}
	domains := c.config.ApplicationConfiguration().GetDomainsConfig()
	var errs = make([]error, 0)
	templateValues, err := populateTemplateValues(c.config.ClusterResourceConfiguration().GetTemplateData())
	if err != nil {
		logger.Warningf(ctx, "Failed to get templatized values specified in config: %v", err)
		errs = append(errs, err)
	}
	domainTemplateValues, err := populateDefaultTemplateValues(c.config.ClusterResourceConfiguration().GetCustomTemplateData())
	if err != nil {
		logger.Warningf(ctx, "Failed to get domain-specific templatized values specified in config: %v", err)
		errs = append(errs, err)
	}

	for _, project := range projects {
		for _, domain := range *domains {
			namespace := common.GetNamespaceName(c.config.NamespaceMappingConfiguration().GetNamespaceTemplate(), project.Identifier, domain.Name)
			customTemplateValues, err := c.getCustomTemplateValues(
				ctx, project.Identifier, domain.ID, domainTemplateValues[domain.ID])
			if err != nil {
				logger.Warningf(ctx, "Failed to get custom template values for %s with err: %v", namespace, err)
				errs = append(errs, err)
			}
			err = c.syncNamespace(ctx, project, domain, namespace, templateValues, customTemplateValues)
			if err != nil {
				logger.Warningf(ctx, "Failed to create cluster resources for namespace [%s] with err: %v", namespace, err)
				c.metrics.ResourceAddErrors.Inc()
				errs = append(errs, err)
			} else {
				c.metrics.ResourcesAdded.Inc()
				logger.Debugf(ctx, "Successfully created kubernetes resources for [%s]", namespace)
			}
		}
	}
	if len(errs) > 0 {
		return errors.NewCollectedFlyteAdminError(codes.Internal, errs)
	}
	return nil
}

func (c *controller) Run() {
	ctx := context.Background()
	logger.Debugf(ctx, "Running ClusterResourceController")
	interval := c.config.ClusterResourceConfiguration().GetRefreshInterval()
	wait.Forever(func() {
		err := c.Sync(ctx)
		if err != nil {
			logger.Warningf(ctx, "Failed cluster resource creation loop with: %v", err)
		}
	}, interval)
}

func newMetrics(scope promutils.Scope) controllerMetrics {
	return controllerMetrics{
		Scope: scope,
		SyncStarted: scope.MustNewCounter("k8s_resource_syncs",
			"overall count of the number of invocations of the resource controller 'sync' method"),
		KubernetesResourcesCreated: scope.MustNewCounter("k8s_resources_created",
			"overall count of successfully created resources in kubernetes"),
		KubernetesResourcesCreateErrors: scope.MustNewCounter("k8s_resource_create_errors",
			"overall count of errors encountered attempting to create resources in kubernetes"),
		ResourcesAdded: scope.MustNewCounter("resources_added",
			"overall count of successfully added resources for namespaces"),
		ResourceAddErrors: scope.MustNewCounter("resource_add_errors",
			"overall count of errors encountered creating resources for namespaces"),
		TemplateReadErrors: scope.MustNewCounter("template_read_errors",
			"errors encountered reading the yaml template file from the local filesystem"),
		TemplateDecodeErrors: scope.MustNewCounter("template_decode_errors",
			"errors encountered trying to decode yaml template into k8s go struct"),
		AppliedTemplateExists: scope.MustNewCounter("applied_template_exists",
			"Number of times the system to tried to apply an uncached resource the kubernetes reported as "+
				"already existing"),
		TemplateUpdateErrors: scope.MustNewCounter("template_update_errors",
			"Number of times an attempt at updating an already existing kubernetes resource with a template"+
				"file failed"),
		Panics: scope.MustNewCounter("panics",
			"overall count of panics encountered in primary ClusterResourceController loop"),
	}
}

func NewClusterResourceController(db repositories.RepositoryInterface, executionCluster interfaces.ClusterInterface, scope promutils.Scope) Controller {
	config := runtime.NewConfigurationProvider()

	return &controller{
		db:               db,
		config:           config,
		executionCluster: executionCluster,
		resourceManager:  resources.NewResourceManager(db, config.ApplicationConfiguration()),
		poller:           make(chan struct{}),
		metrics:          newMetrics(scope),
		appliedTemplates: make(map[string]map[string]time.Time),
	}
}
