package k8s

import (
	"context"
	"flag"
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
	var kubeconfig *string
	if home := homeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	if k := os.Getenv("KUBECONFIG"); k != "" {
		*kubeconfig = k
	}
	flag.Parse()

	//client, err := k8s.NewInClusterClient()
	var err error
	client, err = loadClient(*kubeconfig)
	if err != nil {
		client, err = k8s.NewInClusterClient()
	}
	if err != nil {
		log.Fatal(err)
	}

}

// func main() {

// }

func Secret(name string) (secret corev1.Secret, err error) {

	err = client.Get(context.Background(), "", name, &secret)
	if err != nil {
		err = fmt.Errorf("get secret err %v", err)
		return
	}
	return

}

func SecretList(ns string) (secrets []*corev1.Secret, err error) {
	var slist corev1.SecretList
	err = client.List(context.Background(), ns, &slist)
	if err != nil {
		err = fmt.Errorf("get secret err %v", err)
		return
	}
	secrets = slist.GetItems()
	// for _, secret := range secrets {
	// 	spew.Dump("secret:", secret)
	// 	// log.Printf("secret=%q maps=%t\n", *node.Metadata.Name, !*node.Spec.Unschedulable)
	// }
	// return nil, nil
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

func loadClient(kubeconfigPath string) (*k8s.Client, error) {
	data, err := ioutil.ReadFile(kubeconfigPath)
	if err != nil {
		return nil, fmt.Errorf("read kubeconfig: %v", err)
	}

	// Unmarshal YAML into a Kubernetes config object.
	var config k8s.Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("unmarshal kubeconfig: %v", err)
	}
	return k8s.NewClient(&config)
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}
