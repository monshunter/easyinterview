package objectstore

import "fmt"

type FactoryConfig struct {
	Provider       string
	FilesystemRoot string
	MinIO          MinIOConfig
}

func NewFromConfig(cfg FactoryConfig) (ObjectStore, error) {
	switch cfg.Provider {
	case "filesystem":
		return NewFilesystemStore(cfg.FilesystemRoot), nil
	case "minio":
		return NewMinIOStore(cfg.MinIO), nil
	default:
		return nil, fmt.Errorf("unsupported object storage provider %q", cfg.Provider)
	}
}
