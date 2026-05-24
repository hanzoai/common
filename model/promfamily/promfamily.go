// Package promfamily holds the prometheus/client_model "dto" coupled
// helpers — EscapeMetricFamily and friends — that used to live inside
// model/metric.go. Moved out so that consumers of model can avoid the
// prometheus/client_model/go transitive dep when they only need the
// metric-name validation, label-pair, escape-scheme surface.
package promfamily

import (
	"github.com/hanzoai/common/model"
	dto "github.com/prometheus/client_model/go"
	"google.golang.org/protobuf/proto"
)

// EscapeMetricFamily escapes the given metric names and labels with the given
// escaping scheme. Returns a new object that uses the same pointers to fields
// when possible and creates new escaped versions so as not to mutate the
// input.
func EscapeMetricFamily(v *dto.MetricFamily, scheme model.EscapingScheme) *dto.MetricFamily {
	if v == nil {
		return nil
	}

	if scheme == model.NoEscaping {
		return v
	}

	out := &dto.MetricFamily{
		Help: v.Help,
		Type: v.Type,
		Unit: v.Unit,
	}

	if v.Name == nil || model.IsValidLegacyMetricName(v.GetName()) {
		out.Name = v.Name
	} else {
		out.Name = proto.String(model.EscapeName(v.GetName(), scheme))
	}
	for _, m := range v.Metric {
		if !metricNeedsEscaping(m) {
			out.Metric = append(out.Metric, m)
			continue
		}

		escaped := &dto.Metric{
			Gauge:       m.Gauge,
			Counter:     m.Counter,
			Summary:     m.Summary,
			Untyped:     m.Untyped,
			Histogram:   m.Histogram,
			TimestampMs: m.TimestampMs,
		}

		for _, l := range m.Label {
			if l.GetName() == model.MetricNameLabel {
				if l.Value == nil || model.IsValidLegacyMetricName(l.GetValue()) {
					escaped.Label = append(escaped.Label, l)
					continue
				}
				escaped.Label = append(escaped.Label, &dto.LabelPair{
					Name:  proto.String(model.MetricNameLabel),
					Value: proto.String(model.EscapeName(l.GetValue(), scheme)),
				})
				continue
			}
			if l.Name == nil || model.IsValidLegacyMetricName(l.GetName()) {
				escaped.Label = append(escaped.Label, l)
				continue
			}
			escaped.Label = append(escaped.Label, &dto.LabelPair{
				Name:  proto.String(model.EscapeName(l.GetName(), scheme)),
				Value: l.Value,
			})
		}
		out.Metric = append(out.Metric, escaped)
	}
	return out
}

func metricNeedsEscaping(m *dto.Metric) bool {
	for _, l := range m.Label {
		if l.GetName() == model.MetricNameLabel && !model.IsValidLegacyMetricName(l.GetValue()) {
			return true
		}
		if !model.IsValidLegacyMetricName(l.GetName()) {
			return true
		}
	}
	return false
}
