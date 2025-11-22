package lo

import (
	"context"
	"os"
	"time"
	"log"
	//"github.com/looplab/fsm"
	bolt "go.etcd.io/bbolt"

	"github.com/balaji-balu/margo-hello-world/ent"
	"github.com/balaji-balu/margo-hello-world/internal/gitmanager"
	"github.com/balaji-balu/margo-hello-world/internal/natsbroker"
	//"github.com/balaji-balu/margo-hello-world/internal/fsmloader"
	Db "github.com/balaji-balu/margo-hello-world/internal/lo/db"
	"github.com/balaji-balu/margo-hello-world/internal/lo/heartbeat"
	"github.com/balaji-balu/margo-hello-world/internal/lo/reconciler"
	"github.com/balaji-balu/margo-hello-world/internal/lo/watcher"
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
	Store 		*Db.DbStore
	boltstore 	*reconciler.BoltStore
	//inMemStore	*reconciler.InMemoryStore
	monitor 	*heartbeat.Monitor	
	Mgr     	*gitmanager.Manager
	Watcher 	*watcher.Watcher
	db 			*ent.Client
	eventCh     chan Event
	//logger      *zap.Logger
	currentMode string
	cancelFunc  context.CancelFunc // for stopping running process
}

func New(
	ctx context.Context,
	siteID, natsURL, 
	repo string,
	boltz *bolt.DB,
	db *ent.Client,
	nc *natsbroker.Broker,
	gitmgr *gitmanager.Manager,
	//logger *zap.Logger,
	metrics_port string,
) *LocalOrchestrator {
	rb := NewResultBus()

	log.Println("LocalOrchestrator.new enter ")
	logger.InitLogger(true)

	//logger.Info("boltzpath:", "path", boltzpath)

	inMemStore := reconciler.NewInMemoryStore()

	boltstore, err:= reconciler.NewBoltStore(boltz, inMemStore)
	if err != nil {
		log.Println("reconciler create error")
		return nil
	}
	monitor := heartbeat.NewMonitor(10*time.Second, 3, boltstore) // EN heartbeat every ~10 sec, max 3 misses
	monitor.Start()

	metrics.Init("lo")
	metrics.StartServer(metrics_port)

	//inMemStore := reconciler.NewInMemoryStore()
	na := actuators.NewNatsActuator(nc, siteID, 30)
	//r := localorch.NewHTTPReporter("api/v1/co/deploy/status", 30)
	reconcile := reconciler.NewReconciler(boltstore, na, nil)

	log.Println("LocalOrchestrator.new exiting  ")
	return &LocalOrchestrator{
		Config: LoConfig{
			//Owner: cfg..Owner,
			Repo:    repo, //cfg.Git.Repo,
			NatsUrl: natsURL,//cfg.NATS.URL,
			Token:   os.Getenv("GITHUB_TOKEN"),
			Site:    siteID, //cfg.Server.Site,
		},
		//logger: logger,
		rb:     rb,
		eventCh: make(chan Event, 20),
		RootCtx: ctx,
		db:      db,
		nc:      nc,
		Mgr: 	gitmgr,
		reconcile: reconcile,
		//Store: store,
		monitor: monitor,
		boltstore: boltstore,
	}
}

func (l *LocalOrchestrator) Start(coURL string) {
	log.Println("LO Starting ")
	go l.StartEventDispatcher(l.RootCtx)

	go l.StartNetworkMonitor(l.RootCtx)

	l.MonitorHealthandStatusFromEN(l.monitor, coURL)
}

