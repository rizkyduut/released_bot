package dbadapter

type DBAdapter interface {
	AddService(string) error
	DeleteService(string) error
	DeployService(string, string) error
	GetService(string) (string, error)

	AddGroupService(string, string) error
	DeleteGroupService(string, string) error
	GetAllServiceInGroup(string) ([]string, error)

	AddGroup(string) error
	DeleteGroup(string) error
	GetAllGroup() ([]string, error)

	GetGroupServicesList(gn string) ([]string, error)
}

type Config struct {
	Host string
	Password string
}

