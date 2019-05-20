// template ops relate files
package template

// type K8s struct {
// 	Name     string
// 	Template string
// 	Config   string
// }

// // fetch config-deploy

// var (
// 	base string
// )

// // template: php.v1/k8s/online.yaml  // the name can be anything
// // template: php.v1/k8s/pre.yaml
// // config: _ops/config/online.yaml
// // config: _ops/config/pre.yaml
// func NewK8s(template, config string) {
// 	//read two yaml
// 	// merge them

// }

// we can't merge, as it's not a correct yaml, ( can we get full before? )
//expand all variable?

// https://github.com/drone/envsubst

// let config be a shell env file? to apply?

// only do the copy, let human modify and store it// by verify it first?
// 	//how to verify, if it need much env setting? // doing a real things to test ( or example project )

// much can be a setting ( env)
// if block need to change, try using a new template
//say multiple nfs item (just skip, to let human modify it), we can only improve the test

// the yaml can change anytime

// can we just do the copy, and finalize it there
// later may re-generate it with  overwrite setting

// verify it's working
// final result store into _ops/final?
