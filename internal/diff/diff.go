package diff

import (
	"fmt"
	"reflect"
)

// ResourceDiff represents differences between local and remote resources
type ResourceDiff struct {
	ToCreate []Resource
	ToUpdate []Resource
	ToDelete []Resource
}

// Resource represents a generic APISIX resource
type Resource struct {
	Type string
	ID   string
	Data map[string]interface{}
}

// DiffResult contains all resource differences
type DiffResult struct {
	Routes        ResourceDiff
	Services      ResourceDiff
	Upstreams     ResourceDiff
	Consumers     ResourceDiff
	SSLs          ResourceDiff
	GlobalRules   ResourceDiff
	PluginConfigs ResourceDiff
	StreamRoutes  ResourceDiff
}

// HasChanges returns true if there are any differences
func (d *DiffResult) HasChanges() bool {
	return len(d.Routes.ToCreate) > 0 || len(d.Routes.ToUpdate) > 0 || len(d.Routes.ToDelete) > 0 ||
		len(d.Services.ToCreate) > 0 || len(d.Services.ToUpdate) > 0 || len(d.Services.ToDelete) > 0 ||
		len(d.Upstreams.ToCreate) > 0 || len(d.Upstreams.ToUpdate) > 0 || len(d.Upstreams.ToDelete) > 0 ||
		len(d.Consumers.ToCreate) > 0 || len(d.Consumers.ToUpdate) > 0 || len(d.Consumers.ToDelete) > 0 ||
		len(d.SSLs.ToCreate) > 0 || len(d.SSLs.ToUpdate) > 0 || len(d.SSLs.ToDelete) > 0 ||
		len(d.GlobalRules.ToCreate) > 0 || len(d.GlobalRules.ToUpdate) > 0 || len(d.GlobalRules.ToDelete) > 0 ||
		len(d.PluginConfigs.ToCreate) > 0 || len(d.PluginConfigs.ToUpdate) > 0 || len(d.PluginConfigs.ToDelete) > 0 ||
		len(d.StreamRoutes.ToCreate) > 0 || len(d.StreamRoutes.ToUpdate) > 0 || len(d.StreamRoutes.ToDelete) > 0
}

// CompareResources compares local and remote resources
func CompareResources(localResources, remoteResources []Resource) ResourceDiff {
	diff := ResourceDiff{
		ToCreate: []Resource{},
		ToUpdate: []Resource{},
		ToDelete: []Resource{},
	}

	// Create maps for quick lookup
	localMap := make(map[string]Resource)
	remoteMap := make(map[string]Resource)

	for _, r := range localResources {
		localMap[r.ID] = r
	}

	for _, r := range remoteResources {
		remoteMap[r.ID] = r
	}

	// Find resources to create or update
	for id, local := range localMap {
		if remote, exists := remoteMap[id]; exists {
			// Resource exists, check if it needs update
			if !resourcesEqual(local.Data, remote.Data) {
				diff.ToUpdate = append(diff.ToUpdate, local)
			}
		} else {
			// Resource doesn't exist remotely, needs to be created
			diff.ToCreate = append(diff.ToCreate, local)
		}
	}

	// Find resources to delete
	for id, remote := range remoteMap {
		if _, exists := localMap[id]; !exists {
			diff.ToDelete = append(diff.ToDelete, remote)
		}
	}

	return diff
}

// resourcesEqual compares two resources, ignoring system fields
func resourcesEqual(local, remote map[string]interface{}) bool {
	// Fields to ignore in comparison
	ignoreFields := map[string]bool{
		"create_time": true,
		"update_time": true,
	}

	// Create copies without ignored fields
	localCopy := make(map[string]interface{})
	remoteCopy := make(map[string]interface{})

	for k, v := range local {
		if !ignoreFields[k] {
			localCopy[k] = v
		}
	}

	for k, v := range remote {
		if !ignoreFields[k] {
			remoteCopy[k] = v
		}
	}

	return reflect.DeepEqual(localCopy, remoteCopy)
}

// PrintDiff prints the diff result in a human-readable format
func PrintDiff(result *DiffResult) {
	if !result.HasChanges() {
		fmt.Println("No changes detected")
		return
	}

	fmt.Println("Changes to be applied:")
	fmt.Println()

	printResourceDiff("Routes", result.Routes)
	printResourceDiff("Services", result.Services)
	printResourceDiff("Upstreams", result.Upstreams)
	printResourceDiff("Consumers", result.Consumers)
	printResourceDiff("SSLs", result.SSLs)
	printResourceDiff("Global Rules", result.GlobalRules)
	printResourceDiff("Plugin Configs", result.PluginConfigs)
	printResourceDiff("Stream Routes", result.StreamRoutes)
}

func printResourceDiff(resourceType string, diff ResourceDiff) {
	if len(diff.ToCreate) == 0 && len(diff.ToUpdate) == 0 && len(diff.ToDelete) == 0 {
		return
	}

	fmt.Printf("%s:\n", resourceType)

	if len(diff.ToCreate) > 0 {
		fmt.Printf("  + Create (%d):\n", len(diff.ToCreate))
		for _, r := range diff.ToCreate {
			fmt.Printf("    + %s\n", r.ID)
		}
	}

	if len(diff.ToUpdate) > 0 {
		fmt.Printf("  ~ Update (%d):\n", len(diff.ToUpdate))
		for _, r := range diff.ToUpdate {
			fmt.Printf("    ~ %s\n", r.ID)
		}
	}

	if len(diff.ToDelete) > 0 {
		fmt.Printf("  - Delete (%d):\n", len(diff.ToDelete))
		for _, r := range diff.ToDelete {
			fmt.Printf("    - %s\n", r.ID)
		}
	}

	fmt.Println()
}
