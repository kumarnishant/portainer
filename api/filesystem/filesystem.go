package filesystem

import (
	"bytes"
	"io/ioutil"

	"github.com/portainer/portainer"

	"io"
	"os"
	"path"
)

const (
	// TLSStorePath represents the subfolder where TLS files are stored in the file store folder.
	TLSStorePath = "tls"
	// LDAPStorePath represents the subfolder where LDAP TLS files are stored in the TLSStorePath.
	LDAPStorePath = "ldap"
	// TLSCACertFile represents the name on disk for a TLS CA file.
	TLSCACertFile = "ca.pem"
	// TLSCertFile represents the name on disk for a TLS certificate file.
	TLSCertFile = "cert.pem"
	// TLSKeyFile represents the name on disk for a TLS key file.
	TLSKeyFile = "key.pem"
	// ComposeStorePath represents the subfolder where compose files are stored in the file store folder.
	ComposeStorePath = "compose"
	// ComposeFileDefaultName represents the default name of a compose file.
	ComposeFileDefaultName = "docker-compose.yml"
)

// Service represents a service for managing files and directories.
type Service struct {
	dataStorePath string
	fileStorePath string
}

// NewService initializes a new service. It creates a data directory and a directory to store files
// inside this directory if they don't exist.
func NewService(dataStorePath, fileStorePath string) (*Service, error) {
	service := &Service{
		dataStorePath: dataStorePath,
		fileStorePath: path.Join(dataStorePath, fileStorePath),
	}

	err := os.MkdirAll(dataStorePath, 0755)
	if err != nil {
		return nil, err
	}

	err = service.createDirectoryInStore(TLSStorePath)
	if err != nil {
		return nil, err
	}

	err = service.createDirectoryInStore(ComposeStorePath)
	if err != nil {
		return nil, err
	}

	return service, nil
}

// RemoveDirectory removes a directory on the filesystem.
func (service *Service) RemoveDirectory(directoryPath string) error {
	return os.RemoveAll(directoryPath)
}

// GetStackProjectPath returns the absolute path on the FS for a stack based
// on its identifier.
func (service *Service) GetStackProjectPath(stackIdentifier string) string {
	return path.Join(service.fileStorePath, ComposeStorePath, stackIdentifier)
}

// StoreStackFileFromString creates a subfolder in the ComposeStorePath and stores a new file using the content from a string.
// It returns the path to the folder where the file is stored.
func (service *Service) StoreStackFileFromString(stackIdentifier, fileName, stackFileContent string) (string, error) {
	stackStorePath := path.Join(ComposeStorePath, stackIdentifier)
	err := service.createDirectoryInStore(stackStorePath)
	if err != nil {
		return "", err
	}

	composeFilePath := path.Join(stackStorePath, fileName)
	data := []byte(stackFileContent)
	r := bytes.NewReader(data)

	err = service.createFileInStore(composeFilePath, r)
	if err != nil {
		return "", err
	}

	return path.Join(service.fileStorePath, stackStorePath), nil
}

// StoreStackFileFromReader creates a subfolder in the ComposeStorePath and stores a new file using the content from an io.Reader.
// It returns the path to the folder where the file is stored.
func (service *Service) StoreStackFileFromReader(stackIdentifier, fileName string, r io.Reader) (string, error) {
	stackStorePath := path.Join(ComposeStorePath, stackIdentifier)
	err := service.createDirectoryInStore(stackStorePath)
	if err != nil {
		return "", err
	}

	composeFilePath := path.Join(stackStorePath, fileName)

	err = service.createFileInStore(composeFilePath, r)
	if err != nil {
		return "", err
	}

	return path.Join(service.fileStorePath, stackStorePath), nil
}

// StoreTLSFile creates a folder in the TLSStorePath and stores a new file with the content from r.
func (service *Service) StoreTLSFile(folder string, fileType portainer.TLSFileType, r io.Reader) error {
	storePath := path.Join(TLSStorePath, folder)
	err := service.createDirectoryInStore(storePath)
	if err != nil {
		return err
	}

	var fileName string
	switch fileType {
	case portainer.TLSFileCA:
		fileName = TLSCACertFile
	case portainer.TLSFileCert:
		fileName = TLSCertFile
	case portainer.TLSFileKey:
		fileName = TLSKeyFile
	default:
		return portainer.ErrUndefinedTLSFileType
	}

	tlsFilePath := path.Join(storePath, fileName)
	err = service.createFileInStore(tlsFilePath, r)
	if err != nil {
		return err
	}
	return nil
}

// GetPathForTLSFile returns the absolute path to a specific TLS file for an endpoint.
func (service *Service) GetPathForTLSFile(folder string, fileType portainer.TLSFileType) (string, error) {
	var fileName string
	switch fileType {
	case portainer.TLSFileCA:
		fileName = TLSCACertFile
	case portainer.TLSFileCert:
		fileName = TLSCertFile
	case portainer.TLSFileKey:
		fileName = TLSKeyFile
	default:
		return "", portainer.ErrUndefinedTLSFileType
	}
	return path.Join(service.fileStorePath, TLSStorePath, folder, fileName), nil
}

// DeleteTLSFiles deletes a folder in the TLS store path.
func (service *Service) DeleteTLSFiles(folder string) error {
	storePath := path.Join(service.fileStorePath, TLSStorePath, folder)
	err := os.RemoveAll(storePath)
	if err != nil {
		return err
	}
	return nil
}

// DeleteTLSFile deletes a specific TLS file from a folder.
func (service *Service) DeleteTLSFile(folder string, fileType portainer.TLSFileType) error {
	var fileName string
	switch fileType {
	case portainer.TLSFileCA:
		fileName = TLSCACertFile
	case portainer.TLSFileCert:
		fileName = TLSCertFile
	case portainer.TLSFileKey:
		fileName = TLSKeyFile
	default:
		return portainer.ErrUndefinedTLSFileType
	}

	filePath := path.Join(service.fileStorePath, TLSStorePath, folder, fileName)

	err := os.Remove(filePath)
	if err != nil {
		return err
	}
	return nil
}

// GetFileContent returns a string content from file.
func (service *Service) GetFileContent(filePath string) (string, error) {
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	return string(content), nil
}

// createDirectoryInStore creates a new directory in the file store
func (service *Service) createDirectoryInStore(name string) error {
	path := path.Join(service.fileStorePath, name)
	return os.MkdirAll(path, 0700)
}

// createFile creates a new file in the file store with the content from r.
func (service *Service) createFileInStore(filePath string, r io.Reader) error {
	path := path.Join(service.fileStorePath, filePath)

	out, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, r)
	if err != nil {
		return err
	}

	return nil
}
