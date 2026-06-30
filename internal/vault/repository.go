package vault

type Options struct {
	Path          string
	Recipients    []string
	IdentityPaths []string
}

type Identity struct {
	Value     string
	Recipient string
}

type Repository interface {
	Path() string
	Load() (*Store, error)
	Save(*Store) error
	Read(func(*Store) error) error
	Update(func(*Store) error) error
}

type PlaintextMigrator interface {
	MigratePlaintext(sourcePath string, force bool) error
}

type Provider interface {
	Open(Options) (Repository, error)
	ReadOrCreateIdentity(path string) (Identity, error)
	CheckStatus(report *Report, options Options)
	CheckDoctor(report *Report, options Options)
}
