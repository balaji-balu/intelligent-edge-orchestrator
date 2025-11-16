package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql"
	_ "github.com/lib/pq" // enables the 'postgres' driver
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"go.uber.org/zap"
	
	"github.com/balaji-balu/margo-hello-world/ent"
	"github.com/balaji-balu/margo-hello-world/internal/api"
	"github.com/balaji-balu/margo-hello-world/internal/config"
	"github.com/balaji-balu/margo-hello-world/internal/gitmanager"
	//"github.com/balaji-balu/margo-hello-world/internal/fsmloader"
	"github.com/balaji-balu/margo-hello-world/internal/metrics"
	"github.com/balaji-balu/margo-hello-world/internal/co"
	"github.com/balaji-balu/margo-hello-world/internal/logger"

)

func init() {
	err := godotenv.Load("./.env") // relative path to project root
	if err != nil {
		log.Println("No .env file found, reading from system environment")
	}
	
}

func main() {
	ctx := context.Background()
	logger, _ := logger.New("production", "co")

	log.Println(logger)
	configPath := flag.String("config", "./configs/co.yaml", "path to config file")
	flag.Parse()

	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		logger.Error(ctx, "error loading config:", err)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	metrics_port := os.Getenv("METRICS_PORT")
	if metrics_port == "" {
		metrics_port = "9200"
	}

	// msg := fmt.Sprintf("✅ Loaded config: site=%s, port=%s metrics-port=%s",
	// 	cfg.Server.Site, port, metrics_port)
	// log.Println("msg", msg, ctx)
	logger.Info(ctx, "conf done")

    // // Use NewProduction() for JSON, performance, and sampled logging
    // logger, err := zap.NewProduction()
    // if err != nil {
    //     // If Zap fails, fall back to standard log or panic
    //     log.Fatalf("can't initialize zap logger: %v", err)
    // }
    // defer logger.Sync() // Ensure all buffered logs are written

	// Inside main()
	// Redirect all calls from the standard library's 'log' package to Zap.
	// This is the single most important step for converting existing logs immediately.
	//zap.RedirectStdLog(logger)

	grpcPort := flag.String("grpc", ":50051", "CO gRPC listen address")
	//httpPort := flag.String("http", ":8080", "CO HTTP listen address")
	//loAddr := flag.String("lo", "localhost:50052", "Local Orchestrator address")
	flag.Parse()

	//log.Println("config:", *config, "node:", node)

	// machine, err := fsmloader.LoadFSM("./configs/fsm.yaml", "CO")
	// if err != nil {
	// 	log.Fatalf("failed to load FSM: %v", err)
	// }

	// fmt.Println("CO initial:", machine.Current())
	// _ = machine.Event(ctx, "send_request", )
	// _ = machine.Event(ctx, "complete")
	// _ = machine.Event(ctx, "reset")
	dsn := os.Getenv("DATABASE_URL")
	logger.Info(ctx, "[CO] connecting to postgres at", zap.String("dsn:", dsn))

	var drv *sql.Driver
	var err1 error
	for i := 1; i <= 10; i++ {
		drv, err1 = sql.Open(dialect.Postgres, dsn)
		if err1 == nil {
			if err1 = drv.DB().Ping(); err1 == nil {
				logger.Info(ctx, "✅ Connected to Postgres")
				break
			}
		}
		logger.Info(ctx, "⏳ Waiting for Postgres \n", zap.Int("attempt...", i))
		time.Sleep(3 * time.Second)
	}
	if err1 != nil {
		logger.Error(ctx, "❌ Failed to connect to Postgres after retries.", err)
	}

	client := ent.NewClient(ent.Driver(drv))
	defer client.Close()

	metrics.Init("co")
	metrics.StartServer(metrics_port)

	gitm := gitmanager.NewManager()

	gitm.Register(gitmanager.RepoConfig{
		Name: "app-registry",
		Mode: gitmanager.GitRemote, // or GitLocal
		RemoteURL: "https://github.com/edge-orchestration-platform/app-registry.git",
		Branch: "main",
		WorkingPath: "/tmp/app-registry",
	})

	gitm.Register(gitmanager.RepoConfig{
		Name: "deployments",
		Mode: gitmanager.GitLocal, //GitRemote,
		//RemoteURL: "https://github.com/edge-orchestration-platform/deployments.git",
		LocalPath: "/home/balaji/local-deployments",
		Branch: "main",
		Token: os.Getenv("GITHUB_TOKEN"),
		WorkingPath: "/tmp/deployments-co",
	})
	if err := gitm.InitRepo("deployments"); err != nil {
    	logger.Error(ctx, "Git initrepo failed", err)
	}
	c := co.NewCO(gitm, "app-registry", "deployments")

	router := api.NewRouter(client, c, cfg)
	log.Println("CO API running on :", port)
	if err := router.Run(fmt.Sprintf(":%s", port)); err != nil {
		logger.Error(ctx, "Router init failed", err)
	}

	// Start gRPC server for callbacks from LO
	go func() {
		lis, err := net.Listen("tcp", *grpcPort)
		if err != nil {
			logger.Error(ctx, "[CO] failed to listen:", err)
		}
		s := grpc.NewServer()
		//pb.RegisterCentralOrchestratorServer(s, &server{})
		log.Printf("[CO] gRPC listening on %s", *grpcPort)
		if err := s.Serve(lis); err != nil {
			log.Fatalf("[CO] serve: %v", err)
		}
	}()
}


