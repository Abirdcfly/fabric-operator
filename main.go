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

package main

import (
	"os"
	"path/filepath"
	"time"

	config "github.com/IBM-Blockchain/fabric-operator/operatorconfig"
	"github.com/IBM-Blockchain/fabric-operator/pkg/command"
	cainit "github.com/IBM-Blockchain/fabric-operator/pkg/initializer/ca"
	ccbinit "github.com/IBM-Blockchain/fabric-operator/pkg/initializer/chaincodebuild"
	chaninit "github.com/IBM-Blockchain/fabric-operator/pkg/initializer/channel"
	fedinit "github.com/IBM-Blockchain/fabric-operator/pkg/initializer/federation"
	netinit "github.com/IBM-Blockchain/fabric-operator/pkg/initializer/network"
	ordererinit "github.com/IBM-Blockchain/fabric-operator/pkg/initializer/orderer"
	orginit "github.com/IBM-Blockchain/fabric-operator/pkg/initializer/organization"
	peerinit "github.com/IBM-Blockchain/fabric-operator/pkg/initializer/peer"

	logf "sigs.k8s.io/controller-runtime/pkg/log"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"

	iamv1alpha1 "github.com/IBM-Blockchain/fabric-operator/api/iam/v1alpha1"
	ibpv1beta1 "github.com/IBM-Blockchain/fabric-operator/api/v1beta1"
	// +kubebuilder:scaffold:imports
)

const (
	defaultConfigs           = "./defaultconfig"
	defaultPeerDef           = "./definitions/peer"
	defaultCADef             = "./definitions/ca"
	defaultOrdererDef        = "./definitions/orderer"
	defaultConsoleDef        = "./definitions/console"
	defaultOrganizationDef   = "./definitions/organization"
	defaultFederationDef     = "./definitions/federation"
	defaultVoteDef           = "./definitions/vote"
	defaultNetworkDef        = "./definitions/network"
	defaultChaincodeBuildDef = "./definitions/chaincodebuild"
)

var log = logf.Log.WithName("cmd")

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(ibpv1beta1.AddToScheme(scheme))
	utilruntime.Must(iamv1alpha1.AddToScheme(scheme))
	// +kubebuilder:scaffold:scheme
}

func main() {

	operatorCfg := &config.Config{}

	setDefaultCADefinitions(operatorCfg)
	setDefaultPeerDefinitions(operatorCfg)
	setDefaultOrdererDefinitions(operatorCfg)
	setDefaultConsoleDefinitions(operatorCfg)
	setDefaultOrganizationDefinitions(operatorCfg)
	setDefaultFederationDefinitions(operatorCfg)
	setDefaultVoteDefinitions(operatorCfg)
	setDefaultNetworkDefinitions(operatorCfg)
	setDefaultChannelDefinitions(operatorCfg)
	setDefaultChaincodeBuildDefinitions(operatorCfg)

	operatorCfg.Operator.SetDefaults()

	config.SetOperatorConfigFromEnvironment(operatorCfg)

	if err := command.Operator(operatorCfg); err != nil {
		log.Error(err, "failed to start operator")
		time.Sleep(15 * time.Second)
	}

	// TODO
	// if err = (&ibpca.IBPCAReconciler{
	// 	Client: mgr.GetClient(),
	// 	Log:    ctrl.Log.WithName("controllers").WithName("IBPCA"),
	// 	Scheme: mgr.GetScheme(),
	// }).SetupWithManager(mgr); err != nil {
	// 	setupLog.Error(err, "unable to create controller", "controller", "IBPCA")
	// 	os.Exit(1)
	// }
	// if err = (&controllers.IBPPeerReconciler{
	// 	Client: mgr.GetClient(),
	// 	Log:    ctrl.Log.WithName("controllers").WithName("IBPPeer"),
	// 	Scheme: mgr.GetScheme(),
	// }).SetupWithManager(mgr); err != nil {
	// 	setupLog.Error(err, "unable to create controller", "controller", "IBPPeer")
	// 	os.Exit(1)
	// }
	// if err = (&controllers.IBPOrdererReconciler{
	// 	Client: mgr.GetClient(),
	// 	Log:    ctrl.Log.WithName("controllers").WithName("IBPOrderer"),
	// 	Scheme: mgr.GetScheme(),
	// }).SetupWithManager(mgr); err != nil {
	// 	setupLog.Error(err, "unable to create controller", "controller", "IBPOrderer")
	// 	os.Exit(1)
	// }
	// if err = (&controllers.IBPConsoleReconciler{
	// 	Client: mgr.GetClient(),
	// 	Log:    ctrl.Log.WithName("controllers").WithName("IBPConsole"),
	// 	Scheme: mgr.GetScheme(),
	// }).SetupWithManager(mgr); err != nil {
	// 	setupLog.Error(err, "unable to create controller", "controller", "IBPConsole")
	// 	os.Exit(1)
	// }
	// +kubebuilder:scaffold:builder
}

