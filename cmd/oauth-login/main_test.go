package main

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"
)

type fakeLauncher struct {
	mu    sync.Mutex
	calls []string
	open  func(string)
	err   error
}

func (f *fakeLauncher) Open(rawURL string) error {
	f.mu.Lock()
	f.calls = append(f.calls, rawURL)
	f.mu.Unlock()
	if f.open != nil {
		f.open(rawURL)
	}
	return f.err
}

func (f *fakeLauncher) count() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.calls)
}

type countingWriter struct {
	bytes.Buffer
	writes int
}

func (w *countingWriter) Write(p []byte) (int, error) {
	w.writes++
	return w.Buffer.Write(p)
}

type providerObservation struct {
	authRedirect  string
	tokenRedirect string
	authHost      string
}

type providerFixture struct {
	server       *httptest.Server
	launcher     *fakeLauncher
	observations chan providerObservation
}

func newProviderFixture(t *testing.T, tokenStatus int, tokenBody []byte, launchErr error) *providerFixture {
	t.Helper()
	observations := make(chan providerObservation, 1)
	var mu sync.Mutex
	var authRedirect, authHost string
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	mux.HandleFunc("/authorize", func(w http.ResponseWriter, r *http.Request) {
		redirect := r.URL.Query().Get("redirect_uri")
		if redirect == "" || r.URL.Query().Get("state") == "" || r.URL.Query().Get("code_challenge") == "" {
			http.Error(w, "incomplete authorize request", http.StatusBadRequest)
			return
		}
		parsed, err := url.Parse(redirect)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		mu.Lock()
		authRedirect, authHost = redirect, parsed.Hostname()
		mu.Unlock()
		query := url.Values{"code": {"provider-code"}, "state": {r.URL.Query().Get("state")}}
		http.Redirect(w, r, redirect+"?"+query.Encode(), http.StatusFound)
	})
	mux.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		mu.Lock()
		observation := providerObservation{
			authRedirect: authRedirect, tokenRedirect: r.Form.Get("redirect_uri"), authHost: authHost,
		}
		mu.Unlock()
		select {
		case observations <- observation:
		default:
		}
		w.WriteHeader(tokenStatus)
		_, _ = w.Write(tokenBody)
	})
	launcher := &fakeLauncher{err: launchErr}
	launcher.open = func(authorizeURL string) {
		go func() {
			response, err := http.Get(authorizeURL)
			if err == nil {
				_ = response.Body.Close()
			}
		}()
	}
	return &providerFixture{server: server, launcher: launcher, observations: observations}
}

func (f *providerFixture) close() { f.server.Close() }

func (f *providerFixture) args(extra ...string) []string {
	args := []string{
		"--auth-url", f.server.URL + "/authorize",
		"--token-url", f.server.URL + "/token",
		"--client-id", "test-client",
		"--timeout", "2s",
	}
	return append(args, extra...)
}

func testDependencies(launcher *fakeLauncher) dependencies {
	return dependencies{
		launcher:   launcher,
		entropy:    bytes.NewReader(bytes.Repeat([]byte{0x5a}, 4096)),
		httpClient: http.DefaultClient,
	}
}

func TestMissingRequiredFlags(t *testing.T) {
	// R-1NJA-BC0X
	base := []string{"--auth-url", "https://issuer.invalid/authorize", "--token-url", "https://issuer.invalid/token", "--client-id", "client"}
	for _, missing := range []string{"--auth-url", "--token-url", "--client-id"} {
		t.Run(missing, func(t *testing.T) {
			args := make([]string, 0, len(base)-2)
			for i := 0; i < len(base); i += 2 {
				if base[i] != missing {
					args = append(args, base[i], base[i+1])
				}
			}
			var stdout, stderr bytes.Buffer
			launcher := &fakeLauncher{}
			if got := run(context.Background(), args, &stdout, &stderr, testDependencies(launcher)); got != 1 {
				t.Fatalf("exit = %d, want 1", got)
			}
			if stdout.Len() != 0 || !strings.Contains(stderr.String(), missing) {
				t.Fatalf("stdout=%q stderr=%q; want empty stdout and named flag", stdout.String(), stderr.String())
			}
			if launcher.count() != 0 {
				t.Fatal("launcher invoked during validation")
			}
		})
	}
}

