// Copyright © 2018 Banzai Cloud
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ark

import (
	"github.com/spf13/viper"

	"github.com/banzaicloud/pipeline/internal/ark/providers/amazon"
	"github.com/banzaicloud/pipeline/internal/ark/providers/azure"
	"github.com/banzaicloud/pipeline/internal/ark/providers/google"
	pkgErrors "github.com/banzaicloud/pipeline/pkg/errors"
	"github.com/banzaicloud/pipeline/pkg/providers"
	"github.com/banzaicloud/pipeline/secret"
)

// ChartConfig describes an ARK deployment chart config
type ChartConfig struct {
	Namespace      string
	Chart          string
	Name           string
	Version        string
	ValueOverrides []byte
}

// ValueOverrides descibes values to be overridden in a deployment
type ValueOverrides struct {
	Configuration configuration `json:"configuration"`
	Credentials   credentials   `json:"credentials"`
	Image         image         `json:"image"`
	RBAC          rbac          `json:"rbac"`
}

type rbac struct {
	Create bool `json:"create"`
}

type image struct {
	Repository string `json:"repository"`
	Tag        string `json:"tag"`
	PullPolicy string `json:"pullPolicy"`
}

type credentials struct {
	SecretContents secretContents `json:"secretContents"`
}

type secretContents struct {
	azure.Secret
	Cluster string `json:"cluster,omitempty"`
	Bucket  string `json:"bucket,omitempty"`
}

type configuration struct {
	PersistentVolumeProvider persistentVolumeProvider `json:"persistentVolumeProvider"`
	BackupStorageProvider    backupStorageProvider    `json:"backupStorageProvider"`
	RestoreOnlyMode          bool
}

type persistentVolumeProvider struct {
	Name   string                         `json:"name"`
	Config persistentVolumeProviderConfig `json:"config,omitempty"`
}

type persistentVolumeProviderConfig struct {
	Region     string `json:"region,omitempty"`
	ApiTimeout string `json:"apiTimeout,omitempty"`
}

type backupStorageProvider struct {
	Name   string                      `json:"name"`
	Bucket string                      `json:"bucket"`
	Config backupStorageProviderConfig `json:"config,omitempty"`
}

type backupStorageProviderConfig struct {
	Region           string `json:"region,omitempty"`
	S3ForcePathStyle string `json:"s3ForcePathStyle,omitempty"`
	S3Url            string `json:"s3Url,omitempty"`
	KMSKeyId         string `json:"kmsKeyId,omitempty"`
}

// ConfigRequest describes an ARK config request
type ConfigRequest struct {
	Cloud       string
	CloudSecret *secret.SecretItemResponse
	Location    string

	Bucket       BucketConfig
	BucketSecret *secret.SecretItemResponse

	RBAC        bool
	RestoreMode bool
}

type BucketConfig struct {
	Name     string
	Provider string
	Location string

	AzureBucketConfig
}

type AzureBucketConfig struct {
	StorageAccount string
	ResourceGroup  string
}

// GetChartConfig get a ChartConfig
func GetChartConfig() ChartConfig {

	return ChartConfig{
		Name:      viper.GetString("ark.name"),
		Namespace: viper.GetString("ark.namespace"),
		Chart:     viper.GetString("ark.chart"),
		Version:   viper.GetString("ark.chart_version"),
	}
}

// Get gets helm deployment value overrides
func (req ConfigRequest) Get() (values ValueOverrides, err error) {

	pvp, err := req.getPVPConfig()
	if err != nil {
		return values, err
	}

	bsp, err := req.getBSPConfig()
	if err != nil {
		return values, err
	}

	cred, err := req.getCredentials()
	if err != nil {
		return values, err
	}

	return ValueOverrides{
		Configuration: configuration{
			PersistentVolumeProvider: pvp,
			BackupStorageProvider:    bsp,
			RestoreOnlyMode:          req.RestoreMode,
		},
		RBAC: rbac{
			Create: req.RBAC,
		},
		Credentials: cred,
		Image: image{
			Repository: viper.GetString("ark.image"),
			Tag:        viper.GetString("ark.imagetag"),
			PullPolicy: viper.GetString("ark.pullpolicy"),
		},
	}, nil
}

func (req ConfigRequest) getPVPConfig() (persistentVolumeProvider, error) {

	var config persistentVolumeProvider
	var pvc string

	switch req.Cloud {
	case providers.Amazon:
		pvc = amazon.PersistentVolumeProvider
	case providers.Azure:
		pvc = azure.PersistentVolumeProvider
	case providers.Google:
		pvc = google.PersistentVolumeProvider
	default:
		return config, pkgErrors.ErrorNotSupportedCloudType
	}

	return persistentVolumeProvider{
		Name: pvc,
		Config: persistentVolumeProviderConfig{
			Region:     req.Location,
			ApiTimeout: "3m0s",
		},
	}, nil
}

func (req ConfigRequest) getBSPConfig() (backupStorageProvider, error) {

	var config backupStorageProvider
	var bsp string

	switch req.Bucket.Provider {
	case providers.Amazon:
		bsp = amazon.BackupStorageProvider
	case providers.Azure:
		bsp = azure.BackupStorageProvider
	case providers.Google:
		bsp = google.BackupStorageProvider
	default:
		return config, pkgErrors.ErrorNotSupportedCloudType
	}

	config = backupStorageProvider{
		Name:   bsp,
		Bucket: req.Bucket.Name,
	}

	if req.Bucket.Location != "" {
		config.Config = backupStorageProviderConfig{
			Region: req.Bucket.Location,
		}
	}

	return config, nil
}

func (req ConfigRequest) getCredentials() (credentials, error) {

	var config credentials
	var azureSecret azure.Secret
	var BucketSecretContents, ClusterSecretContents string
	var err error

	switch req.Cloud {
	case providers.Amazon:
		ClusterSecretContents, err = amazon.GetSecret(req.CloudSecret)
		if err != nil {
			return config, err
		}
	case providers.Google:
		ClusterSecretContents, err = google.GetSecret(req.CloudSecret)
		if err != nil {
			return config, err
		}
	case providers.Azure:
		azureSecret = azure.GetSecret(req.CloudSecret)
	default:
		return config, pkgErrors.ErrorNotSupportedCloudType
	}

	switch req.Bucket.Provider {
	case providers.Amazon:
		BucketSecretContents, err = amazon.GetSecret(req.BucketSecret)
		if err != nil {
			return config, err
		}
	case providers.Google:
		BucketSecretContents, err = google.GetSecret(req.BucketSecret)
		if err != nil {
			return config, err
		}
	case providers.Azure:
		azureSecret, err = azure.GetSecretForBucket(req.BucketSecret, req.Bucket.StorageAccount, req.Bucket.ResourceGroup)
		if err != nil {
			return config, err
		}
	default:
		return config, pkgErrors.ErrorNotSupportedCloudType
	}

	return credentials{
		SecretContents: secretContents{
			Secret:  azureSecret,
			Cluster: ClusterSecretContents,
			Bucket:  BucketSecretContents,
		},
	}, err
}
