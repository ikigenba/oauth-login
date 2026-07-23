package main

import (
	"context"
	"crypto/rand"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/ikigenba/oauth-login/internal/browser"
	"github.com/ikigenba/oauth-login/internal/callback"
	"github.com/ikigenba/oauth-login/internal/oauth"
)

var version = "dev"

const helpExample = `
Example:
  oauth-login \
    --auth-url  https://auth.openai.com/oauth/authorize \
    --token-url https://auth.openai.com/oauth/token \
    --client-id app_EMoamEEZ73f0CkXaXp7hrann \
    --scope "openid profile email offline_access" \
    --port 1455 --callback-path /auth/callback \
    > auth.json

Basic authentication:
  --token-header "Authorization=Basic $(printf '%s:%s' "$ID" "$SECRET" | base64 -w0)"
`

const helpFlags = `  --auth-url string
        authorization endpoint (required)
  --token-url string
        token endpoint (required)
  --client-id string
        OAuth client id (required)
  --scope string
        space-separated OAuth scopes
  --client-secret string
        client secret sent in the token request body
  --callback-host string
        host used in the redirect URI (default "localhost")
  --port int
        loopback callback port; 0 chooses an available port (default 0)
  --callback-path string
        callback route and redirect URI path (default "/callback")
  --auth-param key=value
        extra authorize parameter (repeatable)
  --token-param key=value
        extra token parameter (repeatable)
  --token-header key=value
        extra token request header (repeatable)
  --no-browser
        print the authorize URL without opening a browser
  --timeout duration
        maximum time to wait for the callback (default 5m)
  -h
        print help and exit
  -V
        print version and exit
`

type valuesFlag []string

func (v *valuesFlag) String() string { return strings.Join(*v, ",") }

func (v *valuesFlag) Set(value string) error {
	*v = append(*v, value)
	return nil
}

type options struct {
	authURL      string
	tokenURL     string
	clientID     string
	scope        string
	clientSecret string
	callbackHost string
	port         int
	callbackPath string
	authParams   valuesFlag
	tokenParams  valuesFlag
	tokenHeaders valuesFlag
	noBrowser    bool
	timeout      time.Duration
	showVersion  bool
}

type dependencies struct {
	launcher   browser.Launcher
	entropy    io.Reader
	httpClient *http.Client
}

func main() {
	deps := dependencies{
		launcher:   browser.New(),
		entropy:    rand.Reader,
		httpClient: http.DefaultClient,
	}
	os.Exit(run(context.Background(), os.Args[1:], os.Stdout, os.Stderr, deps))
}

func run(ctx context.Context, args []string, stdout, stderr io.Writer, deps dependencies) int {
	opts, help, err := parseOptions(args, stderr)
	if help {
		return 0
	}
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}
	if opts.showVersion {
		_, _ = io.WriteString(stdout, version+"\n")
		return 0
	}

	authParams, err := parseParams("--auth-param", opts.authParams, authReserved)
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}
	tokenParams, err := parseParams("--token-param", opts.tokenParams, tokenReserved)
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}
	tokenHeaders, err := parseParams("--token-header", opts.tokenHeaders, nil)
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}
	if opts.clientSecret != "" && hasHeader(tokenHeaders, "Authorization") {
		fmt.Fprintln(stderr, "error: --client-secret cannot be used with an Authorization --token-header")
		return 1
	}

	server := &callback.Server{Port: opts.port, CallbackPath: opts.callbackPath}
	port, err := server.Listen()
	if err != nil {
		fmt.Fprintf(stderr, "error: listen for callback: %v\n", err)
		return 1
	}

	client := &oauth.Client{
		AuthURL:      opts.authURL,
		TokenURL:     opts.tokenURL,
		ClientID:     opts.clientID,
		ClientSecret: opts.clientSecret,
		RedirectURI:  redirectURI(opts.callbackHost, port, opts.callbackPath),
		Scope:        opts.scope,
		Rand:         deps.entropy,
		HTTPClient:   deps.httpClient,
	}
	session, err := client.NewSession()
	if err != nil {
		fmt.Fprintf(stderr, "error: create OAuth session: %v\n", err)
		return 1
	}
	server.State = session.State
	authorizeURL, err := client.AuthorizeURL(session, authParams)
	if err != nil {
		fmt.Fprintf(stderr, "error: build authorize URL: %v\n", err)
		return 1
	}

	fmt.Fprintf(stderr, "Open this URL to continue:\n%s\n", authorizeURL)
	if !opts.noBrowser {
		if err := deps.launcher.Open(authorizeURL); err != nil {
			fmt.Fprintf(stderr, "note: could not open browser: %v\n", err)
		}
	}

	waitCtx, cancel := context.WithTimeout(ctx, opts.timeout)
	defer cancel()
	result, err := server.Wait(waitCtx)
	if err != nil {
		fmt.Fprintf(stderr, "error: callback failed: %v\n", err)
		return 1
	}
	fmt.Fprintln(stderr, "Callback received; exchanging authorization code.")
	body, err := client.Exchange(waitCtx, session, result.Code, tokenParams, tokenHeaders)
	if err != nil {
		fmt.Fprintf(stderr, "error: token exchange failed: %v\n", err)
		return 1
	}
	if _, err := stdout.Write(body); err != nil {
		fmt.Fprintf(stderr, "error: write token response: %v\n", err)
		return 1
	}
	return 0
}

