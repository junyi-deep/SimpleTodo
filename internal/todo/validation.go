package todo

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

func validateExchange(d Transfer) error {
	series := map[string]bool{}
	for _, s := range d.Series {
		if s.ID == "" || series[s.ID] {
			return fail("VALIDATION_ERROR", "系列 ID 重复或为空")
		}
		series[s.ID] = true
		if err := validateDate(s.StoppedFrom); err != nil {
			return err
		}
	}
	rules := map[string]Rule{}
	bounds := map[string]bool{}
	for _, r := range d.Rules {
		if r.SeriesID == "" || bounds[r.SeriesID+"/"+r.From] {
			return fail("VALIDATION_ERROR", "重复规则边界重复或缺少系列")
		}
		bounds[r.SeriesID+"/"+r.From] = true
		for _, date := range []*string{&r.From, &r.Anchor, r.Until} {
			if err := validateDate(date); err != nil {
				return err
			}
		}
		if parsed, err := time.Parse("2006-01-02", r.GeneratedThrough); err != nil || parsed.Format("2006-01-02") != r.GeneratedThrough {
			return fail("VALIDATION_ERROR", "重复生成游标日期无效")
		}
		if r.Until != nil && *r.Until >= r.From && *r.Until < r.Anchor {
			return fail("VALIDATION_ERROR", "重复规则日期不正确")
		}
		if r.MonthDay < 1 || r.MonthDay > 31 || len(r.Template) == 0 {
			return fail("VALIDATION_ERROR", "重复模板或月日无效")
		}
		if r.Frequency == "weekly" && len(r.Weekdays) == 0 {
			return fail("VALIDATION_ERROR", "每周规则缺少星期")
		}
		for _, w := range r.Weekdays {
			if w < 0 || w > 6 {
				return fail("VALIDATION_ERROR", "星期无效")
			}
		}
		ids := map[string]bool{}
		for _, t := range r.Template {
			if ids[t.ID] {
				return fail("VALIDATION_ERROR", "模板 ID 重复")
			}
			ids[t.ID] = true
		}
		rules[r.ID] = r
	}
	tasks := map[string]Task{}
	for _, t := range d.Tasks {
		tasks[t.ID] = t
		if t.ParentID == nil {
			copy := t
			copy.RemindAt = nil
			if err := validateRoot(&copy, time.Now()); err != nil {
				return err
			}
			if t.SeriesID != nil {
				r, ok := rules[val(t.RuleID)]
				if !ok || r.SeriesID != *t.SeriesID {
					return fail("VALIDATION_ERROR", "任务系列与规则不匹配")
				}
				if err := validateDate(t.OccurrenceDate); err != nil {
					return err
				}
			}
		}
	}
	ids := map[string]bool{}
	active := 0
	for _, f := range d.Focus {
		if f.ID == "" || ids[f.ID] || !contains([]string{"running", "paused", "stopped", "completed"}, f.State) || f.Duration < 60 || f.Duration > 10800 || f.Remaining < 0 || f.Remaining > f.Duration {
			return fail("VALIDATION_ERROR", "专注会话无效")
		}
		ids[f.ID] = true
		if f.State == "running" || f.State == "paused" {
			active++
		}
		if f.TaskID != nil {
			t, ok := tasks[*f.TaskID]
			if !ok || t.ParentID != nil {
				return fail("VALIDATION_ERROR", "专注引用无效任务")
			}
		}
	}
	if active > 1 {
		return fail("VALIDATION_ERROR", "只能保留一个活动专注")
	}
	ids = map[string]bool{}
	for _, c := range d.Countdowns {
		if c.ID == "" || ids[c.ID] || strings.TrimSpace(c.Name) == "" || len([]rune(c.Name)) > 500 || !contains([]string{"none", "daily", "weekly", "monthly"}, c.Frequency) || c.Offset < 0 || c.Offset > 86400 {
			return fail("VALIDATION_ERROR", "事项倒计时无效")
		}
		ids[c.ID] = true
		if err := validateDate(&c.AnchorDate); err != nil {
			return err
		}
		if _, err := time.Parse("15:04", c.AnchorTime); err != nil {
			return fail("VALIDATION_ERROR", "事项时分无效")
		}
	}
	ids = map[string]bool{}
	cycles := map[string]bool{}
	for _, j := range d.Jobs {
		key := fmt.Sprintf("%s/%s/%d/%s", j.SourceType, j.SourceID, j.Generation, j.CycleKey)
		if j.ID == "" || ids[j.ID] || cycles[key] || !contains([]string{"task", "focus", "countdown"}, j.SourceType) || !contains([]string{"pending", "claimed", "submitted", "failed", "uncertain", "cancelled", "missed", "expired", "dismissed"}, j.State) {
			return fail("VALIDATION_ERROR", "提醒记录无效或重复")
		}
		ids[j.ID] = true
		cycles[key] = true
	}
	for k, v := range d.Settings {
		if _, ok := defaults[k]; !ok {
			return fail("VALIDATION_ERROR", "未知设置项")
		}
		if len(v) > 1000 {
			return fail("VALIDATION_ERROR", "设置过长")
		}
		if k == "focusMinutes" {
			n, err := strconv.Atoi(v)
			if err != nil || n < 1 || n > 180 {
				return fail("VALIDATION_ERROR", "专注时长无效")
			}
		}
		switch k {
		case "expandDepth":
			n, er := strconv.Atoi(v)
			if er != nil || n < 1 || n > 5 {
				return fail("VALIDATION_ERROR", "展开层级为 1—5 层")
			}
		case "theme":
			if !contains([]string{"system", "light", "dark"}, v) {
				return fail("VALIDATION_ERROR", "无效主题")
			}
		case "titleClick":
			if !contains([]string{"edit", "complete"}, v) {
				return fail("VALIDATION_ERROR", "点击行为无效")
			}
		case "miniView":
			if !contains([]string{"today", "pending"}, v) {
				return fail("VALIDATION_ERROR", "浮窗视图无效")
			}
		case "dayShowCompleted", "inboxShowCompleted", "dayShowRecurring", "inboxShowRecurring", "autoPostpone", "completedBottom", "tabEntry", "sound", "miniPending", "edgeHide", "onboarding", "animation", "notificationPrompted":
			if v != "true" && v != "false" {
				return fail("VALIDATION_ERROR", "开关值无效")
			}
		}
	}
	return nil
}
