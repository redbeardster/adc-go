package sync

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/api7/adc-go/internal/apisix"
	"github.com/api7/adc-go/internal/declarative"
	"github.com/api7/adc-go/internal/diff"
)

// Syncer handles synchronization between local config and APISIX
type Syncer struct {
	client *apisix.Client
}

// NewSyncer creates a new Syncer
func NewSyncer(client *apisix.Client) *Syncer {
	return &Syncer{client: client}
}

// GetRemoteState fetches current state from APISIX
func (s *Syncer) GetRemoteState() (*declarative.DeclarativeConfig, error) {
	config := &declarative.DeclarativeConfig{
		Version: "1.0",
	}

	// Fetch routes
	routes, err := s.client.GetRoutes()
	if err != nil {
		return nil, fmt.Errorf("failed to get routes: %w", err)
	}
	for _, route := range routes {
		config.Routes = append(config.Routes, convertToRoute(route))
	}

	// Fetch services
	services, err := s.client.GetServices()
	if err != nil {
		return nil, fmt.Errorf("failed to get services: %w", err)
	}
	for _, service := range services {
		config.Services = append(config.Services, convertToService(service))
	}

	// Fetch upstreams
	upstreams, err := s.client.GetUpstreams()
	if err != nil {
		return nil, fmt.Errorf("failed to get upstreams: %w", err)
	}
	for _, upstream := range upstreams {
		config.Upstreams = append(config.Upstreams, convertToUpstream(upstream))
	}

	// Fetch consumers
	consumers, err := s.client.GetConsumers()
	if err != nil {
		return nil, fmt.Errorf("failed to get consumers: %w", err)
	}
	for _, consumer := range consumers {
		config.Consumers = append(config.Consumers, convertToConsumer(consumer))
	}

	// Fetch SSLs
	ssls, err := s.client.GetSSLs()
	if err != nil {
		return nil, fmt.Errorf("failed to get ssls: %w", err)
	}
	for _, ssl := range ssls {
		config.SSLs = append(config.SSLs, convertToSSL(ssl))
	}

	// Fetch global rules
	globalRules, err := s.client.GetGlobalRules()
	if err != nil {
		return nil, fmt.Errorf("failed to get global rules: %w", err)
	}
	for _, rule := range globalRules {
		config.GlobalRules = append(config.GlobalRules, convertToGlobalRule(rule))
	}

	// Fetch plugin configs
	pluginConfigs, err := s.client.GetPluginConfigs()
	if err != nil {
		return nil, fmt.Errorf("failed to get plugin configs: %w", err)
	}
	for _, pc := range pluginConfigs {
		config.PluginConfigs = append(config.PluginConfigs, convertToPluginConfig(pc))
	}

	// Fetch stream routes
	streamRoutes, err := s.client.GetStreamRoutes()
	if err != nil {
		// Stream routes may be disabled, ignore this error
		if !strings.Contains(err.Error(), "stream mode is disabled") {
			return nil, fmt.Errorf("failed to get stream routes: %w", err)
		}
		// Stream mode disabled, skip
	} else {
		for _, sr := range streamRoutes {
			config.StreamRoutes = append(config.StreamRoutes, convertToStreamRoute(sr))
		}
	}

	return config, nil
}

