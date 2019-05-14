package init

// trigger init based on tag or _ops directory( need manual create? )
//let's based on tag text

// create contents in repo

// _ops

type Template struct {
	dockerfile string
	k8sfile    string
}

func New() {

	if tmplfile == "" {
		// using default one
	}

}

// copy file from template
func DockerFile() {

}

// check if dockerfile exist, if not create one from template ( dockertemplate: php default)
//template must be exist, before ( manual written )
//template need to easy testing ( by a curl ), or provide with repo for test the whole?

// init k8s template, with final yaml (for customize, suggest to customize from _ops )?
//prepare k8s template from config-deploy top directory?  template/php.v1.template
// ops can specify different k8stemplate: filename

//copy template to config's projects path, // do we need template? people need customize template?
//let's copy if first

// init deploy.sh // why this?

// update genereated files, verify them
