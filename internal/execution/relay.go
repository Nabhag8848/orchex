package execution

import (
	"context"
	"log"
	"time"

	"github.com/nabhag8848/orchex/internal/db"
	sqlcdb "github.com/nabhag8848/orchex/internal/db/sqlc"
	"github.com/nabhag8848/orchex/internal/queue"
)

// StartRelay sends due outbox jobs to SQS once per second.
func StartRelay(ctx context.Context, store *db.Store, sqs *queue.SQS) {
	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := relay(ctx, store, sqs); err != nil && ctx.Err() == nil {
					log.Printf("relay: %v", err)
				}
			}
		}
	}()
}

func relay(ctx context.Context, store *db.Store, sqs *queue.SQS) error {
	return store.InTx(ctx, func(q *sqlcdb.Queries) error {
		jobs, err := q.LockDueRunNodeJobsOutbox(ctx)
		if err != nil {
			return err
		}

		// batch send & delete later
		for _, job := range jobs {
			if job.RunStatus == sqlcdb.WorkflowRunStatusPending || job.RunStatus == sqlcdb.WorkflowRunStatusRunning {
				if err := sqs.Send(ctx, queue.NodeJobMessage{
					RunID:             job.RunID,
					WorkflowVersionID: job.WorkflowVersionID,
					NodeID:            job.NodeID,
					Attempt:           job.Attempt,
				}); err != nil {
					return err
				}
			}

			if err := q.DeleteRunNodeJobOutbox(ctx, job.ID); err != nil {
				return err
			}
		}
		return nil
	})
}
