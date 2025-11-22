package lo

import (
	"context"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/balaji-balu/margo-hello-world/internal/lo/heartbeat"
	"github.com/balaji-balu/margo-hello-world/internal/lo/reconciler"
	"github.com/balaji-balu/margo-hello-world/internal/lo/logger"
	"github.com/balaji-balu/margo-hello-world/pkg/model"

)

func (l *LocalOrchestrator) MonitorHealthandStatusFromEN(
	monitor *heartbeat.Monitor, coUrl string) {

	// Subscribe to health
	go func() {
		log.Println("lo with siteid:", fmt.Sprintf("health.%s.*", l.Config.Site))
		subHealth := fmt.Sprintf("health.%s.*", l.Config.Site)
		err := l.nc.Subscribe2(subHealth, func(h model.HealthMsg) {
			log.Printf("[LO] health from %s runtime=%s", h.NodeID, h.Runtime)

			host := reconciler.Host{
				ID: h.NodeID,
				Labels: map[string]string{
					"region": "us-east",
					"role":   "worker",
				},
				Status: "alive",
			}
			l.boltstore.AddOrUpdateHost(context.Background(), h.SiteID, &host)
			monitor.Update(h.NodeID)	

			// err := lo.CreateEdgeNode(lo.RootCtx, h)
			// if err != nil {
			// 	log.Printf("[LO] error saving the Node:", err)
			// }
			//nodeCount.Set(float64(len(orchestrator.GetAllNodes(db))))
			//if fsm.GetState() == shared.Discovering  {
			//	fsm.Transition(shared.Running)
			//}
		})
		if err != nil {
			log.Println("subscribe error:", err)
		} else {
			log.Println("subscribed to", subHealth)
		}
		l.nc.Flush()
		log.Println("subscription ready for", subHealth)

	}()

	go func() {
		subStatus := fmt.Sprintf("status.%s.*", l.Config.Site)
		err := l.nc.Subscribe4(subStatus, func(s reconciler.DeploymentComponentStatus) {
			//log.Printf("[LO] status %s from %s: success=%v, msg=%s",
			//	s.DeploymentID, s.NodeID, s.Status, s.Message)
			log.Println("[LO] component state:", s)

			if s.Status == "installed" {
				ctx := context.Background()

				desired, _ := l.boltstore.GetDesiredHashes(ctx,s.DeploymentID)
				desiredHash := desired[s.ComponentName]

				// component installed successfully -> update actual hash
				err := l.boltstore.SetActualHash(
					ctx,
					s.DeploymentID,
					s.HostID,     // THIS IS HOST ID
					s.ComponentName,  
					desiredHash,       // hash returned from agent
				)
				if err != nil {
					logger.Error("failed to update actual hash", "err", err)
				}
			}

			// Forward the status back to CO
			// l.sendStatusToCO(s)			

			//forward to CO
			//forwardToCO(coUrl, s)

		})
		if err != nil {
			log.Fatal("[LO] failed to subscribe to status updates:", err)
		}

	}()

}

func forwardToCO(baseurl string, report model.DeploymentStatus) {
	url := fmt.Sprintf("%s/deployments/%s/status", baseurl, report.DeploymentID)
	payload, err := json.Marshal(report)
	if err != nil {
		log.Println("[LO] failed to marshal report:", err)
		return
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		log.Println("[LO] failed to send report to CO:", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("[LO] CO returned status %d: %s", resp.StatusCode, string(body))
		return
	}

	log.Printf("[LO] ✅ Report forwarded to CO successfully (deployment_id=%s)", 
					report.DeploymentID)
}
