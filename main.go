// Copyright 2017 Luis Pabón <lpabon@gmail.com>
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

package main

import (
	"flag"
	"os"

	//	apierrors "k8s.io/apimachinery/pkg/api/errors"
	//	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	//	"k8s.io/apimachinery/pkg/labels"
	//	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes"
	//	"k8s.io/client-go/pkg/api"
	//	"k8s.io/client-go/pkg/apis/extensions"
	//	"k8s.io/client-go/pkg/apis/extensions/v1beta1"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/heketi/utils"
)

var (
	logger     = utils.NewLogger("pvc-create", utils.LEVEL_INFO)
	kubeconfig string
)

type pvcCreator struct {
	kclient *kubernetes.Clientset
}

func init() {
	flagset := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	flagset.StringVar(&kubeconfig, "kubeconfig", "", "Path to kubeconfig.")
	flagset.Parse(os.Args[1:])
}

func main() {
	logger.Info("Starting")

	// Create object
	p := newPvcCreator(kubeconfig)
	if p == nil {
		return
	}

	// Show version
	ver, err := p.GetVersion()
	if err != nil {
		return
	}
	logger.Info("connection to Kubernetes established. Cluster version %s", ver)

}

func newPvcCreator(kubeconfig string) *pvcCreator {
	p := &pvcCreator{}

	// Setup REST client to Kubernetes
	var err error
	var cfg *restclient.Config
	if len(kubeconfig) != 0 {
		// Get configuration from kubeconfig file
		cfg, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			logger.Err(err)
			return nil
		}
	} else {
		// Running as a container inside Kubernetes
		cfg, err = restclient.InClusterConfig()
		if err != nil {
			logger.Err(err)
			return nil
		}
	}

	// Setup Clientset (High level APIs) for Kube
	p.kclient, err = kubernetes.NewForConfig(cfg)
	if err != nil {
		logger.Err(err)
		return nil
	}

	return p
}

func (p *pvcCreator) GetVersion() (string, error) {
	v, err := p.kclient.Discovery().ServerVersion()
	if err != nil {
		return "", logger.LogError("communicating with server failed: %s", err)
	}

	return v.String(), nil
}
