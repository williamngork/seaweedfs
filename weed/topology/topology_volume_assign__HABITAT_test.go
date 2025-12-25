package topology

import (
	"strings"
	"testing"
	"time"

	"github.com/seaweedfs/seaweedfs/weed/sequence"
	"github.com/seaweedfs/seaweedfs/weed/storage"
	"github.com/seaweedfs/seaweedfs/weed/storage/needle"
	"github.com/seaweedfs/seaweedfs/weed/storage/super_block"
	"github.com/seaweedfs/seaweedfs/weed/storage/types"
)

func TestVolumeAssignmentWithNonExistentDC__HABITAT(t *testing.T) {
	topo := NewTopology("test", sequence.NewMemorySequencer(), 32*1024, 5, false)

	// Add a data node with volumes in a different DC
	dcName := "existing-dc"
	rackName := "rack1"
	serverAddress := "localhost:8080"

	// Create a data node with some volumes
	dn := NewDataNode(serverAddress)
	dn.NodeImpl.value = dn

	// Add a writable volume to the data node
	volumeInfo := storage.VolumeInfo{
		Id:               1,
		Size:             1000,
		Collection:       "test",
		ReplicaPlacement: &super_block.ReplicaPlacement{},
		Version:          needle.CurrentVersion,
		Ttl:              &needle.TTL{},
		DiskType:         types.HardDriveType.String(),
	}
	dn.AddOrUpdateVolume(volumeInfo)

	// Add the data node to the topology with an existing DC
	topo.DataNodeRegistration(dcName, rackName, dn)

	// Create a volume layout with a specific data center preference
	diskType := types.HardDriveType
	rp, _ := super_block.NewReplicaPlacementFromString("000")
	vl := topo.GetVolumeLayout("test", rp, nil, diskType)

	// Register the volume with the topology
	topo.RegisterVolumeLayout(volumeInfo, dn)

	// Try to assign a volume with a non-existent data center
	option := &VolumeGrowOption{
		Collection:       "test",
		ReplicaPlacement: rp,
		DataCenter:       "non-existent-dc",
	}

	// This should return quickly with an error about the data center not existing
	start := time.Now()
	_, _, _, _, err := topo.PickForWrite(1, option, vl)
	elapsed := time.Since(start)

	// The operation should fail fast, not wait for a timeout
	if elapsed > 1*time.Second {
		t.Errorf("operation should fail fast, took %v", elapsed)
	}

	// The error should indicate no matching data node was found
	if err == nil {
		t.Fatal("expected an error when data center does not exist")
	}

	// The error should contain the expected message
	expectedErr := "No writable volumes in DataCenter:non-existent-dc Rack: DataNode:"
	if !strings.Contains(err.Error(), expectedErr) {
		t.Errorf("expected error to contain '%s', got: %v", expectedErr, err)
	}
}
