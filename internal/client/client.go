package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

// Client for Orchard API.
type Client struct {
	BaseURL    *url.URL
	HTTPClient *http.Client
	Token      string
}

func NewClient(baseURL, token string) *Client {
	parsed, _ := url.Parse(strings.TrimSuffix(baseURL, "/"))
	return &Client{
		BaseURL:    parsed,
		HTTPClient: &http.Client{Timeout: 30 * time.Second},
		Token:      token,
	}
}

func (c *Client) doRequest(ctx context.Context, method, path string, body interface{}, result interface{}) error {
	relativeURL, err := url.Parse(path)
	if err != nil {
		return err
	}
	u := c.BaseURL.ResolveReference(relativeURL)
	var reqBody io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reqBody = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, method, u.String(), reqBody)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(bodyBytes))
	}
	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return err
		}
	}
	return nil
}

// ---- Types from OpenAPI ----

type Worker struct {
	Name      string         `json:"name"`
	Resources map[string]int `json:"resources"`
}

type VMMeta struct {
	Name       string `json:"name"`
	Generation int    `json:"generation,omitempty"`
}

type VMSpec struct {
	OS                   string            `json:"os,omitempty"`
	Arch                 string            `json:"arch,omitempty"`
	Runtime              string            `json:"runtime,omitempty"`
	Image                string            `json:"image,omitempty"`
	ImagePullPolicy      string            `json:"imagePullPolicy,omitempty"`
	CPU                  float64           `json:"cpu,omitempty"`
	Memory               float64           `json:"memory,omitempty"`
	DiskSize             float64           `json:"diskSize,omitempty"`
	NetSoftnet           *bool             `json:"netSoftnet,omitempty"`
	NetSoftnetAllow      []string          `json:"netSoftnetAllow,omitempty"`
	NetSoftnetBlock      []string          `json:"netSoftnetBlock,omitempty"`
	Suspendable          bool              `json:"suspendable,omitempty"`
	NetBridged           string            `json:"net-bridged,omitempty"`
	Headless             bool              `json:"headless,omitempty"`
	Nested               bool              `json:"nested,omitempty"`
	Username             string            `json:"username,omitempty"`
	Password             string            `json:"password,omitempty"`
	StartupScript        *StartupScript    `json:"startup_script,omitempty"`
	RestartPolicy        string            `json:"restart_policy,omitempty"`
	RequestedResources   map[string]int    `json:"resources,omitempty"`
	Labels               map[string]string `json:"labels,omitempty"`
	HostDirs             []HostDir         `json:"hostDirs,omitempty"`
	PowerState           string            `json:"powerState,omitempty"`
	LocalName            string            `json:"localName,omitempty"`
	NetSoftnetDeprecated *bool             `json:"net-softnet,omitempty"`
}

// Отдельная структура для ответа от API
type VMResources struct {
	Resources map[string]int `json:"resources,omitempty"`
}

type StartupScript struct {
	ScriptContent string            `json:"script_content,omitempty"`
	Env           map[string]string `json:"env,omitempty"`
}

type HostDir struct {
	Name string `json:"name,omitempty"`
	Path string `json:"path"`
	Ro   bool   `json:"ro,omitempty"`
}

type VMState struct {
	Status             string         `json:"status,omitempty"`
	StatusMessage      string         `json:"status_message,omitempty"`
	Worker             string         `json:"worker,omitempty"`
	ObservedGeneration int            `json:"observedGeneration,omitempty"`
	UsedResources      map[string]int `json:"-"`
}

type VMCreateRequest struct {
	VMMeta
	VMSpec
}

type VM struct {
	VMMeta
	VMSpec
	VMState
}

type ServiceAccount struct {
	Name  string   `json:"name"`
	Token string   `json:"token,omitempty"`
	Roles []string `json:"roles,omitempty"`
}

type ControllerInfo struct {
	Version      string   `json:"version"`
	Commit       string   `json:"commit"`
	Capabilities []string `json:"capabilities"`
}

type ClusterSettings struct {
	HostDirPolicies  []HostDirPolicy `json:"hostDirPolicies,omitempty"`
	SchedulerProfile string          `json:"schedulerProfile,omitempty"`
}

type HostDirPolicy struct {
	PathPrefix string `json:"pathPrefix"`
	Ro         bool   `json:"ro"`
}

// ---- API methods with /v1/ prefix ----

// Controller Info
func (c *Client) GetControllerInfo(ctx context.Context) (*ControllerInfo, error) {
	var info ControllerInfo
	err := c.doRequest(ctx, http.MethodGet, "/v1/controller/info", nil, &info)
	return &info, err
}

