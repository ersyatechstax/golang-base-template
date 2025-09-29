package context

import "context"

type (
	Key string
)

var (
	KeyGBT    = Key("gbt")
	KeyDevice = Key("device")
)

func SetGBT(parent context.Context, value string) context.Context {
	if parent == nil {
		parent = context.Background()
	}
	return context.WithValue(parent, KeyGBT, value)
}

func GetGBT(ctx context.Context) (result string, ok bool) {
	if ctx == nil {
		return "", false
	}
	if val := ctx.Value(KeyGBT); val != nil {
		result, ok = val.(string)
	}
	return
}

func SetDevice(parent context.Context, value string) context.Context {
	if parent == nil {
		parent = context.Background()
	}
	return context.WithValue(parent, KeyDevice, value)
}

func GetDevice(ctx context.Context) (result string, ok bool) {
	if ctx == nil {
		return "", false
	}
	if val := ctx.Value(KeyDevice); val != nil {
		result, ok = val.(string)
	}
	return
}
