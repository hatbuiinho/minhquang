package reminder

import (
	"context"
	"fmt"
	"log"
	"time"

	"reminder/be/internal/device"
	"reminder/be/internal/event"
	"reminder/be/internal/push"
)

type Worker struct {
	events  event.Store
	devices *device.Service
	sender  push.Sender
	now     func() time.Time
	limit   int
	logger  *log.Logger
}

type Config struct {
	BatchSize int
	Logger    *log.Logger
}

func NewWorker(events event.Store, devices *device.Service, sender push.Sender, now func() time.Time, config Config) *Worker {
	limit := config.BatchSize
	if limit <= 0 {
		limit = 25
	}
	logger := config.Logger
	if logger == nil {
		logger = log.Default()
	}

	return &Worker{
		events:  events,
		devices: devices,
		sender:  sender,
		now:     now,
		limit:   limit,
		logger:  logger,
	}
}

func (w *Worker) Run(ctx context.Context, interval time.Duration) {
	if interval <= 0 {
		interval = 30 * time.Second
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		if err := w.ProcessDue(ctx); err != nil {
			w.logger.Printf("process reminder jobs: %v", err)
		}

		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func (w *Worker) ProcessDue(ctx context.Context) error {
	now := w.now().UTC()
	jobs, err := w.events.ClaimDueReminderJobs(ctx, now, w.limit)
	if err != nil {
		return err
	}

	for _, job := range jobs {
		if err := w.processJob(ctx, job, now); err != nil {
			w.logger.Printf("process reminder job %s: %v", job.ID, err)
		}
	}

	return nil
}

func (w *Worker) processJob(ctx context.Context, job event.ReminderJob, now time.Time) error {
	devices, err := w.devices.ListEnabledByUser(ctx, job.UserID)
	if err != nil {
		_ = w.events.MarkReminderJobFailed(ctx, job.ID, now)
		return err
	}
	if len(devices) == 0 {
		_ = w.events.MarkReminderJobFailed(ctx, job.ID, now)
		return fmt.Errorf("no enabled devices for user %s", job.UserID)
	}

	successCount := 0
	for _, item := range devices {
		err := w.sender.Send(ctx, push.Message{
			Token:      item.PushToken,
			Title:      "Nhắc hẹn",
			Body:       job.EventTitle,
			Importance: pushImportance(job.Importance),
			Data: map[string]string{
				"type":       "reminder",
				"event_id":   job.EventID,
				"job_id":     job.ID,
				"importance": string(job.Importance),
			},
		})
		if err != nil {
			w.logger.Printf("send reminder job %s to device %s: %v", job.ID, item.ID, err)
			continue
		}
		successCount++
	}

	if successCount == 0 {
		return w.events.MarkReminderJobFailed(ctx, job.ID, now)
	}

	return w.events.MarkReminderJobSent(ctx, job.ID, now)
}

func pushImportance(value event.ReminderImportance) push.Importance {
	if value == event.ReminderImportanceUrgent {
		return push.ImportanceUrgent
	}

	return push.ImportanceNormal
}
