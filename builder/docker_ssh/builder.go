package docker_ssh

import (
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/builder/docker"
	"github.com/mitchellh/packer/builder/null"
	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/packer"
	"log"
	"reflect"
	"time"
)

const BuilderId = "packer.docker"

type Builder struct {
	config *Config
	runner multistep.Runner
}

func (b *Builder) Prepare(raws ...interface{}) ([]string, error) {

	for k, v := range raws {
		log.Println("Type: ", reflect.TypeOf(v))
		for l, u := range v {
			log.Println("key:", l, "value:", u)
		}
	}

	c, warnings, errs := NewConfig(raws...)
	if errs != nil {
		return warnings, errs
	}
	b.config = c

	return warnings, nil
}

func (b *Builder) Run(ui packer.Ui, hook packer.Hook, cache packer.Cache) (packer.Artifact, error) {
	driver := &DockerDriver{Tpl: b.config.tpl, Ui: ui}
	if err := driver.Verify(); err != nil {
		return nil, err
	}

	steps := []multistep.Step{
		&docker.StepTempDir{},
		&docker.StepPull{},
		&StepRun{},
		&common.StepConnectSSH{
			SSHAddress:     SSHAddress(b.config.Port),
			SSHConfig:      null.SSHConfig(b.config.SSHUsername, b.config.SSHPassword, b.config.SSHPrivateKeyFile),
			SSHWaitTimeout: 1 * time.Minute,
		},
		&common.StepProvision{},
		&StepExport{},
	}

	// // Convert docker_ssh.Config to docker.Config
	// log.Println("ExportPath: ", b.config.ExportPath)
	// log.Println("Image: ", b.config.Image)
	// log.Println("Pull: ", b.config.Pull)
	// log.Println("RunCommand: ", b.config.RunCommand)

	// raw_config := map[string]interface{}{
	// 	"ExportPath": b.config.ExportPath,
	// 	"Image":
	// }

	// docker_config, _, _ := docker.NewConfig()

	// log.Println("ExportPath: ", docker_config.ExportPath)
	// log.Println("Image: ", docker_config.Image)
	// log.Println("Pull: ", docker_config.Pull)
	// log.Println("RunCommand: ", docker_config.RunCommand)

	// Setup the state bag and initial state for the steps
	state := new(multistep.BasicStateBag)
	state.Put("config", b.config)
	state.Put("hook", hook)
	state.Put("ui", ui)

	// Setup the driver that will talk to Docker
	state.Put("driver", driver)

	// Run!
	if b.config.PackerDebug {
		b.runner = &multistep.DebugRunner{
			Steps:   steps,
			PauseFn: common.MultistepDebugFn(ui),
		}
	} else {
		b.runner = &multistep.BasicRunner{Steps: steps}
	}

	b.runner.Run(state)

	// If there was an error, return that
	if rawErr, ok := state.GetOk("error"); ok {
		return nil, rawErr.(error)
	}

	// No errors, must've worked
	artifact := &ExportArtifact{path: b.config.ExportPath}
	return artifact, nil
}

func (b *Builder) Cancel() {
	if b.runner != nil {
		log.Println("Cancelling the step runner...")
		b.runner.Cancel()
	}
}
