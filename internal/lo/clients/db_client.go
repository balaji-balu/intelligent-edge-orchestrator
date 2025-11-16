package lo

type DbStore struct {
	client *ent.Client
}

func NewDbStore(client *ent.Client) *EntStore {
    return &DbStore{client: client}
}

func (s *DbStore) GetDesired(deploymentID string) (DeploymentSpec, error) {
    return GetDeployment(s.db, deploymentID)
}

func (s *DbStore) GetActual(deploymentID string) (DeploymentState, error) {
    // fetch actual state from deployment_state table, aggregate if needed
}

func (s *DbStore) GetActualPerNode(deploymentID string) (map[string]DeploymentState, error) {
    // fetch all component states grouped by node_id
}

func (s *DbStore) GetNodes(siteID string) (map[string]NodeState, error) {
    // fetch all nodes for site_id
}
