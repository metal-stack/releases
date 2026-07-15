package machinery

import (
	"log/slog"
	"os"
	"testing"
	"time"

	apiv2client "github.com/metal-stack/api/go/client"
	adminv2 "github.com/metal-stack/api/go/metalstack/admin/v2"
	apiv2 "github.com/metal-stack/api/go/metalstack/api/v2"
	metalgo "github.com/metal-stack/metal-go"
	"github.com/metal-stack/metal-go/api/client/version"
	"google.golang.org/protobuf/types/known/durationpb"

	"github.com/stretchr/testify/require"
)

func GetV1Client(t *testing.T) metalgo.Client {
	var (
		metalURL   = os.Getenv("METALCTL_API_URL")
		metalHMAC  = os.Getenv("METALCTL_HMAC")
		metalToken = os.Getenv("METALCTL_TOKEN")
	)

	apiv1Client, err := metalgo.NewDriver(metalURL, metalToken, metalHMAC, metalgo.AuthType("Metal-Admin"))
	require.NoError(t, err)

	v, err := apiv1Client.Version().Info(&version.InfoParams{}, nil)
	require.NoError(t, err)
	t.Logf("connected to metal-api at:%s version:%s", metalURL, v.String())
	return apiv1Client
}

func GetV2Client(t *testing.T, log *slog.Logger, tokenRoles *adminv2.TokenServiceCreateRequest) (apiv2client.Client, string) {
	var (
		metalApiV2URL           = os.Getenv("METAL_APIV2_URL")
		adminTokenCreateRequest = &adminv2.TokenServiceCreateRequest{
			// User: new("api-conformance-tests"),
			TokenCreateRequest: &apiv2.TokenServiceCreateRequest{
				Description: "api-conformance-tests",
				Expires:     durationpb.New(time.Hour),
				AdminRole:   apiv2.AdminRole_ADMIN_ROLE_EDITOR.Enum(),
			},
		}
	)

	if tokenRoles == nil {
		tokenRoles = adminTokenCreateRequest
	}

	// FIXME fetch admin token from secret
	apiV2TokenAdmin, err := generateApiServerToken(t.Context(), adminTokenCreateRequest.TokenCreateRequest)
	require.NoError(t, err)

	apiv2ClientAdmin, err := apiv2client.New(&apiv2client.DialConfig{
		BaseURL: metalApiV2URL,
		Token:   apiV2TokenAdmin,
		Log:     log,
	})
	require.NoError(t, err)

	var tenant *apiv2.Tenant
	tenants, err := apiv2ClientAdmin.Adminv2().Tenant().List(t.Context(), &adminv2.TenantServiceListRequest{
		Query: &apiv2.TenantQuery{
			Login: tokenRoles.User,
		},
	})
	require.NoError(t, err)
	if len(tenants.Tenants) == 0 {
		resp, err := apiv2ClientAdmin.Adminv2().Tenant().Create(t.Context(), &adminv2.TenantServiceCreateRequest{Name: *tokenRoles.User})
		require.NoError(t, err)
		require.NotNil(t, resp)
		tenant = resp.Tenant
	} else {
		tenant = tenants.Tenants[0]
	}
	require.NotNil(t, tenant)

	tokenRoles.User = &tenant.Login

	userToken, err := apiv2ClientAdmin.Adminv2().Token().Create(t.Context(), tokenRoles)
	require.NoError(t, err)

	apiv2Client, err := apiv2client.New(&apiv2client.DialConfig{
		BaseURL: metalApiV2URL,
		Token:   userToken.Secret,
		Log:     log,
	})
	require.NoError(t, err)

	v2, err := apiv2Client.Apiv2().Version().Get(t.Context(), &apiv2.VersionServiceGetRequest{})
	require.NoError(t, err)

	t.Logf("connected to metal-apiserver at:%s version:%s", metalApiV2URL, v2.Version)

	return apiv2Client, tenant.Login
}
