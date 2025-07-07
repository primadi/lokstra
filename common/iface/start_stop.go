package iface

type WithStart interface {
	Start() error
}

type WithStop interface {
	Stop() error
}
