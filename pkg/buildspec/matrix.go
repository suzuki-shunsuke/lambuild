package buildspec

type Matrix struct {
	Static  MatrixStatic
	Dynamic MatrixDynamic
}

type LMatrix struct {
	Static  MatrixStatic
	Dynamic LMatrixDynamic
}

type MatrixStatic struct {
	IgnoreFailure bool            `yaml:"ignore-failure,omitempty"`
	Env           MatrixStaticEnv `yaml:",omitempty"`
}

type MatrixFilter struct {
	Value  string
	Filter LambuildFilter
}

type MatrixDynamic struct {
	Buildspec []string         `yaml:",omitempty"`
	Env       MatrixDynamicEnv `yaml:",omitempty"`
}

type LMatrixDynamic struct {
	Buildspec []MatrixFilter   `yaml:",omitempty"`
	Env       MatrixDynamicEnv `yaml:",omitempty"`
}

type MatrixDynamicEnv struct {
	ComputeType []string            `yaml:"compute-type,omitempty"`
	Image       []string            `yaml:",omitempty"`
	Variables   map[string][]string `yaml:",omitempty"`
}

type LMatrixDynamicEnv struct {
	ComputeType []MatrixFilter            `yaml:"compute-type,omitempty"`
	Image       []MatrixFilter            `yaml:",omitempty"`
	Variables   map[string][]MatrixFilter `yaml:",omitempty"`
}

type MatrixStaticEnv struct {
	ComputeType    string            `yaml:"compute-type,omitempty"`
	Image          string            `yaml:",omitempty"`
	Type           string            `yaml:",omitempty"`
	Variables      map[string]string `yaml:",omitempty"`
	PrivilegedMode bool              `yaml:"privileged-mode,omitempty"`
}

func (matrix *Matrix) Empty() bool {
	if len(matrix.Dynamic.Buildspec) != 0 {
		return false
	}
	if len(matrix.Dynamic.Env.ComputeType) != 0 {
		return false
	}
	if len(matrix.Dynamic.Env.Image) != 0 {
		return false
	}
	if len(matrix.Dynamic.Env.Variables) != 0 {
		return false
	}
	return true
}
