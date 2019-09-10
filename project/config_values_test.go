package project

import (
	"fmt"
	"testing"
)

func TestValuesJsonToYaml(t *testing.T) {
	out, err := ValuesJsonToYaml(demojson)
	if err != nil {
		t.Error("convertJsonToYaml err", err)
		return
	}

	fmt.Println("out: ", out)
}

func TestValuesYamlToJson(t *testing.T) {
	out, err := ValuesYamlToJson(demoyaml)
	if err != nil {
		t.Error("convertJsonToYaml err", err)
		return
	}

	fmt.Println("out: ", out)
}

var demoyaml = `
envs:
  EXAMPLE-KEY: EXAMPLE-value
mysql:
- name: 10-107-3307-liuliang
  host: DB_HOST
  port: DB_PORT
  database: DB_DATABASE
  username: DB_USERNAME
  password: DB_PASSWORD
codis:
  REDIS_HOST: 192.168.10.99
  REDIS_PORT: "7201"
  SESSION_REDIS_HOST: codis-proxy-flow-center-loanapi.codis-cluster
  SESSION_REDIS_PORT: "19000"
nfs:
- name: loanapi-public
  path: /data/staticfile_yjr/file_data/openapi
  server: 172.31.83.26
  mountpath: /apps/loanapi/www/Public
  `
var demojson = `
{
	"envs": {
	   "EXAMPLE-KEY": "EXAMPLE-value"
	},
	"mysql": [
	   {
		  "name": "10-107-3307-liuliang",
		  "host": "DB_HOST",
		  "port": "DB_PORT",
		  "database": "DB_DATABASE",
		  "username": "DB_USERNAME",
		  "password": "DB_PASSWORD"
	   }
	],
	"codis": {
	   "SESSION_REDIS_HOST": "codis-proxy-flow-center-loanapi.codis-cluster",
	   "SESSION_REDIS_PORT": "19000",
	   "REDIS_HOST": "192.168.10.99",
	   "REDIS_PORT": "7201"
	},
	"nfs": [
	   {
		  "name": "loanapi-public",
		  "path": "/data/staticfile_yjr/file_data/openapi",
		  "server": "172.31.83.26",
		  "mountPath": "/apps/loanapi/www/Public"
	   }
	]
 }
`