// Cluster Settings
func (c *Client) GetClusterSettings(ctx context.Context) (*ClusterSettings, error) {
	var settings ClusterSettings
	err := c.doRequest(ctx, http.MethodGet, "/v1/cluster-settings", nil, &settings)
	return &settings, err
}

func (c *Client) UpdateClusterSettings(ctx context.Context, settings *ClusterSettings) (*ClusterSettings, error) {
	var updated ClusterSettings
	err := c.doRequest(ctx, http.MethodPut, "/v1/cluster-settings", settings, &updated)
	return &updated, err
}

// Service Accounts
func (c *Client) CreateServiceAccount(ctx context.Context, sa *ServiceAccount) (*ServiceAccount, error) {
	var created ServiceAccount
	err := c.doRequest(ctx, http.MethodPost, "/v1/service-accounts", sa, &created)
	return &created, err
}

func (c *Client) GetServiceAccount(ctx context.Context, name string) (*ServiceAccount, error) {
	var sa ServiceAccount
	err := c.doRequest(ctx, http.MethodGet, "/v1/service-accounts/"+name, nil, &sa)
	return &sa, err
}

func (c *Client) ListServiceAccounts(ctx context.Context) ([]ServiceAccount, error) {
	var list []ServiceAccount
	err := c.doRequest(ctx, http.MethodGet, "/v1/service-accounts", nil, &list)
	return list, err
}

func (c *Client) UpdateServiceAccount(ctx context.Context, name string, sa *ServiceAccount) (*ServiceAccount, error) {
	var updated ServiceAccount
	err := c.doRequest(ctx, http.MethodPut, "/v1/service-accounts/"+name, sa, &updated)
	return &updated, err
}

func (c *Client) DeleteServiceAccount(ctx context.Context, name string) error {
	return c.doRequest(ctx, http.MethodDelete, "/v1/service-accounts/"+name, nil, nil)
}

// Workers
func (c *Client) ListWorkers(ctx context.Context) ([]Worker, error) {
	var list []Worker
	err := c.doRequest(ctx, http.MethodGet, "/v1/workers", nil, &list)
	return list, err
}

func (c *Client) GetWorker(ctx context.Context, name string) (*Worker, error) {
	var w Worker
	err := c.doRequest(ctx, http.MethodGet, "/v1/workers/"+name, nil, &w)
	return &w, err
}

// VMs

func (c *Client) CreateVM(ctx context.Context, vm *VM) (*VM, error) {
	// Логируем запрос в stderr
	b, _ := json.MarshalIndent(vm, "", "  ")
	os.Stderr.WriteString("\n========== CREATE VM REQUEST ==========\n")
	os.Stderr.Write(b)
	os.Stderr.WriteString("\n========================================\n\n")

	reqBody := VMCreateRequest{
		VMMeta: vm.VMMeta,
		VMSpec: vm.VMSpec,
	}

	var created VM
	err := c.doRequest(
		ctx,
		http.MethodPost,
		"/v1/vms",
		&reqBody,
		&created,
	)

	return &created, err
}

func (c *Client) GetVM(ctx context.Context, name string) (*VM, error) {
	var vm VM
	err := c.doRequest(ctx, http.MethodGet, "/v1/vms/"+name, nil, &vm)
	return &vm, err
}

type VMIP struct {
	IP string `json:"ip"`
}

func (c *Client) GetVMIP(ctx context.Context, name string) (string, error) {
	var result VMIP
	query := url.Values{"wait": []string{"2"}}
	err := c.doRequest(ctx, http.MethodGet, "/v1/vms/"+name+"/ip?"+query.Encode(), nil, &result)
	return result.IP, err
}

func (c *Client) ListVMs(ctx context.Context, filter string) ([]VM, error) {
	path := "/v1/vms"
	if filter != "" {
		path += "?filter=" + url.QueryEscape(filter)
	}
	var list []VM
	err := c.doRequest(ctx, http.MethodGet, path, nil, &list)
	return list, err
}

func (c *Client) UpdateVM(ctx context.Context, name string, spec *VMSpec) (*VM, error) {
	var updated VM
	err := c.doRequest(ctx, http.MethodPut, "/v1/vms/"+name, spec, &updated)
	return &updated, err
}

func (c *Client) DeleteVM(ctx context.Context, name string) error {
	return c.doRequest(ctx, http.MethodDelete, "/v1/vms/"+name, nil, nil)
}
