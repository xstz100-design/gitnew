package services

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

// coordRegex 用于判断字符串是否为 "lat,lng" 坐标格式
var coordRegex = regexp.MustCompile(`^[-+]?([1-8]?\d(\.\d+)?|90(\.0+)?),\s*[-+]?(180(\.0+)?|((1[0-7]\d)|([1-9]?\d))(\.\d+)?)$`)

// isCoordString 检查字符串是否为坐标格式
func isCoordString(s string) (lat, lng float64, ok bool) {
	s = strings.TrimSpace(s)
	if !coordRegex.MatchString(s) {
		return 0, 0, false
	}
	parts := strings.SplitN(s, ",", 2)
	if len(parts) != 2 {
		return 0, 0, false
	}
	lat, err1 := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
	lng, err2 := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
	if err1 != nil || err2 != nil {
		return 0, 0, false
	}
	return lat, lng, true
}

// IsCoordString 检查字符串是否为坐标格式（导出版）
func IsCoordString(s string) (lat, lng float64, ok bool) {
	return isCoordString(s)
}

// HaversineKM 计算两个经纬度坐标之间的直线距离（公里）
func HaversineKM(lat1, lng1, lat2, lng2 float64) float64 {
	const R = 6371.0 // 地球半径（公里）
	dLat := (lat2 - lat1) * math.Pi / 180
	dLng := (lng2 - lng1) * math.Pi / 180
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*math.Pi/180)*math.Cos(lat2*math.Pi/180)*
			math.Sin(dLng/2)*math.Sin(dLng/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return R * c
}

// ─────────────────── 距离缓存 ───────────────────

type distanceCacheEntry struct {
	km        float64
	expiresAt time.Time
}

var (
	distanceCacheMu  sync.Mutex
	distanceCacheMap = make(map[string]distanceCacheEntry)
)

const distanceCacheTTL = 7 * 24 * time.Hour // 7 天

func distanceCacheKey(origin, destination string) string {
	return strings.ToLower(strings.TrimSpace(origin)) + "||" + strings.ToLower(strings.TrimSpace(destination))
}

func getDistanceCache(key string) (float64, bool) {
	distanceCacheMu.Lock()
	defer distanceCacheMu.Unlock()
	if e, ok := distanceCacheMap[key]; ok && time.Now().Before(e.expiresAt) {
		return e.km, true
	}
	return 0, false
}

func setDistanceCache(key string, km float64) {
	distanceCacheMu.Lock()
	defer distanceCacheMu.Unlock()
	// 顺手清理过期条目（简单防泄漏）
	for k, e := range distanceCacheMap {
		if time.Now().After(e.expiresAt) {
			delete(distanceCacheMap, k)
		}
	}
	distanceCacheMap[key] = distanceCacheEntry{km: km, expiresAt: time.Now().Add(distanceCacheTTL)}
}

// ─────────────────── Distance Matrix ───────────────────

// GetDistanceKM 通过 Google Maps Distance Matrix API 计算两地距离（公里）。
// 结果缓存 7 天，相同 origin/destination 不重复调用 API。
func GetDistanceKM(origin, destination, apiKey string) (float64, error) {
	if apiKey == "" {
		return 0, fmt.Errorf("Google Maps API key 未配置")
	}
	if origin == "" || destination == "" {
		return 0, fmt.Errorf("origin 或 destination 为空")
	}

	cacheKey := distanceCacheKey(origin, destination)
	if km, ok := getDistanceCache(cacheKey); ok {
		return km, nil
	}

	reqURL := fmt.Sprintf(
		"https://maps.googleapis.com/maps/api/distancematrix/json?origins=%s&destinations=%s&key=%s",
		url.QueryEscape(origin),
		url.QueryEscape(destination),
		apiKey,
	)

	resp, err := http.Get(reqURL) //nolint:gosec
	if err != nil {
		return 0, fmt.Errorf("无法连接 Google Maps API: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Status string `json:"status"`
		Rows   []struct {
			Elements []struct {
				Status   string `json:"status"`
				Distance struct {
					Value int `json:"value"` // 单位：米
				} `json:"distance"`
			} `json:"elements"`
		} `json:"rows"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, fmt.Errorf("解析 Google Maps 响应失败: %w", err)
	}
	if result.Status != "OK" {
		return 0, fmt.Errorf("Google Maps API 返回错误: %s", result.Status)
	}
	if len(result.Rows) == 0 || len(result.Rows[0].Elements) == 0 {
		return 0, fmt.Errorf("Google Maps 未返回距离数据")
	}
	elem := result.Rows[0].Elements[0]
	if elem.Status != "OK" {
		return 0, fmt.Errorf("Google Maps 路径计算失败: %s", elem.Status)
	}

	// 米 → 公里，保留2位小数
	km := float64(elem.Distance.Value) / 1000.0
	setDistanceCache(cacheKey, km)
	return km, nil
}
