package tracing

import (
	"errors"
	"net/http"

	"github.com/uptrace/bunrouter"
	"github.com/uptrace/go-clickhouse/ch"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/uql"
	"go4.org/syncutil"
)

type SpanHandler struct {
	*bunapp.App
}

func NewSpanHandler(app *bunapp.App) *SpanHandler {
	return &SpanHandler{
		App: app,
	}
}

func (h *SpanHandler) ListSpans(w http.ResponseWriter, req bunrouter.Request) error {
	f, err := DecodeSpanFilter(h.App, req)
	if err != nil {
		return err
	}
	disableColumnsAndGroups(f.parts)

	ctx := req.Context()
	spans := make([]*Span, 0)

	q := buildSpanIndexQuerySlow(f, f.Duration().Minutes()).
		ColumnExpr("`span.id` AS id").
		ColumnExpr("`span.trace_id` AS trace_id").
		Apply(f.CHOrder).
		Limit(10).
		Offset(f.Pager.GetOffset())

	count, err := q.ScanAndCount(ctx, &spans)
	if err != nil {
		return err
	}

	var group syncutil.Group

	for _, span := range spans {
		span := span
		group.Go(func() error {
			return SelectSpan(ctx, h.App, span)
		})
	}

	if err := group.Err(); err != nil {
		return err
	}

	return bunrouter.JSON(w, bunrouter.H{
		"spans": spans,
		"count": count,
	})
}

func (h *SpanHandler) ListGroups(w http.ResponseWriter, req bunrouter.Request) error {
	f, err := DecodeSpanFilter(h.App, req)
	if err != nil {
		return err
	}

	ctx := req.Context()
	groups := make([]map[string]any, 0)

	q := buildSpanIndexQuery(f, f.Duration().Minutes()).
		Limit(1000)

	if err := q.Scan(ctx, &groups); err != nil {
		return err
	}

	for _, m := range groups {
		fixJSBigInt(m)
	}

	return bunrouter.JSON(w, bunrouter.H{
		"groups":     groups,
		"queryParts": f.parts,
		"columns":    f.columns(groups),
	})
}

func (h *SpanHandler) Percentiles(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()

	f, err := DecodeSpanFilter(h.App, req)
	if err != nil {
		return err
	}

	groupPeriod := calcGroupPeriod(&f.TimeFilter, 300)
	minutes := groupPeriod.Minutes()

	m := make(map[string]interface{})

	subq := h.CH().NewSelect().
		Model((*SpanIndex)(nil)).
		WithAlias("qsNaN", "quantilesTDigest(0.5, 0.9, 0.99)(`span.duration`)").
		WithAlias("qs", "if(isNaN(qsNaN[1]), [0, 0, 0], qsNaN)").
		ColumnExpr("count() AS count").
		ColumnExpr("count() / ? AS rate", minutes).
		ColumnExpr("countIf(`span.status_code` = 'error') AS errorCount").
		ColumnExpr("countIf(`span.status_code` = 'error') / ? AS errorRate", minutes).
		ColumnExpr("round(qs[1]) AS p50").
		ColumnExpr("round(qs[2]) AS p90").
		ColumnExpr("round(qs[3]) AS p99").
		ColumnExpr("toStartOfInterval(`span.time`, INTERVAL ? minute) AS time", minutes).
		Apply(f.whereClause).
		GroupExpr("time").
		OrderExpr("time ASC").
		Limit(10000)

	if err := h.CH().NewSelect().
		ColumnExpr("groupArray(s.count) AS count").
		ColumnExpr("groupArray(s.rate) AS rate").
		ColumnExpr("groupArray(s.errorCount) AS errorCount").
		ColumnExpr("groupArray(s.errorRate) AS errorRate").
		ColumnExpr("groupArray(s.p50) AS p50").
		ColumnExpr("groupArray(s.p90) AS p90").
		ColumnExpr("groupArray(s.p99) AS p99").
		ColumnExpr("groupArray(s.time) AS time").
		TableExpr("(?) AS s", subq).
		GroupExpr("tuple()").
		Limit(1000).
		Scan(ctx, &m); err != nil {
		return err
	}

	return bunrouter.JSON(w, m)
}

func (h *SpanHandler) Stats(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()

	f, err := DecodeSpanFilter(h.App, req)
	if err != nil {
		return err
	}
	disableColumnsAndGroups(f.parts)

	if f.Column == "" {
		return errors.New("'column' query param is required")
	}

	colName, err := uql.ParseName(f.Column)
	if err != nil {
		return err
	}

	minutes := calcGroupPeriod(&f.TimeFilter, 300).Minutes()
	m := make(map[string]interface{})

	subq := buildSpanIndexQuerySlow(f, minutes)
	subq = uqlColumnSlow(subq, colName, minutes).
		ColumnExpr("toStartOfInterval(`span.time`, toIntervalMinute(?)) AS time", minutes).
		GroupExpr("time").
		OrderExpr(
			"time ASC WITH FILL "+
				"FROM toStartOfInterval(toDateTime(?), toIntervalMinute(?)) "+
				"TO toStartOfInterval(toDateTime(?), toIntervalMinute(?)) "+
				"STEP toIntervalMinute(?)",
			f.TimeGTE, minutes, f.TimeLT, minutes, minutes,
		)

	if err := h.CH().NewSelect().
		ColumnExpr("groupArray(?) AS ?", ch.Ident(f.Column), ch.Ident(f.Column)).
		ColumnExpr("groupArray(s.time) AS time").
		TableExpr("(?) AS s", subq).
		GroupExpr("tuple()").
		Limit(1000).
		Scan(ctx, &m); err != nil {
		return err
	}

	return bunrouter.JSON(w, m)
}
