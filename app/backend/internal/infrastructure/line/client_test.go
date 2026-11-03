package line

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/way-station/tasuki/app/backend/internal/application/auth"
)

const (
	testChannelID     = "1234567890"
	testChannelSecret = "channel-secret"
	testRedirectURI   = "https://tasuki.example/api/v1/auth/line/callback"
	testNonce         = "nonce-1"
)

// The verifier and challenge from RFC 7636 Appendix B.
const (
	testCodeVerifier  = "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk"
	testCodeChallenge = "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM"
)

var _ auth.LineLogin = (*Client)(nil)

func TestAuthorizationURL(t *testing.T) {
	client := NewClient(testChannelID, testChannelSecret, testRedirectURI)

	authorizationURL, err := url.Parse(client.AuthorizationURL("state-1", testNonce, testCodeVerifier))
	if err != nil {
		t.Fatal(err)
	}

	if got := authorizationURL.Scheme + "://" + authorizationURL.Host + authorizationURL.Path; got != authorizeEndpoint {
		t.Errorf("endpoint = %q, want %q", got, authorizeEndpoint)
	}
	want := map[string]string{
		"response_type":         "code",
		"client_id":             testChannelID,
		"redirect_uri":          testRedirectURI,
		"state":                 "state-1",
		"scope":                 "openid profile",
		"nonce":                 testNonce,
		"code_challenge":        testCodeChallenge,
		"code_challenge_method": "S256",
	}
	query := authorizationURL.Query()
	for name, value := range want {
		if got := query.Get(name); got != value {
			t.Errorf("%s = %q, want %q", name, got, value)
		}
	}
	if len(query) != len(want) {
		t.Errorf("query = %v, want only %v", query, want)
	}
}

// validClaims are the claims of an ID token LINE would issue for this login.
func validClaims() jwt.MapClaims {
	return jwt.MapClaims{
		"iss":   issuer,
		"sub":   "U1234",
		"aud":   testChannelID,
		"exp":   time.Now().Add(time.Hour).Unix(),
		"iat":   time.Now().Unix(),
		"nonce": testNonce,
		"name":  "Taro",
	}
}

func sign(t *testing.T, method jwt.SigningMethod, key any, claims jwt.MapClaims) string {
	t.Helper()
	token, err := jwt.NewWithClaims(method, claims).SignedString(key)
	if err != nil {
		t.Fatal(err)
	}
	return token
}

