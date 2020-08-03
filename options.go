package sorm

type option struct {
	forceMaster bool // 如果存在主从读写分离，是否强制走主库查询
	forUpdate   bool // 是否给记录添加forUpdate锁
	forceLoad   bool // 若数据有缓存，是否强制重新查询数据库
	load        bool // 调用Select方法的同时是否查询数据库记录
}

type Option func(o *option)

func newOption() option {
	return option{
		forceMaster: false,
		forUpdate:   false,
		forceLoad:   false,
		load:        false,
	}
}

func ForceMaster() Option {
	return func(o *option) {
		o.forceMaster = true
	}
}

func ForUpdate() Option {
	return func(o *option) {
		o.forUpdate = true
	}
}

func ForceLoad() Option {
	return func(o *option) {
		o.forceLoad = true
	}
}

func Load() Option {
	return func(o *option) {
		o.load = true
	}
}
