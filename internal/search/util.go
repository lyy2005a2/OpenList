package search

import (
	"strings"

	"github.com/OpenListTeam/OpenList/v4/drivers/base"
	"github.com/OpenListTeam/OpenList/v4/drivers/openlist"
	"github.com/OpenListTeam/OpenList/v4/internal/conf"
	"github.com/OpenListTeam/OpenList/v4/internal/driver"
	"github.com/OpenListTeam/OpenList/v4/internal/model"
	"github.com/OpenListTeam/OpenList/v4/internal/op"
	"github.com/OpenListTeam/OpenList/v4/internal/setting"
	"github.com/OpenListTeam/OpenList/v4/pkg/utils"
	log "github.com/sirupsen/logrus"
)

func Progress() (*model.IndexProgress, error) {
	p := setting.GetStr(conf.IndexProgress)
	var progress model.IndexProgress
	err := utils.Json.UnmarshalFromString(p, &progress)
	return &progress, err
}

func WriteProgress(progress *model.IndexProgress) {
	p, err := utils.Json.MarshalToString(progress)
	if err != nil {
		log.Errorf("marshal progress error: %+v", err)
	}
	err = op.SaveSettingItem(&model.SettingItem{
		Key:   conf.IndexProgress,
		Value: p,
		Type:  conf.TypeText,
		Group: model.SINGLE,
		Flag:  model.PRIVATE,
	})
	if err != nil {
		log.Errorf("save progress error: %+v", err)
	}
}

func updateIgnorePaths(customIgnorePaths string) {
	storages := op.GetAllStorages()
	ignorePaths := make([]string, 0)
	var skipDrivers = []string{ "OpenList", "Virtual"}
	v3Visited := make(map[string]bool)
	for _, storage := range storages {
		if utils.SliceContains(skipDrivers, storage.Config().Name) {
			if storage.Config().Name == "OpenList" {
				addition := storage.GetAddition().(*openlist.Addition)
				allowIndexed, visited := v3Visited[addition.Address]
				if !visited {
					url := addition.Address + "/api/public/settings"
					res, err := base.RestyClient.R().Get(url)
					if err == nil {
						log.Debugf("allow_indexed body: %+v", res.String())
						allowIndexed = utils.Json.Get(res.Body(), "data", conf.AllowIndexed).ToString() == "true"
						v3Visited[addition.Address] = allowIndexed
					}
				}
				log.Debugf("%s allow_indexed: %v", addition.Address, allowIndexed)
				if !allowIndexed {
					ignorePaths = append(ignorePaths, storage.GetStorage().MountPath)
				}
			} else {
				ignorePaths = append(ignorePaths, storage.GetStorage().MountPath)
			}
		}
	}
	if customIgnorePaths != "" {
		ignorePaths = append(ignorePaths, strings.Split(customIgnorePaths, "\n")...)
	}
	conf.SlicesMap[conf.IgnorePaths] = ignorePaths
}

func isIgnorePath(path string) bool {
	for _, ignorePath := range conf.SlicesMap[conf.IgnorePaths] {
		if strings.HasPrefix(path, ignorePath) {
			return true
		}
	}
	return false
}

func init() {
	op.RegisterSettingItemHook(conf.IgnorePaths, func(item *model.SettingItem) error {
		updateIgnorePaths(item.Value)
		return nil
	})
	op.RegisterStorageHook(func(typ string, storage driver.Driver) {
		var skipDrivers = []string{ "OpenList", "Virtual"}
		if utils.SliceContains(skipDrivers, storage.Config().Name) {
			updateIgnorePaths(setting.GetStr(conf.IgnorePaths))
		}
	})
}
