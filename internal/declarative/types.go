package declarative

// DeclarativeConfig представляет полный декларативный конфиг
type DeclarativeConfig struct {
	Version string          `yaml:"version" json:"version"`
	Routes  []Route         `yaml:"routes,omitempty" json:"routes,omitempty"`
	Services []Service      `yaml:"services,omitempty" json:"services,omitempty"`
	Consumers []Consumer    `yaml:"consumers,omitempty" json:"consumers,omitempty"`
	SSLs     []SSL          `yaml:"ssls,omitempty" json:"ssls,omitempty"`
	GlobalRules []GlobalRule `yaml:"global_rules,omitempty" json:"global_rules,omitempty"`
	PluginConfigs []PluginConfig `yaml:"plugin_configs,omitempty" json:"plugin_configs,omitempty"`
	StreamRoutes []StreamRoute `yaml:"stream_routes,omitempty" json:"stream_routes,omitempty"`
}

// Route представляет конфигурацию маршрута APISIX
type Route struct {
	ID          string                 `yaml:"id" json:"id"`
	Name        string                 `yaml:"name,omitempty" json:"name,omitempty"`
	Desc        string                 `yaml:"desc,omitempty" json:"desc,omitempty"`
	URI         string                 `yaml:"uri,omitempty" json:"uri,omitempty"`
	URIs        []string               `yaml:"uris,omitempty" json:"uris,omitempty"`
	Host        string                 `yaml:"host,omitempty" json:"host,omitempty"`
	Hosts       []string               `yaml:"hosts,omitempty" json:"hosts,omitempty"`
	RemoteAddr  string                 `yaml:"remote_addr,omitempty" json:"remote_addr,omitempty"`
	RemoteAddrs []string               `yaml:"remote_addrs,omitempty" json:"remote_addrs,omitempty"`
	Methods     []string               `yaml:"methods,omitempty" json:"methods,omitempty"`
	Vars        [][]interface{}        `yaml:"vars,omitempty" json:"vars,omitempty"`
	FilterFunc  string                 `yaml:"filter_func,omitempty" json:"filter_func,omitempty"`
	Plugins     map[string]interface{} `yaml:"plugins,omitempty" json:"plugins,omitempty"`
	Upstream    *Upstream              `yaml:"upstream,omitempty" json:"upstream,omitempty"`
	UpstreamID  string                 `yaml:"upstream_id,omitempty" json:"upstream_id,omitempty"`
	ServiceID   string                 `yaml:"service_id,omitempty" json:"service_id,omitempty"`
	PluginConfigID string              `yaml:"plugin_config_id,omitempty" json:"plugin_config_id,omitempty"`
	EnableWebsocket bool               `yaml:"enable_websocket,omitempty" json:"enable_websocket,omitempty"`
	Timeout     *Timeout               `yaml:"timeout,omitempty" json:"timeout,omitempty"`
}

// Service представляет сервис APISIX
type Service struct {
	ID          string                 `yaml:"id" json:"id"`
	Name        string                 `yaml:"name,omitempty" json:"name,omitempty"`
	Desc        string                 `yaml:"desc,omitempty" json:"desc,omitempty"`
	Upstream    *Upstream              `yaml:"upstream,omitempty" json:"upstream,omitempty"`
	UpstreamID  string                 `yaml:"upstream_id,omitempty" json:"upstream_id,omitempty"`
	Plugins     map[string]interface{} `yaml:"plugins,omitempty" json:"plugins,omitempty"`
	EnableWebsocket bool               `yaml:"enable_websocket,omitempty" json:"enable_websocket,omitempty"`
	Labels      map[string]string      `yaml:"labels,omitempty" json:"labels,omitempty"`
}

// Upstream представляет апстрим APISIX
type Upstream struct {
	ID          string          `yaml:"id" json:"id"`
	Name        string          `yaml:"name,omitempty" json:"name,omitempty"`
	Desc        string          `yaml:"desc,omitempty" json:"desc,omitempty"`
	Type        string          `yaml:"type,omitempty" json:"type,omitempty"`
	HashOn      string          `yaml:"hash_on,omitempty" json:"hash_on,omitempty"`
	Key         string          `yaml:"key,omitempty" json:"key,omitempty"`
	Nodes       map[string]int  `yaml:"nodes" json:"nodes"`
	Retries     *int            `yaml:"retries,omitempty" json:"retries,omitempty"`
	Timeout     *UpstreamTimeout `yaml:"timeout,omitempty" json:"timeout,omitempty"`
	Checks      *HealthCheck    `yaml:"checks,omitempty" json:"checks,omitempty"`
	PassHost    string          `yaml:"pass_host,omitempty" json:"pass_host,omitempty"`
	Scheme      string          `yaml:"scheme,omitempty" json:"scheme,omitempty"`
	Labels      map[string]string `yaml:"labels,omitempty" json:"labels,omitempty"`
}

type UpstreamTimeout struct {
	Connect float64 `yaml:"connect,omitempty" json:"connect,omitempty"`
	Send    float64 `yaml:"send,omitempty" json:"send,omitempty"`
	Read    float64 `yaml:"read,omitempty" json:"read,omitempty"`
}

type Timeout struct {
	Connect float64 `yaml:"connect,omitempty" json:"connect,omitempty"`
	Send    float64 `yaml:"send,omitempty" json:"send,omitempty"`
	Read    float64 `yaml:"read,omitempty" json:"read,omitempty"`
}

type HealthCheck struct {
	Active  *ActiveHealthCheck  `yaml:"active,omitempty" json:"active,omitempty"`
	Passive *PassiveHealthCheck `yaml:"passive,omitempty" json:"passive,omitempty"`
}

