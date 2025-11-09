package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"analytics-service/internal/model"
	"analytics-service/internal/repo"
)

// Aggregator 数据聚合器
type Aggregator struct {
	repo repo.AnalyticsRepo
}

// NewAggregator 创建聚合器
func NewAggregator(repo repo.AnalyticsRepo) *Aggregator {
	return &Aggregator{
		repo: repo,
	}
}

// ProcessVisitEvent 处理访问事件
func (a *Aggregator) ProcessVisitEvent(ctx context.Context, event *model.VisitEvent) error {
	// 转换时间戳
	visitTime := time.Unix(event.Timestamp, 0)
	date := visitTime.Format("2006-01-02")
	hour := visitTime.Format("2006-01-02 15")

	// 并发处理多个聚合任务
	errChan := make(chan error, 5)

	// 1. 每日统计
	go func() {
		daily := &model.AnalyticsDaily{
			ShortCode:      event.ShortCode,
			Date:           date,
			TotalVisits:    1,
			UniqueVisitors: 1, // 简化处理，实际应根据IP去重
		}
		errChan <- a.repo.UpsertDaily(ctx, daily)
	}()

	// 2. 每小时统计
	go func() {
		hourly := &model.AnalyticsHourly{
			ShortCode:  event.ShortCode,
			Hour:       hour,
			VisitCount: 1,
		}
		errChan <- a.repo.UpsertHourly(ctx, hourly)
	}()

	// 3. 浏览器统计
	go func() {
		if event.Browser != "" {
			browser := &model.AnalyticsBrowser{
				ShortCode:  event.ShortCode,
				Browser:    event.Browser,
				VisitCount: 1,
			}
			errChan <- a.repo.UpsertBrowser(ctx, browser)
		} else {
			errChan <- nil
		}
	}()

	// 4. 设备统计
	go func() {
		if event.DeviceType != "" {
			device := &model.AnalyticsDevice{
				ShortCode:  event.ShortCode,
				DeviceType: event.DeviceType,
				VisitCount: 1,
			}
			errChan <- a.repo.UpsertDevice(ctx, device)
		} else {
			errChan <- nil
		}
	}()

	// 5. 操作系统统计
	go func() {
		if event.OS != "" {
			os := &model.AnalyticsOS{
				ShortCode:  event.ShortCode,
				OS:         event.OS,
				VisitCount: 1,
			}
			errChan <- a.repo.UpsertOS(ctx, os)
		} else {
			errChan <- nil
		}
	}()

	// 收集错误
	var errors []error
	for i := 0; i < 5; i++ {
		if err := <-errChan; err != nil {
			errors = append(errors, err)
			log.Printf("⚠️  Aggregation error: %v", err)
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("aggregation failed with %d errors", len(errors))
	}

	log.Printf("✅ Aggregated event for short_code=%s, time=%s", event.ShortCode, visitTime.Format("2006-01-02 15:04:05"))
	return nil
}