// CalculateDiff compares local and remote configurations
func (s *Syncer) CalculateDiff(local, remote *declarative.DeclarativeConfig) *diff.DiffResult {
	result := &diff.DiffResult{}

	// Compare routes
	result.Routes = diff.CompareResources(
		routesToResources(local.Routes),
		routesToResources(remote.Routes),
	)

	// Compare services
	result.Services = diff.CompareResources(
		servicesToResources(local.Services),
		servicesToResources(remote.Services),
	)

	// Compare upstreams
	result.Upstreams = diff.CompareResources(
		upstreamsToResources(local.Upstreams),
		upstreamsToResources(remote.Upstreams),
	)

	// Compare consumers
	result.Consumers = diff.CompareResources(
		consumersToResources(local.Consumers),
		consumersToResources(remote.Consumers),
	)

	// Compare SSLs
	result.SSLs = diff.CompareResources(
		sslsToResources(local.SSLs),
		sslsToResources(remote.SSLs),
	)

	// Compare global rules
	result.GlobalRules = diff.CompareResources(
		globalRulesToResources(local.GlobalRules),
		globalRulesToResources(remote.GlobalRules),
	)

	// Compare plugin configs
	result.PluginConfigs = diff.CompareResources(
		pluginConfigsToResources(local.PluginConfigs),
		pluginConfigsToResources(remote.PluginConfigs),
	)

	// Compare stream routes
	result.StreamRoutes = diff.CompareResources(
		streamRoutesToResources(local.StreamRoutes),
		streamRoutesToResources(remote.StreamRoutes),
	)

	return result
}

// ApplyDiff applies the calculated diff to APISIX
func (s *Syncer) ApplyDiff(diffResult *diff.DiffResult, deleteRemoved bool) error {
	// Apply in dependency order: upstreams -> services -> routes

	// 1. Create/Update Upstreams first
	if err := s.applyUpstreams(diffResult.Upstreams); err != nil {
		return fmt.Errorf("failed to apply upstreams: %w", err)
	}

	// 2. Create/Update Services
	if err := s.applyServices(diffResult.Services); err != nil {
		return fmt.Errorf("failed to apply services: %w", err)
	}

	// 3. Create/Update Routes
	if err := s.applyRoutes(diffResult.Routes); err != nil {
		return fmt.Errorf("failed to apply routes: %w", err)
	}

	// 4. Apply other resources
	if err := s.applyConsumers(diffResult.Consumers); err != nil {
		return fmt.Errorf("failed to apply consumers: %w", err)
	}

	if err := s.applySSLs(diffResult.SSLs); err != nil {
		return fmt.Errorf("failed to apply ssls: %w", err)
	}

	if err := s.applyGlobalRules(diffResult.GlobalRules); err != nil {
		return fmt.Errorf("failed to apply global rules: %w", err)
	}

	if err := s.applyPluginConfigs(diffResult.PluginConfigs); err != nil {
		return fmt.Errorf("failed to apply plugin configs: %w", err)
	}

	if err := s.applyStreamRoutes(diffResult.StreamRoutes); err != nil {
		return fmt.Errorf("failed to apply stream routes: %w", err)
	}

	// Delete resources if requested (in reverse dependency order)
	if deleteRemoved {
		if err := s.deleteRoutes(diffResult.Routes.ToDelete); err != nil {
			return fmt.Errorf("failed to delete routes: %w", err)
		}

		if err := s.deleteServices(diffResult.Services.ToDelete); err != nil {
			return fmt.Errorf("failed to delete services: %w", err)
		}

		if err := s.deleteUpstreams(diffResult.Upstreams.ToDelete); err != nil {
			return fmt.Errorf("failed to delete upstreams: %w", err)
		}

		if err := s.deleteConsumers(diffResult.Consumers.ToDelete); err != nil {
			return fmt.Errorf("failed to delete consumers: %w", err)
		}

		if err := s.deleteSSLs(diffResult.SSLs.ToDelete); err != nil {
			return fmt.Errorf("failed to delete ssls: %w", err)
		}

		if err := s.deleteGlobalRules(diffResult.GlobalRules.ToDelete); err != nil {
			return fmt.Errorf("failed to delete global rules: %w", err)
		}

		if err := s.deletePluginConfigs(diffResult.PluginConfigs.ToDelete); err != nil {
			return fmt.Errorf("failed to delete plugin configs: %w", err)
		}

		if err := s.deleteStreamRoutes(diffResult.StreamRoutes.ToDelete); err != nil {
			return fmt.Errorf("failed to delete stream routes: %w", err)
		}
	}

	return nil
}

