#!/bin/bash

mydir=$(dirname "${BASH_SOURCE[0]}")
TIFLOW_HOME="${mydir}/../../.."

"${TIFLOW_HOME}/tools/bin/oapi-codegen" -generate gin -package openapi -o "${TIFLOW_HOME}/dm/simulator/internal/openapi/gen.server.go" "${TIFLOW_HOME}/dm/simulator/internal/openapi/spec/simulator.yaml"
"${TIFLOW_HOME}/tools/bin/oapi-codegen" -generate types -package openapi -o "${TIFLOW_HOME}/dm/simulator/internal/openapi/gen.types.go" "${TIFLOW_HOME}/dm/simulator/internal/openapi/spec/simulator.yaml"
