package main

import (
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/expfmt"
	"github.com/prometheus/common/promlog"
	"github.com/prometheus/common/version"
	"github.com/prometheus/exporter-toolkit/web"
	"github.com/spf13/cobra"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	namespace = "clash"
)

var (
	clashInfo          = prometheus.NewDesc(prometheus.BuildFQName(namespace, "version", "info"), "Clash version info.", []string{"premium", "version"}, nil)
	clashUp            = prometheus.NewDesc(prometheus.BuildFQName(namespace, "", "up"), "Was the last scrape of Clash successful.", nil, nil)
	proxyDelay         = prometheus.NewDesc(prometheus.BuildFQName(namespace, "proxy", "delay"), "Proxy delay.", []string{"type", "name", "provider"}, nil)
	downloadTotal      = prometheus.NewDesc(prometheus.BuildFQName(namespace, "connection", "download_total"), "Number of bytes that downloaded by clash.", nil, nil)
	uploadTotal        = prometheus.NewDesc(prometheus.BuildFQName(namespace, "connection", "upload_total"), "Number of bytes that uploaded by clash.", nil, nil)
	connectionDownload = prometheus.NewDesc(prometheus.BuildFQName(namespace, "connection", "download"), "Number of bytes for specific connection that downloaded by clash.", nil, nil)
	connectionUpload   = prometheus.NewDesc(prometheus.BuildFQName(namespace, "connection", "upload"), "Number of bytes for specific connection that uploaded by clash.", nil, nil)
)

type Exporter struct {
	mutex sync.RWMutex

	Client         IClient
	testUrl        string
	testUrlTimeout time.Duration

	totalScrapes prometheus.Counter
}

// NewExporter returns an initialized Exporter.
func NewExporter(client IClient, testUrl string, testUrlTimeout time.Duration) (*Exporter, error) {
	return &Exporter{
		Client:         client,
		testUrl:        testUrl,
		testUrlTimeout: testUrlTimeout,
		totalScrapes: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "exporter_scrapes_total",
			Help:      "Current total Clash scrapes.",
		}),
	}, nil
}

func (e *Exporter) Describe(descs chan<- *prometheus.Desc) {
	descs <- clashInfo
	descs <- clashUp
	descs <- proxyDelay
	descs <- downloadTotal
	descs <- uploadTotal
	descs <- connectionDownload
	descs <- connectionUpload
	descs <- e.totalScrapes.Desc()
}

func (e *Exporter) Collect(metrics chan<- prometheus.Metric) {
	e.mutex.Lock() // To protect metrics from concurrent collects.
	defer e.mutex.Unlock()

	up := e.scrape(metrics)
	metrics <- prometheus.MustNewConstMetric(clashUp, prometheus.GaugeValue, up)
	metrics <- e.totalScrapes
}

func (e *Exporter) scrapeVersion(metrics chan<- prometheus.Metric) error {
	v, err := e.Client.GetVersion()
	if err != nil {
		return err
	}
	metrics <- prometheus.MustNewConstMetric(clashInfo, prometheus.GaugeValue, 1, strconv.FormatBool(v.Premium), v.Version)
	return nil
}

func (e *Exporter) scrapeProxies(metrics chan<- prometheus.Metric) error {
	proxies, err := e.Client.GetProxies()
	if err != nil {
		return err
	}
	for proxyName, delay := range GetAllProxyDelay(proxies, IsConnectionProxy, e.Client, e.testUrl, e.testUrlTimeout) {
		proxy := proxies[proxyName]
		metrics <- prometheus.MustNewConstMetric(proxyDelay, prometheus.GaugeValue, float64(delay), proxy.Type, proxy.Name, "")
	}
	return nil
}

func (e *Exporter) scrapeProvidersProxies(metrics chan<- prometheus.Metric) error {
	providers, err := e.Client.GetProvidersProxies()
	if err != nil {
		return err
	}
	wg := sync.WaitGroup{}
	count := 0
	for _, provider := range providers {
		if provider.VehicleType == VehicleTypeHTTP || provider.VehicleType == VehicleTypeFile {
			wg.Add(1)
			count += 1
			go func(provider *Provider) {
				defer wg.Done()
				err = e.Client.ProviderProxiesHealthCheck(provider.Name)
				if err != nil {
					level.Error(logger).Log("msg", "error when do health check", "err", err, "provider", provider.Name)
				}
			}(provider)
		}
	}
	wg.Wait()
	if count == 0 {
		level.Info(logger).Log("msg", "no provider do health check")
		return nil
	}

	providers, err = e.Client.GetProvidersProxies()
	if err != nil {
		return err
	}
	for _, provider := range providers {
		if provider.VehicleType == VehicleTypeHTTP || provider.VehicleType == VehicleTypeFile {
			for _, proxy := range provider.Proxies {
				if IsConnectionProxy(proxy) {
					n := len(proxy.History)
					if n >= 1 && time.Now().Sub(proxy.History[n-1].Time) <= 1*time.Minute {
						delay := proxy.History[n-1].Delay
						if delay == 0 {
							delay = MaxDelay
						}
						metrics <- prometheus.MustNewConstMetric(proxyDelay, prometheus.GaugeValue, float64(delay), proxy.Type, proxy.Name, provider.Name)
					} else {
						level.Error(logger).Log("msg", "provider proxy should have at least one history", "proxy", proxy.Name, "providerName", provider.Name)
					}
				}
			}
		}
	}
	return nil
}

