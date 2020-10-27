package bitbank

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jjjjpppp/bitbank-go-client/v1/models"
	"github.com/jjjjpppp/bitbank-go-client/v1/request"
)

func (c *Client) GetWithdrawalAccounts(ctx context.Context, params request.GetWithdrawalAccountsParams) (*models.WithdrawalAccounts, error) {
	spath := fmt.Sprintf("/user/withdrawal_account")
	queryParam := make(map[string]string)

	// set required param
	queryParam["asset"] = params.Asset

	res, err := c.sendRequest(ctx, "GET", spath, nil, &queryParam)
	if err != nil {
		return nil, err
	}

	var withdrawalAcdounts models.WithdrawalAccounts
	if err := decodeBody(res, &withdrawalAcdounts); err != nil {
		return nil, err
	}

	return &withdrawalAcdounts, nil
}

func (c *Client) RequestWithdrawal(ctx context.Context, params request.RequestWithdrawalParams) (*models.RequestWithdrawal, error) {
	spath := fmt.Sprintf("/user/request_withdrawal")
	bodyTemplate, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	err = json.Indent(&buf, []byte(bodyTemplate), "", "  ")
	if err != nil {
		return nil, err
	}
	body := buf.String()
	res, err := c.sendRequest(ctx, "POST", spath, strings.NewReader(body), nil)
	if err != nil {
		return nil, err
	}

	var requestWithdrawal models.RequestWithdrawal
	if err := decodeBody(res, &requestWithdrawal); err != nil {
		return nil, err
	}

	return &requestWithdrawal, nil
}
