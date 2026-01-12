package dependency

import (
	"testing"

	"github.com/stretchr/testify/require"
	"oaklab.hu/debian/deb822/types/arch"
	"oaklab.hu/debian/deb822/types/version"
)

func TestArchSetString(t *testing.T) {
	tests := []struct {
		name     string
		archSet  ArchSet
		expected string
	}{
		{
			name:     "empty architectures",
			archSet:  ArchSet{},
			expected: "",
		},
		{
			name: "single architecture",
			archSet: ArchSet{
				Architectures: []arch.Arch{arch.MustParse("amd64")},
			},
			expected: "[amd64]",
		},
		{
			name: "multiple architectures",
			archSet: ArchSet{
				Architectures: []arch.Arch{
					arch.MustParse("amd64"),
					arch.MustParse("i386"),
				},
			},
			expected: "[amd64 i386]",
		},
		{
			name: "single architecture with not",
			archSet: ArchSet{
				Not:           true,
				Architectures: []arch.Arch{arch.MustParse("amd64")},
			},
			expected: "[!amd64]",
		},
		{
			name: "multiple architectures with not",
			archSet: ArchSet{
				Not: true,
				Architectures: []arch.Arch{
					arch.MustParse("amd64"),
					arch.MustParse("i386"),
				},
			},
			expected: "[!amd64 !i386]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.archSet.String()
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestVersionRelationString(t *testing.T) {
	tests := []struct {
		name     string
		ver      VersionRelation
		expected string
	}{
		{
			name: "greater than or equal",
			ver: VersionRelation{
				Operator: ">=",
				Version:  version.Version{Epoch: 0, Version: "1.0"},
			},
			expected: "(>= 1.0)",
		},
		{
			name: "less than or equal",
			ver: VersionRelation{
				Operator: "<=",
				Version:  version.Version{Epoch: 0, Version: "2.0"},
			},
			expected: "(<= 2.0)",
		},
		{
			name: "exactly equal",
			ver: VersionRelation{
				Operator: "=",
				Version:  version.Version{Epoch: 0, Version: "1.5"},
			},
			expected: "(= 1.5)",
		},
		{
			name: "strictly earlier",
			ver: VersionRelation{
				Operator: "<<",
				Version:  version.Version{Epoch: 0, Version: "3.0"},
			},
			expected: "(<< 3.0)",
		},
		{
			name: "strictly later",
			ver: VersionRelation{
				Operator: ">>",
				Version:  version.Version{Epoch: 0, Version: "0.5"},
			},
			expected: "(>> 0.5)",
		},
		{
			name: "version with epoch",
			ver: VersionRelation{
				Operator: ">=",
				Version:  version.Version{Epoch: 1, Version: "2.0"},
			},
			expected: "(>= 1:2.0)",
		},
		{
			name: "version with revision",
			ver: VersionRelation{
				Operator: "=",
				Version:  version.Version{Epoch: 0, Version: "1.0", Revision: "1"},
			},
			expected: "(= 1.0-1)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.ver.String()
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestStageString(t *testing.T) {
	tests := []struct {
		name     string
		stage    Stage
		expected string
	}{
		{
			name: "stage without not",
			stage: Stage{
				Not:  false,
				Name: "build",
			},
			expected: "build",
		},
		{
			name: "stage with not",
			stage: Stage{
				Not:  true,
				Name: "build",
			},
			expected: "!build",
		},
		{
			name: "host stage without not",
			stage: Stage{
				Not:  false,
				Name: "host",
			},
			expected: "host",
		},
		{
			name: "host stage with not",
			stage: Stage{
				Not:  true,
				Name: "host",
			},
			expected: "!host",
		},
		{
			name: "empty stage name",
			stage: Stage{
				Not:  false,
				Name: "",
			},
			expected: "",
		},
		{
			name: "empty stage name with not",
			stage: Stage{
				Not:  true,
				Name: "",
			},
			expected: "!",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.stage.String()
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestStageSetString(t *testing.T) {
	tests := []struct {
		name     string
		stageSet StageSet
		expected string
	}{
		{
			name:     "empty stages",
			stageSet: StageSet{},
			expected: "",
		},
		{
			name: "single stage",
			stageSet: StageSet{
				Stages: []Stage{
					{Not: false, Name: "build"},
				},
			},
			expected: "<build>",
		},
		{
			name: "multiple stages",
			stageSet: StageSet{
				Stages: []Stage{
					{Not: false, Name: "build"},
					{Not: false, Name: "host"},
				},
			},
			expected: "<build host>",
		},
		{
			name: "stage with not",
			stageSet: StageSet{
				Stages: []Stage{
					{Not: true, Name: "build"},
				},
			},
			expected: "<!build>",
		},
		{
			name: "multiple stages with mixed not",
			stageSet: StageSet{
				Stages: []Stage{
					{Not: false, Name: "build"},
					{Not: true, Name: "host"},
				},
			},
			expected: "<build !host>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.stageSet.String()
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestPossibilityString(t *testing.T) {
	tests := []struct {
		name        string
		possibility Possibility
		expected    string
	}{
		{
			name: "name only",
			possibility: Possibility{
				Name: "foo",
			},
			expected: "foo",
		},
		{
			name: "name with architecture qualifier",
			possibility: Possibility{
				Name: "foo",
				Arch: &arch.Arch{CPU: "amd64", OS: "any", ABI: "any"},
			},
			expected: "foo:amd64",
		},
		{
			name: "name with architecture restrictions",
			possibility: Possibility{
				Name: "foo",
				Architectures: &ArchSet{
					Architectures: []arch.Arch{arch.MustParse("amd64")},
				},
			},
			expected: "foo [amd64]",
		},
		{
			name: "name with version relation",
			possibility: Possibility{
				Name: "foo",
				Version: &VersionRelation{
					Operator: ">=",
					Version:  version.Version{Version: "1.0"},
				},
			},
			expected: "foo (>= 1.0)",
		},
		{
			name: "name with single stage set",
			possibility: Possibility{
				Name: "foo",
				StageSets: []StageSet{
					{
						Stages: []Stage{{Not: false, Name: "build"}},
					},
				},
			},
			expected: "foo <build>",
		},
		{
			name: "name with multiple stage sets",
			possibility: Possibility{
				Name: "foo",
				StageSets: []StageSet{
					{
						Stages: []Stage{{Not: false, Name: "build"}},
					},
					{
						Stages: []Stage{{Not: false, Name: "host"}},
					},
				},
			},
			expected: "foo <build> <host>",
		},
		{
			name: "complete possibility",
			possibility: Possibility{
				Name: "foo",
				Arch: &arch.Arch{CPU: "amd64", OS: "any", ABI: "any"},
				Architectures: &ArchSet{
					Architectures: []arch.Arch{arch.MustParse("i386")},
				},
				Version: &VersionRelation{
					Operator: ">=",
					Version:  version.Version{Version: "1.0"},
				},
				StageSets: []StageSet{
					{
						Stages: []Stage{{Not: false, Name: "build"}},
					},
				},
			},
			expected: "foo:amd64 [i386] (>= 1.0) <build>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.possibility.String()
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestRelationString(t *testing.T) {
	tests := []struct {
		name     string
		relation Relation
		expected string
	}{
		{
			name:     "empty possibilities",
			relation: Relation{},
			expected: "",
		},
		{
			name: "single possibility",
			relation: Relation{
				Possibilities: []Possibility{
					{Name: "foo"},
				},
			},
			expected: "foo",
		},
		{
			name: "multiple possibilities",
			relation: Relation{
				Possibilities: []Possibility{
					{Name: "foo"},
					{Name: "bar"},
				},
			},
			expected: "foo | bar",
		},
		{
			name: "possibilities with version constraints",
			relation: Relation{
				Possibilities: []Possibility{
					{
						Name: "foo",
						Version: &VersionRelation{
							Operator: ">=",
							Version:  version.Version{Version: "1.0"},
						},
					},
					{
						Name: "bar",
						Version: &VersionRelation{
							Operator: "<=",
							Version:  version.Version{Version: "2.0"},
						},
					},
				},
			},
			expected: "foo (>= 1.0) | bar (<= 2.0)",
		},
		{
			name: "possibilities with architecture restrictions",
			relation: Relation{
				Possibilities: []Possibility{
					{
						Name: "foo",
						Architectures: &ArchSet{
							Architectures: []arch.Arch{arch.MustParse("amd64")},
						},
					},
					{
						Name: "bar",
						Architectures: &ArchSet{
							Architectures: []arch.Arch{arch.MustParse("i386")},
						},
					},
				},
			},
			expected: "foo [amd64] | bar [i386]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.relation.String()
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestDependencyString(t *testing.T) {
	tests := []struct {
		name string
		dep  Dependency
		want string
	}{
		{
			name: "empty dependency",
			dep:  Dependency{},
			want: "",
		},
		{
			name: "single relation single possibility",
			dep: Dependency{
				Relations: []Relation{
					{
						Possibilities: []Possibility{
							{Name: "foo"},
						},
					},
				},
			},
			want: "foo",
		},
		{
			name: "multiple relations",
			dep: Dependency{
				Relations: []Relation{
					{
						Possibilities: []Possibility{
							{Name: "foo"},
						},
					},
					{
						Possibilities: []Possibility{
							{
								Name: "bar",
								Version: &VersionRelation{
									Operator: ">=",
									Version:  version.Version{Version: "1.0"},
								},
							},
							{
								Name: "baz",
								Architectures: &ArchSet{
									Architectures: []arch.Arch{arch.MustParse("amd64")},
								},
							},
						},
					},
				},
			},
			want: "foo, bar (>= 1.0) | baz [amd64]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, tt.dep.String())
		})
	}
}

func TestDependencyMarshalText(t *testing.T) {
	dep := Dependency{
		Relations: []Relation{
			{
				Possibilities: []Possibility{
					{Name: "foo"},
				},
			},
		},
	}

	text, err := dep.MarshalText()
	require.NoError(t, err)
	require.Equal(t, "foo", string(text))
}

func TestDependencyUnmarshalText(t *testing.T) {
	var dep Dependency
	err := dep.UnmarshalText([]byte("foo"))
	require.NoError(t, err)
	expected := Dependency{
		Relations: []Relation{
			{
				Possibilities: []Possibility{
					{
						Name: "foo",
						Architectures: &ArchSet{
							Architectures: []arch.Arch{},
						},
						StageSets: []StageSet{},
					},
				},
			},
		},
	}
	require.Equal(t, expected, dep)
}

func TestSourceString(t *testing.T) {
	tests := []struct {
		name string
		src  Source
		want string
	}{
		{
			name: "without version",
			src: Source{
				Name: "pkg",
			},
			want: "pkg",
		},
		{
			name: "with version",
			src: Source{
				Name: "pkg",
				Version: &version.Version{
					Epoch:    1,
					Version:  "2.0",
					Revision: "1",
				},
			},
			want: "pkg (1:2.0-1)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, tt.src.String())
		})
	}
}

func TestSourceMarshalText(t *testing.T) {
	src := Source{
		Name: "example-package",
		Version: &version.Version{
			Epoch:    0,
			Version:  "1.2.3",
			Revision: "4",
		},
	}

	text, err := src.MarshalText()
	require.NoError(t, err)
	require.Equal(t, "example-package (1.2.3-4)", string(text))
}

func TestSourceUnmarshalText(t *testing.T) {
	var src Source
	err := src.UnmarshalText([]byte("2048-qt (0.1.6-2)"))
	require.NoError(t, err)

	expectedVersion := version.MustParse("0.1.6-2")
	expected := Source{
		Name:    "2048-qt",
		Version: &expectedVersion,
	}

	require.Equal(t, expected, src)
}
