package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"

	"golang.org/x/oauth2/google"
	sqladmin "google.golang.org/api/sqladmin/v1beta4"
)

// PubSubMessage is the payload of a Pub/Sub event.
// See the documentation for more details:
// https://cloud.google.com/pubsub/docs/reference/rest/v1/PubsubMessage
type PubSubMessage struct {
	Data []byte `json:"data"`
}

type MessagePayload struct {
	Instance string
	Project  string
	Action   string
}

// ProcessPubSub consumes and processes a Pub/Sub message.
func ProcessPubSub(ctx context.Context, m PubSubMessage) error {
	var psData MessagePayload
	if err := json.Unmarshal(m.Data, &psData); err != nil {
		return err
	}
	log.Printf("Request received for Cloud SQL instance %s action: %s, %s", psData.Instance, psData.Action, psData.Project)

	// Create an http.Client that uses Application Default Credentials.
	client, err := google.DefaultClient(ctx, sqladmin.CloudPlatformScope)
	if err != nil {
		return err
	}

	// Create the Google Cloud SQL service.
	service, err := sqladmin.New(client)
	if err != nil {
		return err
	}

	// Get the requested start or stop Action.
	action := "UNDEFINED"
	switch psData.Action {
	case "start":
		action = "ALWAYS"
	case "stop":
		action = "NEVER"
	default:
		return errors.New("No valid action provided")
	}

	// See more examples at:
	// https://cloud.google.com/sql/docs/sqlserver/admin-api/rest/v1beta4/instances/patch
	rb := &sqladmin.DatabaseInstance{
		Settings: &sqladmin.Settings{
			ActivationPolicy: action,
		},
	}

	resp, err := service.Instances.Patch(psData.Project, psData.Instance, rb).Context(ctx).Do()
	if err != nil {
		return err
	}
	log.Printf("%#v\n", resp)
	return nil
}
