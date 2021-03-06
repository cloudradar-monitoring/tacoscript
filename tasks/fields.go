package tasks

const (
	TaskTypeCmdRun = "cmd.run"
	FileManaged    = "file.managed"
	PkgInstalled   = "pkg.installed"
	PkgRemoved     = "pkg.removed"
	PkgUpgraded    = "pkg.uptodate"

	NameField       = "name"
	NamesField      = "names"
	CwdField        = "cwd"
	UserField       = "user"
	ShellField      = "shell"
	EnvField        = "env"
	CreatesField    = "creates"
	RequireField    = "require"
	OnlyIf          = "onlyif"
	Unless          = "unless"
	SourceField     = "source"
	SourceHashField = "source_hash"
	MakeDirsField   = "makedirs"
	ReplaceField    = "replace"
	SkipVerifyField = "skip_verify"
	ContentsField   = "contents"
	GroupField      = "group"
	ModeField       = "mode"
	EncodingField   = "encoding"
	Version         = "version"
	Refresh         = "refresh"
)
