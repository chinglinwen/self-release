package k8s

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/ericchiang/k8s"
	corev1 "github.com/ericchiang/k8s/apis/core/v1"
	"github.com/pkg/errors"

	"github.com/ghodss/yaml"
)

var (
	client *k8s.Client
)

func init() {
	var kubeconfig string
	if home := homeDir(); home != "" {
		kubeconfig = filepath.Join(home, ".kube", "config")
	}
	if k := os.Getenv("KUBECONFIG"); k != "" {
		kubeconfig = k
	}
	Init(kubeconfig)
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}

func Init(kubeconfig string) {

	//client, err := k8s.NewInClusterClient()
	var err error
	envconfig := os.Getenv("KUBECONFIG")
	if envconfig != "" {
		client, err = loadClient(envconfig)
		if err == nil {
			log.Printf("loaded kubeconfig from KUBECONFIG env\n")
			return
		}
		log.Printf("try load kubeconfig from env: %v, err: %v\n", envconfig, err)
	}
	client, err = loadClient(kubeconfig)
	if err == nil {
		log.Printf("loaded kubeconfig from: %v\n", kubeconfig)
		return
	}
	log.Printf("try load kubeconfig: %v, err: %v\n", kubeconfig, err)
	client, err = k8s.NewInClusterClient()
	if err != nil {
		log.Printf("last method to load kubeconfig from inside cluster failed\n")
		log.Fatal(err)
	}
}

func loadClient(kubeconfigPath string) (*k8s.Client, error) {
	log.Printf("try load k8s config from: %v\n", kubeconfigPath)
	data, err := ioutil.ReadFile(kubeconfigPath)
	if err != nil {
		return nil, fmt.Errorf("read kubeconfig: %v", err)
	}
	// log.Printf("k8s config: %v\n", string(data))
	// Unmarshal YAML into a Kubernetes config object.
	var config k8s.Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("unmarshal kubeconfig: %v", err)
	}
	return k8s.NewClient(&config)
}

func Service(name string) (service corev1.Service, err error) {
	err = client.Get(context.Background(), "", name, &service)
	if err != nil {
		err = fmt.Errorf("get service err %v", err)
		return
	}
	return
}

func ServiceListAll() (services []*corev1.Service, err error) {
	return ServiceList("")
}

func ServiceListWithLabels(ns string, labels map[string]string) (services []*corev1.Service, err error) {
	options := []k8s.Option{}
	if len(labels) != 0 {
		l := new(k8s.LabelSelector)
		for k, v := range labels {
			l.Eq(k, v)
		}
		options = append(options, l.Selector())
	}
	var slist corev1.ServiceList
	err = client.List(context.Background(), ns, &slist, options...)
	if err != nil {
		err = fmt.Errorf("get secret err %v", err)
		return
	}
	services = slist.GetItems()
	return
}

func ServiceList(ns string) (services []*corev1.Service, err error) {
	var slist corev1.ServiceList
	err = client.List(context.Background(), ns, &slist)
	if err != nil {
		err = fmt.Errorf("get secret err %v", err)
		return
	}
	services = slist.GetItems()
	return
}

func nodeList() error {
	var nodes corev1.NodeList
	if err := client.List(context.Background(), "", &nodes); err != nil {
		return errors.Wrap(err, "client list")
	}
	for _, node := range nodes.Items {
		log.Printf("name=%q schedulable=%t\n", *node.Metadata.Name, !*node.Spec.Unschedulable)
	}
	return nil
}
