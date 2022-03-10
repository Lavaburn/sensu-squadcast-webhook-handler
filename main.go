package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/asaskevich/govalidator"
	"github.com/sensu-community/sensu-plugin-sdk/sensu"
	"github.com/sensu-community/sensu-plugin-sdk/templates"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
)

// Config represents the handler plugin config.
type Config struct {
	sensu.PluginConfig
	APIURL      string
	Message 	string
	Description string
	Template    string
}

// SQEvent is the JSON type for creating a Squadcast incident
type SQEvent struct {	
	Check          	*corev2.Check  `json:"check,omitempty"`
	Entity			*corev2.Entity `json:"entity,omitempty"`
	
	Status    		string         `json:"status"`	
	EventId			string		   `json:"event_id"`
	
	Message   		string         `json:"message,omitempty"`
	Description     string         `json:"description,omitempty"`
}

const (
	apiurl  	= "api-url"
	message 	= "message"
	description	= "description"
	template 	= "template"
)

var (
	plugin = Config{
		PluginConfig: sensu.PluginConfig{
			Name:     "sensu-squadcast-webhook-handler",
			Short:    "sends  sensu  events to squadcast via Incident Webhook",
			Keyspace: "sensu.io/plugins/sensu-squadcast-webhook-handler/config",
		},
	}

	options = []*sensu.PluginConfigOption{
		{
			Path:      apiurl,
			Env:       "SENSU_SQUADCAST_APIURL",
			Argument:  apiurl,
			Shorthand: "a",
			Default:   "",
			Secret:    true,
			Usage:     "The URL for the Squadcast API",
			Value:     &plugin.APIURL,
		},
		{
			Path:      message,
			Env:       "SENSU_SQUADCAST_MESSAGE",
			Argument:  message,
			Shorthand: "m",
			Default:   "{{.Entity.Name}}/{{.Check.Name}}",
			Usage:     "The template to use for the incident name",
			Value:     &plugin.Message,
		},
		{
			Path:      description,
			Env:       "SENSU_SQUADCAST_DESCRIPTION",
			Argument:  description,
			Shorthand: "d",
			Default:   "{{.Check.Output}}",
			Usage:     "The template to use for the incident description",
			Value:     &plugin.Description,
		},
		{
			Path:      template,
			Env:       "SENSU_SQUADCAST_TEMPLATE",
			Argument:  template,
			Shorthand: "t",
			Default:   "",
			Usage:     "The template file to use for the incident description",
			Value:     &plugin.Template,
		},
	}
)

// CheckArgs validates the configuration passed to the handler
func CheckArgs(_ *corev2.Event) error {
	if len(plugin.APIURL) == 0 {
		return errors.New("Missing Squadcast API URL")
	}
	if !govalidator.IsURL(plugin.APIURL) {
		return errors.New("Invalid Squadcast API URL")
	}
	return nil
}

// SendEventToSquadcast sends the event data to configured Squadcast Webhook endpoint
func SendEventToSquadcast(event *corev2.Event) error {
	var status string

	switch eventStatus := event.Check.Status; eventStatus {
	case 0:
		status = "resolve"
	default:
		status = "trigger"
	}

	msgMessage, err := templates.EvalTemplate("Message", plugin.Message, event)
	if err != nil {
		return fmt.Errorf("Failed to evaluate template %s: %v", plugin.Message, err)
	}
	
	descriptionTemplate := ""
	if len(plugin.Template) > 0 {
		templateBytes, fileErr := ioutil.ReadFile(plugin.Template)
		if fileErr != nil {
			return fmt.Errorf("Failed to read specified template file %s", plugin.Template)
		}
		descriptionTemplate = string(templateBytes)
	} else {
		descriptionTemplate = plugin.Description
	}
	
	msgDescription, err := templates.EvalTemplate("description", descriptionTemplate, event)
	if err != nil {
		return fmt.Errorf("Failed to evaluate template %s: %v", descriptionTemplate, err)
	}
	
	sqEvent := SQEvent{
		Check:          event.Check,
		Entity:         event.Entity,
		Status:    		status,
		EventId:        msgMessage,
		Message:        msgMessage,
		Description:	msgDescription,
	}
	
	msgBytes, err := json.Marshal(sqEvent)
	if err != nil {
		return fmt.Errorf("Failed to marshal Squadcast event: %s", err)
	}

	resp, err := http.Post(plugin.APIURL, "application/json", bytes.NewBuffer(msgBytes))
	if err != nil {
		return fmt.Errorf("Post to %s failed: %s", plugin.APIURL, err)
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("POST to %s failed with %v", plugin.APIURL, resp.Status)
	}
	return nil
}

func main() {
	handler := sensu.NewGoHandler(&plugin.PluginConfig, options, CheckArgs, SendEventToSquadcast)
	handler.Execute()
}
