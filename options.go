package sorm

type options struct {
	forceMaster bool
	forUpdate   bool
}

type option func(o *options)

func newOptions() options {
	return options{
		forceMaster: false,
		forUpdate:   false,
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
