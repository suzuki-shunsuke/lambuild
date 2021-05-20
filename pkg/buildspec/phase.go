package buildspec

type Phase struct {
	Commands Commands               `yaml:",omitempty"`
	Finally  Commands               `yaml:",omitempty"`
	Map      map[string]interface{} `yaml:",inline,omitempty"`
}

func (phase *Phase) Filter(param interface{}) (map[string]interface{}, error) {
	m := make(map[string]interface{}, len(phase.Map)+2) //nolint:gomnd
	for k, v := range phase.Map {
		m[k] = v
	}

	if len(phase.Commands) != 0 {
		cmds, err := phase.Commands.Filter(param)
		if err != nil {
			return nil, err
		}
		m["commands"] = cmds
	}

	if len(phase.Finally) != 0 {
		cmds, err := phase.Finally.Filter(param)
		if err != nil {
			return nil, err
		}
		m["finally"] = cmds
	}

	return m, nil
}

func (phases *Phases) Filter(param interface{}) (map[string]interface{}, error) {
	m := make(map[string]interface{}, 4)

	install, err := phases.Install.Filter(param)
	if err != nil {
		return nil, err
	}
	if len(install) != 0 {
		m["install"] = install
	}

	preBuild, err := phases.PreBuild.Filter(param)
	if err != nil {
		return nil, err
	}
	if len(preBuild) != 0 {
		m["pre_build"] = preBuild
	}

	build, err := phases.Build.Filter(param)
	if err != nil {
		return nil, err
	}
	if len(preBuild) != 0 {
		m["build"] = build
	}

	postBuild, err := phases.PostBuild.Filter(param)
	if err != nil {
		return nil, err
	}
	if len(postBuild) != 0 {
		m["post_build"] = postBuild
	}

	return m, nil
}
