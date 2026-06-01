module github.com/gen0sec/workflow-templates

go 1.26.1

require (
	github.com/deepnoodle-ai/workflow v0.0.5
	gopkg.in/yaml.v3 v3.0.1
)

require github.com/deepnoodle-ai/expr v0.0.1 // indirect

// Track the same gen0sec fork the consumer uses; bump when the
// workflow-service upgrades its engine dependency.
replace github.com/deepnoodle-ai/workflow => github.com/gen0sec/workflow v0.0.0-20260523175025-5091591f62a2
