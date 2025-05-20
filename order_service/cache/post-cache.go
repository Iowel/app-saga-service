package cache

import (
	"order_service/model"
)

type IPostCache interface {
	Set(key string, value *model.UserCache)
	Get(key string) *model.UserCache
	GetAll() []*model.UserCache
	Delete(key string)
}
