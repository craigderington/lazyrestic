package model

import "github.com/craigderington/lazyrestic/pkg/types"

// ConfigLoader interface for loading configuration
type ConfigLoader interface {
	LoadOrDefault(path string) *types.ResticConfig
}

// ResticClientFactory interface for creating restic clients
type ResticClientFactory interface {
	NewClient(config types.RepositoryConfig) ResticClient
}

// ResticClient interface for restic operations
type ResticClient interface {
	ListSnapshots() ([]types.Snapshot, error)
	ListFiles(snapshotID string, path string) ([]types.FileNode, error)
	GetRepositoryInfo() (*types.Repository, error)
}
