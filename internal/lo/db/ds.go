package db 

import (
	// "context"
	// "github.com/google/uuid"
	// "github.com/balaji-balu/margo-hello-world/ent"
	// "github.com/balaji-balu/margo-hello-world/ent/deploymentstatus"
	// "github.com/balaji-balu/margo-hello-world/ent/deploymentcomponentstatus"
)

// func xx(ctx context.Context) {
// 	id, err := uuid.Parse(s.DeploymentID)
// 	updated, err := l.db.DeploymentComponentStatus.
// 		Update().
// 		Where(
// 			deploymentcomponentstatus.HasDeploymentWith(deploymentstatus.ID(id)),
// 			deploymentcomponentstatus.NameEQ(s.Name),
// 		).
// 		SetState(s.State).
// 		SetErrorCode(s.Error.Code).
// 		SetErrorMessage(s.Error.Message).
// 		//SetUpdatedAtNow().
// 		Save(ctx)
// 	if err != nil {
// 		return err
// 	}
// }