func setDefaultCADefinitions(cfg *config.Config) {
	cfg.CAInitConfig = &cainit.Config{
		CADefaultConfigPath:    filepath.Join(defaultConfigs, "ca/ca.yaml"),
		TLSCADefaultConfigPath: filepath.Join(defaultConfigs, "ca/tlsca.yaml"),
		DeploymentFile:         filepath.Join(defaultCADef, "deployment.yaml"),
		PVCFile:                filepath.Join(defaultCADef, "pvc.yaml"),
		ServiceFile:            filepath.Join(defaultCADef, "service.yaml"),
		RoleFile:               filepath.Join(defaultCADef, "role.yaml"),
		ServiceAccountFile:     filepath.Join(defaultCADef, "serviceaccount.yaml"),
		RoleBindingFile:        filepath.Join(defaultCADef, "rolebinding.yaml"),
		ConfigMapFile:          filepath.Join(defaultCADef, "configmap-caoverride.yaml"),
		IngressFile:            filepath.Join(defaultCADef, "ingress.yaml"),
		Ingressv1beta1File:     filepath.Join(defaultCADef, "ingressv1beta1.yaml"),
		RouteFile:              filepath.Join(defaultCADef, "route.yaml"),
		SharedPath:             "/tmp/data",
	}
}

func setDefaultPeerDefinitions(cfg *config.Config) {
	cfg.PeerInitConfig = &peerinit.Config{
		OUFile:                 filepath.Join(defaultConfigs, "peer/ouconfig.yaml"),
		InterOUFile:            filepath.Join(defaultConfigs, "peer/ouconfig-inter.yaml"),
		CorePeerFile:           filepath.Join(defaultConfigs, "peer/core.yaml"),
		CorePeerV2File:         filepath.Join(defaultConfigs, "peer/v2/core.yaml"),
		DeploymentFile:         filepath.Join(defaultPeerDef, "deployment.yaml"),
		PVCFile:                filepath.Join(defaultPeerDef, "pvc.yaml"),
		CouchDBPVCFile:         filepath.Join(defaultPeerDef, "couchdb-pvc.yaml"),
		ServiceFile:            filepath.Join(defaultPeerDef, "service.yaml"),
		RoleFile:               filepath.Join(defaultPeerDef, "role.yaml"),
		ServiceAccountFile:     filepath.Join(defaultPeerDef, "serviceaccount.yaml"),
		RoleBindingFile:        filepath.Join(defaultPeerDef, "rolebinding.yaml"),
		FluentdConfigMapFile:   filepath.Join(defaultPeerDef, "fluentd-configmap.yaml"),
		CouchContainerFile:     filepath.Join(defaultPeerDef, "couchdb.yaml"),
		CouchInitContainerFile: filepath.Join(defaultPeerDef, "couchdb-init.yaml"),
		IngressFile:            filepath.Join(defaultPeerDef, "ingress.yaml"),
		Ingressv1beta1File:     filepath.Join(defaultPeerDef, "ingressv1beta1.yaml"),
		CCLauncherFile:         filepath.Join(defaultPeerDef, "chaincode-launcher.yaml"),
		RouteFile:              filepath.Join(defaultPeerDef, "route.yaml"),
		StoragePath:            "/tmp/peerinit",
	}
}

func setDefaultOrdererDefinitions(cfg *config.Config) {
	cfg.OrdererInitConfig = &ordererinit.Config{
		OrdererV2File:      filepath.Join(defaultConfigs, "orderer/v2/orderer.yaml"),
		OrdererV24File:     filepath.Join(defaultConfigs, "orderer/v24/orderer.yaml"),
		OrdererFile:        filepath.Join(defaultConfigs, "orderer/orderer.yaml"),
		ConfigTxFile:       filepath.Join(defaultConfigs, "orderer/configtx.yaml"),
		OUFile:             filepath.Join(defaultConfigs, "orderer/ouconfig.yaml"),
		InterOUFile:        filepath.Join(defaultConfigs, "orderer/ouconfig-inter.yaml"),
		DeploymentFile:     filepath.Join(defaultOrdererDef, "deployment.yaml"),
		PVCFile:            filepath.Join(defaultOrdererDef, "pvc.yaml"),
		ServiceFile:        filepath.Join(defaultOrdererDef, "service.yaml"),
		CMFile:             filepath.Join(defaultOrdererDef, "configmap.yaml"),
		RoleFile:           filepath.Join(defaultOrdererDef, "role.yaml"),
		ServiceAccountFile: filepath.Join(defaultOrdererDef, "serviceaccount.yaml"),
		RoleBindingFile:    filepath.Join(defaultOrdererDef, "rolebinding.yaml"),
		IngressFile:        filepath.Join(defaultOrdererDef, "ingress.yaml"),
		Ingressv1beta1File: filepath.Join(defaultOrdererDef, "ingressv1beta1.yaml"),
		RouteFile:          filepath.Join(defaultOrdererDef, "route.yaml"),
		StoragePath:        "/tmp/ordererinit",
	}
}