type ActiveHealthCheck struct {
	Type           string            `yaml:"type,omitempty" json:"type,omitempty"`
	Timeout        float64           `yaml:"timeout,omitempty" json:"timeout,omitempty"`
	Concurrency    int               `yaml:"concurrency,omitempty" json:"concurrency,omitempty"`
	HTTPPath       string            `yaml:"http_path,omitempty" json:"http_path,omitempty"`
	Host           string            `yaml:"host,omitempty" json:"host,omitempty"`
	Port           int               `yaml:"port,omitempty" json:"port,omitempty"`
	Healthy        *HealthyStatus    `yaml:"healthy,omitempty" json:"healthy,omitempty"`
	Unhealthy      *UnhealthyStatus  `yaml:"unhealthy,omitempty" json:"unhealthy,omitempty"`
	ReqHeaders     []string          `yaml:"req_headers,omitempty" json:"req_headers,omitempty"`
}

type PassiveHealthCheck struct {
	Type      string           `yaml:"type,omitempty" json:"type,omitempty"`
	Healthy   *HealthyStatus   `yaml:"healthy,omitempty" json:"healthy,omitempty"`
	Unhealthy *UnhealthyStatus `yaml:"unhealthy,omitempty" json:"unhealthy,omitempty"`
}

type HealthyStatus struct {
	Interval   int `yaml:"interval,omitempty" json:"interval,omitempty"`
	HTTPStatuses []int `yaml:"http_statuses,omitempty" json:"http_statuses,omitempty"`
	Successes  int `yaml:"successes,omitempty" json:"successes,omitempty"`
}

type UnhealthyStatus struct {
	Interval     int   `yaml:"interval,omitempty" json:"interval,omitempty"`
	HTTPStatuses []int `yaml:"http_statuses,omitempty" json:"http_statuses,omitempty"`
	HTTPFailures int   `yaml:"http_failures,omitempty" json:"http_failures,omitempty"`
	TCPFailures  int   `yaml:"tcp_failures,omitempty" json:"tcp_failures,omitempty"`
	Timeouts     int   `yaml:"timeouts,omitempty" json:"timeouts,omitempty"`
}

// Consumer представляет потребителя APISIX
type Consumer struct {
	Username string                 `yaml:"username" json:"username"`
	Desc     string                 `yaml:"desc,omitempty" json:"desc,omitempty"`
	Labels   map[string]string      `yaml:"labels,omitempty" json:"labels,omitempty"`
	Plugins  map[string]interface{} `yaml:"plugins,omitempty" json:"plugins,omitempty"`
}

// SSL представляет SSL сертификат
type SSL struct {
	ID      string            `yaml:"id" json:"id"`
	Cert    string            `yaml:"cert" json:"cert"`
	Key     string            `yaml:"key" json:"key"`
	Sni     []string          `yaml:"snis,omitempty" json:"snis,omitempty"`
	Labels  map[string]string `yaml:"labels,omitempty" json:"labels,omitempty"`
}

// GlobalRule представляет глобальное правило
type GlobalRule struct {
	ID      string                 `yaml:"id" json:"id"`
	Plugins map[string]interface{} `yaml:"plugins" json:"plugins"`
}

// PluginConfig представляет конфигурацию плагина
type PluginConfig struct {
	ID      string                 `yaml:"id" json:"id"`
	Desc    string                 `yaml:"desc,omitempty" json:"desc,omitempty"`
	Plugins map[string]interface{} `yaml:"plugins" json:"plugins"`
	Labels  map[string]string      `yaml:"labels,omitempty" json:"labels,omitempty"`
}

// StreamRoute представляет потоковый маршрут
type StreamRoute struct {
	ID         string                 `yaml:"id" json:"id"`
	Desc       string                 `yaml:"desc,omitempty" json:"desc,omitempty"`
	RemoteAddr string                 `yaml:"remote_addr,omitempty" json:"remote_addr,omitempty"`
	ServerAddr string                 `yaml:"server_addr,omitempty" json:"server_addr,omitempty"`
	ServerPort int                    `yaml:"server_port,omitempty" json:"server_port,omitempty"`
	Sni        string                 `yaml:"sni,omitempty" json:"sni,omitempty"`
	Upstream   *Upstream              `yaml:"upstream,omitempty" json:"upstream,omitempty"`
	UpstreamID string                 `yaml:"upstream_id,omitempty" json:"upstream_id,omitempty"`
	Plugins    map[string]interface{} `yaml:"plugins,omitempty" json:"plugins,omitempty"`
}

// // Upstreams возвращает все уникальные апстримы из конфигурации
// func (c *DeclarativeConfig) Upstreams() []Upstream {
// 	upstreams := make([]Upstream, 0)
// 	seen := make(map[string]bool)

// 	// Собираем из routes
// 	for _, route := range c.Routes {
// 		if route.Upstream != nil && !seen[route.Upstream.ID] {
// 			upstreams = append(upstreams, *route.Upstream)
// 			seen[route.Upstream.ID] = true
// 		}
// 	}

// 	// Собираем из services
// 	for _, service := range c.Services {
// 		if service.Upstream != nil && !seen[service.Upstream.ID] {
// 			upstreams = append(upstreams, *service.Upstream)
// 			seen[service.Upstream.ID] = true
// 		}
// 	}

// 	// Собираем из stream_routes
// 	for _, streamRoute := range c.StreamRoutes {
// 		if streamRoute.Upstream != nil && !seen[streamRoute.Upstream.ID] {
// 			upstreams = append(upstreams, *streamRoute.Upstream)
// 			seen[streamRoute.Upstream.ID] = true
// 		}
// 	}

// 	return upstreams
// }
