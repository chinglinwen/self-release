package project

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"wen/self-release/git"

	"github.com/chinglinwen/log"

	"github.com/drone/envsubst"
)

// for env compare and env suffix for deployment
const (
	TEST   = "test"
	PRE    = "pre" // TODO: pre branch is same as master branch?
	ONLINE = "online"
)

// func GetEnvFromBranchOrCommitID(project, branch string, fromGitlab bool) string {
// 	if git.BranchIsTag(branch) {
// 		if git.BranchIsOnline(branch) {
// 			return ONLINE
// 		}
// 		if git.BranchIsPre(branch) {
// 			return PRE
// 		}
// 	}
// 	// if it's commitid, and from gitlab, it's a test env
// 	if fromGitlab {
// 		return TEST
// 	}

// 	// if it's from harbor ( means from human, or trx )
// 	//  we detect the last commitid to know the env
// 	//
// 	// check if branch is commitid, manual build should not use this type of tag
// 	if len(branch) == 8 {
// 		// should we check this? compatible with trx?
// 		o, p, err := git.GetLastTagCommitID(project)
// 		if err == nil {
// 			// pre comes out first, online later
// 			if p != "" && p == branch {
// 				return PRE
// 			}
// 			if o != "" && o == branch {
// 				return ONLINE
// 			}
// 		}
// 	}
// 	return TEST
// }

func GetEnvFromBranch(project, branch string) string {
	if git.BranchIsTag(branch) {
		if git.BranchIsOnline(branch) {
			return ONLINE
		}
		if git.BranchIsPre(branch) {
			return PRE
		}
	}
	return TEST
}

func GetCommitIDFromBranch(project, branch string) (commitid string, err error) {
	// not tag maybe it's commitid, or branch name?
	if !git.BranchIsTag(branch) {
		commitid = branch
		return
	}
	// what if tag not exist? filterout by harbor event
	// harbor event don't need this id
	return git.GetCommitIDFromTag(project, branch)
}

func GetProjectName(project string) (namespace, name string, err error) {
	if project == "" {
		err = fmt.Errorf("parse project empty err")
		return
	}
	s := strings.Split(project, "/")
	if len(s) != 2 {
		err = fmt.Errorf("parse project err, invalid project: %v(expect namespace/repo-name)", project)
		return
	}
	return s[0], s[1], nil
}

func getHashByFile(file string) (sum string, err error) {
	s, err := ioutil.ReadFile(file)
	if err != nil {
		err = fmt.Errorf("gethashbyfile err: %v", err)
		return
	}
	sum, err = getHash(string(s))
	return
}

func getHash(filebody string) (sum string, err error) {
	h := sha256.New()
	// if file

	_, err = h.Write([]byte(filebody))
	if err != nil {
		err = fmt.Errorf("gethash err: %v", err)
		return
	}
	sum = hex.EncodeToString(h.Sum(nil))
	return
}

func convertToSubst(templateBody string) string {
	// s := strings.ReplaceAll(templateBody, "{{", "")
	// s = strings.ReplaceAll(s, "}}", "")

	ciEnvPattern := regexp.MustCompile(`({{\s+\$(\w+)\s+}})`)
	s := ciEnvPattern.ReplaceAll([]byte(templateBody), []byte("${${2}}"))
	return string(s)
}

func generateByMap(templateBody string, envMap map[string]string) (string, error) {
	// for compatibilities
	templateBody = convertToSubst(templateBody)

	// inject resource here
	templateBody = injectResource(templateBody, envMap)

	return envsubst.Eval(templateBody, func(k string) string {
		if v, ok := envMap[k]; !ok {
			log.Printf("got unknown env config name: %v\n", k)
			return fmt.Sprintf("UNKNOWN-%v", k)
		} else {
			return v
		}
	})
}

func generateByEnv(templateBody string) (string, error) {
	return envsubst.EvalEnv(templateBody)
}

func injectResource(templateBody string, envMap map[string]string) string {
	if envMap["ENV-RESOURCE"] != "" {
		templateBody = strings.Replace(templateBody, "env:", fmt.Sprintf("env:\n%v", envMap["ENV-RESOURCE"]), -1)
	}
	if envMap["VOLUME-RESOURCE"] != "" {
		templateBody = strings.Replace(templateBody, "volumes:", fmt.Sprintf("volumes:\n%v", envMap["VOLUME-RESOURCE"]), -1)
	}
	if envMap["MOUNT-RESOURCE"] != "" {
		templateBody = strings.Replace(templateBody, "volumeMounts::", fmt.Sprintf("volumeMounts::\n%v", envMap["MOUNT-RESOURCE"]), -1)
	}

	return templateBody
}

func isExist(file string) bool {
	if _, err := os.Stat(file); !os.IsNotExist(err) {
		return true
	}
	return false
}

type resource struct {
	envs    string
	volumes string
	mounts  string
}

func (p *Project) readResources() (r resource, err error) {
	r.envs, err = p.readResource("envs.resource")
	if err != nil {
		log.Printf("%v get resource err: %v, will ignore", p.Project, err)
	}
	r.volumes, err = p.readResource("volumes.resource")
	if err != nil {
		log.Printf("%v get resource err: %v, will ignore", p.Project, err)
	}
	r.mounts, err = p.readResource("mounts.resource")
	if err != nil {
		log.Printf("%v get resource err: %v, will ignore", p.Project, err)
	}
	err = nil
	return
}

func (p *Project) readResource(name string) (s string, err error) {
	f := filepath.Join(p.configConfigPath, "template", name)
	if !p.configrepo.IsExist(f) {
		err = fmt.Errorf("%v is not exist", name)
		return
	}
	b, err := p.configrepo.GetFile(f)
	if err != nil {
		err = fmt.Errorf("try read %v err: %v", name, err)
		return
	}
	s = string(b)
	return
}
