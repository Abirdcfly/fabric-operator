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

package clusterrole_test

import (
	"github.com/IBM-Blockchain/fabric-operator/controllers/mocks"
	"github.com/IBM-Blockchain/fabric-operator/pkg/manager/resources"
	"github.com/IBM-Blockchain/fabric-operator/pkg/manager/resources/clusterrole"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	rbacv1 "k8s.io/api/rbac/v1"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("ClusterRole manager", func() {
	var (
		mockKubeClient *mocks.Client
		manager        *clusterrole.Manager
		instance       metav1.Object
	)

	BeforeEach(func() {
		mockKubeClient = &mocks.Client{}
		manager = &clusterrole.Manager{
			ClusterRoleFile: "../../../../definitions/organization/cluster_role.yaml",
			Client:          mockKubeClient,
			OverrideFunc: func(v1.Object, *rbacv1.ClusterRole, resources.Action) error {
				return nil
			},
			LabelsFunc: func(v1.Object) map[string]string {
				return map[string]string{}
			},
		}

		instance = &metav1.ObjectMeta{}
	})

	Context("reconciles the  instance", func() {
		It("does not try to create cluster role if the get request returns an error other than 'not found'", func() {
			errMsg := "connection refused"
			mockKubeClient.GetReturns(errors.New(errMsg))
			err := manager.Reconcile(instance, false)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(errMsg))
		})

		When("cluster role does not exist", func() {
			BeforeEach(func() {
				notFoundErr := &k8serror.StatusError{
					ErrStatus: metav1.Status{
						Reason: metav1.StatusReasonNotFound,
					},
				}
				mockKubeClient.GetReturns(notFoundErr)
			})

			It("returns an error if fails to load default config", func() {
				manager.ClusterRoleFile = "bad.yaml"
				err := manager.Reconcile(instance, false)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("no such file or directory"))
			})

			It("returns an error if override cluster role value fails", func() {
				manager.OverrideFunc = func(v1.Object, *rbacv1.ClusterRole, resources.Action) error {
					return errors.New("creation override failed")
				}
				err := manager.Reconcile(instance, false)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).Should(ContainSubstring("creation override failed"))
			})

			It("returns an error if the creation of the cluster role fails", func() {
				errMsg := "unable to create cluster role"
				mockKubeClient.CreateReturns(errors.New(errMsg))
				err := manager.Reconcile(instance, false)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal(errMsg))
			})

			It("does not return an error on a successfull cluster role creation", func() {
				err := manager.Reconcile(instance, false)
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})
})
