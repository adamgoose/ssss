package lib

import (
	"context"

	"github.com/defval/di"
	"github.com/spf13/cobra"
)

var App *di.Container

func init() {
	App, _ = di.New()
}

func Apply(options ...di.Option) error {
	return App.Apply(options...)
}

func Provide(constructor di.Constructor, options ...di.ProvideOption) {
	App.Provide(constructor, options...)
}

func ProvideValue(value di.Value, options ...di.ProvideOption) {
	App.ProvideValue(value, options...)
}

func Resolve(ptr di.Pointer, options ...di.ResolveOption) error {
	return App.Resolve(ptr, options...)
}

func AutoResolve[T any](options ...di.ResolveOption) (T, error) {
	v := new(T)
	err := App.Resolve(v, options...)
	return *v, err
}

func MustAutoResolve[T any](options ...di.ResolveOption) T {
	return Must(AutoResolve[T](options...))
}

func Must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

func Invoke(invocation di.Invocation, options ...di.InvokeOption) error {
	return App.Invoke(invocation, options...)
}

func Wrap(options ...di.Option) (*di.Container, error) {
	c, _ := di.New()
	c.AddParent(App)

	if err := c.Apply(options...); err != nil {
		return nil, err
	}

	return c, nil
}

func RunE(runE interface{}) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		c, _ := Wrap(
			di.ProvideValue(cmd),
			di.ProvideValue(args),
			di.ProvideValue(cmd.Context(), di.As(new(context.Context))),
		)
		return c.Invoke(runE)
	}
}
