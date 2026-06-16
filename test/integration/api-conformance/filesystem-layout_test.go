package conformance

import (
	"conformance/machinery"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	adminv2 "github.com/metal-stack/api/go/metalstack/admin/v2"
	"github.com/metal-stack/api/go/metalstack/admin/v2/adminv2connect"
	apiv2 "github.com/metal-stack/api/go/metalstack/api/v2"
	"github.com/metal-stack/api/go/metalstack/api/v2/apiv2connect"
	"github.com/metal-stack/metal-go/api/client/filesystemlayout"
	"github.com/metal-stack/metal-go/api/models"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/durationpb"
)

var (
	apiv2fl1 = &apiv2.FilesystemLayout{
		Meta: &apiv2.Meta{
			// Labels: &apiv2.Labels{
			// 	Labels: map[string]string{
			// 		"purpose": "integration-test",
			// 	},
			// },
		},
		Id:          "test-filesystem-layout",
		Name:        new("Test FilesystemLayout"),
		Description: new("Test FilesystemLayout Description"),
		Filesystems: []*apiv2.Filesystem{
			{
				Device:        "/dev/sda1",
				Format:        apiv2.Format_FORMAT_EXT4,
				Name:          new("rootfs"),
				Description:   new("Root filesystem"),
				Path:          new("/"),
				Label:         new("root"),
				MountOptions:  []string{"defaults", "noatime"},
				CreateOptions: []string{"-L", "root"},
			},
		},
		Disks: []*apiv2.Disk{
			{
				Device: "/dev/sda",
				Partitions: []*apiv2.DiskPartition{
					{
						Number:  1,
						Label:   new("boot"),
						Size:    512,
						GptType: apiv2.GPTType_GPT_TYPE_BOOT.Enum(),
					},
					{
						Number:  2,
						Size:    102400,
						GptType: apiv2.GPTType_GPT_TYPE_LINUX.Enum(),
					},
				},
			},
			{
				Device:     "/dev/sdb",
				Partitions: []*apiv2.DiskPartition{},
			},
			{
				Device:     "/dev/sdc",
				Partitions: []*apiv2.DiskPartition{},
			},
		},
		Raid: []*apiv2.Raid{
			{
				ArrayName:     "/dev/md0",
				Devices:       []string{"/dev/sdb", "/dev/sdc"},
				Level:         apiv2.RaidLevel_RAID_LEVEL_1,
				CreateOptions: []string{"--metadata=1.0"},
				Spares:        1,
			},
		},
		VolumeGroups: []*apiv2.VolumeGroup{
			{
				Name:    "vg0",
				Devices: []string{"/dev/md0"},
				Tags:    []string{"metal-stack"},
			},
		},
		LogicalVolumes: []*apiv2.LogicalVolume{
			{
				Name:        "lv_root",
				VolumeGroup: "vg0",
				Size:        51200,
				LvmType:     apiv2.LVMType_LVM_TYPE_LINEAR,
			},
			{
				Name:        "lv_swap",
				VolumeGroup: "vg0",
				Size:        4096,
				LvmType:     apiv2.LVMType_LVM_TYPE_LINEAR,
			},
		},
		Constraints: &apiv2.FilesystemLayoutConstraints{
			Sizes: []string{"test-size"},
			Images: map[string]string{
				"debian": ">= 13",
			},
		},
	}

	apiv1fl1 = &models.V1FilesystemLayoutResponse{
		ID:          &apiv2fl1.Id,
		Name:        *apiv2fl1.Name,
		Description: *apiv2fl1.Description,
		Constraints: &models.V1FilesystemLayoutConstraints{
			Sizes:  apiv2fl1.Constraints.Sizes,
			Images: apiv2fl1.Constraints.Images,
		},
		Filesystems: []*models.V1Filesystem{
			{
				Device:        new("/dev/sda1"),
				Format:        new("ext4"),
				Label:         "root",
				Path:          "/",
				Mountoptions:  []string{"defaults", "noatime"},
				Createoptions: []string{"-L", "root"},
			},
		},
		Disks: []*models.V1Disk{
			{
				Device: new("/dev/sda"),
				Partitions: []*models.V1DiskPartition{
					{
						Number:  new(int64(1)),
						Label:   "boot",
						Size:    new(int64(512)),
						Gpttype: new("ef00"),
					},
					{
						Number:  new(int64(2)),
						Size:    new(int64(102400)),
						Gpttype: new("8300"),
					},
				},
				Wipeonreinstall: new(bool),
			},
			{
				Device:          new("/dev/sdb"),
				Wipeonreinstall: new(bool),
			},
			{
				Device:          new("/dev/sdc"),
				Wipeonreinstall: new(bool),
			},
		},
		Raid: []*models.V1Raid{
			{
				Arrayname:     new("/dev/md0"),
				Devices:       []string{"/dev/sdb", "/dev/sdc"},
				Level:         new("1"),
				Createoptions: []string{"--metadata=1.0"},
				Spares:        new(int32(1)),
			},
		},
		Volumegroups: []*models.V1VolumeGroup{
			{
				Name:    new("vg0"),
				Devices: []string{"/dev/md0"},
				Tags:    []string{"metal-stack"},
			},
		},
		Logicalvolumes: []*models.V1LogicalVolume{
			{
				Name:        new("lv_root"),
				Volumegroup: new("vg0"),
				Size:        new(int64(51200)),
				Lvmtype:     new("linear"),
			},
			{
				Name:        new("lv_swap"),
				Volumegroup: new("vg0"),
				Size:        new(int64(4096)),
				Lvmtype:     new("linear"),
			},
		},
	}
)

