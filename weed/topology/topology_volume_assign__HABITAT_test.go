package topology

import (
	"testing"

	"github.com/seaweedfs/seaweedfs/weed/sequence"
	"github.com/seaweedfs/seaweedfs/weed/storage/super_block"
	"github.com/seaweedfs/seaweedfs/weed/storage/types"
)

func TestVolumeAssignmentWithNonExistentDC__HABITAT(t *testing.T) {
	topo := NewTopology("test", sequence.NewMemorySequencer(), 32*1024, 5, false)

	// Create a volume layout with a specific data center preference
	diskType := types.HardDriveType
	rp, _ := super_block.NewReplicaPlacementFromString("000")
	vl := topo.GetVolumeLayout("test", rp, nil, diskType)

	// Try to assign a volume with a non-existent data center
	option := &VolumeGrowOption{
		Collection:       "test",
		ReplicaPlacement: rp,
		DataCenter:       "non-existent-dc",
	}

	// This should return quickly with an error about the data center not existing
	_, _, _, _, err := topo.PickForWrite(1, option, vl)
	if err == nil {
		t.Fatal("expected an error when data center does not exist")
	}

	// The error should indicate the data center doesn't exist
	expectedErr := "no matching data node found"
	if err.Error() != expectedErr {
		t.Errorf("expected error '%s', got: %v", expectedErr, err)
	}
}