func TestConflictingClientAuthentication(t *testing.T) {
	// R-1PZ3-2VIB
	base := []string{"--auth-url", "https://issuer.invalid/a", "--token-url", "https://issuer.invalid/t", "--client-id", "client"}
	for _, header := range []string{"Authorization=Basic abc", "authorization=Basic abc", "AUTHORIZATION=Basic abc"} {
		var stdout, stderr bytes.Buffer
		launcher := &fakeLauncher{}
		args := append(append([]string{}, base...), "--client-secret", "secret", "--token-header", header)
		if got := run(context.Background(), args, &stdout, &stderr, testDependencies(launcher)); got != 1 {
			t.Fatalf("exit = %d, want 1", got)
		}
		if stdout.Len() != 0 || !strings.Contains(stderr.String(), "--client-secret") || !strings.Contains(stderr.String(), "--token-header") {
			t.Fatalf("stdout=%q stderr=%q", stdout.String(), stderr.String())
		}
		if launcher.count() != 0 {
			t.Fatal("launcher invoked during validation")
		}
	}
	for _, lone := range [][]string{{"--client-secret", "secret"}, {"--token-header", "Authorization=Basic abc"}} {
		var help bytes.Buffer
		if _, _, err := parseOptions(append(append([]string{}, base...), lone...), &help); err != nil {
			t.Fatalf("authentication method alone rejected: %v", err)
		}
	}
}

func TestRepeatableParameterValidationPrecedesSideEffects(t *testing.T) {
	// R-1R6Z-GN90
	base := []string{"--auth-url", "https://issuer.invalid/a", "--token-url", "https://issuer.invalid/t", "--client-id", "client", "--port", "1"}
	cases := [][]string{
		{"--auth-param", "redirect_uri=http://wrong.invalid"},
		{"--auth-param", "state=x"},
		{"--token-param", "code=x"},
		{"--token-param", "client_secret=x"},
		{"--auth-param", "missing-equals"},
		{"--token-param", "missing-equals"},
		{"--token-header", "missing-equals"},
	}
	for _, extra := range cases {
		var stdout, stderr bytes.Buffer
		launcher := &fakeLauncher{}
		got := run(context.Background(), append(append([]string{}, base...), extra...), &stdout, &stderr, testDependencies(launcher))
		if got != 1 || stdout.Len() != 0 || launcher.count() != 0 {
			t.Fatalf("args=%v exit=%d stdout=%q launches=%d", extra, got, stdout.String(), launcher.count())
		}
		if !strings.Contains(stderr.String(), extra[0]) || strings.Contains(stderr.String(), "listen for callback") {
			t.Fatalf("validation did not precede bind: %q", stderr.String())
		}
	}
}

func TestSuccessfulLoginWritesOnlyExactProviderBody(t *testing.T) {
	// R-1SEV-UEZP
	body := []byte("{\"access_token\":\"raw\"}\x00tail")
	fixture := newProviderFixture(t, http.StatusOK, body, nil)
	defer fixture.close()
	var stdout countingWriter
	var stderr bytes.Buffer
	if got := run(context.Background(), fixture.args(), &stdout, &stderr, testDependencies(fixture.launcher)); got != 0 {
		t.Fatalf("exit=%d stderr=%q", got, stderr.String())
	}
	if !bytes.Equal(stdout.Bytes(), body) || stdout.writes != 1 {
		t.Fatalf("stdout=%q writes=%d, want exact body in one write", stdout.Bytes(), stdout.writes)
	}
}

func TestFailuresNeverWriteStdout(t *testing.T) {
	// R-1TMS-86QE
	tests := []struct {
		name string
		run  func(*bytes.Buffer, *bytes.Buffer) int
	}{
		{"validation", func(out, errOut *bytes.Buffer) int {
			return run(context.Background(), nil, out, errOut, testDependencies(&fakeLauncher{}))
		}},
		{"provider error", callbackFailureRun(t, url.Values{"error": {"access_denied"}})},
		{"state mismatch", callbackFailureRun(t, url.Values{"code": {"code"}, "state": {"wrong"}})},
		{"non-2xx token", tokenFailureRun(t)},
		{"timeout", func(out, errOut *bytes.Buffer) int {
			fixture := newProviderFixture(t, http.StatusOK, []byte("unused"), nil)
			defer fixture.close()
			return run(context.Background(), append(fixture.args(), "--no-browser", "--timeout", "15ms"), out, errOut, testDependencies(fixture.launcher))
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			if got := tt.run(&stdout, &stderr); got != 1 {
				t.Fatalf("exit=%d stderr=%q", got, stderr.String())
			}
			if stdout.Len() != 0 {
				t.Fatalf("stdout=%q, want empty", stdout.String())
			}
		})
	}
}

