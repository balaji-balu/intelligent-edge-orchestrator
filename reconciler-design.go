package reconciler

type ComponentSpec struct {
	Name       string `yaml:"name"`
	Properties struct {
		// Compose profile
		PackageURL string `yaml:"packageLocation,omitempty"`
		KeyURL     string `yaml:"keyLocation,omitempty"`

		// Helm profile
		Repository string `yaml:"repository,omitempty"`
		Revision   string `yaml:"revision,omitempty"`
		Wait       *bool  `yaml:"wait,omitempty"`
		Timeout    string `yaml:"timeout,omitempty"`

		NodeSelector map[string]string `yaml:"nodeSelector,omitempty"`
	} `yaml:"properties"`
}

type ApplicationDeployment struct {
	Metadata struct {
		Annotations struct {
			ID string `yaml:"id"`
		} `yaml:"annotations"`
	} `yaml:"metadata"`

	Spec struct {
		DeploymentProfile struct {
			Type       string          `yaml:"type"`
			Components []ComponentSpec `yaml:"components"`
		} `yaml:"deploymentProfile"`
	} `yaml:"spec"`
}

type DeploymentComponentStatus struct {
	ID            uuid.UUID
	DeploymentID  string
	HostID        string
	ComponentName string
	DesiredHash   string
	ActualHash    string
	Status        string
	Message       string
	LastUpdate    time.Time
}


// Operation types produced by the reconciler.
type OpType string

const (
	OpInstall OpType = "install"
	OpUpdate  OpType = "update"
	OpRemove  OpType = "remove"
)

// Operation is a single action to be executed on a node for a component.
type Operation struct {
	Type         OpType
	DeploymentID string
	Component    ComponentSpec
	DesiredHash  string
	Actual       *DeploymentComponentStatus // may be nil for install
	TargetNode   string
}

type Store interface {

}

type Actuator interface {
	Execute(op Operation) error
}

func ReconcileMulti() {

} 

func diffMulti() {

}

/*
siteinfo
    site-id

hostinfo
    host-id

desired/
    site_id/
        app_id/
            version
            components/
                comp_name/
                    version

hosts/
    host_id → HostState


actual/
    site_id/
       host_id/
           app_id/
               version
               components/
                   comp_name/
                       status
                       last_updated
                       hash

ops/
    site_id/
        host_id/
            op_uuid → OperationStruct

site_state/
    site_id → { last_desired_sync, last_actual_sync }

*/