// Helper methods for applying specific resource types
func (s *Syncer) applyRoutes(resourceDiff diff.ResourceDiff) error {
	for _, r := range resourceDiff.ToCreate {
		if err := s.client.CreateRoute(r.ID, r.Data); err != nil {
			return fmt.Errorf("failed to create route %s: %w", r.ID, err)
		}
		fmt.Printf("✓ Created route: %s\n", r.ID)
	}

	for _, r := range resourceDiff.ToUpdate {
		if err := s.client.UpdateRoute(r.ID, r.Data); err != nil {
			return fmt.Errorf("failed to update route %s: %w", r.ID, err)
		}
		fmt.Printf("✓ Updated route: %s\n", r.ID)
	}

	return nil
}

func (s *Syncer) applyServices(resourceDiff diff.ResourceDiff) error {
	for _, r := range resourceDiff.ToCreate {
		if err := s.client.CreateService(r.ID, r.Data); err != nil {
			return fmt.Errorf("failed to create service %s: %w", r.ID, err)
		}
		fmt.Printf("✓ Created service: %s\n", r.ID)
	}

	for _, r := range resourceDiff.ToUpdate {
		if err := s.client.UpdateService(r.ID, r.Data); err != nil {
			return fmt.Errorf("failed to update service %s: %w", r.ID, err)
		}
		fmt.Printf("✓ Updated service: %s\n", r.ID)
	}

	return nil
}

func (s *Syncer) applyUpstreams(resourceDiff diff.ResourceDiff) error {
	for _, r := range resourceDiff.ToCreate {
		if err := s.client.CreateUpstream(r.ID, r.Data); err != nil {
			return fmt.Errorf("failed to create upstream %s: %w", r.ID, err)
		}
		fmt.Printf("✓ Created upstream: %s\n", r.ID)
	}

	for _, r := range resourceDiff.ToUpdate {
		if err := s.client.UpdateUpstream(r.ID, r.Data); err != nil {
			return fmt.Errorf("failed to update upstream %s: %w", r.ID, err)
		}
		fmt.Printf("✓ Updated upstream: %s\n", r.ID)
	}

	return nil
}

func (s *Syncer) applyConsumers(resourceDiff diff.ResourceDiff) error {
	for _, r := range resourceDiff.ToCreate {
		if err := s.client.CreateConsumer(r.ID, r.Data); err != nil {
			return fmt.Errorf("failed to create consumer %s: %w", r.ID, err)
		}
		fmt.Printf("✓ Created consumer: %s\n", r.ID)
	}

	for _, r := range resourceDiff.ToUpdate {
		if err := s.client.UpdateConsumer(r.ID, r.Data); err != nil {
			return fmt.Errorf("failed to update consumer %s: %w", r.ID, err)
		}
		fmt.Printf("✓ Updated consumer: %s\n", r.ID)
	}

	return nil
}

func (s *Syncer) applySSLs(resourceDiff diff.ResourceDiff) error {
	for _, r := range resourceDiff.ToCreate {
		if err := s.client.CreateSSL(r.ID, r.Data); err != nil {
			return fmt.Errorf("failed to create ssl %s: %w", r.ID, err)
		}
		fmt.Printf("✓ Created ssl: %s\n", r.ID)
	}

	for _, r := range resourceDiff.ToUpdate {
		if err := s.client.UpdateSSL(r.ID, r.Data); err != nil {
			return fmt.Errorf("failed to update ssl %s: %w", r.ID, err)
		}
		fmt.Printf("✓ Updated ssl: %s\n", r.ID)
	}

	return nil
}

func (s *Syncer) applyGlobalRules(resourceDiff diff.ResourceDiff) error {
	for _, r := range resourceDiff.ToCreate {
		if err := s.client.CreateGlobalRule(r.ID, r.Data); err != nil {
			return fmt.Errorf("failed to create global rule %s: %w", r.ID, err)
		}
		fmt.Printf("✓ Created global rule: %s\n", r.ID)
	}

	for _, r := range resourceDiff.ToUpdate {
		if err := s.client.UpdateGlobalRule(r.ID, r.Data); err != nil {
			return fmt.Errorf("failed to update global rule %s: %w", r.ID, err)
		}
		fmt.Printf("✓ Updated global rule: %s\n", r.ID)
	}

	return nil
}

