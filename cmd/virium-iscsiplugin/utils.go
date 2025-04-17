/*
Copyright 2021 The Kubernetes Authors.

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
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/kubernetes-csi/csi-lib-utils/protosanitizer"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	klog "k8s.io/klog/v2"
)

func NewDefaultIdentityServer(d *driver) *IdentityServer {
	return &IdentityServer{
		Driver: d,
	}
}

func NewControllerServer(d *driver) *ControllerServer {
	return &ControllerServer{
		Driver: d,
	}
}

func NewControllerServiceCapability(cap csi.ControllerServiceCapability_RPC_Type) *csi.ControllerServiceCapability {
	return &csi.ControllerServiceCapability{
		Type: &csi.ControllerServiceCapability_Rpc{
			Rpc: &csi.ControllerServiceCapability_RPC{
				Type: cap,
			},
		},
	}
}

func ParseEndpoint(ep string) (string, string, error) {
	if strings.HasPrefix(strings.ToLower(ep), "unix://") || strings.HasPrefix(strings.ToLower(ep), "tcp://") {
		s := strings.SplitN(ep, "://", 2)
		if s[1] != "" {
			return s[0], s[1], nil
		}
	}
	return "", "", fmt.Errorf("invalid endpoint: %v", ep)
}

func logGRPC(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	klog.V(3).Infof("GRPC call: %s", info.FullMethod)
	klog.V(5).Infof("GRPC request: %s", protosanitizer.StripSecrets(req))
	resp, err := handler(ctx, req)
	if err != nil {
		klog.Errorf("GRPC error: %v", err)
	} else {
		klog.V(5).Infof("GRPC response: %s", protosanitizer.StripSecrets(resp))
	}
	return resp, err
}

// isValidVolumeCapabilities validates the given VolumeCapability array is valid
func isValidVolumeCapabilities(volCaps []*csi.VolumeCapability) error {
	if len(volCaps) == 0 {
		return fmt.Errorf("volume capabilities missing in request")
	}
	/* for _, c := range volCaps {
		if c.GetMount() != nil {
			return fmt.Errorf("mount volume capability not supported")
		}
	} */
	return nil
}

func viriumHttpClient(method string, url string, jsonData []byte) ([]byte, error) {
	// Step 2: Make the HTTP POST request
	// Create custom HTTP client with timeout
	timeout := time.Duration(50 * time.Second)
	client := &http.Client{
		Timeout: timeout,
	}

	// Build the HTTP request manually
	httpReq, err := http.NewRequest(method, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %v", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	// Send the request
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to call API: %v", err)
	}
	defer resp.Body.Close()

	// Read all data into memory
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if method == "POST" {
		// We expect HTTP 201 response
		if resp.StatusCode != http.StatusCreated {
			return nil, fmt.Errorf("API error(%d): %s", resp.StatusCode, string(body))
		}
	} else if method == "DELETE" {
		// We expect HTTP 200 response
		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
			return nil, fmt.Errorf("API error(%d): %s", resp.StatusCode, string(body))
		}
	}

	return body, nil
}
