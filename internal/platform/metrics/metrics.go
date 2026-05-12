package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	LoginsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "app_logins_total",
		Help: "The total number of successful user logins",
	})

	RegistrationsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "app_registrations_total",
		Help: "The total number of successful user registrations",
	})

	MangasUploadedTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "app_mangas_uploaded_total",
		Help: "The total number of mangas created/uploaded",
	})
)
