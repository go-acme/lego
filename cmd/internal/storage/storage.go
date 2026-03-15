package storage

type Storage struct {
	Certificate *CertificatesStorage
	Account     *AccountsStorage
	Archiver    *Archiver
}

func New(basePath string) *Storage {
	return &Storage{
		Certificate: NewCertificatesStorage(basePath),
		Account:     NewAccountsStorage(basePath),
		Archiver:    NewArchiver(basePath),
	}
}