func (e *Exporter) scrapeConnections(metrics chan<- prometheus.Metric) error {
	s, err := e.Client.GetConnections()
	if err != nil {
		return err
	}
	metrics <- prometheus.MustNewConstMetric(downloadTotal, prometheus.CounterValue, float64(s.DownloadTotal))
	metrics <- prometheus.MustNewConstMetric(uploadTotal, prometheus.CounterValue, float64(s.UploadTotal))
	// TODO: 连接不是一直存在的，相同目的地址可能有多个连接，需要修改 clash 数据
	//for _, c := range s.Connections {
	//	metrics <- prometheus.MustNewConstMetric(connectionDownload, prometheus.CounterValue, float64(c.DownloadTotal))
	//	metrics <- prometheus.MustNewConstMetric(connectionUpload, prometheus.CounterValue, float64(c.UploadTotal))
	//}
	return nil
}

func (e *Exporter) scrape(metrics chan<- prometheus.Metric) (up float64) {
	e.totalScrapes.Inc()
	errors := make([]error, 0)
	scrapes := []func(metrics chan<- prometheus.Metric) error{
		e.scrapeVersion, e.scrapeProxies, e.scrapeProvidersProxies, e.scrapeConnections,
	}
	wg := sync.WaitGroup{}
	wg.Add(len(scrapes))
	for _, fn := range scrapes {
		go func(fn func(metrics chan<- prometheus.Metric) error) {
			defer wg.Done()
			err := fn(metrics)
			if err != nil {
				level.Warn(logger).Log("msg", "error when get proxy delay", "err", err)
			}
		}(fn)
	}
	wg.Wait()
	if len(errors) > 0 {
		level.Error(logger).Log("msg", "error when scrape clash", "err", errors)
		return 0
	}
	return 1
}

func CollectToText(c prometheus.Collector) (string, error) {
	buf := &strings.Builder{}
	enc := expfmt.NewEncoder(buf, expfmt.FmtText)
	reg := prometheus.NewPedanticRegistry()
	reg.MustRegister(c)
	mfs, err := reg.Gather()
	if err != nil {
		return "", err
	}
	for _, mf := range mfs {
		err = enc.Encode(mf)
		if err != nil {
			return "", err
		}
	}
	return buf.String(), nil
}

var (
	listenAddress string
	tlsConfigPath string
	metricsPath   string

	externalController string
	secret             string
	testUrl            string
	testUrlTimeout     time.Duration

	logger = promlog.New(&promlog.Config{})
	cmd    = &cobra.Command{
		Use: "clash_exporter",
	}
)

func init() {
	cmd.Flags().StringVar(&listenAddress, "web.listen-address", ":9877", "Address to listen on for web interface and telemetry")
	cmd.Flags().StringVar(&metricsPath, "web.telemetry-path", "/metrics", "Path under which to expose metrics")
	cmd.Flags().StringVar(&tlsConfigPath, "web.config.file", "", "[EXPERIMENTAL] Path to configuration file that can enable TLS or authentication")

	cmd.Flags().StringVar(&externalController, "clash.external-controller", "http://127.0.0.1:9090/", "RESTful web API listening address")
	cmd.Flags().StringVar(&secret, "clash.secret", "", "Secret for the RESTful API")
	cmd.Flags().StringVar(&testUrl, "clash.test-url", DefaultTestUrl, "")
	cmd.Flags().DurationVar(&testUrlTimeout, "clash.test-url-timeout", DefaultTestUrlTimeout, "")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		client, err := NewClient(externalController, secret)
		if err != nil {
			return err
		}
		c, err := NewExporter(client, testUrl, testUrlTimeout)
		if err != nil {
			return err
		}
		prometheus.MustRegister(version.NewCollector("clash_exporter"))
		prometheus.MustRegister(c)
		level.Info(logger).Log("msg", "Listening on address", "address", listenAddress)
		http.Handle(metricsPath, promhttp.Handler())
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte(`<html>
             <head><title>Clash Exporter</title></head>
             <body>
             <h1>Clash Exporter</h1>
             <p><a href='` + metricsPath + `'>Metrics</a></p>
             </body>
             </html>`))
		})
		srv := &http.Server{Addr: listenAddress}
		err = web.ListenAndServe(srv, tlsConfigPath, logger)
		if err != nil {
			level.Error(logger).Log("msg", "Error starting HTTP server", "err", err)
		}
		return err
	}
}

func main() {
	level.Info(logger).Log("msg", "Starting clash_exporter", "version", version.Info())
	level.Info(logger).Log("msg", "Build context", "context", version.BuildContext())
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