func (s *Syncer) applyPluginConfigs(resourceDiff diff.ResourceDiff) error {
	for _, r := range resourceDiff.ToCreate {
		if err := s.client.CreatePluginConfig(r.ID, r.Data); err != nil {
			return fmt.Errorf("failed to create plugin config %s: %w", r.ID, err)
		}
		fmt.Printf("✓ Created plugin config: %s\n", r.ID)
	}

	for _, r := range resourceDiff.ToUpdate {
		if err := s.client.UpdatePluginConfig(r.ID, r.Data); err != nil {
			return fmt.Errorf("failed to update plugin config %s: %w", r.ID, err)
		}
		fmt.Printf("✓ Updated plugin config: %s\n", r.ID)
	}

	return nil
}

func (s *Syncer) applyStreamRoutes(resourceDiff diff.ResourceDiff) error {
	for _, r := range resourceDiff.ToCreate {
		if err := s.client.CreateStreamRoute(r.ID, r.Data); err != nil {
			return fmt.Errorf("failed to create stream route %s: %w", r.ID, err)
		}
		fmt.Printf("✓ Created stream route: %s\n", r.ID)
	}

	for _, r := range resourceDiff.ToUpdate {
		if err := s.client.UpdateStreamRoute(r.ID, r.Data); err != nil {
			return fmt.Errorf("failed to update stream route %s: %w", r.ID, err)
		}
		fmt.Printf("✓ Updated stream route: %s\n", r.ID)
	}

	return nil
}

// Delete methods
func (s *Syncer) deleteRoutes(resources []diff.Resource) error {
	for _, r := range resources {
		if err := s.client.DeleteRoute(r.ID); err != nil {
			return fmt.Errorf("failed to delete route %s: %w", r.ID, err)
		}
		fmt.Printf("✓ Deleted route: %s\n", r.ID)
	}
	return nil
}

func (s *Syncer) deleteServices(resources []diff.Resource) error {
	for _, r := range resources {
		if err := s.client.DeleteService(r.ID); err != nil {
			return fmt.Errorf("failed to delete service %s: %w", r.ID, err)
		}
		fmt.Printf("✓ Deleted service: %s\n", r.ID)
	}
	return nil
}

func (s *Syncer) deleteUpstreams(resources []diff.Resource) error {
	for _, r := range resources {
		if err := s.client.DeleteUpstream(r.ID); err != nil {
			return fmt.Errorf("failed to delete upstream %s: %w", r.ID, err)
		}
		fmt.Printf("✓ Deleted upstream: %s\n", r.ID)
	}
	return nil
}

func (s *Syncer) deleteConsumers(resources []diff.Resource) error {
	for _, r := range resources {
		if err := s.client.DeleteConsumer(r.ID); err != nil {
			return fmt.Errorf("failed to delete consumer %s: %w", r.ID, err)
		}
		fmt.Printf("✓ Deleted consumer: %s\n", r.ID)
	}
	return nil
}

func (s *Syncer) deleteSSLs(resources []diff.Resource) error {
	for _, r := range resources {
		if err := s.client.DeleteSSL(r.ID); err != nil {
			return fmt.Errorf("failed to delete ssl %s: %w", r.ID, err)
		}
		fmt.Printf("✓ Deleted ssl: %s\n", r.ID)
	}
	return nil
}

func (s *Syncer) deleteGlobalRules(resources []diff.Resource) error {
	for _, r := range resources {
		if err := s.client.DeleteGlobalRule(r.ID); err != nil {
			return fmt.Errorf("failed to delete global rule %s: %w", r.ID, err)
		}
		fmt.Printf("✓ Deleted global rule: %s\n", r.ID)
	}
	return nil
}

