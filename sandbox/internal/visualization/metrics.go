package visualization

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/databases/visualization_db"
)

// Unique visitors are distinct ips, always counted on the whole ip, whoever
// reads them: a viewer reads masked ips, never other counts.

// Summary is the cards of the dashboard.
type Summary struct {
	// Views is how many views the filter keeps.
	Views int
	// Unique is how many distinct ips made them.
	Unique int
	// ViewsPerVisitor is Views over Unique, two decimals.
	ViewsPerVisitor float64
	// NotFound is how many of them were not-found views.
	NotFound int
	// Bots is how many views of bots the filter keeps, left out of the rest
	// unless bots are asked for.
	Bots int
}

// Bucket is one bar of the views over time.
type Bucket struct {
	// Start is the instant the bucket starts.
	Start int64
	// Views is how many views it holds.
	Views int
	// Unique is how many distinct ips made them.
	Unique int
}

// TopRow is one row of a top table.
type TopRow struct {
	// Value is what the row counts: a page, a country code, a device, a
	// browser, an os, a referrer ("" for a direct visit), an ip.
	Value string
	// Views is how many views carry it.
	Views int
	// Unique is how many distinct ips made them.
	Unique int
	// Percent is Views over the views of the table, one decimal.
	Percent float64
}

// Tops is every top table of the dashboard.
type Tops struct {
	Pages     []TopRow
	Countries []TopRow
	Devices   []TopRow
	Browsers  []TopRow
	Os        []TopRow
	Referrers []TopRow
	// NotFound is the paths that answered the not-found page, its percent
	// over the not-found views.
	NotFound []TopRow
	// Visitors is the ips that made the most views, masked to a viewer.
	Visitors []TopRow
}

// Metrics is what the dashboard and api-visualization-metrics answer.
type Metrics struct {
	Summary  Summary
	Bucket   string
	Timeline []Bucket
	Top      Tops
}

// ChooseBucket is the bucket asked — hour, day or month — or, for "", the one
// the period of filter calls for: hours up to 2 days, days up to 90, months
// above. Anything else is refused under bucket.
func ChooseBucket(sandbox *api.Sandbox, filter Filter, asked string) (string, *Refusal) {
	switch sandbox.Deps.StringsDeps.TrimSpace(asked) {
	case BucketHour:
		return BucketHour, nil
	case BucketDay:
		return BucketDay, nil
	case BucketMonth:
		return BucketMonth, nil
	case "":
	default:
		return "", refuse("bucket", "bucket is hour, day or month.")
	}
	switch {
	case filter.Days <= 2:
		return BucketHour, nil
	case filter.Days <= 90:
		return BucketDay, nil
	}
	return BucketMonth, nil
}

// Measure is every metric of selection: the cards, the views over time in
// bucket, and the first top rows of each table — the ips masked unless root.
func Measure(sandbox *api.Sandbox, selection Selection, bucket string, top int, root bool) Metrics {
	return Metrics{
		Summary:  Summarize(sandbox, selection),
		Bucket:   bucket,
		Timeline: Timeline(sandbox, selection, bucket),
		Top:      Top(sandbox, selection, top, root),
	}
}

// Summarize is the cards of selection.
func Summarize(sandbox *api.Sandbox, selection Selection) Summary {
	summary := Summary{Views: len(selection.Views), Bots: selection.Bots}
	ips := map[string]bool{}
	for _, view := range selection.Views {
		ips[view.Ip] = true
		if view.Status == StatusNotFound {
			summary.NotFound++
		}
	}
	summary.Unique = len(ips)
	if summary.Unique > 0 {
		summary.ViewsPerVisitor = round(float64(summary.Views)/float64(summary.Unique), 100)
	}
	return summary
}

// Timeline is every bucket of the period of selection, empty ones included,
// oldest first.
func Timeline(sandbox *api.Sandbox, selection Selection, bucket string) []Bucket {
	buckets := []Bucket{}
	index := map[int64]int{}
	for start := BucketStart(sandbox, selection.Filter.AtMin, bucket); start <= selection.Filter.AtMax; start = NextBucket(sandbox, start, bucket) {
		index[start] = len(buckets)
		buckets = append(buckets, Bucket{Start: start})
	}
	ips := make([]map[string]bool, len(buckets))
	for _, view := range selection.Views {
		at, ok := index[BucketStart(sandbox, view.At, bucket)]
		if !ok {
			continue
		}
		buckets[at].Views++
		if ips[at] == nil {
			ips[at] = map[string]bool{}
		}
		ips[at][view.Ip] = true
	}
	for at := range buckets {
		buckets[at].Unique = len(ips[at])
	}
	return buckets
}

// Top is the first count rows of every top table of selection.
func Top(sandbox *api.Sandbox, selection Selection, count int, root bool) Tops {
	views := selection.Views
	notFound := []visualization_db.VisualizationViewRecord{}
	for _, view := range views {
		if view.Status == StatusNotFound {
			notFound = append(notFound, view)
		}
	}
	tops := Tops{
		Pages:     rank(sandbox, views, count, func(view visualization_db.VisualizationViewRecord) string { return view.Page }),
		Countries: rank(sandbox, views, count, func(view visualization_db.VisualizationViewRecord) string { return view.Country }),
		Devices:   rank(sandbox, views, count, func(view visualization_db.VisualizationViewRecord) string { return view.Device }),
		Browsers:  rank(sandbox, views, count, func(view visualization_db.VisualizationViewRecord) string { return view.Browser }),
		Os:        rank(sandbox, views, count, func(view visualization_db.VisualizationViewRecord) string { return view.Os }),
		Referrers: rank(sandbox, views, count, func(view visualization_db.VisualizationViewRecord) string { return view.Referrer }),
		NotFound:  rank(sandbox, notFound, count, func(view visualization_db.VisualizationViewRecord) string { return view.Path }),
		Visitors:  rank(sandbox, views, count, func(view visualization_db.VisualizationViewRecord) string { return view.Ip }),
	}
	for index := range tops.Visitors {
		tops.Visitors[index].Value = ShownIp(sandbox, tops.Visitors[index].Value, root)
	}
	return tops
}

// rank counts views by the value of key, and answers the count values with
// the most views — then the most distinct ips, then by value — each with its
// percent of every view.
func rank(sandbox *api.Sandbox, views []visualization_db.VisualizationViewRecord, count int, key func(view visualization_db.VisualizationViewRecord) string) []TopRow {
	rows := []TopRow{}
	position := map[string]int{}
	ips := []map[string]bool{}
	for _, view := range views {
		value := key(view)
		at, ok := position[value]
		if !ok {
			at = len(rows)
			position[value] = at
			rows = append(rows, TopRow{Value: value})
			ips = append(ips, map[string]bool{})
		}
		rows[at].Views++
		ips[at][view.Ip] = true
	}
	for at := range rows {
		rows[at].Unique = len(ips[at])
		rows[at].Percent = round(float64(rows[at].Views)*100/float64(len(views)), 10)
	}
	sandbox.Deps.SortDeps.Slice(rows, func(i int, j int) bool {
		if rows[i].Views != rows[j].Views {
			return rows[i].Views > rows[j].Views
		}
		if rows[i].Unique != rows[j].Unique {
			return rows[i].Unique > rows[j].Unique
		}
		return rows[i].Value < rows[j].Value
	})
	if count > 0 && len(rows) > count {
		rows = rows[:count]
	}
	return rows
}

// round is value rounded to the nearest 1/scale.
func round(value float64, scale float64) float64 {
	return float64(int64(value*scale+0.5)) / scale
}
