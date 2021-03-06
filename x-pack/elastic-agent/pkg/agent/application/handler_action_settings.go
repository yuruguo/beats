// Copyright Elasticsearch B.V. and/or licensed to Elasticsearch B.V. under one
// or more contributor license agreements. Licensed under the Elastic License;
// you may not use this file except in compliance with the Elastic License.

package application

import (
	"context"
	"fmt"

	"github.com/elastic/beats/v7/x-pack/elastic-agent/pkg/agent/application/info"
	"github.com/elastic/beats/v7/x-pack/elastic-agent/pkg/agent/errors"
	"github.com/elastic/beats/v7/x-pack/elastic-agent/pkg/core/logger"
	"github.com/elastic/beats/v7/x-pack/elastic-agent/pkg/fleetapi"
)

// handlerSettings handles settings change coming from fleet and updates log level.
type handlerSettings struct {
	log       *logger.Logger
	reexec    reexecManager
	agentInfo *info.AgentInfo
}

// Handle handles SETTINGS action.
func (h *handlerSettings) Handle(ctx context.Context, a action, acker fleetAcker) error {
	h.log.Debugf("handlerUpgrade: action '%+v' received", a)
	action, ok := a.(*fleetapi.ActionSettings)
	if !ok {
		return fmt.Errorf("invalid type, expected ActionSettings and received %T", a)
	}

	if !isSupportedLogLevel(action.LogLevel) {
		return fmt.Errorf("invalid log level, expected debug|info|warning|error and received '%s'", action.LogLevel)
	}

	if err := h.agentInfo.LogLevel(action.LogLevel); err != nil {
		return errors.New("failed to update log level", err)
	}

	if err := acker.Ack(ctx, a); err != nil {
		h.log.Errorf("failed to acknowledge SETTINGS action with id '%s'", action.ActionID)
	} else if err := acker.Commit(ctx); err != nil {
		h.log.Errorf("failed to commit acker after acknowledging action with id '%s'", action.ActionID)
	}

	h.reexec.ReExec()
	return nil
}

func isSupportedLogLevel(level string) bool {
	return level == "error" || level == "debug" || level == "info" || level == "warning"
}
