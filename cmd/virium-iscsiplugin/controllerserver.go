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

package main

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

type VolumeRequest struct {
	InitiatorName string `json:"initiator_name"`
	Capacity      int64  `json:"capacity"`
}

type VolumeResponse struct {
	VolumeID     string `json:"id"`
	TargetPortal string `json:"targetPortal"`
	Iqn          string `json:"iqn"`
	Lun          string `json:"lun"`
}

type DeleteVolumeRequest struct {
	VolumeID string `json:"id"`
}

type VolumeResizeRequest struct {
	VolumeID string `json:"id"`
	Capacity int64  `json:"capacity"`
}

type SnapshotRequest struct {
	VolumeID string `json:"id"`
}

type DeleteSnapshotRequest struct {
	SnapshotID string `json:"id"`
}

func (cs *ControllerServer) CreateVolume(ctx context.Context, req *csi.CreateVolumeRequest) (*csi.CreateVolumeResponse, error) {
	fmt.Println("Creating Volume via REST API for:", req.Name)

	//
	// reqCapacity := req.GetCapacityRange().GetRequiredBytes()

	// XXX FIXME XXX
	/* if req.GetVolumeContentSource() != nil {
		if err := cs.copyVolume(ctx, req, nfsVol); err != nil {
			return nil, err
		}
	} */

	// Step 1: Prepare request payload
	apiURL := fmt.Sprintf("%s/api/volumes/create", cs.Driver.apiURL)
	payload := VolumeRequest{
		InitiatorName: cs.Driver.initiatorName,
		Capacity:      req.CapacityRange.RequiredBytes,
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

	// emptyList, _ := json.Marshal([]string{}) // gives: []byte(`[]`)
	// 	"portal":            string(emptyList), // portal: "[]"

	// Step 4: Return CSI-compatible volume response
	return &csi.CreateVolumeResponse{
		Volume: &csi.Volume{
			VolumeId:      volResp.VolumeID,
			CapacityBytes: req.CapacityRange.RequiredBytes,
			VolumeContext: map[string]string{
				"targetPortal":      volResp.TargetPortal,
				"iqn":               volResp.Iqn,
				"lun":               volResp.Lun,
				"iscsiInterface":    "default",
				"discoveryCHAPAuth": "true",
				"sessionCHAPAuth":   "false",
			},
		},
	}, nil

}

func (cs *ControllerServer) DeleteVolume(ctx context.Context, req *csi.DeleteVolumeRequest) (*csi.DeleteVolumeResponse, error) {
	volumeID := req.GetVolumeId()
	if volumeID == "" {
		return nil, fmt.Errorf("volume ID is required")
	}
	fmt.Println("Deleting Volume via REST API:", volumeID)

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

	fmt.Println("Volume successfully deleted:", volumeID)
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
	fmt.Println("Creating snapshot via REST API for:", req.Name)

	// Step 1: Prepare request payload
	apiURL := fmt.Sprintf("%s/api/snapshot/create", cs.Driver.apiURL)
	payload := SnapshotRequest{
		VolumeID: req.Name,
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

	// Step 4: Return CSI-compatible volume response
	return &csi.CreateSnapshotResponse{
		Snapshot: &csi.Snapshot{
			SnapshotId:     volResp.VolumeID,
			SourceVolumeId: req.Name,
			CreationTime:   timestamppb.Now(),
			ReadyToUse:     true,
		},
	}, nil
}

func (cs *ControllerServer) DeleteSnapshot(ctx context.Context, req *csi.DeleteSnapshotRequest) (*csi.DeleteSnapshotResponse, error) {
	snapshotID := req.GetSnapshotId()
	if snapshotID == "" {
		return nil, fmt.Errorf("snapshot ID is required")
	}
	fmt.Println("Deleting Volume via REST API:", snapshotID)

	// Step 1: Prepare request payload
	apiURL := fmt.Sprintf("%s/api/snapshot/delete", cs.Driver.apiURL)
	payload := DeleteSnapshotRequest{
		SnapshotID: snapshotID,
	}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	_, err = viriumHttpClient("DELETE", apiURL, jsonData)
	if err != nil {
		return nil, fmt.Errorf("API request failed: %v", err)
	}

	fmt.Println("Snapshot successfully deleted:", snapshotID)
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

	// klog.V(2).Infof("ControllerExpandVolume(%s) successfully, currentQuota: %d bytes", req.VolumeId, volSizeBytes)

	return &csi.ControllerExpandVolumeResponse{CapacityBytes: req.GetCapacityRange().GetRequiredBytes()}, nil
}

func (cs *ControllerServer) ControllerGetVolume(ctx context.Context, req *csi.ControllerGetVolumeRequest) (*csi.ControllerGetVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}
