package k8s

import (
	"context"
	"fmt"

	corev1 "github.com/ericchiang/k8s/apis/core/v1"
)

func Secret(name string) (secret corev1.Secret, err error) {
	err = client.Get(context.Background(), "", name, &secret)
	if err != nil {
		err = fmt.Errorf("get secret err %v", err)
		return
	}
	return
}

// empty ns means all namespaces
// using key to filter secrets which has such key defined
func SecretListAllWithHasKey(ns, key string, excludens []string) (secrets []*corev1.Secret, err error) {
	ss, err := SecretList(ns)
	if err != nil {
		return
	}
	for _, v := range ss {
		ns := v.GetMetadata().GetNamespace()
		if contains(excludens, ns) {
			continue
		}
		d := v.GetData()
		if _, ok := d[key]; ok {
			secrets = append(secrets, v)
		}
	}
	return
}

func contains(excludens []string, ns string) bool {
	if excludens == nil {
		return false
	}
	for _, v := range excludens {
		if ns == v {
			return true
		}
	}
	return false
}

func SecretListAll() (secrets []*corev1.Secret, err error) {
	return SecretList("")
}

func SecretList(ns string) (secrets []*corev1.Secret, err error) {
	var slist corev1.SecretList
	err = client.List(context.Background(), ns, &slist)
	if err != nil {
		err = fmt.Errorf("get secret err %v", err)
		return
	}
	secrets = slist.GetItems()
	return
}
