package lo

import (
	"context"
	"os"
	"time"
	//"log"
	"net/http"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	
	"github.com/balaji-balu/margo-hello-world/pkg/model"
	"github.com/balaji-balu/margo-hello-world/internal/gitmanager"
	"github.com/balaji-balu/margo-hello-world/internal/natsbroker"
	"github.com/balaji-balu/margo-hello-world/internal/lo/heartbeat"
	"github.com/balaji-balu/margo-hello-world/internal/lo/reconciler"
	"github.com/balaji-balu/margo-hello-world/internal/lo/watcher"
	"github.com/balaji-balu/margo-hello-world/internal/lo/boltstore"
	"github.com/balaji-balu/margo-hello-world/internal/metrics"
	"github.com/balaji-balu/margo-hello-world/internal/lo/actuators"
	"github.com/balaji-balu/margo-hello-world/internal/lo/logger"	
)

type EventType string

const (
	EventGitPolled      = "EventGitPolled"
	EventNetworkChange  = "EventNetworkChange"
	EventDeployComplete = "EventDeployComplete"
)

type Event struct {
	Name string
	Data interface{}
	Time time.Time
}

type GitPolledPayload struct {
	Commit      string
	//Deployments []gitobserver.DeploymentChange
	Deployments []watcher.DeploymentChange
}

type LoConfig struct {
	Owner   string
	Repo    string
	Token   string
	Path    string
	NatsUrl string
	Site    string
}

type LocalOrchestrator struct {
	Config  LoConfig
	//Journal Journal
	//EOPort  string
	//Hosturls []string
	Hosts []string
	//FSM   *fsm.FSM

	rb 			*ResultBus

	RootCtx 	context.Context
	nc      	*natsbroker.Broker

	reconcile  	*reconciler.Reconciler
	//Store 		*Db.DbStore
	store 	*boltstore.StateStore
	//inMemStore	*reconciler.InMemoryStore
	monitor 	*heartbeat.Monitor	
	Mgr     	*gitmanager.Manager
	Watcher 	*watcher.Watcher
	//db 			*ent.Client
	eventCh     chan Event
	log      *zap.SugaredLogger
	currentMode string
	cancelFunc  context.CancelFunc // for stopping running process
}

func NewLO(
	ctx context.Context,
	siteID string, 
	boltDb string,
	natsURL, 
	repo string,
	//boltz *bolt.DB,
	//db *ent.Client,
	nc *natsbroker.Broker,
	gitmgr *gitmanager.Manager,
	metrics_port string,
	log *zap.SugaredLogger,
) *LocalOrchestrator {

	rb := NewResultBus()

	log.Debugw("LocalOrchestrator.new enter ")
	logger.InitLogger(true)

	store, err:= boltstore.NewStateStore(boltDb)
	if err != nil {
		log.Errorf("store create error", "err", err)
		return nil
	}
	monitor := heartbeat.NewMonitor(10*time.Second, 3, store) // EN heartbeat every ~10 sec, max 3 misses
	monitor.Start()

	metrics.Init("lo")
	metrics.StartServer(metrics_port)

	//inMemStore := reconciler.NewInMemoryStore()
	na := actuators.NewNatsActuator(store, nc, siteID, 30)
	//r := localorch.NewHTTPReporter("api/v1/co/deploy/status", 30)
	reconcile := reconciler.NewReconciler(store, na)

	log.Debugw("LocalOrchestrator.new exiting  ")
	return &LocalOrchestrator{
		Config: LoConfig{
			//Owner: cfg..Owner,
			Repo:    repo, //cfg.Git.Repo,
			NatsUrl: natsURL,//cfg.NATS.URL,
			Token:   os.Getenv("GITHUB_TOKEN"),
			Site:    siteID, //cfg.Server.Site,
		},
		log: log,
		rb:     rb,
		eventCh: make(chan Event, 20),
		RootCtx: ctx,
		//db:      db,
		nc:      nc,
		Mgr: 	gitmgr,
		reconcile: reconcile,
		//Store: store,
		monitor: monitor,
		store: store,
	}
}

func (l *LocalOrchestrator) Start(coURL string) {
	
	go l.StartEventDispatcher(l.RootCtx)

	go l.StartNetworkMonitor(l.RootCtx)

	l.MonitorHealthandStatusFromEN(l.monitor, coURL)
}

func (l *LocalOrchestrator) HandlerGetActual(c *gin.Context) {
	actual, err := l.store.GetActual()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, actual)	
}

func (l *LocalOrchestrator) HandlerGetHosts(c *gin.Context) {
	hosts, err := l.store.LoadAllHosts()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, hosts)
}

func (l *LocalOrchestrator) RegisterERA(c *gin.Context) {
    var req struct {
        HostID string `json:"host_id"`
    }

    if err := c.BindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
        return
    }     

    if req.HostID == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "hostID missing"})
        return
    }

	// store host id 
	if err := l.store.AddOrUpdateHost(model.Host{ID: req.HostID}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	} 

	// return site id
	c.JSON(http.StatusOK, l.Config.Site)
}