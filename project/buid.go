package project

// now it has repo store local

// build image, by dockerfile, if not exist using template to generate one
// doing in init?

// some will be combined, how to import?

// for env, we may define many template ( from secret ), or just directly writtin in yaml?
//can we evolve to use a single env? it's compatible
//keep the indent of original

//template with config?

//define a general template, how to do the setting?
// config-online.env for different branch

// as a opt in process, along with manual copy
// genereate base, for change, but easier modify?

// a modified template

//how to merge? using git merge
//https://github.com/mikefarah/yq

//   https://github.com/pixelb/crudini
//   # merge an ini file from another ini
//   crudini --merge config_file < another.ini

// https://ini.unknwon.io/

// provide an api /url to post the form for verify?

// human or hook-listener or cli do the call

// make it general for all template?

// build assume file is ready? ( verified, and tested )

//  resources:
//  requests:
//    cpu: 0.5
//    memory: 512M
//  limits:
//    cpu: 2
//    memory: 4G

// fetch

// some pre-steps?

// build image

// https://knative.dev/docs/serving/samples/source-to-url-go/

// provide a templat build script?  using local one, or embed

// what info they provide
// project, or just dockerfile? target name

// var configRepo = "yunwei/config-deploy" //"http://g.haodai.net/yunwei/config-deploy.git"

var buildBody = `
#build.sh
echo start building ...
env
ls -la .
echo build ok
`
