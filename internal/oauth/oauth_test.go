package oauth

import (
	"bytes"
	"context"
	"encoding/base64"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"
)

func testClient() *Client {
	return &Client{
		AuthURL:     "https://issuer.invalid/authorize",
		ClientID:    "client-id",
		RedirectURI: "http://127.0.0.1:8080/callback",
	}
}

func testSession() *Session {
	return &Session{State: "test-state", CodeVerifier: strings.Repeat("v", 43)}
}

// R-0EF0-5TL8
func TestVerifierUsesAllowedLengthAndCharactersAndRejectsInvalidValues(t *testing.T) {
	client := testClient()
	client.Rand = bytes.NewReader(make([]byte, 96))
	session, err := client.NewSession()
	if err != nil {
		t.Fatalf("NewSession() error = %v", err)
	}
	if length := len(session.CodeVerifier); length < 43 || length > 128 {
		t.Fatalf("verifier length = %d, want 43..128", length)
	}
	for _, char := range session.CodeVerifier {
		if !strings.ContainsRune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-._~", char) {
			t.Fatalf("verifier contains disallowed character %q", char)
		}
	}

	for _, verifier := range []string{strings.Repeat("a", 42), strings.Repeat("a", 129), strings.Repeat("a", 42) + "+"} {
		if got, err := client.AuthorizeURL(&Session{State: "state", CodeVerifier: verifier}, nil); err == nil || got != "" {
			t.Errorf("AuthorizeURL(verifier length %d) = %q, %v; want empty URL and error", len(verifier), got, err)
		}
	}
}

// R-0FMW-JLBX
func TestStateContainsEntropyDrawnFromRandAndShortStateDrawFails(t *testing.T) {
	entropy := make([]byte, 96)
	for i := range entropy {
		entropy[i] = byte(i)
	}
	client := testClient()
	client.Rand = bytes.NewReader(entropy)
	session, err := client.NewSession()
	if err != nil {
		t.Fatalf("NewSession() error = %v", err)
	}
	decoded, err := base64.RawURLEncoding.DecodeString(session.State)
	if err != nil {
		t.Fatalf("decode state: %v", err)
	}
	if len(decoded) < 16 {
		t.Fatalf("decoded state has %d bytes, want at least 16", len(decoded))
	}
	if !bytes.Equal(decoded, entropy[64:]) {
		t.Fatalf("state bytes = %v, want bytes supplied by Rand %v", decoded, entropy[64:])
	}

	client.Rand = bytes.NewReader(make([]byte, 64+15))
	if session, err := client.NewSession(); err == nil || session != nil {
		t.Fatalf("NewSession() with 15 state bytes = %#v, %v; want nil and error", session, err)
	}
}

// R-0GUS-XD2M
func TestAuthorizeURLAlwaysUsesS256(t *testing.T) {
	got := mustAuthorizeURL(t, testClient(), testSession(), nil)
	query := got.Query()
	if values := query["code_challenge_method"]; !reflect.DeepEqual(values, []string{"S256"}) {
		t.Fatalf("code_challenge_method = %v, want exactly [S256]", values)
	}
	if strings.Contains(got.RawQuery, "plain") {
		t.Fatalf("authorize query %q contains plain", got.RawQuery)
	}
}

// R-0I2P-B4TB
func TestAuthorizeURLMatchesRFC7636AppendixBChallenge(t *testing.T) {
	session := &Session{
		State:        "state",
		CodeVerifier: "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk",
	}
	got := mustAuthorizeURL(t, testClient(), session, nil).Query().Get("code_challenge")
	const want = "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM"
	if got != want {
		t.Fatalf("code_challenge = %q, want %q", got, want)
	}
	if strings.Contains(got, "=") {
		t.Fatalf("code_challenge %q contains padding", got)
	}
}

// R-0JAL-OWK0
func TestSuccessiveDefaultRandomSessionsDiffer(t *testing.T) {
	client := testClient()
	first, err := client.NewSession()
	if err != nil {
		t.Fatalf("first NewSession() error = %v", err)
	}
	second, err := client.NewSession()
	if err != nil {
		t.Fatalf("second NewSession() error = %v", err)
	}
	if first.State == second.State {
		t.Fatalf("successive state values are equal: %q", first.State)
	}
	if first.CodeVerifier == second.CodeVerifier {
		t.Fatalf("successive verifier values are equal: %q", first.CodeVerifier)
	}
}

