package main

import (
	"github.com/gobuffalo/envy"
	"github.com/rollbar/rollbar-go"

	"github.com/silinternational/wecarry-api/actions"
)

var GitCommitHash string

// main is the starting point for your Buffalo application.
// You can feel free and add to this `main` method, change
// what it does, etc...
// All we ask is that, at some point, you make sure to
// call `app.Serve()`, unless you don't want to start your
// application that is. :)
func main() {

	// init rollbar
	rollbar.SetToken(envy.Get("ROLLBAR_TOKEN", ""))
	rollbar.SetEnvironment(envy.Get("GO_ENV", "development"))
	rollbar.SetCodeVersion(GitCommitHash)
	rollbar.SetServerRoot(envy.Get("ROLLBAR_SERVER_ROOT", "github.com/silinternational/wecarry-api"))

	app := actions.App()
	rollbar.WrapAndWait(func() {
		if err := app.Serve(); err != nil {
			panic(err)
		}
	})

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
