/*
Copyright 2016 The Kubernetes Authors All rights reserved.

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

package localkube

import (
	"time"

	"github.com/coreos/etcd/embed"
	"github.com/golang/glog"
	"fmt"
)

const (
	// EtcdName is the name of the extra-config component for etcd
	EtcdName                        = "etcd"
	DefaultListenPeerURLs           = "http://0.0.0.0:2380"
	DefaultListenClientURLs         = "http://0.0.0.0:2379"
	DefaultInitialAdvertisePeerURLs = "http://127.0.0.1:2380"
	DefaultAdvertiseClientURLs      = "http://%s:2379"
)

// EtcdServer is a Server which manages an Etcd cluster
type EtcdServer struct {
	Etcd   *embed.Etcd
	Config *embed.Config
}

// NewEtcd creates a new default etcd Server using 'dataDir' for persistence. Panics if could not be configured.
func (lk LocalkubeServer) NewEtcd(ip string, dataDir string) (*EtcdServer, error) {
	cfg := embed.NewConfig()
	cfg.Dir = dataDir

	lpurl, _ := url.Parse(DefaultListenPeerURLs)
	apurl, _ := url.Parse(DefaultInitialAdvertisePeerURLs)
	lcurl, _ := url.Parse(DefaultListenClientURLs)
	acurl, _ := url.Parse(fmt.Sprintf(DefaultAdvertiseClientURLs, ip))

	cfg.LPUrls = []url.URL{*lpurl}
	cfg.APUrls = []url.URL{*apurl}
	cfg.LCUrls = []url.URL{*lcurl}
	cfg.ACUrls = []url.URL{*acurl}

	lk.SetExtraConfigForComponent(EtcdName, &cfg)
	return &EtcdServer{
		Config: cfg,
	}, nil
}

// Start starts the etcd server and listening for client connections
func (e *EtcdServer) Start() {
	var err error
	e.Etcd, err = embed.StartEtcd(e.Config)
	if err != nil {
		glog.Fatalf("Error starting up etcd: %s", err)
	}

	select {
	case <-e.Etcd.Server.ReadyNotify():
		glog.Infoln("Etcd server is ready")
	case <-time.After(60 * time.Second):
		e.Etcd.Server.Stop() // trigger a shutdown
		glog.Fatalf("Etcd took too long to start")
	}
}

// Stop closes all connections and stops the Etcd server
func (e *EtcdServer) Stop() {
	if e.Etcd != nil {
		e.Etcd.Server.Stop()
	}
}

// Name returns the servers unique name
func (e EtcdServer) Name() string {
	return e.Config.Name
}
