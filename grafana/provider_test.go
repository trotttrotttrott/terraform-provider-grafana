package grafana

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// To run these acceptance tests, you will need a Grafana server.
// Grafana can be downloaded here: http://grafana.org/download/
//
// The tests will need an API key to authenticate with the server. To create
// one, use the menu for one of your installation's organizations (The
// "Main Org." is fine if you've just done a fresh installation to run these
// tests) to reach the "API Keys" admin page.
//
// Giving the API key the Admin role is the easiest way to ensure enough
// access is granted to run all of the tests.
//
// Once you've created the API key, set the GRAFANA_URL and GRAFANA_AUTH
// environment variables to the Grafana base URL and the API key respectively,
// and then run:
//    make testacc TEST=./builtin/providers/grafana

var testAccProviders map[string]*schema.Provider
var testAccProvider *schema.Provider

func init() {
	testAccProvider = Provider()
	testAccProviders = map[string]*schema.Provider{
		"grafana": testAccProvider,
	}
}

func TestProvider(t *testing.T) {
	if err := Provider().InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProvider_impl(t *testing.T) {
	var _ *schema.Provider = Provider()
}

func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("GRAFANA_URL"); v == "" {
		t.Fatal("GRAFANA_URL must be set for acceptance tests")
	}
	if v := os.Getenv("GRAFANA_AUTH"); v == "" {
		t.Fatal("GRAFANA_AUTH must be set for acceptance tests")
	}
	if v := os.Getenv("GRAFANA_ORG_ID"); v == "" {
		t.Fatal("GRAFANA_ORG_ID must be set for acceptance tests")
	}
}