func callbackFailureRun(t *testing.T, callbackQuery url.Values) func(*bytes.Buffer, *bytes.Buffer) int {
	t.Helper()
	return func(out, errOut *bytes.Buffer) int {
		fixture := newProviderFixture(t, http.StatusOK, []byte("unused"), nil)
		defer fixture.close()
		fixture.launcher.open = func(authorizeURL string) {
			parsed, _ := url.Parse(authorizeURL)
			redirect := parsed.Query().Get("redirect_uri")
			go func() {
				response, err := http.Get(redirect + "?" + callbackQuery.Encode())
				if err == nil {
					_ = response.Body.Close()
				}
			}()
		}
		return run(context.Background(), fixture.args(), out, errOut, testDependencies(fixture.launcher))
	}
}

func tokenFailureRun(t *testing.T) func(*bytes.Buffer, *bytes.Buffer) int {
	t.Helper()
	return func(out, errOut *bytes.Buffer) int {
		fixture := newProviderFixture(t, http.StatusBadRequest, []byte("rejected"), nil)
		defer fixture.close()
		return run(context.Background(), fixture.args(), out, errOut, testDependencies(fixture.launcher))
	}
}

func TestHumanOutputStaysOnStderr(t *testing.T) {
	// R-1UUO-LYH3
	fixture := newProviderFixture(t, http.StatusOK, []byte("token-file"), nil)
	defer fixture.close()
	var stdout, stderr bytes.Buffer
	if got := run(context.Background(), fixture.args(), &stdout, &stderr, testDependencies(fixture.launcher)); got != 0 {
		t.Fatalf("exit=%d stderr=%q", got, stderr.String())
	}
	if stdout.String() != "token-file" || strings.Contains(stdout.String(), "Open this URL") || strings.Contains(stdout.String(), "Callback received") {
		t.Fatalf("redirected stdout=%q", stdout.String())
	}
	if !strings.Contains(stderr.String(), "Open this URL") || !strings.Contains(stderr.String(), "Callback received") {
		t.Fatalf("stderr lacks human output: %q", stderr.String())
	}
}

func TestExitCodesAreZeroOrOne(t *testing.T) {
	// R-1W2K-ZQ7S
	fixture := newProviderFixture(t, http.StatusOK, []byte("ok"), nil)
	defer fixture.close()
	var stdout, stderr bytes.Buffer
	if got := run(context.Background(), fixture.args(), &stdout, &stderr, testDependencies(fixture.launcher)); got != 0 {
		t.Fatalf("successful exit=%d", got)
	}
	stdout.Reset()
	stderr.Reset()
	if got := run(context.Background(), nil, &stdout, &stderr, testDependencies(&fakeLauncher{})); got != 1 {
		t.Fatalf("failure exit=%d", got)
	}
}

func TestHelpDocumentsCompleteFlagSurfaceAndExamples(t *testing.T) {
	// R-1XAH-DHYH
	var stdout, stderr bytes.Buffer
	if got := run(context.Background(), []string{"--help"}, &stdout, &stderr, testDependencies(&fakeLauncher{})); got != 0 {
		t.Fatalf("help exit=%d", got)
	}
	if stdout.Len() != 0 {
		t.Fatalf("help wrote stdout: %q", stdout.String())
	}
	for _, text := range []string{
		"--auth-url", "--token-url", "--client-id", "--scope", "--client-secret", "--callback-host",
		"--port", "--callback-path", "--auth-param", "--token-param", "--token-header", "--no-browser", "--timeout",
		"https://auth.openai.com/oauth/authorize", "https://auth.openai.com/oauth/token", "app_EMoamEEZ73f0CkXaXp7hrann",
		"--port 1455", "--callback-path /auth/callback", "Authorization=Basic",
	} {
		if !strings.Contains(stderr.String(), text) {
			t.Errorf("help does not contain %q", text)
		}
	}
}

func TestCallbackHostMatchesAcrossAuthorizationAndExchange(t *testing.T) {
	// R-1YID-R9P6
	for _, host := range []string{"localhost", "127.0.0.1"} {
		t.Run(host, func(t *testing.T) {
			fixture := newProviderFixture(t, http.StatusOK, []byte("ok"), nil)
			defer fixture.close()
			extra := []string(nil)
			if host != "localhost" {
				extra = []string{"--callback-host", host}
			}
			var stdout, stderr bytes.Buffer
			if got := run(context.Background(), fixture.args(extra...), &stdout, &stderr, testDependencies(fixture.launcher)); got != 0 {
				t.Fatalf("exit=%d stderr=%q", got, stderr.String())
			}
			observation := <-fixture.observations
			if observation.authHost != host || observation.authRedirect != observation.tokenRedirect {
				t.Fatalf("observation=%+v, want host %q and identical redirects", observation, host)
			}
		})
	}
}

