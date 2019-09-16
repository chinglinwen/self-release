package nfs

import (
	"testing"

	"github.com/davecgh/go-spew/spew"
)

func TestRun(t *testing.T) {
	r := Parse(demotext)
	spew.Dump("result", r)
}

var demotext = `
# bigdata
/home/business_big_data/ 192.168.10.0/24(insecure,rw,no_root_squash,no_all_squash,sync)

# nginxLB
/data/data_other/nginxLB 172.31.83.0/24(insecure,rw,no_root_squash,no_all_squash,sync,anonuid=7373,anongid=7373)
`
