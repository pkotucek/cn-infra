// Copyright (c) 2018 Cisco and/or its affiliates.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package bolt

import (
	"log"
	"github.com/ligato/cn-infra/db/keyval/kvproto"
	"github.com/ligato/cn-infra/flavors/local"
	"github.com/ligato/cn-infra/datasync/resync"
	"github.com/ligato/cn-infra/db/keyval"
	"github.com/boltdb/bolt"
)

// Config represents configuration for Bolt plugin.
type Config struct {
	DbPath         	string `json:"db-path"`
	BucketSeparator	string `json:"bucket-separator"`
}

// Plugin implements bolt plugin.
type Plugin struct {
	Deps

	// Plugin is disabled if there is no config file available
	disabled bool

	// Bolt DB encapsulation
	client *Client

	// Read/Write proto modelled data
	protoWrapper *kvproto.ProtoWrapper
}

// Deps lists dependencies of the etcd plugin.
// If injected, etcd plugin will use StatusCheck to signal the connection status.
type Deps struct {
	local.PluginInfraDeps
	Resync *resync.Plugin
}

// Disabled returns *true* if the plugin is not in use due to missing configuration.
func (plugin *Plugin) Disabled() bool {
	return plugin.disabled
}

func (plugin *Plugin) getConfig() (*Config, error) {
	var cfg Config
	found, err := plugin.PluginConfig.GetValue(&cfg)
	if err != nil {
		return nil, err
	}
	if !found {
		plugin.Log.Info("Bolt config not found, skip loading this plugin")
		plugin.disabled = true
		return nil, nil
	}
	return &cfg, nil
}

func (plugin *Plugin) Init() (err error) {
	cfg, err := plugin.getConfig()
	if err != nil || plugin.disabled {
		return err
	}

	plugin.client = &Client{}
	plugin.client.db_path, err = bolt.Open(cfg.DbPath, 432, nil)
	plugin.client.bucket_separator = cfg.BucketSeparator
	if err != nil {
		log.Fatal(err)
		plugin.disabled = true
		return  err
	}

	plugin.protoWrapper = kvproto.NewProtoWrapperWithSerializer(plugin.client, &keyval.SerializerJSON{})

	plugin.Log.Infof("Bolt DB started %v", cfg.DbPath)
	return nil
}

func (plugin *Plugin) Close() error {
	if !plugin.disabled {
		plugin.client.db_path.Close()
	}
	return nil
}

//// NewBroker creates new instance of prefixed broker that provides API with arguments of type proto.Message.
//func (plugin *Plugin) NewBroker(keyPrefix string) keyval.ProtoBroker {
//	return plugin.protoWrapper.NewBroker(keyPrefix)
//}

//// NewWatcher creates new instance of prefixed broker that provides API with arguments of type proto.Message.
//func (plugin *Plugin) NewWatcher(keyPrefix string) keyval.ProtoWatcher {
//	return plugin.protoWrapper.NewWatcher(keyPrefix)
//}