package webhook

import (
	"context"
	"time"

	"github.com/open-policy-agent/gatekeeper/v3/pkg/metrics"
	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
)

const (
	validationRequestCountMetricName    = "validation_request_count"
	validationRequestDurationMetricName = "validation_request_duration_seconds"

	mutationRequestCountMetricName    = "mutation_request_count"
	mutationRequestDurationMetricName = "mutation_request_duration_seconds"
)

var (
	validationResponseTimeInSecM = stats.Float64(
		validationRequestDurationMetricName,
		"The response time in seconds",
		stats.UnitSeconds)

	mutationResponseTimeInSecM = stats.Float64(
		mutationRequestDurationMetricName,
		"The response time in seconds",
		stats.UnitSeconds)

	admissionStatusKey = tag.MustNewKey("admission_status")
	admissionDryRunKey = tag.MustNewKey("admission_dryrun")
	mutationStatusKey  = tag.MustNewKey("mutation_status")
)

func init() {
	if err := register(); err != nil {
		panic(err)
	}
}

// StatsReporter reports webhook metrics.
type StatsReporter interface {
	ReportValidationRequest(ctx context.Context, response requestResponse, isDryRun string, d time.Duration) error
	ReportMutationRequest(ctx context.Context, response requestResponse, d time.Duration) error
}

// reporter implements StatsReporter interface.
type reporter struct{}

// newStatsReporter creaters a reporter for webhook metrics.
func newStatsReporter() (StatsReporter, error) {
	return &reporter{}, nil
}

func (r *reporter) ReportValidationRequest(ctx context.Context, response requestResponse, isDryRun string, d time.Duration) error {
	mutators := []tag.Mutator{tag.Insert(admissionStatusKey, string(response)), tag.Insert(admissionDryRunKey, isDryRun)}
	return r.reportRequest(ctx, mutators, validationResponseTimeInSecM.M(d.Seconds()))
}

func (r *reporter) ReportMutationRequest(ctx context.Context, response requestResponse, d time.Duration) error {
	mutators := []tag.Mutator{tag.Insert(mutationStatusKey, string(response))}
	return r.reportRequest(ctx, mutators, mutationResponseTimeInSecM.M(d.Seconds()))
}

// Captures req count metric, recording the count and the duration.
func (r *reporter) reportRequest(ctx context.Context, mutators []tag.Mutator, m stats.Measurement) error {
	ctx, err := tag.New(
		ctx,
		mutators...,
	)
	if err != nil {
		return err
	}

	return metrics.Record(ctx, m)
}

func register() error {
	views := []*view.View{
		{
			Name:        validationRequestCountMetricName,
			Description: "The number of requests that are routed to validation webhook",
			Measure:     validationResponseTimeInSecM,
			Aggregation: view.Count(),
			TagKeys:     []tag.Key{admissionStatusKey, admissionDryRunKey},
		},
		{
			Name:        validationRequestDurationMetricName,
			Description: validationResponseTimeInSecM.Description(),
			Measure:     validationResponseTimeInSecM,
			Aggregation: view.Distribution(0.001, 0.002, 0.003, 0.004, 0.005, 0.006, 0.007, 0.008, 0.009, 0.01, 0.02, 0.03, 0.04, 0.05, 0.06, 0.07, 0.08, 0.09, 0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8, 0.9, 1, 1.5, 2, 2.5, 3),
			TagKeys:     []tag.Key{admissionStatusKey},
		},
		{
			Name:        mutationRequestCountMetricName,
			Description: "The number of requests that are routed to mutation webhook",
			Measure:     mutationResponseTimeInSecM,
			Aggregation: view.Count(),
			TagKeys:     []tag.Key{mutationStatusKey},
		},
		{
			Name:        mutationRequestDurationMetricName,
			Description: mutationResponseTimeInSecM.Description(),
			Measure:     mutationResponseTimeInSecM,
			Aggregation: view.Distribution(0.001, 0.002, 0.003, 0.004, 0.005, 0.006, 0.007, 0.008, 0.009, 0.01, 0.02, 0.03, 0.04, 0.05, 0.06, 0.07, 0.08, 0.09, 0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8, 0.9, 1, 1.5, 2, 2.5, 3),
			TagKeys:     []tag.Key{mutationStatusKey},
		},
	}
	return view.Register(views...)
}
