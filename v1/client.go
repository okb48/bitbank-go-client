package bitbank

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/jjjjpppp/bitbank-go-client/v1/models"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"path"
	"runtime"
	"time"
)

const (
	baseUrl string = "https://public.bitbank.cc"
	version string = "0.0.1"
)

type Client struct {
	URL        *url.URL
	ApiTokenID string
	ApiSecret  string
	HTTPClient *http.Client
	Logger     *log.Logger
	testServer *httptest.Server
}

func NewClient(apiTokenID string, apiSecret string, logger *log.Logger) (*Client, error) {
	if len(apiTokenID) == 0 {
		return nil, fmt.Errorf("apiTokenID is not set")
	}

	if len(apiSecret) == 0 {
		return nil, fmt.Errorf("apiSecret is not set")
	}

	url, err := url.ParseRequestURI(baseUrl)
	if err != nil {
		return nil, err
	}

	var discardLogger = log.New(ioutil.Discard, "", log.LstdFlags)
	if logger == nil {
		logger = discardLogger
	}

	client := &http.Client{Timeout: time.Duration(10) * time.Second}
	return &Client{URL: url, ApiTokenID: apiTokenID, ApiSecret: apiSecret, HTTPClient: client, Logger: logger}, nil

}

func (c *Client) GetTicker(ctx context.Context, pair string) (*models.Ticker, error) {
	spath := fmt.Sprintf("/%s/ticker", pair)
	res, err := c.sendRequest(ctx, "GET", spath, nil, nil)
	if err != nil {
		return nil, err
	}

	var ticker models.Ticker
	if err := decodeBody(res, &ticker); err != nil {
		return nil, err
	}

	return &ticker, nil
}

func (c *Client) GetDepth(ctx context.Context, pair string) (*models.Depth, error) {
	spath := fmt.Sprintf("/%s/depth", pair)
	res, err := c.sendRequest(ctx, "GET", spath, nil, nil)
	if err != nil {
		return nil, err
	}

	var depth models.Depth
	if err := decodeBody(res, &depth); err != nil {
		return nil, err
	}

	return &depth, nil
}

func (c *Client) GetTransactions(ctx context.Context, pair string) (*models.Transactions, error) {
	spath := fmt.Sprintf("/%s/transactions", pair)
	res, err := c.sendRequest(ctx, "GET", spath, nil, nil)
	if err != nil {
		return nil, err
	}

	var transaction models.Transactions
	if err := decodeBody(res, &transaction); err != nil {
		return nil, err
	}

	return &transaction, nil
}

func (c *Client) GetTransactionsByYMD(ctx context.Context, pair, ymdString string) (*models.Transactions, error) {
	spath := fmt.Sprintf("/%s/transactions/%s", pair, ymdString)
	res, err := c.sendRequest(ctx, "GET", spath, nil, nil)
	if err != nil {
		return nil, err
	}

	var transaction models.Transactions
	if err := decodeBody(res, &transaction); err != nil {
		return nil, err
	}

	return &transaction, nil
}

func (c *Client) GetCandlesticks(ctx context.Context, pair, candleType, ymdString string) (*models.Candlesticks, error) {
	spath := fmt.Sprintf("/%s/candlestick/%s/%s", pair, candleType, ymdString)
	res, err := c.sendRequest(ctx, "GET", spath, nil, nil)
	if err != nil {
		return nil, err
	}

	var candlesticks models.Candlesticks
	if err := decodeBody(res, &candlesticks); err != nil {
		return nil, err
	}

	return &candlesticks, nil
}

func (c *Client) newRequest(ctx context.Context, method, spath string, body io.Reader, queryParam *map[string]string) (*http.Request, error) {

	// swith client url for unit test
	if c.testServer != nil {
		url, _ := url.ParseRequestURI(c.testServer.URL)
		*c.URL = *url
	}

	u := *c.URL
	u.Path = path.Join(c.URL.Path, spath)

	// build QueryParameter
	if queryParam != nil {
		q := u.Query()
		for k, v := range *queryParam {
			q.Set(k, v)
		}
		u.RawQuery = q.Encode()
	}

	userAgent := fmt.Sprintf("GoClient/%s (%s)", version, runtime.Version())
	accessNonce := fmt.Sprintf("%d", time.Now().Unix())
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"path":  spath + u.RawQuery,
		"nonce": accessNonce,
	})

	tokenString, err := token.SignedString([]byte(c.ApiSecret))
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(method, u.String(), body)
	if err != nil {
		return nil, err
	}

	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("ACCESS-KEY", c.ApiTokenID)
	req.Header.Set("ACCESS-NONCE", accessNonce)
	req.Header.Set("ACCESS-SIGNATURE", tokenString)

	return req, nil
}

func (c *Client) sendRequest(ctx context.Context, method, spath string, body io.Reader, queryParam *map[string]string) (*http.Response, error) {
	req, err := c.newRequest(ctx, method, spath, body, queryParam)
	c.Logger.Printf("Request:  %s \n", httpRequestLog(req))
	if err != nil {
		c.Logger.Printf("err: %#v \n", err)
		return nil, err
	}

	res, err := c.HTTPClient.Do(req)
	c.Logger.Printf("Response: %s \n", httpResponseLog(res))
	if err != nil {
		c.Logger.Printf("err: %#v \n", err)
		return nil, err
	}

	if res.StatusCode != 200 {
		c.Logger.Printf("err: %#v \n", err)
		return nil, fmt.Errorf("faild to get data. status: %s", res.Status)
	}
	return res, nil
}

func decodeBody(resp *http.Response, out interface{}) error {
	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)
	return decoder.Decode(out)
}

func httpResponseLog(resp *http.Response) string {
	b, _ := httputil.DumpResponse(resp, true)
	return string(b)
}
func httpRequestLog(req *http.Request) string {
	b, _ := httputil.DumpRequest(req, true)
	return string(b)
}