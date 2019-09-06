package resource

import (
	"fmt"
	"strconv"
	"wen/self-release/pkg/k8s"

	corev1 "github.com/ericchiang/k8s/apis/core/v1"
)

type Mysql struct {
	Name string `json:"name"` //"10-107-3307-liuliang", we only need name list
	// Host     string //"DB_HOST",
	// Port     string //"DB_PORT",
	// Database string //"DB_DATABASE",
	// Username string //"DB_USERNAME",
	// password string //"DB_PASSWORD", we dont send password
}

type Codis struct {
	Host string `json:"host"`
	Port string `json:"port"`
}

type Nfs struct {
	name   string `json:"name"`   //"loanapi-public",
	path   string `json:"path"`   //"/data/staticfile_yjr/file_data/openapi",
	server string `json:"server"` //"172.31.83.26",
	// mountPath string // "/apps/loanapi/www/Public", user provide this
}

type Resource struct {
	Mysql []Mysql `json:"mysql"`
	Codis []Codis `json:"codis"`
}

func GetResource(ns string) (r Resource, err error) {
	mysqls, err := k8s.SecretListAllWithHasKey(ns, "database", nil)
	if err != nil {
		return
	}
	r.Mysql, err = convertMysql(mysqls)
	if err != nil {
		return
	}

	l := map[string]string{"codis-component": "proxy"}
	codises, err := k8s.ServiceListWithLabels("codis-cluster", l)
	if err != nil {
		return
	}
	r.Codis, err = convertCodis(codises)
	if err != nil {
		return
	}
	return
}

func convertMysql(secrets []*corev1.Secret) (ms []Mysql, err error) {
	for _, v := range secrets {

		m, e := decodeMysql(v)
		if e != nil {

			err = fmt.Errorf("decode %v/%v err: %v", v.GetMetadata().GetNamespace(), v.GetMetadata().GetName(), e)
			return
		}
		ms = append(ms, m)
	}
	return
}

func convertCodis(services []*corev1.Service) (ms []Codis, err error) {
	for _, v := range services {
		var port int32
		for _, x := range v.GetSpec().GetPorts() {
			if x.GetName() == "proxy" {
				port = x.GetPort()
			}
		}
		host := fmt.Sprintf("%v.%v", v.GetMetadata().GetName(), v.GetMetadata().GetNamespace())
		m := Codis{
			Host: host,
			Port: strconv.Itoa(int(port)),
		}
		ms = append(ms, m)
	}
	return
}

func decodeMysql(s *corev1.Secret) (m Mysql, err error) {
	name := s.GetMetadata().GetName()
	d := s.GetData()
	if d == nil {
		err = fmt.Errorf("empty data")
		return
	}

	if v, ok := d["database"]; !ok || len(v) == 0 {
		err = fmt.Errorf("database is empty or not exist")
		return
	}
	if v, ok := d["host"]; !ok || len(v) == 0 {
		err = fmt.Errorf("host is empty or not exist")
		return
	}
	if v, ok := d["port"]; !ok || len(v) == 0 {
		err = fmt.Errorf("port is empty or not exist")
		return
	}
	if v, ok := d["username"]; !ok || len(v) == 0 {
		err = fmt.Errorf("username is empty or not exist")
		return
	}
	if v, ok := d["password"]; !ok || len(v) == 0 {
		err = fmt.Errorf("password is empty or not exist")
		return
	}
	m = Mysql{Name: name}
	return
}