// R-0KII-2OAP
func TestNilRandDefaultsToCryptoRandAndShortVerifierReadErrors(t *testing.T) {
	client := testClient()
	session, err := client.NewSession()
	if err != nil {
		t.Fatalf("NewSession() with nil Rand error = %v", err)
	}
	if session.State == "" || session.CodeVerifier == "" {
		t.Fatalf("NewSession() with nil Rand returned zero-valued secrets: %#v", session)
	}

	client.Rand = bytes.NewReader(make([]byte, 63))
	if session, err := client.NewSession(); err == nil || session != nil {
		t.Fatalf("NewSession() with short Rand = %#v, %v; want nil and error", session, err)
	}
}

// R-0LQE-GG1E
func TestAuthorizeURLHasExactlyRequiredEncodedParameters(t *testing.T) {
	client := testClient()
	client.AuthURL = "https://issuer.invalid/oauth/authorize"
	client.ClientID = "client id&value"
	client.RedirectURI = "http://[::1]:8123/callback path?from=a&next=/home"
	session := &Session{State: "state +&", CodeVerifier: strings.Repeat("v", 43)}
	got := mustAuthorizeURL(t, client, session, nil)

	if got.Scheme != "https" || got.Host != "issuer.invalid" || got.Path != "/oauth/authorize" {
		t.Fatalf("authorize URL base = %s://%s%s, want configured base", got.Scheme, got.Host, got.Path)
	}
	want := map[string]string{
		"response_type":         "code",
		"client_id":             client.ClientID,
		"redirect_uri":          client.RedirectURI,
		"state":                 session.State,
		"code_challenge":        "7w_YNF9DSfIdPf_pRjSq646_kPr-2-o9NAl16JGghdM",
		"code_challenge_method": "S256",
	}
	if len(got.Query()) != len(want) {
		t.Fatalf("query = %v, want exactly %d keys", got.Query(), len(want))
	}
	for key, value := range want {
		if values := got.Query()[key]; !reflect.DeepEqual(values, []string{value}) {
			t.Errorf("query[%q] = %v, want [%q]", key, values, value)
		}
	}
	if !strings.Contains(got.RawQuery, "redirect_uri=") || strings.Contains(got.RawQuery, "callback path") {
		t.Fatalf("redirect_uri is not percent encoded in raw query %q", got.RawQuery)
	}
}

// R-0MYA-U7S3
func TestAuthorizeURLIncludesOnlyNonemptyScope(t *testing.T) {
	client := testClient()
	client.Scope = "openid profile custom:value"
	if got := mustAuthorizeURL(t, client, testSession(), nil).Query()["scope"]; !reflect.DeepEqual(got, []string{client.Scope}) {
		t.Fatalf("nonempty scope = %v, want [%q]", got, client.Scope)
	}
	client.Scope = ""
	if got, present := mustAuthorizeURL(t, client, testSession(), nil).Query()["scope"]; present {
		t.Fatalf("empty scope emitted as %v", got)
	}
}

// R-0O67-7ZIS
func TestAuthorizeURLPreservesExtraParametersAndRepeats(t *testing.T) {
	extra := []Param{{Key: "audience", Value: "one & two"}, {Key: "prompt", Value: "select account"}, {Key: "audience", Value: "second/value"}}
	got := mustAuthorizeURL(t, testClient(), testSession(), extra)
	if values := got.Query()["audience"]; !reflect.DeepEqual(values, []string{"one & two", "second/value"}) {
		t.Fatalf("audience values = %v, want both occurrences", values)
	}
	if values := got.Query()["prompt"]; !reflect.DeepEqual(values, []string{"select account"}) {
		t.Fatalf("prompt values = %v, want one exact occurrence", values)
	}
}

