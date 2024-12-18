package api

import (
	"context"
)

type HealthcheckOutput struct {
	Body HealthcheckMessage
}

type HealthcheckMessage struct {
	Status string `json:"status"`
}

// healthcheckHandler handles healthcheck requests to determine if the
// application is running and available.
func (app *Application) healthcheckHandler(ctx context.Context, input *struct{}) (*HealthcheckOutput, error) {
	resp := &HealthcheckOutput{
		Body: HealthcheckMessage{
			Status: "available",
		},
	}

	return resp, nil
}
