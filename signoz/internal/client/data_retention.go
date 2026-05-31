package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/SigNoz/terraform-provider-signoz/signoz/internal/model"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

const (
	// dataRetentionPath - URL path for data retention settings API.
	dataRetentionPath = "api/v1/settings/ttl"
)

// dataRetentionResponse maps the GET response envelope.
type dataRetentionResponse struct {
	Status string               `json:"status"`
	Data   model.DataRetention  `json:"data"`
}

// GetDataRetention returns the current data retention settings.
func (c *Client) GetDataRetention(ctx context.Context) (*model.DataRetention, error) {
	reqURL, err := url.JoinPath(c.hostURL.String(), dataRetentionPath)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(ctx, req)
	if err != nil {
		return nil, err
	}

	var bodyObj dataRetentionResponse
	err = json.Unmarshal(body, &bodyObj)
	if err != nil {
		return nil, err
	}

	if bodyObj.Status != "success" {
		tflog.Error(ctx, "GetDataRetention: error while fetching data retention settings", map[string]any{
			"status": bodyObj.Status,
		})
		return nil, fmt.Errorf("error while fetching data retention settings: status=%s", bodyObj.Status)
	}

	tflog.Debug(ctx, "GetDataRetention: data retention settings fetched", map[string]any{"data": bodyObj.Data})

	return &bodyObj.Data, nil
}

// SetDataRetention sets data retention for a single signal type.
func (c *Client) SetDataRetention(ctx context.Context, payload *model.DataRetentionUpdate) error {
	rb, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	reqURL, err := url.JoinPath(c.hostURL.String(), dataRetentionPath)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, reqURL, strings.NewReader(string(rb)))
	if err != nil {
		return err
	}

	body, err := c.doRequest(ctx, req)
	if err != nil {
		return err
	}

	var bodyObj signozResponse
	err = json.Unmarshal(body, &bodyObj)
	if err != nil {
		return err
	}

	if bodyObj.Status != "success" {
		tflog.Error(ctx, "SetDataRetention: error while setting data retention", map[string]any{
			"error": bodyObj.Error,
			"type":  payload.Type,
		})
		return fmt.Errorf("error while setting data retention for %s: %s", payload.Type, bodyObj.Error)
	}

	tflog.Debug(ctx, "SetDataRetention: data retention set", map[string]any{"type": payload.Type})

	return nil
}