func setDefaultConsoleDefinitions(cfg *config.Config) {
	cfg.ConsoleInitConfig = &config.ConsoleConfig{
		DeploymentFile:           filepath.Join(defaultConsoleDef, "deployment.yaml"),
		PVCFile:                  filepath.Join(defaultConsoleDef, "pvc.yaml"),
		ServiceFile:              filepath.Join(defaultConsoleDef, "service.yaml"),
		DeployerServiceFile:      filepath.Join(defaultConsoleDef, "deployer-service.yaml"),
		CMFile:                   filepath.Join(defaultConsoleDef, "configmap.yaml"),
		ConsoleCMFile:            filepath.Join(defaultConsoleDef, "console-configmap.yaml"),
		DeployerCMFile:           filepath.Join(defaultConsoleDef, "deployer-configmap.yaml"),
		RoleFile:                 filepath.Join(defaultConsoleDef, "role.yaml"),
		ServiceAccountFile:       filepath.Join(defaultConsoleDef, "serviceaccount.yaml"),
		RoleBindingFile:          filepath.Join(defaultConsoleDef, "rolebinding.yaml"),
		IngressFile:              filepath.Join(defaultConsoleDef, "ingress.yaml"),
		Ingressv1beta1File:       filepath.Join(defaultConsoleDef, "ingressv1beta1.yaml"),
		RouteFile:                filepath.Join(defaultConsoleDef, "route.yaml"),
		NetworkPolicyIngressFile: filepath.Join(defaultConsoleDef, "networkpolicy-ingress.yaml"),
		NetworkPolicyDenyAllFile: filepath.Join(defaultConsoleDef, "networkpolicy-denyall.yaml"),
	}
}

func setDefaultOrganizationDefinitions(cfg *config.Config) {
	cfg.OrganizationInitConfig = &orginit.Config{
		IAMEnabled:             os.Getenv("OPERATOR_USER_TYPE") != "sa",
		AdminRoleFile:          filepath.Join(defaultOrganizationDef, "admin_role.yaml"),
		ClientRoleFile:         filepath.Join(defaultOrganizationDef, "client_role.yaml"),
		RoleBindingFile:        filepath.Join(defaultOrganizationDef, "role_binding.yaml"),
		ClusterRoleFile:        filepath.Join(defaultOrganizationDef, "cluster_role.yaml"),
		ClusterRoleBindingFile: filepath.Join(defaultOrganizationDef, "cluster_role_binding.yaml"),
		CAFile:                 filepath.Join(defaultOrganizationDef, "ca.yaml"),
		NodeOUConfigFile:       filepath.Join(defaultOrganizationDef, "ou_config.yaml"),
		StoragePath:            "/tmp/orginit",
	}
}

func setDefaultFederationDefinitions(cfg *config.Config) {
	cfg.FederationInitConfig = &fedinit.Config{}
}

func setDefaultVoteDefinitions(cfg *config.Config) {
	cfg.VoteConfig = &config.VoteConfig{
		RoleFile:           filepath.Join(defaultVoteDef, "role.yaml"),
		ServiceAccountFile: filepath.Join(defaultVoteDef, "serviceaccount.yaml"),
		RoleBindingFile:    filepath.Join(defaultVoteDef, "rolebinding.yaml"),
	}
}

func setDefaultNetworkDefinitions(cfg *config.Config) {
	cfg.NetworkInitConfig = &netinit.Config{
		OrdererFile: filepath.Join(defaultNetworkDef, "orderer.yaml"),
	}
}

func setDefaultChannelDefinitions(cfg *config.Config) {
	cfg.ChannelInitConfig = &chaninit.Config{
		StoragePath: "/tmp/chaninit",
	}
}

func setDefaultChaincodeBuildDefinitions(cfg *config.Config) {
	cfg.ChaincodeBuildInitConfig = &ccbinit.Config{
		PipelinRunNamespace: os.Getenv("PIPELINE_RUN_NAMESPACE"),
		MinioHost:           os.Getenv("MINIO_HOST"),
		MinioAccessKey:      os.Getenv("MINIO_ACCESS_KEY"),
		MinioSecretKey:      os.Getenv("MINIO_SECRET_KEY"),
		PipelineRunPVCFile:  filepath.Join(defaultChaincodeBuildDef, "pipelinerun_pvc.yaml"),
		PipelineRunFile:     filepath.Join(defaultChaincodeBuildDef, "pipelinerun.yaml"),
	}
}
