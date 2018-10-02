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

package azure

import (
	"context"

	mgmtStorage "github.com/Azure/azure-sdk-for-go/services/storage/mgmt/2017-10-01/storage"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure/auth"
	pkgSecret "github.com/banzaicloud/pipeline/pkg/secret"
	"github.com/banzaicloud/pipeline/secret"
	"github.com/pkg/errors"
)

// Secret describes values for Azure access
type Secret struct {
	// For general access
	ClientID       string `json:"AZURE_CLIENT_ID,omitempty"`
	ClientSecret   string `json:"AZURE_CLIENT_SECRET,omitempty"`
	SubscriptionID string `json:"AZURE_SUBSCRIPTION_ID,omitempty"`
	TenantID       string `json:"AZURE_TENANT_ID,omitempty"`

	// For bucket access
	ResourceGroup  string `json:"AZURE_RESOURCE_GROUP,omitempty"`
	StorageAccount string `json:"AZURE_STORAGE_ACCOUNT_ID,omitempty"`
	StorageKey     string `json:"AZURE_STORAGE_KEY,omitempty"`
}

// GetSecretForBucket gets formatted secret for ARK backup bucket
func GetSecretForBucket(secret *secret.SecretItemResponse, storageAccount string, resourceGroup string) (Secret, error) {

	s := GetSecret(secret)
	s.StorageAccount = storageAccount
	s.ResourceGroup = resourceGroup

	key, err := s.getStorageAccountKey()
	if err != nil {
		return Secret{}, err
	}
	s.StorageKey = key

	return s, nil
}

// GetSecret gets formatted secret for ARK
func GetSecret(secret *secret.SecretItemResponse) Secret {

	return Secret{
		ClientID:       secret.Values[pkgSecret.AzureClientId],
		ClientSecret:   secret.Values[pkgSecret.AzureClientSecret],
		SubscriptionID: secret.Values[pkgSecret.AzureSubscriptionId],
		TenantID:       secret.Values[pkgSecret.AzureTenantId],
	}
}

func (s Secret) getStorageAccountKey() (string, error) {
	client, err := s.createStorageAccountClient()
	if err != nil {
		return "", errors.WithStack(err)
	}

	keys, err := client.ListKeys(context.TODO(), s.ResourceGroup, s.StorageAccount)
	if err != nil {
		return "", errors.WithStack(err)
	}

	key := (*keys.Keys)[0].Value
	if key != nil {
		return *key, nil
	}

	return "", nil
}

func (s Secret) createStorageAccountClient() (*mgmtStorage.AccountsClient, error) {
	accountClient := mgmtStorage.NewAccountsClient(s.SubscriptionID)

	authorizer, err := s.newAuthorizer()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	accountClient.Authorizer = authorizer

	return &accountClient, nil
}

func (s Secret) newAuthorizer() (autorest.Authorizer, error) {
	authorizer, err := auth.NewClientCredentialsConfig(
		s.ClientID,
		s.ClientSecret,
		s.TenantID).Authorizer()

	if err != nil {
		return nil, errors.WithStack(err)
	}

	return authorizer, nil
}
