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

package v1beta1

import (
	"context"
	"fmt"
	"math/rand"
	"reflect"
	"strings"
	"time"

	authenticationv1 "k8s.io/api/authentication/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

// log is for logging in this package.
var ccLogger = logf.Log.WithName("chaincode-resource")

const (
	minLength, maxLength = 10, 30

	// https://github.com/hyperledger/fabric/blob/main/core/chaincode/persistence/chaincode_package.go#L248
	// remove the underscore, the name of the pod does not support the use of underscores.
	alnum = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890-"
)

func genLabel() string {
	seed := rand.NewSource(time.Now().UnixNano())
	rr := rand.New(seed)

	// total a-zA-Z0-9
	base := 62
	targetLength := rr.Intn(21) + minLength

	buf := strings.Builder{}
	for i := 0; i < targetLength; i++ {
		pickItem := rr.Intn(base)
		buf.WriteByte(alnum[pickItem])
		if i == 0 {
			// -
			base += 1
		}
	}
	return buf.String()
}

//+kubebuilder:webhook:path=/mutate-ibp-com-v1beta1-chaincode,mutating=true,failurePolicy=fail,sideEffects=None,groups=ibp.com,resources=chaincodes,verbs=create;update,versions=v1beta1,name=chaincode.mutate.webhook,admissionReviewVersions=v1

var _ defaulter = &Chaincode{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *Chaincode) Default(ctx context.Context, client client.Client, user authenticationv1.UserInfo) {
	ccLogger.Info("default", "name", r.Name, "user", user.String())
	if r.Spec.Label == "" {
		r.Spec.Label = genLabel()
	}
}

//+kubebuilder:webhook:path=/validate-ibp-com-v1beta1-chaincode,mutating=false,failurePolicy=fail,sideEffects=None,groups=ibp.com,resources=chaincodes,verbs=create;update;delete,versions=v1beta1,name=chaincode.validate.webhook,admissionReviewVersions=v1

var _ validator = &Chaincode{}

// checkChAndEp Check if both ch and ep are present
func (r *Chaincode) checkChAndEp(c client.Client) error {
	_, err := r.GetChannel(c)
	if err != nil {
		return err
	}
	ep := &EndorsePolicy{}
	return c.Get(context.TODO(), types.NamespacedName{Name: r.Spec.EndorsePolicyRef.Name}, ep)
}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Chaincode) ValidateCreate(ctx context.Context, client client.Client, user authenticationv1.UserInfo) error {
	ccLogger.Info("validate create", "name", r.Name, "user", user.String())
	if err := checkChaincodeBuildImage(ctx, client, r.Spec.ExternalBuilder); err != nil {
		ccLogger.Error(err, "")
		return err
	}
	return r.checkChAndEp(client)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *Chaincode) ValidateUpdate(ctx context.Context, client client.Client, old runtime.Object, user authenticationv1.UserInfo) error {
	ccLogger.Info("validate update", "name", r.Name, "user", user.String())
	oldcc := old.(*Chaincode)
	if err := checkChaincodeBuildImage(ctx, client, r.Spec.ExternalBuilder); err != nil {
		ccLogger.Error(err, "")
		return err
	}

	// If the information relating to the mirror of the chaincode has changed,
	// it can only be updated if the previous version of the chaincode is in a pending phase.
	if r.Spec.Version != oldcc.Spec.Version || r.Spec.ExternalBuilder != oldcc.Spec.ExternalBuilder || !reflect.DeepEqual(r.Spec.Images, oldcc.Spec.Images) {
		if oldcc.Status.Phase != ChaincodePhasePending {
			return fmt.Errorf("please upgrade via proposal")
		}
	}
	return oldcc.checkChAndEp(client)
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *Chaincode) ValidateDelete(ctx context.Context, c client.Client, user authenticationv1.UserInfo) error {
	ccLogger.Info("validate delete", "name", r.Name, "user", user.String())
	if err := r.checkChAndEp(c); err != nil {
		return err
	}
	if r.Status.Phase != ChaincodePhaseUnapproved {
		return fmt.Errorf("it can only be deleted if the vote is not approved")
	}
	return nil
}

func checkChaincodeBuildImage(ctx context.Context, c client.Client, chaincodebuildName string) error {
	ccb := &ChaincodeBuild{}
	if err := c.Get(ctx, types.NamespacedName{Name: chaincodebuildName}, ccb); err != nil {
		return err
	}

	return ccb.HasImage()
}
