package gauth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	calendar "google.golang.org/api/calendar/v3"
)

var (
	ErrNoCredentials = errors.New("no Google client credentials; run 'qarz setup'")
	ErrNoToken       = errors.New("not signed in to Google; run 'qarz setup'")
)

func HTTPClient(ctx context.Context, credentialsPath, tokenPath string) (*http.Client, error) {
	cfg, err := oauthConfig(credentialsPath)
	if err != nil {
		return nil, err
	}

	token, err := loadToken(tokenPath)
	if err != nil {
		return nil, err
	}

	return cfg.Client(ctx, token), nil
}

func Login(ctx context.Context, credentialsPath, tokenPath string, out io.Writer) error {
	cfg, err := oauthConfig(credentialsPath)
	if err != nil {
		return err
	}

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return fmt.Errorf("open a local port for the sign-in redirect: %w", err)
	}
	defer listener.Close()

	addr, ok := listener.Addr().(*net.TCPAddr)
	if !ok {
		return fmt.Errorf("unexpected local address %s for the sign-in redirect", listener.Addr())
	}
	cfg.RedirectURL = fmt.Sprintf("http://127.0.0.1:%d", addr.Port)

	state, err := randomState()
	if err != nil {
		return err
	}

	fmt.Fprintln(out, "Open this URL in your browser and approve access:")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "  "+cfg.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.ApprovalForce))
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Waiting for the redirect...")

	code, err := awaitCode(ctx, listener, state)
	if err != nil {
		return err
	}

	token, err := cfg.Exchange(ctx, code)
	if err != nil {
		return fmt.Errorf("exchange the sign-in code for a token: %w", err)
	}
	if token.RefreshToken == "" {
		return errors.New("Google returned no refresh token; revoke qarz's access in your Google account and run setup again")
	}

	return saveToken(tokenPath, token)
}

func HasToken(tokenPath string) bool {
	_, err := loadToken(tokenPath)
	return err == nil
}

func oauthConfig(credentialsPath string) (*oauth2.Config, error) {
	raw, err := os.ReadFile(credentialsPath)
	if os.IsNotExist(err) {
		return nil, ErrNoCredentials
	}
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", credentialsPath, err)
	}

	cfg, err := google.ConfigFromJSON(raw, calendar.CalendarScope)
	if err != nil {
		return nil, fmt.Errorf("read %s as a Google desktop client: %w", credentialsPath, err)
	}
	return cfg, nil
}

func awaitCode(ctx context.Context, listener net.Listener, state string) (string, error) {
	type result struct {
		code string
		err  error
	}
	results := make(chan result, 2)

	report := func(res result) {
		select {
		case results <- res:
		default:
		}
	}

	server := &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()

		if got := query.Get("state"); got != state {
			http.Error(w, "unexpected state", http.StatusBadRequest)
			report(result{err: errors.New("the sign-in redirect carried an unexpected state value")})
			return
		}
		if desc := query.Get("error"); desc != "" {
			http.Error(w, desc, http.StatusBadRequest)
			report(result{err: fmt.Errorf("Google refused the sign-in: %s", desc)})
			return
		}

		code := query.Get("code")
		if code == "" {
			http.Error(w, "no code", http.StatusBadRequest)
			report(result{err: errors.New("the sign-in redirect carried no code")})
			return
		}

		fmt.Fprintln(w, "qarz is signed in. You can close this tab.")
		report(result{code: code})
	})}

	go func() {
		if err := server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			report(result{err: fmt.Errorf("serve the sign-in redirect: %w", err)})
		}
	}()
	defer server.Close()

	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case res := <-results:
		return res.code, res.err
	}
}

func loadToken(path string) (*oauth2.Token, error) {
	raw, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil, ErrNoToken
	}
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}

	var token oauth2.Token
	if err := json.Unmarshal(raw, &token); err != nil {
		return nil, fmt.Errorf("read %s: %w (delete it and run 'qarz setup')", path, err)
	}
	if token.RefreshToken == "" && !token.Valid() {
		return nil, ErrNoToken
	}
	return &token, nil
}

func saveToken(path string, token *oauth2.Token) error {
	raw, err := json.MarshalIndent(token, "", "  ")
	if err != nil {
		return fmt.Errorf("encode the Google token: %w", err)
	}
	if err := os.WriteFile(path, raw, 0o600); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}

func randomState() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", fmt.Errorf("generate a sign-in state value: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b[:]), nil
}

func CredentialsUsable(path string) bool {
	_, err := oauthConfig(path)
	return err == nil
}
