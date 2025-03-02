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

package configmap

import (
	"context"
	"fmt"

	k8sclient "github.com/IBM-Blockchain/fabric-operator/pkg/k8s/controllerclient"
	"github.com/IBM-Blockchain/fabric-operator/pkg/manager/resources"
	"github.com/IBM-Blockchain/fabric-operator/pkg/operatorerrors"
	"github.com/IBM-Blockchain/fabric-operator/pkg/util"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var log = logf.Log.WithName("configmap_manager")

type Manager struct {
	Client        k8sclient.Client
	Scheme        *runtime.Scheme
	ConfigMapFile string
	Name          string
	Options       map[string]interface{}

	LabelsFunc   func(v1.Object) map[string]string
	OverrideFunc func(v1.Object, *corev1.ConfigMap, resources.Action, map[string]interface{}) error
}

func (m *Manager) GetName(instance v1.Object) string {
	if m.Name != "" {
		return fmt.Sprintf("%s-%s", instance.GetName(), m.Name)
	}
	return instance.GetName()
}

func (m *Manager) Reconcile(instance v1.Object, update bool) error {
	name := m.GetName(instance)
	configMap := &corev1.ConfigMap{}
	err := m.Client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: instance.GetNamespace()}, configMap)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			log.Info(fmt.Sprintf("Creating configmap '%s'", name))
			configMap, err := m.GetConfigMapBasedOnCRFromFile(instance)
			if err != nil {
				return err
			}

			err = m.Client.Create(context.TODO(), configMap, k8sclient.CreateOption{
				Owner:  instance,
				Scheme: m.Scheme,
			})
			if err != nil {
				return err
			}
			return nil
		}
		return err
	}

	if update {
		if m.OverrideFunc != nil {
			log.Info(fmt.Sprintf("Updating configmap '%s'", name))
			err := m.OverrideFunc(instance, configMap, resources.Update, m.Options)
			if err != nil {
				return err
			}

			err = m.Client.Update(context.TODO(), configMap, k8sclient.UpdateOption{
				Owner:  instance,
				Scheme: m.Scheme,
			})
			if err != nil {
				return err
			}
			return nil
		}
	}

	// TODO: If needed, update logic for servie goes here

	return nil
}

func (m *Manager) GetConfigMapBasedOnCRFromFile(instance v1.Object) (*corev1.ConfigMap, error) {
	configMap, err := util.GetConfigMapFromFile(m.ConfigMapFile)
	if err != nil {
		log.Error(err, fmt.Sprintf("Error reading configmap configuration file: %s", m.ConfigMapFile))
		return nil, err
	}

	configMap.Name = m.GetName(instance)
	configMap.Namespace = instance.GetNamespace()
	configMap.Labels = m.LabelsFunc(instance)

	return m.BasedOnCR(instance, configMap)
}

func (m *Manager) BasedOnCR(instance v1.Object, configMap *corev1.ConfigMap) (*corev1.ConfigMap, error) {
	if m.OverrideFunc != nil {
		err := m.OverrideFunc(instance, configMap, resources.Create, m.Options)
		if err != nil {
			return nil, operatorerrors.New(operatorerrors.InvalidConfigMapCreateRequest, err.Error())
		}
	}

	return configMap, nil
}

func (m *Manager) Get(instance v1.Object) (client.Object, error) {
	if instance == nil {
		return nil, nil // Instance has not been reconciled yet
	}

	name := m.GetName(instance)
	cm := &corev1.ConfigMap{}
	err := m.Client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: instance.GetNamespace()}, cm)
	if err != nil {
		return nil, err
	}

	return cm, nil
}

func (m *Manager) Exists(instance v1.Object) bool {
	_, err := m.Get(instance)

	return err == nil
}

func (m *Manager) Delete(instance v1.Object) error {
	cm, err := m.Get(instance)
	if err != nil {
		if !k8serrors.IsNotFound(err) {
			return err
		}
	}

	if cm == nil {
		return nil
	}

	err = m.Client.Delete(context.TODO(), cm)
	if err != nil {
		if !k8serrors.IsNotFound(err) {
			return err
		}
	}

	return nil
}

func (m *Manager) CheckState(instance v1.Object) error {
	// NO-OP
	return nil
}

func (m *Manager) RestoreState(instance v1.Object) error {
	// NO-OP
	return nil
}

func (m *Manager) SetCustomName(name string) {
	// NO-OP
}