func TestCallbackTimeoutDefaultAndOverride(t *testing.T) {
	// R-1ZQA-51FV
	base := []string{"--auth-url", "https://issuer.invalid/a", "--token-url", "https://issuer.invalid/t", "--client-id", "client"}
	opts, _, err := parseOptions(base, io.Discard)
	if err != nil || opts.timeout != 5*time.Minute {
		t.Fatalf("default timeout=%s err=%v", opts.timeout, err)
	}
	opts, _, err = parseOptions(append(base, "--timeout", "37s"), io.Discard)
	if err != nil || opts.timeout != 37*time.Second {
		t.Fatalf("explicit timeout=%s err=%v", opts.timeout, err)
	}
	fixture := newProviderFixture(t, http.StatusOK, []byte("unused"), nil)
	defer fixture.close()
	var stdout, stderr bytes.Buffer
	started := time.Now()
	if got := run(context.Background(), append(fixture.args(), "--no-browser", "--timeout", "20ms"), &stdout, &stderr, testDependencies(fixture.launcher)); got != 1 {
		t.Fatalf("timeout exit=%d", got)
	}
	if elapsed := time.Since(started); elapsed > time.Second || !strings.Contains(stderr.String(), "deadline exceeded") {
		t.Fatalf("elapsed=%s stderr=%q", elapsed, stderr.String())
	}
}

func TestRedirectURIUsesActuallyBoundPortAndConfiguredPath(t *testing.T) {
	// R-20Y6-IT6K
	if got := redirectURI("localhost", 1455, "/auth/callback"); got != "http://localhost:1455/auth/callback" {
		t.Fatalf("fixed redirect=%q", got)
	}
	fixture := newProviderFixture(t, http.StatusOK, []byte("ok"), nil)
	defer fixture.close()
	var stdout, stderr bytes.Buffer
	if got := run(context.Background(), fixture.args(), &stdout, &stderr, testDependencies(fixture.launcher)); got != 0 {
		t.Fatalf("exit=%d stderr=%q", got, stderr.String())
	}
	observation := <-fixture.observations
	parsed, err := url.Parse(observation.authRedirect)
	if err != nil || parsed.Port() == "" || parsed.Port() == "0" || parsed.Path != "/callback" {
		t.Fatalf("ephemeral redirect=%q err=%v", observation.authRedirect, err)
	}
}

func TestLauncherFailureIsNonFatal(t *testing.T) {
	// R-1JVL-60SU
	fixture := newProviderFixture(t, http.StatusOK, []byte("ok-after-launch-error"), errors.New("launcher failed"))
	defer fixture.close()
	var stdout, stderr bytes.Buffer
	if got := run(context.Background(), fixture.args(), &stdout, &stderr, testDependencies(fixture.launcher)); got != 0 {
		t.Fatalf("exit=%d stderr=%q", got, stderr.String())
	}
	if stdout.String() != "ok-after-launch-error" || !strings.Contains(stderr.String(), "launcher failed") {
		t.Fatalf("stdout=%q stderr=%q", stdout.String(), stderr.String())
	}
}

func TestNoBrowserSuppressesLauncherButPrintsURL(t *testing.T) {
	// R-1L3H-JSJJ
	fixture := newProviderFixture(t, http.StatusOK, []byte("unused"), nil)
	defer fixture.close()
	var stdout, stderr bytes.Buffer
	if got := run(context.Background(), append(fixture.args(), "--no-browser", "--timeout", "15ms"), &stdout, &stderr, testDependencies(fixture.launcher)); got != 1 {
		t.Fatalf("exit=%d", got)
	}
	if fixture.launcher.count() != 0 || !strings.Contains(stderr.String(), fixture.server.URL+"/authorize") {
		t.Fatalf("launches=%d stderr=%q", fixture.launcher.count(), stderr.String())
	}
}

func TestUnsupportedLauncherErrorStillAllowsLogin(t *testing.T) {
	// R-1MBD-XKA8
	fixture := newProviderFixture(t, http.StatusOK, []byte("manual-flow-ok"), errors.New("no supported launcher"))
	defer fixture.close()
	var stdout, stderr bytes.Buffer
	if got := run(context.Background(), fixture.args(), &stdout, &stderr, testDependencies(fixture.launcher)); got != 0 {
		t.Fatalf("exit=%d stderr=%q", got, stderr.String())
	}
	if stdout.String() != "manual-flow-ok" || !strings.Contains(stderr.String(), "no supported launcher") {
		t.Fatalf("stdout=%q stderr=%q", stdout.String(), stderr.String())
	}
}