func TestFilesystemLayout(t *testing.T) {
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	v1Client := machinery.GetV1Client(t)
	v2Client, _ := machinery.GetV2Client(t, log, &adminv2.TokenServiceCreateRequest{
		User: new("fsl-conformance-test"),
		TokenCreateRequest: &apiv2.TokenServiceCreateRequest{
			Description: "filesystem-layout-conformance-tests",
			Expires:     durationpb.New(10 * time.Minute),
			Permissions: []*apiv2.MethodPermission{
				{
					Subject: "*",
					Methods: []string{
						apiv2connect.FilesystemServiceGetProcedure,
						apiv2connect.VersionServiceGetProcedure,
						apiv2connect.TokenServiceListProcedure,
						adminv2connect.FilesystemServiceCreateProcedure,
						adminv2connect.FilesystemServiceDeleteProcedure,
					},
				},
			},
		},
	})

	deleteFilesystemLayoutFn := func() {
		_, err := v1Client.Filesystemlayout().DeleteFilesystemLayout(&filesystemlayout.DeleteFilesystemLayoutParams{ID: apiv2fl1.Id}, nil)
		require.NoError(t, err)
	}

	defer func() {
		deleteFilesystemLayoutFn()
	}()

	v1fl, err := v1Client.Filesystemlayout().CreateFilesystemLayout(&filesystemlayout.CreateFilesystemLayoutParams{
		Body: &models.V1FilesystemLayoutCreateRequest{
			ID:             &apiv2fl1.Id,
			Name:           *apiv2fl1.Name,
			Description:    *apiv2fl1.Description,
			Filesystems:    apiv1fl1.Filesystems,
			Disks:          apiv1fl1.Disks,
			Raid:           apiv1fl1.Raid,
			Volumegroups:   apiv1fl1.Volumegroups,
			Logicalvolumes: apiv1fl1.Logicalvolumes,
			Constraints:    apiv1fl1.Constraints,
		},
	}, nil)
	require.NoError(t, err)
	require.NotNil(t, v1fl)

	resp, err := v2Client.Apiv2().Filesystem().Get(t.Context(), &apiv2.FilesystemServiceGetRequest{Id: apiv2fl1.Id})
	require.NoError(t, err)

	if diff := cmp.Diff(
		apiv2fl1, resp.FilesystemLayout,
		protocmp.Transform(),
		protocmp.IgnoreFields(
			&apiv2.Filesystem{}, "description", "name", // These properties are not stored in the apiv1 model
		),
		protocmp.IgnoreFields(
			&apiv2.Meta{}, "created_at", "updated_at",
		),
	); diff != "" {
		t.Errorf("filesystemlayout create and get () diff: %s", diff)
	}

	deleteFilesystemLayoutFn()

	_, err = v2Client.Adminv2().Filesystem().Create(t.Context(), &adminv2.FilesystemServiceCreateRequest{FilesystemLayout: apiv2fl1})
	require.NoError(t, err)

	v1resp, err := v1Client.Filesystemlayout().GetFilesystemLayout(&filesystemlayout.GetFilesystemLayoutParams{ID: *apiv1fl1.ID}, nil)
	require.NoError(t, err)

	if diff := cmp.Diff(
		apiv1fl1, v1resp.Payload,
	); diff != "" {
		t.Errorf("filesystemlayout create and get () diff: %s", diff)
	}
}
