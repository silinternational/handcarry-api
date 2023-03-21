package main

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"time"

	buffalo "github.com/gobuffalo/buffalo/runtime"
	"github.com/gobuffalo/buffalo/servers"
	"github.com/rollbar/rollbar-go"

	"github.com/silinternational/wecarry-api/actions"
	"github.com/silinternational/wecarry-api/domain"
)

// main is the starting point for your Buffalo application.
// You can feel free and add to this `main` method, change
// what it does, etc...
// All we ask is that, at some point, you make sure to
// call `app.Serve()`, unless you don't want to start your
// application that is. :)
func main() {
	// init rollbar
	rollbar.SetToken(domain.Env.RollbarToken)
	rollbar.SetEnvironment(domain.Env.GoEnv)
	rollbar.SetCodeVersion(domain.Env.CommitID)
	rollbar.SetServerRoot(domain.Env.RollbarServerRoot)

	srv, err := getServer()
	if err != nil {
		domain.ErrLogger.Printf(err.Error())
		os.Exit(1)
	}

	app := actions.App()

	fmt.Printf("Go version: %s\n", runtime.Version())
	fmt.Printf("Buffalo version: %s\n", buffalo.Version)
	fmt.Printf("Buffalo build info: %s\n", buffalo.Build())
	fmt.Printf("Commit hash: %s\n", domain.Env.CommitID)

	rollbar.WrapAndWait(func() {
		if err := app.Serve(srv); err != nil {
			if err.Error() != "context canceled" {
				panic(err)
			}
			os.Exit(0)
		}
	})
}

func getServer() (servers.Server, error) {
	const (
		certFile = "cert.pem"
		keyFile  = "key.pem"
	)

	if domain.Env.DisableTLS {
		return servers.New(), nil
	}

	err := generateCert(certFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("generate cert: %w", err)
	}

	cfg, err := tlsConfig(certFile, keyFile)
	if err != nil {
		return servers.New(), fmt.Errorf("get TLS config: %w", err)
	}
	listener, err := tls.Listen("tcp", fmt.Sprintf(":%d", domain.Env.ServerPort), cfg)
	if err != nil {
		return servers.New(), fmt.Errorf("get TLS listener: %w", err)
	}

	return servers.WrapListener(&http.Server{ReadHeaderTimeout: time.Second * 15}, listener), nil
}

func tlsConfig(certFile, keyFile string) (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("load cert/key files: %w", err)
	}

	config := tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
	}
	return &config, nil
}

/*
# Notes about `main.go`

## SSL Support

We recommend placing your application behind a proxy, such as
Apache or Nginx and letting them do the SSL heavy lifting
for you. https://gobuffalo.io/en/docs/proxy

## Buffalo Build

When `buffalo build` is run to compile your binary, this `main`
function will be at the heart of that binary. It is expected
that your `main` function will start your application using
the `app.Serve()` method.

*/
