/*
Copyright 2017 The Kubernetes Authors.

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
	"flag"
	"os"

	klog "k8s.io/klog/v2"
)

var (
	endpoint      = flag.String("endpoint", "unix:///csi/csi.sock", "CSI endpoint")
	apiURL        = flag.String("apiurl", "http://virium-isci-fqdn.domain.tld:8787", "Virium api url")
	initiatorName = flag.String("initiatorname", "iqn.2025-04.net.virer.virium:target1", "iSCSI initiator name identifier")
	api_username  = flag.String("api_username", "", "api_username")
	api_password  = flag.String("api_password", "", "api_password")
)

func main() {
	klog.InitFlags(nil)
	_ = flag.Set("logtostderr", "true")
	flag.Parse()
	handle()
	os.Exit(0)
}

func handle() {
	d := NewDriver(*endpoint, *apiURL, *initiatorName, *api_username, *api_password)
	d.Run()
}
