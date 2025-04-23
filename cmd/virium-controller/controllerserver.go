/*
Copyright 2019 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	klog "k8s.io/klog/v2"
)

type ControllerServer struct {
	Driver *driver
	csi.UnimplementedControllerServer
}

// Volume Request :
type VolumeRequest struct {
	InitiatorName string               `json:"initiator_name"`
	Capacity      int64                `json:"capacity"`
	ContentSource *VolumeContentSource `json:"content_source,omitempty"`
}
type VolumeContentSource struct {
	Type struct {
		Snapshot *SnapshotSource `json:"Snapshot,omitempty"`
		Volume   *VolumeSource   `json:"Volume,omitempty"`
	} `json:"Type"`
}
type SnapshotSource struct {
	SnapshotID string `json:"snapshot_id"`
}
type VolumeSource struct {
	VolumeID string `json:"volume_id"`
}

// Volume Request ^^

type VolumeResponse struct {
	VolumeID          string `json:"volume_id"`
	TargetPortal      string `json:"targetPortal"`
	Iqn               string `json:"iqn"`
	Lun               string `json:"lun"`
	DiscoveryCHAPAuth string `json:"discoveryCHAPAuth"`
	SessionCHAPAuth   string `json:"sessionCHAPAuth"`
}

type DeleteVolumeRequest struct {
	VolumeID string `json:"volume_id"`
}

type VolumeResizeRequest struct {
	VolumeID string `json:"volume_id"`
	Capacity int64  `json:"capacity"`
}

type SnapshotRequest struct {
	Name     string `json:"name"`
	VolumeID string `json:"source_volume_id"`
}

type SnapshotResponse struct {
	VolumeID string `json:"snapshot_id"`
	Capacity int64  `json:"capacity"`
}

type DeleteSnapshotRequest struct {
	SnapshotID string `json:"snapshot_id"`
}

func (cs *ControllerServer) CreateVolume(ctx context.Context, req *csi.CreateVolumeRequest) (*csi.CreateVolumeResponse, error) {
	klog.V(1).Info("Creating Volume via API for:", req.Name)

	// Step 1: Prepare request payload
	apiURL := fmt.Sprintf("%s/api/volumes/create", cs.Driver.apiURL)
	payload := VolumeRequest{
		InitiatorName: cs.Driver.initiatorName,
		Capacity:      req.CapacityRange.RequiredBytes,
	}
	src := req.VolumeContentSource
	if src != nil {
		klog.V(5).Info("Content source requested", src)
		switch src := req.VolumeContentSource.Type.(type) {
		case *csi.VolumeContentSource_Snapshot:
			payload.ContentSource = &VolumeContentSource{
				Type: struct {
					Snapshot *SnapshotSource `json:"Snapshot,omitempty"`
					Volume   *VolumeSource   `json:"Volume,omitempty"`
				}{
					Snapshot: &SnapshotSource{SnapshotID: src.Snapshot.SnapshotId},
				},
			}
		case *csi.VolumeContentSource_Volume:
			payload.ContentSource = &VolumeContentSource{
				Type: struct {
					Snapshot *SnapshotSource `json:"Snapshot,omitempty"`
					Volume   *VolumeSource   `json:"Volume,omitempty"`
				}{
					Volume: &VolumeSource{VolumeID: src.Volume.VolumeId},
				},
			}
		default:
			fmt.Println("ContentSource Unknown type:", src)
		}
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	resp, err := viriumHttpClient("POST", apiURL, jsonData)
	if err != nil {
		return nil, fmt.Errorf("API request failed: %v", err)
	}

	var volResp VolumeResponse
	if err := json.NewDecoder(bytes.NewReader(resp)).Decode(&volResp); err != nil {
		return nil, fmt.Errorf("failed to parse volume response: %v", err)
	}

	portals := []string{}
	portals = append(portals, volResp.TargetPortal)
	portalList, _ := json.Marshal(portals)

	klog.V(1).Info("Volume created successfully", req.Name)

	// Step 4: Return CSI-compatible volume response
	ret_value := &csi.CreateVolumeResponse{
		Volume: &csi.Volume{
			VolumeId:      volResp.VolumeID,
			CapacityBytes: req.CapacityRange.RequiredBytes,
			VolumeContext: map[string]string{
				"portals":           string(portalList), // portal: "[]"
				"targetPortal":      volResp.TargetPortal,
				"iqn":               volResp.Iqn,
				"lun":               volResp.Lun,
				"interface":         "default",
				"discoveryCHAPAuth": volResp.DiscoveryCHAPAuth,
				"sessionCHAPAuth":   volResp.SessionCHAPAuth,
			},
		},
	}
	if src != nil {
		ret_value.Volume.ContentSource = &csi.VolumeContentSource{
			Type: &csi.VolumeContentSource_Volume{
				Volume: &csi.VolumeContentSource_VolumeSource{
					VolumeId: src.GetVolume().GetVolumeId(),
				},
			},
		}
	}

	klog.V(1).Infof("Volume creation payload %+v\n", ret_value)
	return ret_value, nil

}

func (cs *ControllerServer) DeleteVolume(ctx context.Context, req *csi.DeleteVolumeRequest) (*csi.DeleteVolumeResponse, error) {
	volumeID := req.GetVolumeId()
	if volumeID == "" {
		return nil, fmt.Errorf("volume ID is required")
	}
	klog.V(1).Info("Deleting Volume via API:", volumeID)

	// Step 1: Prepare request payload
	apiURL := fmt.Sprintf("%s/api/volumes/delete", cs.Driver.apiURL)
	payload := DeleteVolumeRequest{
		VolumeID: volumeID,
	}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	_, err = viriumHttpClient("DELETE", apiURL, jsonData)
	if err != nil {
		return nil, fmt.Errorf("API request failed: %v", err)
	}

	klog.V(1).Info("Volume successfully deleted", volumeID)
	return &csi.DeleteVolumeResponse{}, nil
}

func (cs *ControllerServer) ControllerPublishVolume(ctx context.Context, req *csi.ControllerPublishVolumeRequest) (*csi.ControllerPublishVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

func (cs *ControllerServer) ControllerUnpublishVolume(ctx context.Context, req *csi.ControllerUnpublishVolumeRequest) (*csi.ControllerUnpublishVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

func (cs *ControllerServer) ValidateVolumeCapabilities(ctx context.Context, req *csi.ValidateVolumeCapabilitiesRequest) (*csi.ValidateVolumeCapabilitiesResponse, error) {
	if len(req.GetVolumeId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Volume ID missing in request")
	}
	if err := isValidVolumeCapabilities(req.GetVolumeCapabilities()); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &csi.ValidateVolumeCapabilitiesResponse{
		Confirmed: &csi.ValidateVolumeCapabilitiesResponse_Confirmed{
			VolumeCapabilities: req.GetVolumeCapabilities(),
		},
		Message: "",
	}, nil
}

func (cs *ControllerServer) ListVolumes(ctx context.Context, req *csi.ListVolumesRequest) (*csi.ListVolumesResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

func (cs *ControllerServer) GetCapacity(ctx context.Context, req *csi.GetCapacityRequest) (*csi.GetCapacityResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

// ControllerGetCapabilities implements the default GRPC callout.
// Default supports all capabilities.
func (cs *ControllerServer) ControllerGetCapabilities(ctx context.Context, req *csi.ControllerGetCapabilitiesRequest) (*csi.ControllerGetCapabilitiesResponse, error) {
	klog.V(5).Infof("Using default ControllerGetCapabilities")

	return &csi.ControllerGetCapabilitiesResponse{
		Capabilities: cs.Driver.cscap,
	}, nil
}

func (cs *ControllerServer) CreateSnapshot(ctx context.Context, req *csi.CreateSnapshotRequest) (*csi.CreateSnapshotResponse, error) {
	klog.V(1).Info("Creating snapshot via API for:", req.Name)

	// Step 1: Prepare request payload
	apiURL := fmt.Sprintf("%s/api/snapshot/create", cs.Driver.apiURL)
	payload := SnapshotRequest{
		VolumeID: req.SourceVolumeId,
		Name:     req.Name,
	}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	resp, err := viriumHttpClient("POST", apiURL, jsonData)
	if err != nil {
		return nil, fmt.Errorf("API request failed: %v", err)
	}

	var volResp SnapshotResponse
	if err := json.NewDecoder(bytes.NewReader(resp)).Decode(&volResp); err != nil {
		return nil, fmt.Errorf("failed to parse volume response: %v", err)
	}
	klog.V(1).Info("Snapshot created successfully, snapshotId:", volResp.VolumeID)
	// Step 4: Return CSI-compatible volume response
	return &csi.CreateSnapshotResponse{
		Snapshot: &csi.Snapshot{
			SnapshotId:     volResp.VolumeID,
			SourceVolumeId: req.Name,
			CreationTime:   timestamppb.Now(),
			ReadyToUse:     true,
			SizeBytes:      volResp.Capacity,
		},
	}, nil
}

func (cs *ControllerServer) DeleteSnapshot(ctx context.Context, req *csi.DeleteSnapshotRequest) (*csi.DeleteSnapshotResponse, error) {
	klog.V(5).Infof("Delete snap req: %+v", req)
	klog.V(5).Infof("Delete snap snapId %s", req.SnapshotId)
	klog.V(5).Infof("Delete snap Secret %+v", req.GetSecrets())
	if len(req.GetSnapshotId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Snapshot ID is required for deletion")
	}
	klog.V(1).Info("Deleting Volume via API:", req.SnapshotId)

	// Step 1: Prepare request payload
	apiURL := fmt.Sprintf("%s/api/snapshot/delete", cs.Driver.apiURL)
	payload := DeleteSnapshotRequest{
		SnapshotID: req.SnapshotId,
	}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	_, err = viriumHttpClient("DELETE", apiURL, jsonData)
	if err != nil {
		return nil, fmt.Errorf("API request failed: %v", err)
	}

	klog.V(1).Info("Snapshot successfully deleted:", req.SnapshotId)
	return &csi.DeleteSnapshotResponse{}, nil
}

func (cs *ControllerServer) ListSnapshots(ctx context.Context, req *csi.ListSnapshotsRequest) (*csi.ListSnapshotsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

func (cs *ControllerServer) ControllerExpandVolume(ctx context.Context, req *csi.ControllerExpandVolumeRequest) (*csi.ControllerExpandVolumeResponse, error) {
	if len(req.GetVolumeId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Volume ID missing in request")
	}

	if req.GetCapacityRange() == nil {
		return nil, status.Error(codes.InvalidArgument, "Capacity Range missing in request")
	}
	klog.V(1).Info("Expand Volume", req.GetVolumeId())
	volSizeBytes := int64(req.GetCapacityRange().GetRequiredBytes())
	// Step 1: Prepare request payload
	apiURL := fmt.Sprintf("%s/api/volumes/resize", cs.Driver.apiURL)
	payload := VolumeResizeRequest{
		VolumeID: req.GetVolumeId(),
		Capacity: volSizeBytes,
	}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	resp, err := viriumHttpClient("POST", apiURL, jsonData)
	if err != nil {
		return nil, fmt.Errorf("API request failed: %v", err)
	}

	var volResp VolumeResponse
	if err := json.NewDecoder(bytes.NewReader(resp)).Decode(&volResp); err != nil {
		return nil, fmt.Errorf("failed to parse volume response: %v", err)
	}

	klog.V(1).Infof("Expand Volume %s successfully, currentQuota: %d bytes", req.VolumeId, volSizeBytes)

	return &csi.ControllerExpandVolumeResponse{CapacityBytes: req.GetCapacityRange().GetRequiredBytes()}, nil
}

func (cs *ControllerServer) ControllerGetVolume(ctx context.Context, req *csi.ControllerGetVolumeRequest) (*csi.ControllerGetVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}
