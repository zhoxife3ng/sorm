package sorm

type options struct {
	forceMaster bool // 如果存在主从读写分离，是否强制走主库查询
	forUpdate   bool // 是否给记录添加forUpdate锁
	forceLoad   bool // 若数据有缓存，是否强制重新查询数据库
	load        bool // 调用Select方法的同时是否查询数据库记录
}

type option func(o *options)

func newOptions() options {
	return options{
		forceMaster: false,
		forUpdate:   false,
		forceLoad:   false,
		load:        false,
	}
}

func ForceMaster() option {
	return func(o *options) {
		o.forceMaster = true
	}
}

func ForUpdate() option {
	return func(o *options) {
		o.forUpdate = true
	}
}

func ForceLoad() option {
	return func(o *options) {
		o.forceLoad = true
	}
}

func Load() option {
	return func(o *options) {
		o.load = true
	}
}
