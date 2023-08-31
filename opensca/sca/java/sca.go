package java

import (
	"context"
	"io"
	"strings"

	"github.com/xmirrorsecurity/opensca-cli/opensca/model"
	"github.com/xmirrorsecurity/opensca-cli/opensca/sca/filter"
)

type Sca struct {
	NotUseMvn bool
}

func (sca Sca) Language() model.Language {
	return model.Lan_Java
}

func (sca Sca) Filter(relpath string) bool {
	return filter.JavaPom(relpath)
}

func (sca Sca) Sca(ctx context.Context, parent *model.File, files []*model.File) []*model.DepGraph {

	// jar包中的pom仅读取pom自身信息 不获取子依赖
	if strings.Contains(parent.Relpath, ".jar") {
		var deps []*model.DepGraph
		for _, file := range files {
			if !filter.JavaPom(file.Relpath) {
				continue
			}
			file.OpenReader(func(reader io.Reader) {
				p := ReadPom(reader)
				if !p.Check() {
					return
				}
				deps = append(deps, &model.DepGraph{
					Vendor:  p.GroupId,
					Name:    p.ArtifactId,
					Version: p.Version,
					Path:    file.Relpath,
				})
			})
		}
		return deps
	}

	// 调用mvn
	if !sca.NotUseMvn {
		deps := MvnTree(parent)
		if len(deps) > 0 {
			return deps
		}
	}

	// 静态解析
	poms := []*Pom{}
	for _, file := range files {
		if filter.JavaPom(file.Relpath) {
			file.OpenReader(func(reader io.Reader) {
				pom := ReadPom(reader)
				pom.File = file
				poms = append(poms, pom)
			})
		}
	}
	return ParsePoms(poms)
}

type MvnRepo struct {
	Url      string `json:"url" xml:"url"`
	Username string
	Password string
}

var defaultRepo []MvnRepo

func RegisterRepo(repos ...MvnRepo) {
	defaultRepo = append(repos, repos...)
}
