package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
)

// OverviewData represents aggregated overview stats.
type OverviewData struct {
	TotalUsage  int `json:"total_usage"`
	TotalTokens int `json:"total_tokens"`
	TotalModels int `json:"total_models"`
	TotalDays   int `json:"total_days"`
}

// GetOverview returns aggregate usage stats.
func (h *Handler) GetOverview(w http.ResponseWriter, r *http.Request) {
	var o OverviewData
	err := h.db.QueryRow(`
		SELECT COALESCE(SUM(usage_count), 0), COALESCE(SUM(total_tokens), 0), COUNT(DISTINCT model_id), COUNT(DISTINCT date)
		FROM usage_stats`).Scan(&o.TotalUsage, &o.TotalTokens, &o.TotalModels, &o.TotalDays)
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	respondOK(w, o)
}

// DailyStat represents a daily usage statistic.
type DailyStat struct {
	Date        string `json:"date"`
	UsageCount  int    `json:"usage_count"`
	TotalTokens int    `json:"total_tokens"`
}

// GetDaily returns daily stats with optional date range filter.
func (h *Handler) GetDaily(w http.ResponseWriter, r *http.Request) {
	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")

	query := "SELECT date, COALESCE(SUM(usage_count), 0), COALESCE(SUM(total_tokens), 0) FROM usage_stats"
	args := []interface{}{}

	where := ""
	if startDate != "" {
		where += " date >= ?"
		args = append(args, startDate)
	}
	if endDate != "" {
		if where != "" {
			where += " AND"
		}
		where += " date <= ?"
		args = append(args, endDate)
	}
	if where != "" {
		query += " WHERE" + where
	}
	query += " GROUP BY date ORDER BY date ASC"

	rows, err := h.db.Query(query, args...)
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	defer rows.Close()

	var list []DailyStat
	for rows.Next() {
		var d DailyStat
		if err := rows.Scan(&d.Date, &d.UsageCount, &d.TotalTokens); err != nil {
			respondFail(w, 500, err.Error())
			return
		}
		list = append(list, d)
	}
	respondOK(w, list)
}

// ModelStat represents a usage statistic grouped by model.
type ModelStat struct {
	ModelName    string  `json:"model_name"`
	ProviderName *string `json:"provider_name"`
	UsageCount   int     `json:"usage_count"`
	TotalTokens  int     `json:"total_tokens"`
}

// GetModelStats returns usage stats grouped by model.
func (h *Handler) GetModelStats(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.Query(`
		SELECT model_name, provider_name, COALESCE(SUM(usage_count), 0), COALESCE(SUM(total_tokens), 0)
		FROM usage_stats
		GROUP BY model_name, provider_name
		ORDER BY usage_count DESC`)
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	defer rows.Close()

	var list []ModelStat
	for rows.Next() {
		var m ModelStat
		if err := rows.Scan(&m.ModelName, &m.ProviderName, &m.UsageCount, &m.TotalTokens); err != nil {
			respondFail(w, 500, err.Error())
			return
		}
		list = append(list, m)
	}
	respondOK(w, list)
}

// MonthlyStat represents a monthly usage statistic.
type MonthlyStat struct {
	Month       string `json:"month"`
	UsageCount  int    `json:"usage_count"`
	TotalTokens int    `json:"total_tokens"`
}

// GetMonthly returns monthly stats.
func (h *Handler) GetMonthly(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.Query(`
		SELECT strftime('%Y-%m', date) as month, COALESCE(SUM(usage_count), 0), COALESCE(SUM(total_tokens), 0)
		FROM usage_stats
		GROUP BY month
		ORDER BY month ASC`)
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	defer rows.Close()

	var list []MonthlyStat
	for rows.Next() {
		var m MonthlyStat
		if err := rows.Scan(&m.Month, &m.UsageCount, &m.TotalTokens); err != nil {
			respondFail(w, 500, err.Error())
			return
		}
		list = append(list, m)
	}
	respondOK(w, list)
}

// RecordUsageRequest is the request body for recording usage.
type RecordUsageRequest struct {
	Date         string  `json:"date"`
	ModelID      *int    `json:"model_id"`
	ModelName    string  `json:"model_name"`
	ProviderName *string `json:"provider_name"`
	TotalTokens  int     `json:"total_tokens"`
}

// RecordUsage records a usage event (upsert).
func (h *Handler) RecordUsage(w http.ResponseWriter, r *http.Request) {
	var req RecordUsageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondFail(w, 400, "请求体格式错误: "+err.Error())
		return
	}

	if req.Date == "" {
		respondFail(w, 400, "日期不能为空")
		return
	}

	// Upsert: try to update existing record, insert if not found
	var existingID int
	err := h.db.QueryRow("SELECT id FROM usage_stats WHERE date = ? AND model_id = ?", req.Date, req.ModelID).Scan(&existingID)
	switch err {
	case sql.ErrNoRows:
		// Insert new record
		_, err = h.db.Exec("INSERT INTO usage_stats (date, model_id, model_name, provider_name, usage_count, total_tokens) VALUES (?, ?, ?, ?, 1, ?)",
			req.Date, req.ModelID, req.ModelName, req.ProviderName, req.TotalTokens)
	case nil:
		// Update existing record
		_, err = h.db.Exec("UPDATE usage_stats SET usage_count = usage_count + 1, total_tokens = total_tokens + ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
			req.TotalTokens, existingID)
	}
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	respondOK(w, "记录成功")
}
