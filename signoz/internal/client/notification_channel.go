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
	// channelPath - URL path for notification channel APIs.
	channelPath = "api/v1/channels"
)

// channelResponse - Maps the response data of GetChannel and CreateChannel.
type channelResponse struct {
	Status    string                       `json:"status"`
	Error     string                       `json:"error"`
	ErrorType string                       `json:"errorType"`
	Data      model.NotificationChannel    `json:"data"`
}

// GetChannel - Returns specific notification channel.
func (c *Client) GetChannel(ctx context.Context, channelID string) (*model.NotificationChannel, error) {
	url, err := url.JoinPath(c.hostURL.String(), channelPath, channelID)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(ctx, req)
	if err != nil {
		return nil, err
	}

	var bodyObj channelResponse
	err = json.Unmarshal(body, &bodyObj)
	if err != nil {
		return nil, err
	}

	if bodyObj.Status != "success" || bodyObj.Error != "" {
		tflog.Error(ctx, "GetChannel: error while fetching channel", map[string]any{
			"error": bodyObj.Error,
			"type":  bodyObj.ErrorType,
			"data":  bodyObj.Data,
		})

		return &model.NotificationChannel{}, fmt.Errorf("error while fetching channel: %s", bodyObj.Error)
	}

	tflog.Debug(ctx, "GetChannel: channel fetched", map[string]any{"channel": bodyObj.Data})

	return &bodyObj.Data, nil
}

// CreateChannel - Creates a new notification channel.
func (c *Client) CreateChannel(ctx context.Context, channelPayload *model.NotificationChannel) (*model.NotificationChannel, error) {
	rb, err := json.Marshal(channelPayload)
	if err != nil {
		return nil, err
	}

	url, err := url.JoinPath(c.hostURL.String(), channelPath)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(string(rb)))
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(ctx, req)
	if err != nil {
		return nil, err
	}

	var bodyObj channelResponse
	err = json.Unmarshal(body, &bodyObj)
	if err != nil {
		return nil, err
	}

	if bodyObj.Status != "success" || bodyObj.Error != "" {
		tflog.Error(ctx, "CreateChannel: error while creating channel", map[string]any{
			"error":     bodyObj.Error,
			"errorType": bodyObj.ErrorType,
			"data":      bodyObj.Data,
		})
		return nil, fmt.Errorf("error while creating channel: %s", bodyObj.Error)
	}

	tflog.Debug(ctx, "CreateChannel: channel created", map[string]any{"channel": bodyObj.Data})

	return &bodyObj.Data, nil
}

// UpdateChannel - Updates an existing notification channel.
func (c *Client) UpdateChannel(ctx context.Context, channelID string, channelPayload *model.NotificationChannel) error {
	rb, err := json.Marshal(channelPayload)
	if err != nil {
		return err
	}

	url, err := url.JoinPath(c.hostURL.String(), channelPath, channelID)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPut, url, strings.NewReader(string(rb)))
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

	if bodyObj.Status != "success" || bodyObj.Error != "" {
		tflog.Error(ctx, "UpdateChannel: error while updating channel", map[string]any{
			"error":     bodyObj.Error,
			"errorType": bodyObj.ErrorType,
			"data":      bodyObj.Data,
		})
		return fmt.Errorf("error while updating channel: %s", bodyObj.Error)
	}

	tflog.Debug(ctx, "UpdateChannel: channel updated", map[string]any{"channelID": channelID})

	return nil
}

// DeleteChannel - Deletes an existing notification channel.
func (c *Client) DeleteChannel(ctx context.Context, channelID string) error {
	url, err := url.JoinPath(c.hostURL.String(), channelPath, channelID)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodDelete, url, nil)
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

	if bodyObj.Status != "success" || bodyObj.Error != "" {
		tflog.Error(ctx, "DeleteChannel: error while deleting channel", map[string]any{
			"error":     bodyObj.Error,
			"errorType": bodyObj.ErrorType,
			"data":      bodyObj.Data,
		})
		return fmt.Errorf("error while deleting channel: %s", bodyObj.Error)
	}

	tflog.Debug(ctx, "DeleteChannel: channel deleted", map[string]any{"channelID": channelID})

	return nil
}
