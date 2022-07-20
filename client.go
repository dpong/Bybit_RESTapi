package bybitapi

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

type Client struct {
	key, secret        string
	subaccount         string
	client             *http.Client
	spotPrivateChannel *spotPrivateChannelBranch
}

func New(key, secret, subaccount string) *Client {
	hc := &http.Client{
		Timeout: 10 * time.Second,
	}
	return &Client{
		key:        key,
		secret:     secret,
		subaccount: subaccount,
		client:     hc,
	}
}

func (p *Client) getSigned(param string) string {
	sig := hmac.New(sha256.New, []byte(p.secret))
	sig.Write([]byte(param))
	signature := hex.EncodeToString(sig.Sum(nil))
	return signature
}

func (p *Client) newRequest(product, method, spath string, body []byte, params *map[string]string, auth bool) (*http.Request, error) {
	q := url.Values{}
	if params != nil {
		for k, v := range *params {
			q.Add(k, v)
		}
	}
	var timestamp int64
	if auth {
		timestamp = time.Now().UnixNano() / 1e6
		q.Add("api_key", p.key)
		q.Add("timestamp", strconv.Itoa(int(timestamp)))
	}
	host := HostHub(product)

	url, err := p.sign(host, method, spath, &q, auth)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(method, url, strings.NewReader(string(body)))
	if err != nil {
		return nil, err
	}
	switch method {
	case "POST":
		switch product {
		case ProductPerp:
			req.Header.Set("Content-Type", "application/json")
		case ProductSpot:
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		}
	default:
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	return req, nil
}

func (c *Client) sendRequest(product, method, spath string, body []byte, params *map[string]string, auth bool) (*http.Response, error) {
	req, err := c.newRequest(product, method, spath, body, params, auth)
	if err != nil {
		return nil, err
	}
	res, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != 200 {
		//c.Logger.Printf("status: %s", res.Status)
		buf := new(bytes.Buffer)
		buf.ReadFrom(res.Body)
		s := buf.String()
		return nil, fmt.Errorf("faild to get data. status: %s, with error: %s", res.Status, s)
	}
	return res, nil
}

func HostHub(product string) (host string) {
	host = "api.bybit.com"
	return host
}

func decode(res *http.Response, out interface{}) error {
	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)
	err := json.Unmarshal([]byte(body), &out)
	if err == nil {
		return nil
	}
	return err
}

func (c *Client) sign(host, method, spath string, q *url.Values, auth bool) (string, error) {
	var buffer bytes.Buffer
	buffer.WriteString("https://")
	buffer.WriteString(host)
	buffer.WriteString(spath)
	if (*q).Encode() == "" {
		return buffer.String(), nil
	}
	buffer.WriteString("?")
	par := q.Encode()
	//par := q.QueryEscape(",")
	buffer.WriteString(par)
	if !auth {
		return buffer.String(), nil
	}
	signature := c.getSigned(par)
	buffer.WriteString("&sign=")
	buffer.WriteString(signature)
	return buffer.String(), nil
}

type ApiInfoResponse struct {
	RetCode int    `json:"ret_code"`
	RetMsg  string `json:"ret_msg"`
	ExtCode string `json:"ext_code"`
	Result  []struct {
		APIKey      string    `json:"api_key"`
		Type        string    `json:"type"`
		UserID      int       `json:"user_id"`
		InviterID   int       `json:"inviter_id"`
		Ips         []string  `json:"ips"`
		Note        string    `json:"note"`
		Permissions []string  `json:"permissions"`
		CreatedAt   time.Time `json:"created_at"`
		ExpiredAt   time.Time `json:"expired_at"`
		ReadOnly    bool      `json:"read_only"`
	} `json:"result"`
	ExtInfo          interface{} `json:"ext_info"`
	TimeNow          string      `json:"time_now"`
	RateLimitStatus  int         `json:"rate_limit_status"`
	RateLimitResetMs int64       `json:"rate_limit_reset_ms"`
	RateLimit        int         `json:"rate_limit"`
}

func (p *Client) ApiInfo() (result *ApiInfoResponse, err error) {
	res, err := p.sendRequest(ProductPerp, http.MethodGet, "/v2/private/account/api-key", nil, nil, true)
	if err != nil {
		return nil, err
	}
	// in Close()
	err = decode(res, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}
