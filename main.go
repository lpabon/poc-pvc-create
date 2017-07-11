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
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	//	"k8s.io/apimachinery/pkg/labels"
	//	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	//	"k8s.io/client-go/pkg/apis/extensions"
	//	"k8s.io/client-go/pkg/apis/extensions/v1beta1"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/heketi/utils"
)

const (
	pvcCreatorAnnotationSizeAvailable = "pvc-create.alpha.kubernetes.io/size-available"
	demoPvcName                       = "pvc-create-sample-pvc"
)

var (
	logger       = utils.NewLogger("pvc-create", utils.LEVEL_INFO)
	kubeconfig   string
	storageClass string
)

type pvcCreator struct {
	kclient *kubernetes.Clientset
}

func init() {
	flagset := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	flagset.StringVar(&kubeconfig, "kubeconfig", "", "Path to kubeconfig.")
	flagset.StringVar(&storageClass,
		"storageclass",
		"",
		"StorageClass example: gluster.qm.gluster. Check your system for storage classes")
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

	// List PVCs on default namespace
	p.ListPvcs("default")

	// If storage class provided, create a pvc
	if len(storageClass) != 0 {
		err = p.CreatePVC("default", storageClass)
		if err != nil {
			return
		}

		// List
		p.ListPvcs("default")

		err = p.DeletePVC("default")
		if err != nil {
			return
		}
		// List
		p.ListPvcs("default")

	}

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

func (p *pvcCreator) ListPvcs(namespace string) error {

	pvcs := p.kclient.Core().PersistentVolumeClaims(namespace)
	list, err := pvcs.List(meta.ListOptions{})
	if err != nil {
		return logger.Err(err)
	}

	logger.Info("Number of PVCs: %d", len(list.Items))
	for _, pvc := range list.Items {
		logger.Info("PVC: %s\n"+
			"\tVolume name: %s\n"+
			"\tSize Available: %s\n",
			pvc.GetName(),
			pvc.Spec.VolumeName,
			pvc.Annotations[pvcCreatorAnnotationSizeAvailable])
	}

	return nil
}

func (p *pvcCreator) CreatePVC(namespace, storageClass string) error {
	pvc := &v1.PersistentVolumeClaim{
		ObjectMeta: meta.ObjectMeta{
			Name:      demoPvcName,
			Namespace: namespace,
			Labels: map[string]string{
				"createBy": "pvc-create",
			},
			Annotations: map[string]string{
				pvcCreatorAnnotationSizeAvailable:         "10",
				"volume.beta.kubernetes.io/storage-class": storageClass,
			},
		},
		Spec: v1.PersistentVolumeClaimSpec{
			AccessModes: []v1.PersistentVolumeAccessMode{
				v1.ReadWriteMany,
			},
			Resources: v1.ResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceName(v1.ResourceStorage): resource.MustParse("10"),
				},
			},
		},
	}

	// Get an API to submit the PVC
	pvcs := p.kclient.Core().PersistentVolumeClaims(namespace)

	// Submit PVC
	logger.Info("PVC %s submitted", pvc.GetName())
	_, err := pvcs.Create(pvc)
	if apierrors.IsAlreadyExists(err) {
		logger.Warning("pvc already exists")
		return nil
	} else if err != nil {
		logger.Err(err)
	} else {
		logger.Info("Waiting for PVC to be bound")
		return wait.Poll(3*time.Second, 10*time.Minute, func() (bool, error) {

			// Get the PVC state from Kube
			p, err := pvcs.Get(pvc.GetName(), meta.GetOptions{})
			if err != nil {
				return false, logger.Err(err)
			}

			// Check if the pvc is bound
			if p.Status.Phase == v1.ClaimBound {
				return true, nil
			}

			// Not bound yet
			return false, nil
		})
	}

	return nil
}

func (p *pvcCreator) DeletePVC(namespace string) error {

	// Get an API to submit the PVC
	pvcs := p.kclient.Core().PersistentVolumeClaims(namespace)

	// Delete the PVC
	logger.Info("Deleting PVC %s", demoPvcName)
	pvcs.Delete(demoPvcName, &meta.DeleteOptions{})

	// Wait until it is deleted
	return wait.Poll(3*time.Second, 10*time.Minute, func() (bool, error) {

		// Get the PVC state from Kube
		_, err := pvcs.Get(demoPvcName, meta.GetOptions{})
		if err != nil {
			// Does not exist, so we are done
			return true, nil
		}

		// Still there
		return false, nil
	})
}
