package buildbuddy_enterprise

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"testing"

	"github.com/buildbuddy-io/buildbuddy/enterprise/server/testutil/testredis"
	"github.com/buildbuddy-io/buildbuddy/server/testutil/app"
	"github.com/buildbuddy-io/buildbuddy/server/testutil/testport"
	"github.com/buildbuddy-io/buildbuddy/server/util/proto"
	"github.com/stretchr/testify/require"

	ctxpb "github.com/buildbuddy-io/buildbuddy/proto/context"
	uspb "github.com/buildbuddy-io/buildbuddy/proto/user"
)

var (
	webdriverTarget = flag.String("webdriver_target", "local", "Target that should be tested by the webdriver. Should be 'local' or 'remote'. For remote targets, you must also set the --remote_app_endpoint and --remote_sso_slug flags.")

	remoteAppEndpoint = flag.String("remote_app_endpoint", "", "For remote targets, the endpoint to reach the app (Ex. https://app.buildbuddy.dev, or https://app.buildbuddy.io).")
	remoteSSOSlug     = flag.String("remote_sso_slug", "", "For remote targets, the SSO slug to login. Self-auth must be enabled for these targets.")
)

var (
	// set by x_defs in BUILD file
	DefaultConfig         string
	NoAuthConfig          string
	buildbuddyRunfilePath string
)

func Run(t *testing.T, args ...string) *app.App {
	return RunWithConfig(t, DefaultAppConfig(t), DefaultConfig, args...)
}

func RunWithConfig(t *testing.T, appConfig *app.App, configPath string, args ...string) *app.App {
	redisTarget := testredis.Start(t).Target
	commandArgs := []string{
		"--app_directory=/enterprise/app",
		"--app.default_redis_target=" + redisTarget,
		"--telemetry_port=-1",
	}
	commandArgs = append(commandArgs, args...)
	return app.RunWithApp(
		t,
		appConfig,
		/* commandPath= */ buildbuddyRunfilePath,
		commandArgs,
		/* configPath= */ configPath,
	)
}

func DefaultAppConfig(t *testing.T) *app.App {
	return &app.App{
		HttpPort:       testport.FindFree(t),
		GRPCPort:       testport.FindFree(t),
		MonitoringPort: testport.FindFree(t),
	}
}

// remote represents a handle on a remote BuildBuddy enterprise server.
type remote struct{}

func (d *remote) HTTPURL() string {
	return *remoteAppEndpoint
}

func (d *remote) SSOSlug() string {
	return *remoteSSOSlug
}

type WebTarget interface {
	HTTPURL() string
}

func SetupWebTarget(t *testing.T) WebTarget {
	switch *webdriverTarget {
	case "local":
		return Run(t, "--cache.detailed_stats_enabled=true", "--app.user_owned_keys_enabled=true")
	case "remote":
		return &remote{}
	default:
		require.FailNowf(t, "invalid target", "%s is an invalid target for the webdriver invocation tests.", *webdriverTarget)
		return nil
	}
}

// WebClient is a lightweight client for testing enterprise functionality that
// is only supported via the Web UI.
//
// In particular, it is needed for BuildBuddyService RPCs that require admin
// rights, because regular gRPCs currently only support API key authentication
// (or JWT authentication where the JWTs are derived from API keys), but API
// keys are currently assigned the default role (developer).
type WebClient struct {
	t *testing.T
	a *app.App

	// HTTPClient is the underlying HTTP client, which has the login cookies for
	// the user set in the cookie jar.
	HTTPClient *http.Client
	// RequestContext is used for protolet RPCs.
	RequestContext *ctxpb.RequestContext
}

// RPC makes a gRPC request over HTTP using the Web client's login cookies.
func (c *WebClient) RPC(method string, req proto.Message, res proto.Message) error {
	return rpcOverHTTP(c.t, c.a, c.HTTPClient, method, req, res)
}

// LoginAsDefaultSelfAuthUser logs in as the default self-auth user and returns
// a WebClient that can be used to make Web RPCs as that user.
func LoginAsDefaultSelfAuthUser(t *testing.T, a *app.App) *WebClient {
	jar, err := cookiejar.New(nil /*=options*/)
	require.NoError(t, err)
	client := &http.Client{Jar: jar}

	escapedAppURL := url.QueryEscape(a.HTTPURL())
	loginURL := fmt.Sprintf(
		"%s/login/?issuer_url=%s&redirect_url=%s",
		a.HTTPURL(), escapedAppURL, escapedAppURL)

	res, err := client.Get(loginURL)
	require.NoError(t, err)
	io.ReadAll(res.Body)
	res.Body.Close()

	u, err := url.Parse(a.HTTPURL())
	require.NoError(t, err)
	cookies := jar.Cookies(u)
	jwt := ""
	for _, c := range cookies {
		if c.Name == "Authorization" {
			jwt = c.Value
		}
	}
	require.NotEmpty(t, jwt, "missing 'Authorization' cookie. All cookies: %v", cookies)

	// Note: the client should have the Authorization cookie in the jar at this
	// point, so all subsequent rpcOverHTTP requests here will include the JWT.
	// We need to use rpcOverHTTP for the user bootstrap since we rely on the
	// user token which is only populated in the HTTP auth flow.
	uRes := &uspb.GetUserResponse{}
	if err := rpcOverHTTP(t, a, client, "GetUser", nil /*=req*/, uRes); err != nil {
		// Create user if one doesn't exist.
		require.Contains(t, err.Error(), "not found")

		err := rpcOverHTTP(t, a, client, "CreateUser", nil /*=req*/, nil /*=res*/)
		require.NoError(t, err)

		err = rpcOverHTTP(t, a, client, "GetUser", nil /*=req*/, uRes)
		require.NoError(t, err)
	}

	return &WebClient{
		t:          t,
		a:          a,
		HTTPClient: client,
		RequestContext: &ctxpb.RequestContext{
			UserId:  uRes.DisplayUser.UserId,
			GroupId: uRes.SelectedGroupId,
		},
	}
}

func rpcOverHTTP(t *testing.T, a *app.App, client *http.Client, method string, req proto.Message, res proto.Message) error {
	reqBytes := []byte{}
	if req != nil {
		var err error
		reqBytes, err = proto.Marshal(req)
		require.NoError(t, err)
	}
	httpReq, err := http.NewRequest("GET", fmt.Sprintf("%s/rpc/BuildBuddyService/%s", a.HTTPURL(), method), bytes.NewReader(reqBytes))
	require.NoError(t, err)
	httpReq.Header.Add("Content-Type", "application/proto")

	httpRes, err := client.Do(httpReq)
	require.NoError(t, err)

	defer httpRes.Body.Close()
	resBytes, err := io.ReadAll(httpRes.Body)
	require.NoError(t, err)

	if httpRes.StatusCode >= 400 {
		return errors.New(string(resBytes))
	}

	if res != nil {
		err = proto.Unmarshal(resBytes, res)
		require.NoError(t, err)
	}
	return nil
}
