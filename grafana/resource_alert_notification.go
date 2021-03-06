package grafana

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"

	gapi "github.com/grafana/grafana-api-golang-client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var (
	ErrFrequencyMustBeSet = errors.New("frequency must be set when send_reminder is set to 'true'")
)

func ResourceAlertNotification() *schema.Resource {
	return &schema.Resource{
		Create: CreateAlertNotification,
		Update: UpdateAlertNotification,
		Delete: DeleteAlertNotification,
		Read:   ReadAlertNotification,

		Schema: map[string]*schema.Schema{
			"type": {
				Type:     schema.TypeString,
				Required: true,
			},

			"name": {
				Type:     schema.TypeString,
				Required: true,
			},

			"is_default": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			"send_reminder": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			"frequency": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},

			"settings": {
				Type:      schema.TypeMap,
				Optional:  true,
				Sensitive: true,
			},

			"uid": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"disable_resolve_message": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
		},
	}
}

func CreateAlertNotification(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gapi.Client)

	alertNotification, err := makeAlertNotification(d)
	if err != nil {
		return err
	}

	id, err := client.NewAlertNotification(alertNotification)
	if err != nil {
		return err
	}

	d.SetId(strconv.FormatInt(id, 10))

	return ReadAlertNotification(d, meta)
}

func UpdateAlertNotification(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gapi.Client)

	alertNotification, err := makeAlertNotification(d)
	if err != nil {
		return err
	}

	return client.UpdateAlertNotification(alertNotification)
}

func ReadAlertNotification(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gapi.Client)

	idStr := d.Id()
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return fmt.Errorf("Invalid id: %#v", idStr)
	}

	alertNotification, err := client.AlertNotification(id)
	if err != nil {
		if err.Error() == "404 Not Found" {
			log.Printf("[WARN] removing datasource %s from state because it no longer exists in grafana", d.Get("name").(string))
			d.SetId("")
			return nil
		}
		return err
	}

	settings := map[string]interface{}{}
	for k, v := range alertNotification.Settings.(map[string]interface{}) {
		boolVal, ok := v.(bool)
		if ok && boolVal {
			settings[k] = "true"
		} else if ok && !boolVal {
			settings[k] = "false"
		} else {
			settings[k] = v
		}
	}

	d.SetId(strconv.FormatInt(alertNotification.ID, 10))
	d.Set("is_default", alertNotification.IsDefault)
	d.Set("name", alertNotification.Name)
	d.Set("type", alertNotification.Type)
	d.Set("settings", settings)
	d.Set("uid", alertNotification.UID)
	d.Set("disable_resolve_message", alertNotification.DisableResolveMessage)
	d.Set("send_reminder", alertNotification.SendReminder)
	d.Set("frequency", alertNotification.Frequency)

	return nil
}

func DeleteAlertNotification(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gapi.Client)

	idStr := d.Id()
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return fmt.Errorf("Invalid id: %#v", idStr)
	}

	return client.DeleteAlertNotification(id)
}

func makeAlertNotification(d *schema.ResourceData) (*gapi.AlertNotification, error) {
	idStr := d.Id()
	var id int64
	var err error
	if idStr != "" {
		id, err = strconv.ParseInt(idStr, 10, 64)
	}

	settings := map[string]interface{}{}
	for k, v := range d.Get("settings").(map[string]interface{}) {
		strVal, ok := v.(string)
		if ok && strVal == "true" {
			settings[k] = true
		} else if ok && strVal == "false" {
			settings[k] = false
		} else {
			settings[k] = v
		}
	}

	sendReminder := d.Get("send_reminder").(bool)
	frequency := d.Get("frequency").(string)

	if sendReminder {
		if frequency == "" {
			return nil, ErrFrequencyMustBeSet
		}

		if _, err := time.ParseDuration(frequency); err != nil {
			return nil, err
		}
	}

	return &gapi.AlertNotification{
		ID:                    id,
		Name:                  d.Get("name").(string),
		Type:                  d.Get("type").(string),
		IsDefault:             d.Get("is_default").(bool),
		DisableResolveMessage: d.Get("disable_resolve_message").(bool),
		UID:                   d.Get("uid").(string),
		SendReminder:          sendReminder,
		Frequency:             frequency,
		Settings:              settings,
	}, err
}