// R-0PE3-LR9H
func TestAuthorizeURLRejectsEveryReservedExtraAndNamesIt(t *testing.T) {
	reserved := []string{"response_type", "client_id", "redirect_uri", "state", "code_challenge", "code_challenge_method", "scope"}
	for _, key := range reserved {
		t.Run(key, func(t *testing.T) {
			got, err := testClient().AuthorizeURL(testSession(), []Param{{Key: key, Value: "override"}})
			if err == nil || got != "" {
				t.Fatalf("AuthorizeURL() = %q, %v; want empty URL and error", got, err)
			}
			if !strings.Contains(err.Error(), key) {
				t.Fatalf("error %q does not name reserved key %q", err, key)
			}
			if key == "redirect_uri" {
				for _, flag := range []string{"--callback-host", "--port", "--callback-path"} {
					if !strings.Contains(err.Error(), flag) {
						t.Errorf("redirect_uri error %q does not point to %s", err, flag)
					}
				}
			}
		})
	}
}

// R-0QLZ-ZJ06
func TestExchangePostsExactlyRequiredFormFields(t *testing.T) {
	client, captured := exchangeServer(t, http.StatusOK, []byte("ok"))
	client.ClientID = "client & id"
	client.RedirectURI = "http://[::1]/callback?x=a&b=c"
	session := &Session{CodeVerifier: "verifier/value"}
	if _, err := client.Exchange(context.Background(), session, "code + value", nil, nil); err != nil {
		t.Fatalf("Exchange() error = %v", err)
	}
	request := <-captured
	if request.method != http.MethodPost {
		t.Fatalf("method = %q, want POST", request.method)
	}
	if request.contentType != "application/x-www-form-urlencoded" {
		t.Fatalf("Content-Type = %q", request.contentType)
	}
	want := url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {"code + value"},
		"code_verifier": {session.CodeVerifier},
		"redirect_uri":  {client.RedirectURI},
		"client_id":     {client.ClientID},
	}
	if !reflect.DeepEqual(request.form, want) {
		t.Fatalf("form = %v, want exactly %v", request.form, want)
	}
}

// R-0RTW-DAQV
func TestExchangeIncludesOnlyConfiguredClientSecret(t *testing.T) {
	client, captured := exchangeServer(t, http.StatusOK, []byte("ok"))
	client.ClientSecret = "secret & value"
	if _, err := client.Exchange(context.Background(), testSession(), "code", nil, nil); err != nil {
		t.Fatalf("Exchange() with secret error = %v", err)
	}
	if got := (<-captured).form["client_secret"]; !reflect.DeepEqual(got, []string{client.ClientSecret}) {
		t.Fatalf("client_secret = %v, want exact configured secret", got)
	}

	client.ClientSecret = ""
	if _, err := client.Exchange(context.Background(), testSession(), "code", nil, nil); err != nil {
		t.Fatalf("Exchange() without secret error = %v", err)
	}
	if got, present := (<-captured).form["client_secret"]; present {
		t.Fatalf("empty client_secret emitted as %v", got)
	}
}

// R-0U9P-4U89
func TestExchangeAppliesTokenHeadersUnaltered(t *testing.T) {
	client, captured := exchangeServer(t, http.StatusOK, []byte("ok"))
	headers := []Param{{Key: "Authorization", Value: "Basic abc+/="}, {Key: "X-Custom", Value: "exact value"}}
	if _, err := client.Exchange(context.Background(), testSession(), "code", nil, headers); err != nil {
		t.Fatalf("Exchange() error = %v", err)
	}
	got := <-captured
	for _, header := range headers {
		if value := got.header.Get(header.Key); value != header.Value {
			t.Errorf("header %s = %q, want %q", header.Key, value, header.Value)
		}
	}
}

// R-0VHL-ILYY
func TestExchangeRejectsReservedTokenParametersBeforeRequest(t *testing.T) {
	client, captured := exchangeServer(t, http.StatusOK, []byte("ok"))
	for _, key := range []string{"grant_type", "code", "code_verifier", "redirect_uri", "client_id", "client_secret"} {
		t.Run(key, func(t *testing.T) {
			payload, err := client.Exchange(context.Background(), testSession(), "code", []Param{{Key: key, Value: "override"}}, nil)
			if err == nil || len(payload) != 0 {
				t.Fatalf("Exchange() = %q, %v; want zero bytes and error", payload, err)
			}
			if !strings.Contains(err.Error(), key) {
				t.Fatalf("error %q does not name reserved key %q", err, key)
			}
			select {
			case request := <-captured:
				t.Fatalf("server received request for reserved key %q: %#v", key, request)
			default:
			}
		})
	}
}

