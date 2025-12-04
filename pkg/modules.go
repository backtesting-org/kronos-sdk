package packages

import (
	"github.com/backtesting-org/kronos-sdk/pkg/activity"
	"github.com/backtesting-org/kronos-sdk/pkg/adapters"
	"github.com/backtesting-org/kronos-sdk/pkg/analytics"
	"github.com/backtesting-org/kronos-sdk/pkg/events"
	"github.com/backtesting-org/kronos-sdk/pkg/executor"
	"github.com/backtesting-org/kronos-sdk/pkg/health"
	"github.com/backtesting-org/kronos-sdk/pkg/inference/features"
	"github.com/backtesting-org/kronos-sdk/pkg/ingestors"
	"github.com/backtesting-org/kronos-sdk/pkg/lifecycle"
	"github.com/backtesting-org/kronos-sdk/pkg/plugin"
	"github.com/backtesting-org/kronos-sdk/pkg/registry"
	"github.com/backtesting-org/kronos-sdk/pkg/runtime"
	"github.com/backtesting-org/kronos-sdk/pkg/signal"
	"github.com/backtesting-org/kronos-sdk/pkg/stores"
	"go.uber.org/fx"
)

var Module = fx.Options(
	activity.Module,
	adapters.Module,
	analytics.Module,
	events.Module,
	features.Module,
	health.Module,
	ingestors.Module,
	lifecycle.Module,
	plugin.Module,
	registry.Module,
	runtime.Module,
	signal.Module,
	stores.Module,
	executor.Module,
)
