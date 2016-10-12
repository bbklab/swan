package scheduler

import "github.com/Dataman-Cloud/swan/types"

// type Registry interface {
// 	// Register a new task by ID in the memory registry
// 	Register(string, *types.Task) error
//
// 	// Tasks returns all tasks in the registry
// 	Tasks() ([]*types.Task, error)
//
// 	// Delete removes the task by ID from the registry
// 	Delete(string) error
//
// 	// Fetch returns a specific task in the registry by ID
// 	Fetch(string) (*types.Task, error)
//
// 	// Update finds the task in the registry for the ID and updates it's data
// 	Update(string, *types.Task) error
// }

type Registry interface {
	// RegisterFrameworkId is used to register the frameworkId in consul.
	RegisterFrameworkID(string) error

	// FrameworkIDHasRegistered is used to check whether the frameworkId has registered in consul.
	FrameworkIDHasRegistered(string) (bool, error)

	// RegisterApplication is used to register a application in consul.
	RegisterApplication(*types.Application) error

	// ListApplications is used to list all applications belong to a cluster or a user.
	ListApplications() ([]*types.Application, error)

	// FetchApplication is used to fetch application from consul by application id.
	FetchApplication(id string) (*types.Application, error)

	// DeleteApplication is used to delete application from consul by application id.
	DeleteApplication(id string) error
}