// R-0WPH-WDPN
func TestExchangeReturnsSuccessfulBodyByteForByte(t *testing.T) {
	want := []byte("{\n  \"z\": 1,  \"a\" : [ 2,3 ]\n}\n")
	client, _ := exchangeServer(t, http.StatusOK, want)
	got, err := client.Exchange(context.Background(), testSession(), "code", nil, nil)
	if err != nil {
		t.Fatalf("Exchange() error = %v", err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("body = %q, want byte-for-byte %q", got, want)
	}
}

// R-0XXE-A5GC
func TestExchangeReturnsNon2xxStatusAndDescriptionWithoutPayload(t *testing.T) {
	body := []byte(`{"error":"invalid_grant","error_description":"the code expired"}`)
	client, _ := exchangeServer(t, http.StatusBadRequest, body)
	got, err := client.Exchange(context.Background(), testSession(), "code", nil, nil)
	if err == nil || len(got) != 0 {
		t.Fatalf("Exchange() = %q, %v; want zero bytes and error", got, err)
	}
	if !strings.Contains(err.Error(), "400 Bad Request") || !strings.Contains(err.Error(), "the code expired") {
		t.Fatalf("error = %q, want status and provider description", err)
	}
}

// R-0Z5A-NX71
func TestExchangeReturnsNonJSONSuccessUnchanged(t *testing.T) {
	want := []byte("definitely not JSON\x00\n")
	client, _ := exchangeServer(t, http.StatusCreated, want)
	got, err := client.Exchange(context.Background(), testSession(), "code", nil, nil)
	if err != nil {
		t.Fatalf("Exchange() error = %v", err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("body = %q, want unchanged %q", got, want)
	}
}

// R-10D7-1OXQ
func TestExchangeTransportFailureReturnsNoPayload(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		hijacker, ok := writer.(http.Hijacker)
		if !ok {
			t.Error("response writer does not support hijacking")
			return
		}
		connection, _, err := hijacker.Hijack()
		if err != nil {
			t.Errorf("Hijack() error = %v", err)
			return
		}
		connection.Close()
	}))
	defer server.Close()
	client := testClient()
	client.TokenURL = server.URL

	got, err := client.Exchange(context.Background(), testSession(), "code", nil, nil)
	if err == nil || len(got) != 0 {
		t.Fatalf("Exchange() = %q, %v; want zero bytes and transport error", got, err)
	}
}

func mustAuthorizeURL(t *testing.T, client *Client, session *Session, extra []Param) *url.URL {
	t.Helper()
	raw, err := client.AuthorizeURL(session, extra)
	if err != nil {
		t.Fatalf("AuthorizeURL() error = %v", err)
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		t.Fatalf("parse AuthorizeURL() result: %v", err)
	}
	return parsed
}

type capturedRequest struct {
	method      string
	contentType string
	form        url.Values
	header      http.Header
}

func exchangeServer(t *testing.T, status int, body []byte) (*Client, <-chan capturedRequest) {
	t.Helper()
	captured := make(chan capturedRequest, 10)
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		requestBody, err := io.ReadAll(request.Body)
		if err != nil {
			t.Errorf("read request body: %v", err)
		}
		form, err := url.ParseQuery(string(requestBody))
		if err != nil {
			t.Errorf("parse request form: %v", err)
		}
		captured <- capturedRequest{
			method:      request.Method,
			contentType: request.Header.Get("Content-Type"),
			form:        form,
			header:      request.Header.Clone(),
		}
		writer.WriteHeader(status)
		if _, err := writer.Write(body); err != nil {
			t.Errorf("write response: %v", err)
		}
	}))
	t.Cleanup(server.Close)
	client := testClient()
	client.TokenURL = server.URL
	return client, captured
}
