package main

import (
	"context"
	"flag"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/vladkanatov/terraform-provider-orchard/internal/provider"
)

// Version and commit of the provider. Both are set at build time via ldflags
// (see .goreleaser.yml); these are fallback values for local development.
var version = "0.0.1"
var commit = "unknown"

func main() {
	var debug bool
	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	opts := providerserver.ServeOpts{
		Address: "registry.terraform.io/vladkanatov/orchard",
		Debug:   debug,
	}

	err := providerserver.Serve(context.Background(), provider.New(version), opts)
	if err != nil {
		log.Fatal(err)
	}
}