func parseOptions(args []string, output io.Writer) (options, bool, error) {
	var opts options
	fs := flag.NewFlagSet("oauth-login", flag.ContinueOnError)
	fs.SetOutput(output)
	fs.StringVar(&opts.authURL, "auth-url", "", "authorization endpoint (required)")
	fs.StringVar(&opts.tokenURL, "token-url", "", "token endpoint (required)")
	fs.StringVar(&opts.clientID, "client-id", "", "OAuth client id (required)")
	fs.StringVar(&opts.scope, "scope", "", "space-separated OAuth scopes")
	fs.StringVar(&opts.clientSecret, "client-secret", "", "client secret sent in the token request body")
	fs.StringVar(&opts.callbackHost, "callback-host", "localhost", "host used in the redirect URI")
	fs.IntVar(&opts.port, "port", 0, "loopback callback port (0 chooses an available port)")
	fs.StringVar(&opts.callbackPath, "callback-path", "/callback", "callback route and redirect URI path")
	fs.Var(&opts.authParams, "auth-param", "extra authorize parameter as key=value (repeatable)")
	fs.Var(&opts.tokenParams, "token-param", "extra token parameter as key=value (repeatable)")
	fs.Var(&opts.tokenHeaders, "token-header", "extra token request header as key=value (repeatable)")
	fs.BoolVar(&opts.noBrowser, "no-browser", false, "print the authorize URL without opening a browser")
	fs.DurationVar(&opts.timeout, "timeout", 5*time.Minute, "maximum time to wait for the callback")
	fs.BoolVar(&opts.showVersion, "V", false, "print version and exit")
	fs.Usage = func() {
		fmt.Fprintf(output, "Usage: oauth-login [flags]\n\nFlags:\n")
		io.WriteString(output, helpFlags)
		io.WriteString(output, helpExample)
	}
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return opts, true, nil
		}
		return opts, false, err
	}
	if opts.showVersion {
		return opts, false, nil
	}
	for _, required := range []struct {
		name  string
		value string
	}{{"--auth-url", opts.authURL}, {"--token-url", opts.tokenURL}, {"--client-id", opts.clientID}} {
		if required.value == "" {
			return opts, false, fmt.Errorf("missing required flag %s", required.name)
		}
	}
	if opts.timeout <= 0 {
		return opts, false, errors.New("--timeout must be greater than zero")
	}
	return opts, false, nil
}

var authReserved = map[string]bool{
	"response_type": true, "client_id": true, "redirect_uri": true,
	"state": true, "code_challenge": true, "code_challenge_method": true, "scope": true,
}

var tokenReserved = map[string]bool{
	"grant_type": true, "code": true, "code_verifier": true, "redirect_uri": true,
	"client_id": true, "client_secret": true,
}

func parseParams(flagName string, values []string, reserved map[string]bool) ([]oauth.Param, error) {
	params := make([]oauth.Param, 0, len(values))
	for _, value := range values {
		key, val, ok := strings.Cut(value, "=")
		if !ok || key == "" {
			return nil, fmt.Errorf("%s value %q must be key=value", flagName, value)
		}
		if reserved[key] {
			return nil, fmt.Errorf("%s key %q is reserved", flagName, key)
		}
		params = append(params, oauth.Param{Key: key, Value: val})
	}
	return params, nil
}

func hasHeader(headers []oauth.Param, name string) bool {
	for _, header := range headers {
		if strings.EqualFold(header.Key, name) {
			return true
		}
	}
	return false
}

func redirectURI(host string, port int, path string) string {
	return (&url.URL{
		Scheme: "http",
		Host:   net.JoinHostPort(host, strconv.Itoa(port)),
		Path:   path,
	}).String()
}
