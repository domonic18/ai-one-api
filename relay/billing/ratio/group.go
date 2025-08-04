package ratio

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/songquanpeng/one-api/common/logger"
)

var groupRatioLock sync.RWMutex
var GroupRatio = map[string]float64{
	"default": 1,
	"vip":     1,
	"svip":    1,
}

func GroupRatio2JSONString() string {
	jsonBytes, err := json.Marshal(GroupRatio)
	if err != nil {
		logger.SysError("error marshalling model ratio: " + err.Error())
	}
	return string(jsonBytes)
}

func UpdateGroupRatioByJSONString(jsonStr string) error {
	groupRatioLock.Lock()
	defer groupRatioLock.Unlock()
	GroupRatio = make(map[string]float64)
	return json.Unmarshal([]byte(jsonStr), &GroupRatio)
}

func GetGroupRatio(name string) float64 {
	groupRatioLock.RLock()
	defer groupRatioLock.RUnlock()
	ratio, ok := GroupRatio[name]
	if !ok {
		ctx := context.Background()
		logger.Debugf(ctx, "group ratio not found, using default: %s", name)
		return 1
	}
	return ratio
}
