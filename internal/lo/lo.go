package lo

import (
	"context"
	"os"
	"fmt"
	"time"
	"log"
	"path/filepath"
	"net/http"
	"github.com/gin-gonic/gin"
	//"github.com/looplab/fsm"
	//bolt "go.etcd.io/bbolt"

	//"github.com/balaji-balu/margo-hello-world/ent"
	"github.com/balaji-balu/margo-hello-world/internal/gitmanager"
	"github.com/balaji-balu/margo-hello-world/internal/natsbroker"
	//"github.com/balaji-balu/margo-hello-world/internal/fsmloader"
	//Db "github.com/balaji-balu/margo-hello-world/internal/lo/db"
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
	//logger      *zap.Logger
	currentMode string
	cancelFunc  context.CancelFunc // for stopping running process
}

func makeDirName(siteID string) (string) {
	filename := "bolt.db"
	// Resolve home directory
	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	// Build directory path: ~/.lo/<siteid>
	dirPath := filepath.Join(home, ".lo", siteID)

	// // Create the directory (including parents)
	err = os.MkdirAll(dirPath, 0o755)
	if err != nil {
		panic(err)
	}

	// Full file path
	filePath := filepath.Join(dirPath, filename)

	// // Create or open file for writing
	// f, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	// if err != nil {
	// 	panic(err)
	// }
	// defer f.Close()

	// Write something
	// _, err = f.WriteString("{\"hello\":\"world\"}")
	// if err != nil {
	// 	panic(err)
	// }

	fmt.Println("Created:", filePath)

	return filePath
}

func New(
	ctx context.Context,
	siteID, natsURL, 
	repo string,
	//boltz *bolt.DB,
	//db *ent.Client,
	nc *natsbroker.Broker,
	gitmgr *gitmanager.Manager,
	//logger *zap.Logger,
	metrics_port string,
) *LocalOrchestrator {
	rb := NewResultBus()

	log.Println("LocalOrchestrator.new enter ")
	logger.InitLogger(true)

	//logger.Info("boltzpath:", "path", boltzpath)

	//inMemStore := reconciler.NewInMemoryStore()

	store, err:= boltstore.NewStateStore(makeDirName(siteID))
	if err != nil {
		log.Fatalf("store create error", err)
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
	log.Println("LO Starting ")
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