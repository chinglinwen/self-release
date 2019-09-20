package project

import (
	"testing"
)

func TestValuesFileWriteAll(t *testing.T) {
	repo, err := NewValuesRepo("xindaiquan/base-service")
	if err != nil {
		t.Error("new repo", err)
		return
	}
	all, err := ParseAllValuesJson(demojsonall)
	if err != nil {
		t.Error("ParseAllValuesJson err", err)
		return
	}
	err = repo.ValuesFileWriteAll(all)
	if err != nil {
		t.Error("write", err)
		return
	}
}
func TestParseAllValuesJson(t *testing.T) {
	all, err := ParseAllValuesJson(demojsonall)
	if err != nil {
		t.Error("ParseAllValuesJson err", err)
		return
	}

	pretty(all)
}

// func pretty(a interface{}) {
// 	b, _ := json.MarshalIndent(a, "", "  ")
// 	fmt.Println("pretty: ", string(b))
// }

// func TestValuesJsonToYaml(t *testing.T) {
// 	out, err := ValuesJsonToYaml(demojson)
// 	if err != nil {
// 		t.Error("convertJsonToYaml err", err)
// 		return
// 	}

// 	fmt.Println("out: ", out)
// }

// func TestValuesYamlToJson(t *testing.T) {
// 	out, err := ValuesYamlToJson(demoyaml)
// 	if err != nil {
// 		t.Error("convertJsonToYaml err", err)
// 		return
// 	}

// 	fmt.Println("out: ", out)
// }

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

var demojsonall = `
{
	"online": {
	  "envs": {
		"EXAMPLE-KEY": "EXAMPLE-value",
		"b": "b2"
	  },
	  "codis": {
		"REDIS_HOST": "192.168.10.99",
		"REDIS_PORT": "7201",
		"SESSION_REDIS_HOST": "codis-proxy-flow-center-loanapi.codis-cluster",
		"SESSION_REDIS_PORT": "19000"
	  },
	  "mysql": [
		{
		  "name": "10-107-3307-liuliang",
		  "host": "DB_HOST",
		  "port": "DB_PORT",
		  "database": "DB_DATABASE",
		  "username": "DB_USERNAME",
		  "password": "DB_PASSWORD",
		  "id": 1
		}
	  ],
	  "nfs": [
		{
		  "name": "loanapi-public",
		  "path": "/data/staticfile_yjr/file_data/openapi",
		  "server": "172.31.83.26",
		  "mountPath": "/apps/loanapi/www/Publicdemo",
		  "id": 1
		}
	  ]
	},
	"pre": {
	  "envs": {},
	  "codis": {},
	  "mysql": [],
	  "nfs": []
	},
	"test": {
	  "envs": {
		"a": "a2"
	  },
	  "codis": {},
	  "mysql": [],
	  "nfs": []
	}
  }
  `
var demojson2 = `
{
	"online": {
	  "existMysql": [
		{
		  "name": "10-107-3307-liuliang",
		  "host": "DB_HOST",
		  "port": "DB_PORT",
		  "database": "DB_DATABASE",
		  "username": "DB_USERNAME",
		  "password": "DB_PASSWORD",
		  "id": 1
		}
	  ],
	  "existEnvs": [
		{
		  "name": "EXAMPLE-KEY",
		  "value": "EXAMPLE-value",
		  "id": 1
		},
		{
		  "name": "a",
		  "value": "b",
		  "id": 2
		}
	  ],
	  "existRedis": [
		{
		  "name": "codis-proxy-flow-center-loanapi.codis-cluster",
		  "host": "codis-proxy-flow-center-loanapi.codis-cluster",
		  "hostkey": "SESSION_REDIS_HOST",
		  "port": "19000",
		  "portkey": "SESSION_REDIS_PORT",
		  "id": 1
		},
		{
		  "name": "192.168.10.99",
		  "host": "192.168.10.99",
		  "hostkey": "REDIS_HOST",
		  "port": "7201",
		  "portkey": "REDIS_PORT",
		  "id": 2
		}
	  ],
	  "existNfs": [
		{
		  "name": "loanapi-public",
		  "path": "/data/staticfile_yjr/file_data/openapi",
		  "server": "172.31.83.26",
		  "mountPath": "/apps/loanapi/www/Publicdemo",
		  "id": 1
		}
	  ]
	},
	"pre": {
	  "existMysql": [],
	  "existEnvs": [],
	  "existRedis": [],
	  "existNfs": []
	},
	"test": {
	  "existMysql": [],
	  "existEnvs": [],
	  "existRedis": [],
	  "existNfs": []
	}
  }
  `