func (s *Syncer) deletePluginConfigs(resources []diff.Resource) error {
	for _, r := range resources {
		if err := s.client.DeletePluginConfig(r.ID); err != nil {
			return fmt.Errorf("failed to delete plugin config %s: %w", r.ID, err)
		}
		fmt.Printf("✓ Deleted plugin config: %s\n", r.ID)
	}
	return nil
}

func (s *Syncer) deleteStreamRoutes(resources []diff.Resource) error {
	for _, r := range resources {
		if err := s.client.DeleteStreamRoute(r.ID); err != nil {
			return fmt.Errorf("failed to delete stream route %s: %w", r.ID, err)
		}
		fmt.Printf("✓ Deleted stream route: %s\n", r.ID)
	}
	return nil
}

// Conversion helpers
func routesToResources(routes []declarative.Route) []diff.Resource {
	resources := make([]diff.Resource, 0, len(routes))
	for _, route := range routes {
		data, _ := structToMap(route)
		resources = append(resources, diff.Resource{
			Type: "route",
			ID:   route.ID,
			Data: data,
		})
	}
	return resources
}

func servicesToResources(services []declarative.Service) []diff.Resource {
	resources := make([]diff.Resource, 0, len(services))
	for _, service := range services {
		data, _ := structToMap(service)
		resources = append(resources, diff.Resource{
			Type: "service",
			ID:   service.ID,
			Data: data,
		})
	}
	return resources
}

func upstreamsToResources(upstreams []declarative.Upstream) []diff.Resource {
	resources := make([]diff.Resource, 0, len(upstreams))
	for _, upstream := range upstreams {
		data, _ := structToMap(upstream)
		resources = append(resources, diff.Resource{
			Type: "upstream",
			ID:   upstream.ID,
			Data: data,
		})
	}
	return resources
}

func consumersToResources(consumers []declarative.Consumer) []diff.Resource {
	resources := make([]diff.Resource, 0, len(consumers))
	for _, consumer := range consumers {
		data, _ := structToMap(consumer)
		resources = append(resources, diff.Resource{
			Type: "consumer",
			ID:   consumer.Username,
			Data: data,
		})
	}
	return resources
}

func sslsToResources(ssls []declarative.SSL) []diff.Resource {
	resources := make([]diff.Resource, 0, len(ssls))
	for _, ssl := range ssls {
		data, _ := structToMap(ssl)
		resources = append(resources, diff.Resource{
			Type: "ssl",
			ID:   ssl.ID,
			Data: data,
		})
	}
	return resources
}

func globalRulesToResources(rules []declarative.GlobalRule) []diff.Resource {
	resources := make([]diff.Resource, 0, len(rules))
	for _, rule := range rules {
		data, _ := structToMap(rule)
		resources = append(resources, diff.Resource{
			Type: "global_rule",
			ID:   rule.ID,
			Data: data,
		})
	}
	return resources
}

func pluginConfigsToResources(configs []declarative.PluginConfig) []diff.Resource {
	resources := make([]diff.Resource, 0, len(configs))
	for _, config := range configs {
		data, _ := structToMap(config)
		resources = append(resources, diff.Resource{
			Type: "plugin_config",
			ID:   config.ID,
			Data: data,
		})
	}
	return resources
}

func streamRoutesToResources(routes []declarative.StreamRoute) []diff.Resource {
	resources := make([]diff.Resource, 0, len(routes))
	for _, route := range routes {
		data, _ := structToMap(route)
		resources = append(resources, diff.Resource{
			Type: "stream_route",
			ID:   route.ID,
			Data: data,
		})
	}
	return resources
}

func structToMap(v interface{}) (map[string]interface{}, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	return result, nil
}

