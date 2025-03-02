//go:build pkcs11
// +build pkcs11

/*
 * Copyright contributors to the Hyperledger Fabric Operator project
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at:
 *
 * 	  http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package reenroller

import (
	commonapi "github.com/IBM-Blockchain/fabric-operator/pkg/apis/common"
	"github.com/bestchains/fabric-ca/lib"
	"github.com/hyperledger/fabric/bccsp/factory"
	"github.com/hyperledger/fabric/bccsp/pkcs11"
)

func GetClient(client *lib.Client, bccsp *commonapi.BCCSP) *lib.Client {
	if bccsp != nil {
		if bccsp.PKCS11 != nil {
			client.Config.CSP = &factory.FactoryOpts{
				ProviderName: bccsp.ProviderName,
				Pkcs11Opts: &pkcs11.PKCS11Opts{
					SecLevel:   bccsp.PKCS11.SecLevel,
					HashFamily: bccsp.PKCS11.HashFamily,
					Ephemeral:  bccsp.PKCS11.Ephemeral,
					Library:    bccsp.PKCS11.Library,
					Label:      bccsp.PKCS11.Label,
					Pin:        bccsp.PKCS11.Pin,
					SoftVerify: bccsp.PKCS11.SoftVerify,
					Immutable:  bccsp.PKCS11.Immutable,
				},
			}

			if bccsp.PKCS11.FileKeyStore != nil {
				client.Config.CSP.Pkcs11Opts.FileKeystore = &pkcs11.FileKeystoreOpts{
					KeyStorePath: bccsp.PKCS11.FileKeyStore.KeyStorePath,
				}
			}
		}
	}

	return client
}
