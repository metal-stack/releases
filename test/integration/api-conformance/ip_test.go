package conformance

import (
	"conformance/machinery"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/metal-stack/api/go/enum"
	adminv2 "github.com/metal-stack/api/go/metalstack/admin/v2"
	apiv2 "github.com/metal-stack/api/go/metalstack/api/v2"

	"github.com/metal-stack/api/go/metalstack/api/v2/apiv2connect"
	"github.com/metal-stack/metal-go/api/client/ip"
	"github.com/metal-stack/metal-go/api/models"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/durationpb"
)

const (
	sampleProject   = "00000000-0000-0000-0000-000000000001"
	internetMinilab = "internet-mini-lab"
)

func TestIP(t *testing.T) {
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	v1Client := machinery.GetV1Client(t)
	v2Client, tenantId := machinery.GetV2Client(t, log, &adminv2.TokenServiceCreateRequest{
		User: new("ip-conformance-tests-admin"),
		TokenCreateRequest: &apiv2.TokenServiceCreateRequest{
			Description: "ip-conformance-tests-user",
			Expires:     durationpb.New(10 * time.Minute),
			Permissions: []*apiv2.MethodPermission{
				{
					Subject: "*",
					Methods: []string{
						apiv2connect.IPServiceGetProcedure,
						apiv2connect.IPServiceCreateProcedure,
						apiv2connect.IPServiceDeleteProcedure,
						apiv2connect.ProjectServiceCreateProcedure,
						apiv2connect.VersionServiceGetProcedure,
					},
				},
			},
		},
	})

	p1, err := v2Client.Apiv2().Project().Create(t.Context(), &apiv2.ProjectServiceCreateRequest{Login: tenantId, Name: "p1"})
	require.NoError(t, err)
	var ipstring string

	deleteIPFn := func() {
		if ipstring == "" {
			return
		}
		_, err := v1Client.IP().FreeIP(&ip.FreeIPParams{ID: ipstring}, nil)
		require.NoError(t, err)
	}

	defer func() {
		deleteIPFn()
	}()

	v1ipresp, err := v1Client.IP().AllocateIP(&ip.AllocateIPParams{
		Body: &models.V1IPAllocateRequest{
			Description: "ip-conformance",
			Networkid:   new(internetMinilab),
			Projectid:   &p1.Project.Uuid,
		},
	}, nil)
	require.NoError(t, err)
	require.NotNil(t, v1ipresp)

	v1ip := v1ipresp.Payload
	ipstring = *v1ip.Ipaddress

	v2ipresp, err := v2Client.Apiv2().IP().Get(t.Context(), &apiv2.IPServiceGetRequest{Ip: ipstring, Project: p1.Project.Uuid})
	require.NoError(t, err)
	v2ip := v2ipresp.Ip

	v2ipType, err := enum.GetStringValue(v2ip.Type)
	require.NoError(t, err)
	require.NotNil(t, v1ip.Allocationuuid)
	require.NotNil(t, v1ip.Networkid)
	require.NotNil(t, v1ip.Projectid)
	require.NotNil(t, v1ip.Type)
	require.Equal(t, ipstring, v2ip.Ip)
	require.Equal(t, *v1ip.Allocationuuid, v2ip.Uuid)
	require.Equal(t, v1ip.Description, v2ip.Description)
	require.Equal(t, *v1ip.Networkid, v2ip.Network)
	require.Equal(t, *v1ip.Projectid, v2ip.Project)
	require.Equal(t, *v1ip.Type, *v2ipType)

	deleteIPFn()

	v2ipresp2, err := v2Client.Apiv2().IP().Create(t.Context(), &apiv2.IPServiceCreateRequest{
		Description: new("ip-conformance"),
		Network:     internetMinilab,
		Project:     p1.Project.Uuid,
	})
	require.NoError(t, err)

	apiv1resp2, err := v1Client.IP().FindIP(&ip.FindIPParams{ID: v2ipresp2.Ip.Ip}, nil)
	v2ip = v2ipresp2.Ip
	ipstring = v2ip.Ip
	v1ip = apiv1resp2.Payload
	require.NoError(t, err)
	require.NotNil(t, v1ip.Allocationuuid)
	require.NotNil(t, v1ip.Networkid)
	require.NotNil(t, v1ip.Projectid)
	require.NotNil(t, v1ip.Type)
	require.NotNil(t, v1ip.Ipaddress)
	require.Equal(t, ipstring, *v1ip.Ipaddress)
	require.Equal(t, *v1ip.Allocationuuid, v2ip.Uuid)
	require.Equal(t, v1ip.Description, v2ip.Description)
	require.Equal(t, *v1ip.Networkid, v2ip.Network)
	require.Equal(t, *v1ip.Projectid, v2ip.Project)
	require.Equal(t, *v1ip.Type, *v2ipType)
}