// Conversion from API responses to declarative types
func convertToRoute(data map[string]interface{}) declarative.Route {
	route := declarative.Route{}
	if id, ok := data["id"].(string); ok {
		route.ID = id
	}
	if name, ok := data["name"].(string); ok {
		route.Name = name
	}
	if desc, ok := data["desc"].(string); ok {
		route.Desc = desc
	}
	if uri, ok := data["uri"].(string); ok {
		route.URI = uri
	}
	if plugins, ok := data["plugins"].(map[string]interface{}); ok {
		route.Plugins = plugins
	}
	if upstreamID, ok := data["upstream_id"].(string); ok {
		route.UpstreamID = upstreamID
	}
	return route
}

func convertToService(data map[string]interface{}) declarative.Service {
	service := declarative.Service{}
	if id, ok := data["id"].(string); ok {
		service.ID = id
	}
	if name, ok := data["name"].(string); ok {
		service.Name = name
	}
	if desc, ok := data["desc"].(string); ok {
		service.Desc = desc
	}
	if upstreamID, ok := data["upstream_id"].(string); ok {
		service.UpstreamID = upstreamID
	}
	if plugins, ok := data["plugins"].(map[string]interface{}); ok {
		service.Plugins = plugins
	}
	return service
}

func convertToUpstream(data map[string]interface{}) declarative.Upstream {
	upstream := declarative.Upstream{}
	if id, ok := data["id"].(string); ok {
		upstream.ID = id
	}
	if name, ok := data["name"].(string); ok {
		upstream.Name = name
	}
	if nodes, ok := data["nodes"].(map[string]interface{}); ok {
		upstream.Nodes = make(map[string]int)
		for k, v := range nodes {
			if weight, ok := v.(float64); ok {
				upstream.Nodes[k] = int(weight)
			}
		}
	}
	return upstream
}

func convertToConsumer(data map[string]interface{}) declarative.Consumer {
	consumer := declarative.Consumer{}
	if username, ok := data["username"].(string); ok {
		consumer.Username = username
	}
	if desc, ok := data["desc"].(string); ok {
		consumer.Desc = desc
	}
	if plugins, ok := data["plugins"].(map[string]interface{}); ok {
		consumer.Plugins = plugins
	}
	return consumer
}

func convertToSSL(data map[string]interface{}) declarative.SSL {
	ssl := declarative.SSL{}
	if id, ok := data["id"].(string); ok {
		ssl.ID = id
	}
	if cert, ok := data["cert"].(string); ok {
		ssl.Cert = cert
	}
	if key, ok := data["key"].(string); ok {
		ssl.Key = key
	}
	if snis, ok := data["snis"].([]interface{}); ok {
		for _, sni := range snis {
			if s, ok := sni.(string); ok {
				ssl.Sni = append(ssl.Sni, s)
			}
		}
	}
	return ssl
}

func convertToGlobalRule(data map[string]interface{}) declarative.GlobalRule {
	rule := declarative.GlobalRule{}
	if id, ok := data["id"].(string); ok {
		rule.ID = id
	}
	if plugins, ok := data["plugins"].(map[string]interface{}); ok {
		rule.Plugins = plugins
	}
	return rule
}

func convertToPluginConfig(data map[string]interface{}) declarative.PluginConfig {
	pc := declarative.PluginConfig{}
	if id, ok := data["id"].(string); ok {
		pc.ID = id
	}
	if desc, ok := data["desc"].(string); ok {
		pc.Desc = desc
	}
	if plugins, ok := data["plugins"].(map[string]interface{}); ok {
		pc.Plugins = plugins
	}
	return pc
}

func convertToStreamRoute(data map[string]interface{}) declarative.StreamRoute {
	sr := declarative.StreamRoute{}
	if id, ok := data["id"].(string); ok {
		sr.ID = id
	}
	if desc, ok := data["desc"].(string); ok {
		sr.Desc = desc
	}
	if upstreamID, ok := data["upstream_id"].(string); ok {
		sr.UpstreamID = upstreamID
	}
	if plugins, ok := data["plugins"].(map[string]interface{}); ok {
		sr.Plugins = plugins
	}
	return sr

}
