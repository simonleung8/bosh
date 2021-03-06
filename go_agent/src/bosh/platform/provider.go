package platform

import (
	boshdisk "bosh/platform/disk"
	boshsys "bosh/system"
	"errors"
	"fmt"
)

type provider struct {
	platforms map[string]Platform
}

func NewProvider(fs boshsys.FileSystem) (p provider) {
	runner := boshsys.ExecCmdRunner{}
	partitioner := boshdisk.NewSfdiskPartitioner(runner)
	formatter := boshdisk.NewLinuxFormatter(runner)

	p.platforms = map[string]Platform{
		"ubuntu": newUbuntuPlatform(fs, runner, partitioner, formatter),
		"dummy":  newDummyPlatform(),
	}
	return
}

func (p provider) Get(name string) (plat Platform, err error) {
	plat, found := p.platforms[name]

	if !found {
		err = errors.New(fmt.Sprintf("Platform %s could not be found", name))
	}
	return
}
