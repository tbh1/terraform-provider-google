package google

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"google.golang.org/api/compute/v1"
)

func dataSourceGoogleComputeRouterStatus() *schema.Resource {
	routeElemSchema := datasourceSchemaFromResourceSchema(resourceComputeRoute().Schema)

	return &schema.Resource{
		Read: dataSourceComputeRouterRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: "Name of the router to query.",
				Required:    true,
			},
			"project": {
				Type:        schema.TypeString,
				Description: "Project ID of the target router.",
				Required:    false,
			},
			"region": {
				Type:        schema.TypeString,
				Description: "Region of the target router.",
				Required:    false,
			},
			"network": {
				Type:        schema.TypeString,
				Description: "URI of the network to which this router belongs.",
				Computed:    true,
			},
			"best_routes": {
				Type:        schema.TypeList,
				Description: "Best routes for this router's network.",
				Elem:        routeElemSchema,
				Computed:    true,
				Required:    false,
			},
			"best_routes_for_router": {
				Type:        schema.TypeList,
				Description: "Best routes learned by this router.",
				Elem:        routeElemSchema,
				Computed:    true,
				Required:    false,
			},
		},
	}
}

func dataSourceComputeRouterStatusRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	userAgent, err := generateUserAgentString(d, config.userAgent)
	if err != nil {
		return err
	}

	region := config.Region
	if r, ok := d.GetOk("region"); ok {
		region = r.(string)
	}

	project, err := getProject(d, config)
	if err != nil {
		return err
	}

	resp, err := config.NewComputeClient(userAgent).Routers.GetRouterStatus(project, region, d.Id()).Do()
	if err != nil {
		return err
	}

	status := resp.Result

	if err := d.Set("network", status.Network); err != nil {
		return fmt.Errorf("Error setting network: %s", err)
	}

	if err := mapRoutes(d, "best_routes", status.BestRoutes); err != nil {
		return fmt.Errorf("Error setting best_routes: %s", err)
	}

	if err := mapRoutes(d, "best_routes_for_router", status.BestRoutesForRouter); err != nil {
		return fmt.Errorf("Error setting best_routes_for_router: %s", err)
	}

	return nil
}

func mapRoutes(d *schema.ResourceData, field string, routes []*compute.Route) error {
	results := make([]map[string]interface{}, len(routes))

	for _, route := range routes {
		output := make(map[string]interface{})
		output["dest_range"] = route.DestRange
		output["name"] = route.Name
		output["network"] = route.Network
		output["description"] = route.Description
		output["next_hop_gateway"] = route.NextHopGateway
		output["next_hop_ilb"] = route.NextHopIlb
		output["next_hop_ip"] = route.NextHopIp
		output["next_hop_vpn_tunnel"] = route.NextHopVpnTunnel
		output["priority"] = route.Priority
		output["tags"] = route.Tags
		output["next_hop_network"] = route.NextHopNetwork
	}

	return d.Set(field, results)
}
