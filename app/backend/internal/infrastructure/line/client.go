// Package line calls LINE Login v2.1 (OpenID Connect with PKCE, ADR-0007).
package line

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/way-station/tasuki/app/backend/internal/application/auth"
)

const (
	authorizeEndpoint = "https://access.line.me/oauth2/v2.1/authorize"
	// exchangeEndpoint is what LINE calls the "issue access token" endpoint.
	exchangeEndpoint = "https://api.line.me/oauth2/v2.1/token"
	// issuer is the iss claim of every ID token LINE issues.
	issuer = "https://access.line.me"

	requestTimeout = 10 * time.Second
	// maxResponseBytes bounds how much of the token endpoint's response is read.
	maxResponseBytes = 1 << 20
)

// Client implements auth.LineLogin for the web login of one LINE Login channel.
type Client struct {
	channelID     string
	channelSecret string
	redirectURI   string

	authorizeEndpoint string
	exchangeEndpoint  string
	httpClient        *http.Client
}

// NewClient returns a Client for the channel. redirectURI is the callback URL registered
// in the LINE Developers Console; LINE checks that both requests send the same one.
func NewClient(channelID, channelSecret, redirectURI string) *Client {
	return &Client{
		channelID:         channelID,
		channelSecret:     channelSecret,
		redirectURI:       redirectURI,
		authorizeEndpoint: authorizeEndpoint,
		exchangeEndpoint:  exchangeEndpoint,
		httpClient:        &http.Client{Timeout: requestTimeout},
	}
}

// AuthorizationURL returns LINE's login page asking for the openid and profile scopes.
// The PKCE code challenge is the SHA-256 of codeVerifier in base64url (RFC 7636, S256).
func (client *Client) AuthorizationURL(state, nonce, codeVerifier string) string {
	challenge := sha256.Sum256([]byte(codeVerifier))
	query := url.Values{
		"response_type":         {"code"},
		"client_id":             {client.channelID},
		"redirect_uri":          {client.redirectURI},
		"state":                 {state},
		"scope":                 {"openid profile"},
		"nonce":                 {nonce},
		"code_challenge":        {base64.RawURLEncoding.EncodeToString(challenge[:])},
		"code_challenge_method": {"S256"},
	}
	// Encode writes a space as "+"; LINE's documentation writes the scope with "%20".
	// A literal "+" is already escaped as "%2B", so only spaces are replaced.
	return client.authorizeEndpoint + "?" + strings.ReplaceAll(query.Encode(), "+", "%20")
}

// Authenticate exchanges the code at the token endpoint and verifies the ID token in the
// response.
func (client *Client) Authenticate(ctx context.Context, code, codeVerifier, nonce string) (auth.LineIdentity, error) {
	idToken, err := client.exchange(ctx, code, codeVerifier)
	if err != nil {
		return auth.LineIdentity{}, err
	}
	return client.verify(idToken, nonce)
}

func (client *Client) exchange(ctx context.Context, code, codeVerifier string) (string, error) {
	form := url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {code},
		"redirect_uri":  {client.redirectURI},
		"client_id":     {client.channelID},
		"client_secret": {client.channelSecret},
		"code_verifier": {codeVerifier},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, client.exchangeEndpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return "", fmt.Errorf("line: build token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("line: token request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("line: token endpoint returned %d", resp.StatusCode)
	}

	var body struct {
		IDToken string `json:"id_token"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, maxResponseBytes)).Decode(&body); err != nil {
		return "", fmt.Errorf("line: decode token response: %w", err)
	}
	if body.IDToken == "" {
		return "", errors.New("line: token response has no id_token")
	}
	return body.IDToken, nil
}

// idTokenClaims are the claims tasuki reads from LINE's ID token.
type idTokenClaims struct {
	jwt.RegisteredClaims
	Nonce string `json:"nonce"`
	// Name is the display name, present because the profile scope is requested.
	Name string `json:"name"`
}

// verify checks the ID token as OpenID Connect Core 1.0 section 3.1.3.7 requires. Web
// login signs it with HS256 and the channel secret; any other alg is refused, so a token
// with alg "none" or an alg chosen by an attacker never passes (RFC 8725 section 3.1).
func (client *Client) verify(idToken, nonce string) (auth.LineIdentity, error) {
	var claims idTokenClaims
	_, err := jwt.ParseWithClaims(idToken, &claims,
		func(*jwt.Token) (any, error) { return []byte(client.channelSecret), nil },
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithIssuer(issuer),
		jwt.WithAudience(client.channelID),
		jwt.WithExpirationRequired(),
	)
	if err != nil {
		return auth.LineIdentity{}, fmt.Errorf("line: verify ID token: %w", err)
	}
	// The nonce ties the token to the login this browser started, so a token taken from
	// another login cannot be replayed.
	if nonce == "" || claims.Nonce != nonce {
		return auth.LineIdentity{}, errors.New("line: ID token nonce does not match")
	}
	if claims.Subject == "" {
		return auth.LineIdentity{}, errors.New("line: ID token has no sub")
	}
	return auth.LineIdentity{UserID: claims.Subject, DisplayName: claims.Name}, nil
}
