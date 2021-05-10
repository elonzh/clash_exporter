package main

import (
	"github.com/davecgh/go-spew/spew"
	"os"
	"testing"
)

func TestClient(t *testing.T) {
	baseUrl := os.Getenv("CLASH_EXTERNAL_CONTROLLER")
	if baseUrl == "" {
		t.Skip("CLASH_EXTERNAL_CONTROLLER is not provided")
	}
	client, _ := NewClient(baseUrl, os.Getenv("CLASH_SECRET"))
	version, err := client.GetVersion()
	if err != nil || version.Version == "" {
		t.Fail()
		t.Errorf("GetVersion failed because of error = %v, rv = %v", err, spew.Sprint(version))
	}

	proxies, err := client.GetProxies()
	if err != nil || len(proxies) == 0 {
		t.Fail()
		t.Errorf("GetConnections failed because of error = %v, rv = %v", err, spew.Sprint(proxies))
	}

	delays := GetAllProxyDelay(proxies, IsConnectionProxy, client, DefaultTestUrl, DefaultTestUrlTimeout)
	if len(delays) == 0 {
		t.Fail()
		t.Errorf("GetAllProxyDelay failed because of error = %v, rv = %v", err, spew.Sprint(delays))
	}

	connections, err := client.GetConnections()
	if err != nil || connections.DownloadTotal == 0 || connections.UploadTotal == 0 {
		t.Fail()
		t.Errorf("GetConnections failed because of error = %v, rv = %v", err, spew.Sprint(connections))
	}
}
