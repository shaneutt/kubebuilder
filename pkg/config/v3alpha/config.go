/*
Copyright 2021 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v3alpha

import (
	"fmt"
	"strings"

	"sigs.k8s.io/yaml"

	"sigs.k8s.io/kubebuilder/v3/pkg/config"
	"sigs.k8s.io/kubebuilder/v3/pkg/model/resource"
	"sigs.k8s.io/kubebuilder/v3/pkg/model/stage"
)

// Version is the config.Version for project configuration 3-alpha
var Version = config.Version{Number: 3, Stage: stage.Alpha}

type cfg struct {
	// Version
	Version config.Version `json:"version"`

	// String fields
	Domain     string `json:"domain,omitempty"`
	Repository string `json:"repo,omitempty"`
	Name       string `json:"projectName,omitempty"`
	Layout     string `json:"layout,omitempty"`

	// Boolean fields
	MultiGroup      bool `json:"multigroup,omitempty"`
	ComponentConfig bool `json:"componentConfig,omitempty"`

	// Resources
	Resources []resource.Resource `json:"resources,omitempty"`

	// Plugins
	Plugins PluginConfigs `json:"plugins,omitempty"`
}

// PluginConfigs holds a set of arbitrary plugin configuration objects mapped by plugin key.
// TODO: do not export this once internalconfig has merged with config
type PluginConfigs map[string]pluginConfig

// pluginConfig is an arbitrary plugin configuration object.
type pluginConfig interface{}

// New returns a new config.Config
func New() config.Config {
	return &cfg{Version: Version}
}

func init() {
	config.Register(Version, New)
}

// GetVersion implements config.Config
func (c cfg) GetVersion() config.Version {
	return c.Version
}

// GetDomain implements config.Config
func (c cfg) GetDomain() string {
	return c.Domain
}

// SetDomain implements config.Config
func (c *cfg) SetDomain(domain string) error {
	c.Domain = domain
	return nil
}

// GetRepository implements config.Config
func (c cfg) GetRepository() string {
	return c.Repository
}

// SetRepository implements config.Config
func (c *cfg) SetRepository(repository string) error {
	c.Repository = repository
	return nil
}

// GetProjectName implements config.Config
func (c cfg) GetProjectName() string {
	return c.Name
}

// SetProjectName implements config.Config
func (c *cfg) SetProjectName(name string) error {
	c.Name = name
	return nil
}

// GetLayout implements config.Config
func (c cfg) GetLayout() string {
	return c.Layout
}

// SetLayout implements config.Config
func (c *cfg) SetLayout(layout string) error {
	c.Layout = layout
	return nil
}

// IsMultiGroup implements config.Config
func (c cfg) IsMultiGroup() bool {
	return c.MultiGroup
}

// SetMultiGroup implements config.Config
func (c *cfg) SetMultiGroup() error {
	c.MultiGroup = true
	return nil
}

// ClearMultiGroup implements config.Config
func (c *cfg) ClearMultiGroup() error {
	c.MultiGroup = false
	return nil
}

// IsComponentConfig implements config.Config
func (c cfg) IsComponentConfig() bool {
	return c.ComponentConfig
}

// SetComponentConfig implements config.Config
func (c *cfg) SetComponentConfig() error {
	c.ComponentConfig = true
	return nil
}

// ClearComponentConfig implements config.Config
func (c *cfg) ClearComponentConfig() error {
	c.ComponentConfig = false
	return nil
}

// ResourcesLength implements config.Config
func (c cfg) ResourcesLength() int {
	return len(c.Resources)
}

// HasResource implements config.Config
func (c cfg) HasResource(gvk resource.GVK) bool {
	gvk.Domain = "" // Version 3 alpha does not include domain per resource

	for _, res := range c.Resources {
		if gvk.IsEqualTo(res.GVK) {
			return true
		}
	}

	return false
}

// GetResource implements config.Config
func (c cfg) GetResource(gvk resource.GVK) (resource.Resource, error) {
	gvk.Domain = "" // Version 3 alpha does not include domain per resource

	for _, res := range c.Resources {
		if gvk.IsEqualTo(res.GVK) {
			return res.Copy(), nil
		}
	}

	return resource.Resource{}, config.UnknownResource{GVK: gvk}
}

// GetResources implements config.Config
func (c cfg) GetResources() ([]resource.Resource, error) {
	resources := make([]resource.Resource, 0, len(c.Resources))
	for _, res := range c.Resources {
		resources = append(resources, res.Copy())
	}

	return resources, nil
}

func discardNonIncludedFields(res *resource.Resource) {
	res.Domain = "" // Version 3 alpha does not include domain per resource
	res.Plural = "" // Version 3 alpha does not include plural forms
	res.Path = ""   // Version 3 alpha does not include paths
	if res.API != nil {
		res.API.Namespaced = false // Version 3 alpha does not include if the api was namespaced
	}
	res.Controller = false // Version 3 alpha does not include if the controller was scaffolded
	if res.Webhooks != nil {
		res.Webhooks.Defaulting = false // Version 3 alpha does not include if the defaulting webhook was scaffolded
		res.Webhooks.Validation = false // Version 3 alpha does not include if the validation webhook was scaffolded
		res.Webhooks.Conversion = false // Version 3 alpha does not include if the conversion webhook was scaffolded
	}
}

// AddResource implements config.Config
func (c *cfg) AddResource(res resource.Resource) error {
	// As res is passed by value it is already a shallow copy, but we need to make a deep copy
	res = res.Copy()

	discardNonIncludedFields(&res) // Version 3 alpha does not include several fields from the Resource model

	if !c.HasResource(res.GVK) {
		c.Resources = append(c.Resources, res)
	}
	return nil
}

// UpdateResource implements config.Config
func (c *cfg) UpdateResource(res resource.Resource) error {
	// As res is passed by value it is already a shallow copy, but we need to make a deep copy
	res = res.Copy()

	discardNonIncludedFields(&res) // Version 3 alpha does not include several fields from the Resource model

	for i, r := range c.Resources {
		if res.GVK.IsEqualTo(r.GVK) {
			return c.Resources[i].Update(res)
		}
	}

	c.Resources = append(c.Resources, res)
	return nil
}

// HasGroup implements config.Config
func (c cfg) HasGroup(group string) bool {
	// Return true if the target group is found in the tracked resources
	for _, r := range c.Resources {
		if strings.EqualFold(group, r.Group) {
			return true
		}
	}

	// Return false otherwise
	return false
}

// IsCRDVersionCompatible implements config.Config
func (c cfg) IsCRDVersionCompatible(crdVersion string) bool {
	return c.resourceAPIVersionCompatible("crd", crdVersion)
}

// IsWebhookVersionCompatible implements config.Config
func (c cfg) IsWebhookVersionCompatible(webhookVersion string) bool {
	return c.resourceAPIVersionCompatible("webhook", webhookVersion)
}

func (c cfg) resourceAPIVersionCompatible(verType, version string) bool {
	for _, res := range c.Resources {
		var currVersion string
		switch verType {
		case "crd":
			if res.API != nil {
				currVersion = res.API.CRDVersion
			}
		case "webhook":
			if res.Webhooks != nil {
				currVersion = res.Webhooks.WebhookVersion
			}
		}
		if currVersion != "" && version != currVersion {
			return false
		}
	}

	return true
}

// DecodePluginConfig implements config.Config
func (c cfg) DecodePluginConfig(key string, configObj interface{}) error {
	if len(c.Plugins) == 0 {
		return nil
	}

	// Get the object blob by key and unmarshal into the object.
	if pluginConfig, hasKey := c.Plugins[key]; hasKey {
		b, err := yaml.Marshal(pluginConfig)
		if err != nil {
			return fmt.Errorf("failed to convert extra fields object to bytes: %w", err)
		}
		if err := yaml.Unmarshal(b, configObj); err != nil {
			return fmt.Errorf("failed to unmarshal extra fields object: %w", err)
		}
	}

	return nil
}

// EncodePluginConfig will return an error if used on any project version < v3.
func (c *cfg) EncodePluginConfig(key string, configObj interface{}) error {
	// Get object's bytes and set them under key in extra fields.
	b, err := yaml.Marshal(configObj)
	if err != nil {
		return fmt.Errorf("failed to convert %T object to bytes: %s", configObj, err)
	}
	var fields map[string]interface{}
	if err := yaml.Unmarshal(b, &fields); err != nil {
		return fmt.Errorf("failed to unmarshal %T object bytes: %s", configObj, err)
	}
	if c.Plugins == nil {
		c.Plugins = make(map[string]pluginConfig)
	}
	c.Plugins[key] = fields
	return nil
}

// Marshal implements config.Config
func (c cfg) Marshal() ([]byte, error) {
	for i, r := range c.Resources {
		// If API is empty, omit it (prevents `api: {}`).
		if r.API != nil && r.API.IsEmpty() {
			c.Resources[i].API = nil
		}
		// If Webhooks is empty, omit it (prevents `webhooks: {}`).
		if r.Webhooks != nil && r.Webhooks.IsEmpty() {
			c.Resources[i].Webhooks = nil
		}
	}

	content, err := yaml.Marshal(c)
	if err != nil {
		return nil, config.MarshalError{Err: err}
	}

	return content, nil
}

// Unmarshal implements config.Config
func (c *cfg) Unmarshal(b []byte) error {
	if err := yaml.UnmarshalStrict(b, c); err != nil {
		return config.UnmarshalError{Err: err}
	}

	return nil
}
