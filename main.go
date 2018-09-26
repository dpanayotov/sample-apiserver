/*
Copyright 2016 The Kubernetes Authors.

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
	"github.com/dpanayotov/sample-apiserver/apiserver"
	"github.com/dpanayotov/sample-apiserver/pkg/postgres"
	_ "github.com/lib/pq"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	genericapiserver "k8s.io/apiserver/pkg/server"
	genericoptions "k8s.io/apiserver/pkg/server/options"
	"k8s.io/apiserver/pkg/storage/storagebackend/factory"
)

const GroupName = "service.manager.io"
const defaultEtcdPathPrefix = "/registry/wardle.kubernetes.io"

var (
	Scheme = runtime.NewScheme()
	Codecs = serializer.NewCodecFactory(Scheme)

	SchemeGroupVersion = schema.GroupVersion{Group: GroupName, Version: runtime.APIVersionInternal}
)

func main() {
	factory.Register("postgres", postgres.NewPostgresStorage(), postgres.NewPostgresHealthCheck())
	postgres.Register("postgres", NewPostgreSQL())

	recommendedOptions := genericoptions.NewRecommendedOptions(defaultEtcdPathPrefix, apiserver.Codecs.LegacyCodec(SchemeGroupVersion))
	serverConfig := genericapiserver.NewRecommendedConfig(apiserver.Codecs)
	if err := recommendedOptions.ApplyTo(serverConfig, apiserver.Scheme); err != nil {
		panic(err)
	}
	complete := serverConfig.Complete()
	cmd := &cobra.Command{
		Short: "Launch a wardle API server",
		Long:  "Launch a wardle API server",
		RunE: func(c *cobra.Command, args []string) error {
			server, err := complete.New("sample-apiserver", genericapiserver.NewEmptyDelegate())
			if err != nil {
				return err
			}

			stopCh := make(chan struct{})
			if err := server.PrepareRun().Run(stopCh); err != nil {
				return err
			}
			return nil
		},
	}

	flags := cmd.Flags()
	recommendedOptions.AddFlags(flags)

	if err := cmd.Execute(); err != nil {
		panic(err)
	}
}

func NewPostgreSQL() *Generic {
	return &Generic{
		CleanupSQL: "delete from key_value where ttl > 0 and ttl < ?",
		GetSQL:     "select name, value, revision from key_value where name = ?",
		ListSQL:    "select name, value, revision from key_value where name like ?",
		CreateSQL:  "insert into key_value(name, value, revision, ttl) values(?, ?, 1, ?)",
		DeleteSQL:  "delete from key_value where name = ? and revision = ?",
		UpdateSQL:  "update key_value set value = ?, revision = ? where name = ? and revision = ?",
	}
}
