package ovh

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceIpLoadbalancingHttpFarmServer() *schema.Resource {
	return &schema.Resource{
		Create: resourceIpLoadbalancingHttpFarmServerCreate,
		Read:   resourceIpLoadbalancingHttpFarmServerRead,
		Update: resourceIpLoadbalancingHttpFarmServerUpdate,
		Delete: resourceIpLoadbalancingHttpFarmServerDelete,
		Importer: &schema.ResourceImporter{
			State: resourceIpLoadbalancingHttpFarmServerImportState,
		},

		Schema: map[string]*schema.Schema{
			"service_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"farm_id": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},
			"display_name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"address": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
					ip := v.(string)
					if net.ParseIP(ip).To4() == nil {
						errors = append(errors, fmt.Errorf("Address %s is not an IPv4", ip))
					}
					return
				},
			},
			"ssl": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"cookie": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"port": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"proxy_protocol_version": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
					err := validateStringEnum(v.(string), []string{"v1", "v2", "v2-ssl", "v2-ssl-cn"})
					if err != nil {
						errors = append(errors, err)
					}
					return
				},
			},
			"chain": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"weight": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  1,
			},
			"probe": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"backup": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"status": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
					err := validateStringEnum(v.(string), []string{"active", "inactive"})
					if err != nil {
						errors = append(errors, err)
					}
					return
				},
			},
		},
	}
}

func resourceIpLoadbalancingHttpFarmServerImportState(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	givenId := d.Id()
	splitId := strings.SplitN(givenId, "/", 3)
	if len(splitId) != 3 {
		return nil, fmt.Errorf("Import Id is not service_name/farm_id/server id formatted")
	}
	serviceName := splitId[0]
	farmId, err := strconv.Atoi(splitId[1])
	if err != nil {
		return nil, fmt.Errorf("Couldn't cast farmId %d to int: %s", farmId, err.Error())
	}
	serverId := splitId[2]

	d.SetId(serverId)
	d.Set("farm_id", farmId)
	d.Set("service_name", serviceName)

	results := make([]*schema.ResourceData, 1)
	results[0] = d
	return results, nil
}

func resourceIpLoadbalancingHttpFarmServerCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	newBackendServer := &IpLoadbalancingFarmServerCreateOpts{
		DisplayName:          getNilStringPointerFromData(d, "display_name"),
		Address:              d.Get("address").(string),
		Port:                 getNilIntPointerFromData(d, "port"),
		ProxyProtocolVersion: getNilStringPointerFromData(d, "proxy_protocol_version"),
		Chain:                getNilStringPointerFromData(d, "chain"),
		Weight:               getNilIntPointerFromData(d, "weight"),
		Probe:                getNilBoolPointer(d.Get("probe")),
		Ssl:                  getNilBoolPointer(d.Get("ssl")),
		Backup:               getNilBoolPointer(d.Get("backup")),
		Status:               d.Get("status").(string),
	}

	service := d.Get("service_name").(string)
	farmid := d.Get("farm_id").(int)
	r := &IpLoadbalancingFarmServer{}
	endpoint := fmt.Sprintf("/ipLoadbalancing/%s/http/farm/%d/server", service, farmid)

	err := config.OVHClient.Post(endpoint, newBackendServer, r)
	if err != nil {
		return fmt.Errorf("calling POST %s with %d:\n\t %s", endpoint, farmid, err.Error())
	}

	//set id
	d.SetId(fmt.Sprintf("%d", r.ServerId))

	return resourceIpLoadbalancingHttpFarmServerRead(d, meta)
}

func resourceIpLoadbalancingHttpFarmServerRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	service := d.Get("service_name").(string)
	farmid := d.Get("farm_id").(int)
	r := &IpLoadbalancingFarmServer{}

	endpoint := fmt.Sprintf("/ipLoadbalancing/%s/http/farm/%d/server/%s", service, farmid, d.Id())

	err := config.OVHClient.Get(endpoint, r)
	if err != nil {
		return fmt.Errorf("calling GET %s :\n\t %q", endpoint, err)
	}

	// set resource attributes
	for k, v := range r.ToMap() {
		d.Set(k, v)
	}

	return nil
}

func resourceIpLoadbalancingHttpFarmServerUpdate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	update := &IpLoadbalancingFarmServerUpdateOpts{
		DisplayName:          getNilStringPointerFromData(d, "display_name"),
		Address:              getNilStringPointerFromData(d, "address"),
		Port:                 getNilIntPointerFromData(d, "port"),
		ProxyProtocolVersion: getNilStringPointerFromData(d, "proxy_protocol_version"),
		Chain:                getNilStringPointerFromData(d, "chain"),
		Weight:               getNilIntPointerFromData(d, "weight"),
		Probe:                getNilBoolPointer(d.Get("probe")),
		Ssl:                  getNilBoolPointer(d.Get("ssl")),
		Backup:               getNilBoolPointer(d.Get("backup")),
		Status:               getNilStringPointerFromData(d, "status"),
	}

	service := d.Get("service_name").(string)
	farmid := d.Get("farm_id").(int)
	r := &IpLoadbalancingFarmServer{}
	endpoint := fmt.Sprintf("/ipLoadbalancing/%s/http/farm/%d/server/%s", service, farmid, d.Id())

	err := config.OVHClient.Put(endpoint, update, r)
	if err != nil {
		return fmt.Errorf("calling PUT %s with %d:\n\t %s", endpoint, farmid, err.Error())
	}
	return resourceIpLoadbalancingHttpFarmServerRead(d, meta)
}

func resourceIpLoadbalancingHttpFarmServerDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	service := d.Get("service_name").(string)
	farmid := d.Get("farm_id").(int)

	r := &IpLoadbalancingFarmServer{}
	endpoint := fmt.Sprintf("/ipLoadbalancing/%s/http/farm/%d/server/%s", service, farmid, d.Id())

	err := config.OVHClient.Delete(endpoint, r)
	if err != nil {
		return fmt.Errorf("calling DELETE %s :\n\t %s", endpoint, err.Error())
	}

	d.SetId("")
	return nil
}