// tokenServer stands in for LINE's token endpoint, answering with status and body.
func tokenServer(t *testing.T, status int, body string) (*Client, *url.Values) {
	t.Helper()
	received := &url.Values{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost || req.Header.Get("Content-Type") != "application/x-www-form-urlencoded" {
			t.Errorf("request = %s %s, want a form POST", req.Method, req.Header.Get("Content-Type"))
		}
		if err := req.ParseForm(); err != nil {
			t.Error(err)
		}
		*received = req.PostForm
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(server.Close)

	client := NewClient(testChannelID, testChannelSecret, testRedirectURI)
	client.exchangeEndpoint = server.URL
	return client, received
}

func idTokenResponse(t *testing.T, idToken string) string {
	t.Helper()
	body, err := json.Marshal(map[string]any{"access_token": "access", "token_type": "Bearer", "id_token": idToken})
	if err != nil {
		t.Fatal(err)
	}
	return string(body)
}

func TestAuthenticate(t *testing.T) {
	t.Run("returns the LINE user of a valid ID token", func(t *testing.T) {
		idToken := sign(t, jwt.SigningMethodHS256, []byte(testChannelSecret), validClaims())
		client, received := tokenServer(t, http.StatusOK, idTokenResponse(t, idToken))

		identity, err := client.Authenticate(context.Background(), "code-1", testCodeVerifier, testNonce)
		if err != nil {
			t.Fatal(err)
		}

		if identity != (auth.LineIdentity{UserID: "U1234", DisplayName: "Taro"}) {
			t.Errorf("identity = %+v", identity)
		}
		want := url.Values{
			"grant_type":    {"authorization_code"},
			"code":          {"code-1"},
			"redirect_uri":  {testRedirectURI},
			"client_id":     {testChannelID},
			"client_secret": {testChannelSecret},
			"code_verifier": {testCodeVerifier},
		}
		for name := range want {
			if received.Get(name) != want.Get(name) {
				t.Errorf("%s = %q, want %q", name, received.Get(name), want.Get(name))
			}
		}
	})

	t.Run("fails when the token endpoint refuses the code", func(t *testing.T) {
		client, _ := tokenServer(t, http.StatusBadRequest, `{"error":"invalid_grant"}`)

		if _, err := client.Authenticate(context.Background(), "used", testCodeVerifier, testNonce); err == nil {
			t.Error("Authenticate() succeeded, want an error")
		}
	})

	t.Run("fails when the response has no ID token", func(t *testing.T) {
		client, _ := tokenServer(t, http.StatusOK, `{"access_token":"access"}`)

		if _, err := client.Authenticate(context.Background(), "code-1", testCodeVerifier, testNonce); err == nil {
			t.Error("Authenticate() succeeded, want an error")
		}
	})

	t.Run("fails when the response is not JSON", func(t *testing.T) {
		client, _ := tokenServer(t, http.StatusOK, "<html>")

		if _, err := client.Authenticate(context.Background(), "code-1", testCodeVerifier, testNonce); err == nil {
			t.Error("Authenticate() succeeded, want an error")
		}
	})
}

func TestAuthenticateRefusesInvalidIDTokens(t *testing.T) {
	withClaim := func(name string, value any) jwt.MapClaims {
		claims := validClaims()
		claims[name] = value
		return claims
	}
	withoutClaim := func(name string) jwt.MapClaims {
		claims := validClaims()
		delete(claims, name)
		return claims
	}
	secret := []byte(testChannelSecret)

	tests := []struct {
		name    string
		idToken func(t *testing.T) string
	}{
		{"signed with another secret", func(t *testing.T) string {
			return sign(t, jwt.SigningMethodHS256, []byte("another-secret"), validClaims())
		}},
		{"alg none", func(t *testing.T) string {
			return sign(t, jwt.SigningMethodNone, jwt.UnsafeAllowNoneSignatureType, validClaims())
		}},
		{"alg HS512 with the channel secret", func(t *testing.T) string {
			return sign(t, jwt.SigningMethodHS512, secret, validClaims())
		}},
		{"another issuer", func(t *testing.T) string {
			return sign(t, jwt.SigningMethodHS256, secret, withClaim("iss", "https://attacker.example"))
		}},
		{"no issuer", func(t *testing.T) string {
			return sign(t, jwt.SigningMethodHS256, secret, withoutClaim("iss"))
		}},
		{"another channel", func(t *testing.T) string {
			return sign(t, jwt.SigningMethodHS256, secret, withClaim("aud", "9999999999"))
		}},
		{"expired", func(t *testing.T) string {
			return sign(t, jwt.SigningMethodHS256, secret, withClaim("exp", time.Now().Add(-time.Minute).Unix()))
		}},
		{"no expiry", func(t *testing.T) string {
			return sign(t, jwt.SigningMethodHS256, secret, withoutClaim("exp"))
		}},
		{"another nonce", func(t *testing.T) string {
			return sign(t, jwt.SigningMethodHS256, secret, withClaim("nonce", "nonce-2"))
		}},
		{"no nonce", func(t *testing.T) string {
			return sign(t, jwt.SigningMethodHS256, secret, withoutClaim("nonce"))
		}},
		{"no subject", func(t *testing.T) string {
			return sign(t, jwt.SigningMethodHS256, secret, withoutClaim("sub"))
		}},
		{"not a JWT", func(*testing.T) string { return "not-a-jwt" }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client, _ := tokenServer(t, http.StatusOK, idTokenResponse(t, test.idToken(t)))

			identity, err := client.Authenticate(context.Background(), "code-1", testCodeVerifier, testNonce)
			if err == nil {
				t.Errorf("Authenticate() = %+v, want an error", identity)
			}
		})
	}
}
