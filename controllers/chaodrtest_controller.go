package controllers

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os/exec"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrr "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	chaosdrv1 "github.com/harrisin2037/chaos-dr-validator/api/v1"
	"github.com/harrisin2037/chaos-dr-validator/internal/backup"
	"github.com/harrisin2037/chaos-dr-validator/internal/chaos"
	sidecarproto "github.com/harrisin2037/chaos-dr-validator/internal/proto/sidecar"
	"github.com/harrisin2037/chaos-dr-validator/internal/velero"
)

var (
	drTestSuccess = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "chaosdr_test_success",
		Help: "Success status of ChaosDRTest",
	})
	backupDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "chaosdr_backup_duration_seconds",
		Help:    "Duration of backup operation",
		Buckets: prometheus.LinearBuckets(1, 5, 10),
	})
	restoreDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "chaosdr_restore_duration_seconds",
		Help:    "Duration of restore operation",
		Buckets: prometheus.LinearBuckets(1, 5, 10),
	})
)

// ChaosDRTestReconciler reconciles a ChaosDRTest object
type ChaosDRTestReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=chaosdr.io,resources=chaodrtests,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=chaosdr.io,resources=chaodrtests/status,verbs=get;update;patch
//+kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch;delete

func (r *ChaosDRTestReconciler) Reconcile(ctx context.Context, req ctrr.Request) (ctrr.Result, error) {
	log := log.FromContext(ctx)
	var backupClient backup.BackupClient
	backupTool := "velero"
	if backupTool == "restic" {
		backupClient = &backup.ResticClient{}
	} else {
		backupClient = &velero.VeleroClient{}
	}

	cr := &chaosdrv1.ChaosDRTest{}
	if err := r.Get(ctx, req.NamespacedName, cr); err != nil {
		if errors.IsNotFound(err) {
			return ctrr.Result{}, nil
		}
		log.Error(err, "unable to fetch ChaosDRTest")
		return ctrr.Result{}, err
	}

	// Step 1: Trigger backup
	start := time.Now()
	backupName := "dr-backup-" + req.Name
	log.Info("Starting ChaosDRTest reconciliation")
	if err := backupClient.CreateBackup(backupName, cr.Spec.AppSelector); err != nil {
		cr.Status.ErrorMessage = err.Error()
		r.Status().Update(ctx, cr)
		return ctrr.Result{}, err
	}
	backupDuration.Observe(time.Since(start).Seconds())
	log.Info("Ending ChaosDRTest reconciliation")
	cr.Status.BackupName = backupName

	// Step 2: Inject chaos (pod-delete)
	chaosName := "chaos-" + req.Name
	if err := chaos.ApplyChaosExperiment(ctx, r.Client, cr, chaosName, cr.Spec.ChaosType); err != nil {
		cr.Status.ErrorMessage = err.Error()
		r.Status().Update(ctx, cr)
		return ctrr.Result{}, err
	}

	// Wait for chaos to complete (simplified for prototype)
	time.Sleep(30 * time.Second)

	// Step 3: Restore to sandbox namespace
	restoreName := "dr-restore-" + req.Name
	sandboxNs := "sandbox-" + req.Name
	if err := backupClient.CreateRestore(backupName, sandboxNs); err != nil {
		cr.Status.ErrorMessage = err.Error()
		r.Status().Update(ctx, cr)
		return ctrr.Result{}, err
	}
	cr.Status.RestoreName = restoreName

	// Step 4: Validate app
	if err := r.validateApp(ctx, cr); err != nil {
		cr.Status.ErrorMessage = err.Error()
		cr.Status.Success = false
		drTestSuccess.Set(0)
		r.Status().Update(ctx, cr)
		return ctrr.Result{}, err
	}

	// Step 5: Call Rust sidecar for data proof
	if err := r.storeValidationProof(ctx, cr); err != nil {
		cr.Status.ErrorMessage = err.Error()
		cr.Status.Success = false
		drTestSuccess.Set(0)
		r.Status().Update(ctx, cr)
		return ctrr.Result{}, err
	}

	// Step 6: Update status
	cr.Status.Success = true
	cr.Status.ErrorMessage = ""
	drTestSuccess.Set(1)
	if err := r.Status().Update(ctx, cr); err != nil {
		log.Error(err, "unable to update status")
		return ctrr.Result{}, err
	}

	return ctrr.Result{}, nil
}

func (r *ChaosDRTestReconciler) validateApp(ctx context.Context, cr *chaosdrv1.ChaosDRTest) error {
	log := log.FromContext(ctx)
	cfg := cr.Spec.ValidationConfig

	if cfg.Script != "" {
		cmd := exec.Command("bash", "-c", cfg.Script)
		output, err := cmd.CombinedOutput()
		if err != nil {
			log.Error(err, "validation script failed", "output", string(output))
			return err
		}
	}

	if cfg.APIEndpoint != "" {
		resp, err := http.Get(cfg.APIEndpoint)
		if err != nil {
			log.Error(err, "API validation failed")
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode != cfg.ExpectedStatusCode {
			return fmt.Errorf("unexpected status code: got %d, expected %d", resp.StatusCode, cfg.ExpectedStatusCode)
		}
	}

	if cfg.DatabaseQuery != nil {
		// Example: Validate database query (e.g., using SQL driver)
		db, err := sql.Open("mysql", cfg.DatabaseQuery.ConnectionString)
		if err != nil {
			return err
		}
		defer db.Close()
		rows, err := db.Query(cfg.DatabaseQuery.Query)
		if err != nil {
			return err
		}
		defer rows.Close()
		rowCount := 0
		for rows.Next() {
			rowCount++
		}
		if rowCount != cfg.DatabaseQuery.ExpectedRows {
			return fmt.Errorf("unexpected row count: got %d, expected %d", rowCount, cfg.DatabaseQuery.ExpectedRows)
		}
	}

	return nil
}

func (r *ChaosDRTestReconciler) storeValidationProof(ctx context.Context, cr *chaosdrv1.ChaosDRTest) error {
	conn, err := grpc.Dial("sidecar:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	defer conn.Close()

	client := sidecarproto.NewDataValidatorClient(conn)
	resp, err := client.ValidateData(ctx, &sidecarproto.DataRequest{
		Data:   []byte("validation-data"), // Mock data for prototype
		Bucket: "backups",
		Object: "proof-" + cr.Name,
	})
	if err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf("sidecar validation failed")
	}
	return nil
}

func (r *ChaosDRTestReconciler) SetupWithManager(mgr ctrr.Manager) error {
	return ctrr.NewControllerManagedBy(mgr).
		For(&chaosdrv1.ChaosDRTest{}).
		Complete(r)
}
