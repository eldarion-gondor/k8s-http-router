package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"time"

	kapi "github.com/GoogleCloudPlatform/kubernetes/pkg/api"
	kclient "github.com/GoogleCloudPlatform/kubernetes/pkg/client"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/client/cache"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/controller/framework"
	kcontrollerFramework "github.com/GoogleCloudPlatform/kubernetes/pkg/controller/framework"
	kSelector "github.com/GoogleCloudPlatform/kubernetes/pkg/fields"
	"golang.org/x/net/context"
)

const (
	// Resync period for the kube controller loop.
	resyncPeriod = 5 * time.Minute
)

type SyncHandler interface {
	SetHost(host string, target *url.URL)
	RemoveHost(host string)
}

type kubernetesServiceConfig struct {
	Config struct {
	} `json:"config,omitempty"`
	Hosts []string `json:"hosts,omitempty"`
}

type kubernetesDataStore struct {
	kubClient *kclient.Client
	srvCache  []*kapi.Service
}

func NewKubernetesDataStore() (*kubernetesDataStore, error) {
	masterHost := os.Getenv("KUBERNETES_RO_SERVICE_HOST")
	if masterHost == "" {
		log.Fatalf("KUBERNETES_RO_SERVICE_HOST is not defined")
	}
	masterPort := os.Getenv("KUBERNETES_RO_SERVICE_PORT")
	if masterPort == "" {
		log.Fatalf("KUBERNETES_RO_SERVICE_PORT is not defined")
	}
	config := &kclient.Config{
		Host:    fmt.Sprintf("http://%s:%s", masterHost, masterPort),
		Version: "v1beta3",
	}
	kubClient, err := kclient.New(config)
	if err != nil {
		return nil, err
	}
	return &kubernetesDataStore{
		kubClient: kubClient,
	}, nil
}

func (ds *kubernetesDataStore) createServiceLW() *cache.ListWatch {
	return cache.NewListWatchFromClient(ds.kubClient, "services", kapi.NamespaceAll, kSelector.Everything())
}

func (ds *kubernetesDataStore) Watch(ctx context.Context, sh SyncHandler) {
	var serviceController *kcontrollerFramework.Controller
	_, serviceController = framework.NewInformer(
		ds.createServiceLW(),
		&kapi.Service{},
		resyncPeriod,
		framework.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				if s, ok := obj.(*kapi.Service); ok {
					ds.setService(s, sh)
				}
			},
			DeleteFunc: func(obj interface{}) {
				if s, ok := obj.(*kapi.Service); ok {
					ds.removeService(s, sh)
				}
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				if s, ok := newObj.(*kapi.Service); ok {
					ds.setService(s, sh)
				}
			},
		},
	)
	serviceController.Run(ctx.Done())
}

func (ds *kubernetesDataStore) setService(service *kapi.Service, sh SyncHandler) error {
	if _, ok := service.Annotations["router"]; !ok {
		return nil
	}
	var res kubernetesServiceConfig
	if err := json.Unmarshal([]byte(service.Annotations["router"]), &res); err != nil {
		return nil
	}
	for i := range res.Hosts {
		host := res.Hosts[i]
		if len(service.Spec.Ports) == 0 {
			continue
		}
		target, err := url.Parse(
			fmt.Sprintf(
				"http://%s:%d",
				service.Spec.PortalIP,
				service.Spec.Ports[0].Port,
			),
		)
		if err != nil {
			return fmt.Errorf("cannot parse URL: %s", err)
		}
		sh.SetHost(host, target)
	}
	log.Printf("watcher: added %s", service.Name)
	return nil
}

func (ds *kubernetesDataStore) removeService(service *kapi.Service, sh SyncHandler) error {
	if _, ok := service.Annotations["router"]; !ok {
		return nil
	}
	var res kubernetesServiceConfig
	if err := json.Unmarshal([]byte(service.Annotations["router"]), &res); err != nil {
		return nil
	}
	for i := range res.Hosts {
		host := res.Hosts[i]
		sh.RemoveHost(host)
	}
	log.Printf("watcher: removed %s", service.Name)
	return nil
}
