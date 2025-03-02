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

package baseproposal

import (
	"context"
	"fmt"
	"os"
	"sync"

	current "github.com/IBM-Blockchain/fabric-operator/api/v1beta1"
	config "github.com/IBM-Blockchain/fabric-operator/operatorconfig"
	k8sclient "github.com/IBM-Blockchain/fabric-operator/pkg/k8s/controllerclient"
	"github.com/IBM-Blockchain/fabric-operator/pkg/offering/common"
	bcrbac "github.com/IBM-Blockchain/fabric-operator/pkg/rbac"
	"github.com/pkg/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/pointer"
	k8sruntime "sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var log = logf.Log.WithName("base_proposal")

type Override interface{}

//go:generate counterfeiter -o mocks/update.go -fake-name Update . Update

type Update interface {
}

type Proposal interface {
	PreReconcileChecks(instance *current.Proposal) (bool, error)
	ReconcileManagers(ctx context.Context, instance *current.Proposal) error
	Reconcile(instance *current.Proposal) (common.Result, error)
}

var _ Proposal = &BaseProposal{}

type BaseProposal struct {
	Client k8sclient.Client
	Scheme *runtime.Scheme
	Config *config.Config

	RBACManager *bcrbac.Manager

	Override Override
}

func New(client k8sclient.Client, scheme *runtime.Scheme, config *config.Config, o Override) *BaseProposal {
	p := &BaseProposal{
		Client:   client,
		Scheme:   scheme,
		Config:   config,
		Override: o,
	}

	p.CreateManagers()
	return p
}

func (p *BaseProposal) CreateManagers() {
	p.RBACManager = bcrbac.NewRBACManager(p.Client, nil)
}

func (p *BaseProposal) PreReconcileChecks(instance *current.Proposal) (bool, error) {
	// todo add
	return false, nil
}

func (p *BaseProposal) Reconcile(instance *current.Proposal) (result common.Result, err error) {
	log.Info("Reconciling...")
	if instance.Status.Phase == current.ProposalPending {

	} else if instance.Status.Phase == current.ProposalVoting {

	} else if instance.Status.Phase == current.ProposalFinished {
	}

	if err = p.ReconcileManagers(context.TODO(), instance); err != nil {
		return common.Result{}, errors.Wrap(err, "failed to reconcile managers")
	}

	return common.Result{}, nil
}

func (c *BaseProposal) ReconcileManagers(ctx context.Context, instance *current.Proposal) (err error) {
	if err = c.ReconcileOwnerReference(instance); err != nil {
		return errors.Wrap(err, "failed OwerReference reconciliation")
	}

	if err = c.ReconcileRBAC(instance); err != nil {
		return errors.Wrap(err, "failed RBAC reconciliation")
	}
	if err = c.ReconcileVote(ctx, instance); err != nil {
		return errors.Wrap(err, "failed Vote reconciliation")
	}
	return
}

func (c *BaseProposal) ReconcileOwnerReference(instance *current.Proposal) (err error) {
	fed := &current.Federation{}
	err = c.Client.Get(context.TODO(), types.NamespacedName{Name: instance.Spec.Federation}, fed)
	if err != nil {
		return errors.Wrap(err, "get proposal's federation")
	}
	ownerReference := bcrbac.OwnerReference(bcrbac.Federation, fed)

	var exist bool
	for _, reference := range instance.OwnerReferences {
		if reference.UID == ownerReference.UID {
			exist = true
			break
		}
	}
	if !exist {
		instance.OwnerReferences = append(instance.OwnerReferences, ownerReference)

		err = c.Client.Update(context.TODO(), instance)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *BaseProposal) ReconcileRBAC(instance *current.Proposal) (err error) {
	if err = c.RBACManager.Reconcile(bcrbac.Proposal, instance, bcrbac.ResourceCreate); err != nil {
		return errors.Wrap(err, "failed sync rbac")
	}
	return
}

func (c *BaseProposal) ValidateSpec(instance *current.Proposal) error {
	return nil
}

func (c *BaseProposal) GetLabels(instance v1.Object) map[string]string {
	label := os.Getenv("OPERATOR_LABEL_PREFIX")
	if label == "" {
		label = "fabric"
	}

	return map[string]string{
		"app":                          instance.GetName(),
		"creator":                      label,
		"release":                      "operator",
		"helm.sh/chart":                "ibm-" + label,
		"app.kubernetes.io/name":       label,
		"app.kubernetes.io/instance":   label + "Proposal",
		"app.kubernetes.io/managed-by": label + "-operator",
	}
}

func (c *BaseProposal) ReconcileVote(ctx context.Context, instance *current.Proposal) (err error) {
	return c.CreateVoteIfNotExists(ctx, instance)
}

func (c *BaseProposal) CreateVoteIfNotExists(ctx context.Context, instance *current.Proposal) error {
	organizations, err := instance.GetCandidateOrganizations(ctx, c.Client)
	if err != nil {
		log.Error(err, "cant get organizations for proposal:"+instance.GetName())
		return err
	}
	wg := sync.WaitGroup{}
	for _, org := range organizations {
		wg.Add(1)
		go func(orgName string) {
			defer func() {
				wg.Done()
			}()
			org := &current.Organization{ObjectMeta: v1.ObjectMeta{Name: orgName}}
			vote := &current.Vote{
				ObjectMeta: v1.ObjectMeta{
					Name:      instance.GetVoteName(org.Name),
					Namespace: org.GetUserNamespace(),
					Labels:    instance.GetVoteLabel(),
				},
				Spec: current.VoteSpec{
					ProposalName:     instance.GetName(),
					OrganizationName: org.Name,
					Decision:         nil,
					Description:      "",
				},
			}
			if err = c.Client.Get(ctx, k8sruntime.ObjectKeyFromObject(vote), vote); err != nil {
				if k8sruntime.IgnoreNotFound(err) == nil {
					log.Info(fmt.Sprintf("not find vote in org:%s, crate now.", orgName))
					if err = c.Client.Create(ctx, vote, k8sclient.CreateOption{Owner: instance, Scheme: c.Scheme}); err != nil {
						log.Error(err, "Error create vote")
					}
				} else {
					log.Error(err, fmt.Sprintf("Error getting vote in org:%s", orgName))
					// todo return error
				}
			} else {
				if org.Name == instance.Spec.InitiatorOrganization {
					if pointer.BoolDeref(vote.Spec.Decision, false) {
						return
					}
					vote.Spec.Decision = pointer.Bool(true)
					if err = c.Client.Patch(ctx, vote, nil, k8sclient.PatchOption{Resilient: &k8sclient.ResilientPatch{Retry: 3, Into: &current.Vote{}, Strategy: k8sruntime.MergeFrom}}); err != nil {
						log.Error(err, "Error patch vote")
					}
				}
			}
		}(org)
	}
	wg.Wait()
	return nil
}
