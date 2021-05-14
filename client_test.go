package main

import (
	"github.com/davecgh/go-spew/spew"
	"os"
	"testing"
)

func TestClient(t *testing.T) {
	// TODO: mock response
	baseUrl := os.Getenv("CLASH_EXTERNAL_CONTROLLER")
	if baseUrl == "" {
		t.Skip("CLASH_EXTERNAL_CONTROLLER is not provided")
	}

	client, _ := NewClient(baseUrl, os.Getenv("CLASH_SECRET"))

	t.Run("GetVersion", func(t *testing.T) {
		t.Parallel()
		version, err := client.GetVersion()
		if err != nil || version.Version == "" {
			t.Fail()
			t.Errorf("GetVersion failed because of error = %v, rv = %v", err, spew.Sprint(version))
		}
	})

	t.Run("GetProxies and GetAllProxyDelay", func(t *testing.T) {
		t.Parallel()
		proxies, err := client.GetProxies()
		if err != nil || len(proxies) == 0 {
			t.Fail()
			t.Errorf("GetConnections failed because of error = %v, rv = %v", err, spew.Sprint(proxies))
		}
		delays := GetAllProxyDelay(proxies, nil, client, DefaultTestUrl, DefaultTestUrlTimeout)
		if len(delays) == 0 {
			t.Fail()
			t.Errorf("GetAllProxyDelay failed because of error = %v, rv = %v", err, spew.Sprint(delays))
		}
	})

	t.Run("GetProvidersProxies and ProviderProxiesHealthCheck", func(t *testing.T) {
		t.Parallel()
		providers, err := client.GetProvidersProxies()
		if err != nil || len(providers) == 0 {
			t.Fail()
			t.Errorf("GetProvidersProxies failed because of error = %v, rv = %v", err, spew.Sprint(providers))
		}
		for _, v := range providers {
			err = client.ProviderProxiesHealthCheck(v.Name)
			if err != nil {
				t.Fail()
				t.Errorf("ProviderProxiesHealthCheck failed because of error = %v, provider = %v", err, v)
			}
		}
	})

	t.Run("GetConnections", func(t *testing.T) {
		t.Parallel()
		connections, err := client.GetConnections()
		if err != nil || connections.DownloadTotal == 0 || connections.UploadTotal == 0 {
			t.Fail()
			t.Errorf("GetConnections failed because of error = %v, rv = %v", err, spew.Sprint(connections))
		}
	})
}
